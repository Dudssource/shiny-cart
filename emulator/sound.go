package emulator

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// Square 1
	NR10 = 0xFF10
	NR11 = 0xFF11
	NR12 = 0xFF12
	NR13 = 0xFF13
	NR14 = 0xFF14

	// Square 2
	NR21 = 0xFF16
	NR22 = 0xFF17
	NR23 = 0xFF18
	NR24 = 0xFF19

	// Wave
	NR30 = 0xFF1A
	NR31 = 0xFF1B
	NR32 = 0xFF1C
	NR33 = 0xFF1D
	NR34 = 0xFF1E

	// Noise
	NR41 = 0xFF20
	NR42 = 0xFF21
	NR43 = 0xFF22
	NR44 = 0xFF23

	// Control/status
	NR50 = 0xFF24
	NR51 = 0xFF25
	NR52 = 0xFF26
)

var dutyWaveformTable = map[uint8]uint8{
	0x1: 0b00000001, // 12.5%
	0x2: 0b00000011, // 25%
	0x3: 0b00001111, // 50%
	0x4: 0b11111100, // 75%
}

var noiseChannelDivisorTable = map[uint8]uint8{
	0: 8,
	1: 16,
	2: 32,
	3: 48,
	4: 64,
	5: 80,
	6: 96,
	7: 112,
}

type Sound struct {
	stream rl.AudioStream
	mem    memoryArea

	// DIV lastCycleValue, used for falling edge and ticking the frame-sequencer
	divLastValue uint8

	// channel 1 - square wave with sweep
	sCH1Buf            []float32
	timerSCH1          uint16 // channel SCH1 timer (frequency)
	dutyPositionSCH1   uint8  // channel SCH1 wave duty position
	frameSeqStepSCH1   int    // frame-sequencer SCH1
	lengthCounterSCH1  int    // length counter SCH1
	disableSCH1        bool   // disable SCH1
	volumeSCH1         int    // volume for SCH1 (controlled by envelope)
	volumeTimerSCH1    int    // volume timer for SCH1
	sweepTimerSCH1     int    // frequency sweep timer for SCH1
	sweepEnableSCH1    bool   // sweep enabled for SCH1
	sweepShadowRegSCH1 uint16 // shadow sweep register for SCH1

	// channel 2 - square wave
	sCH2Buf           []float32
	timerSCH2         uint16 // channel SCH2 timer (frequency)
	dutyPositionSCH2  uint8  // channel SCH2 wave duty position
	frameSeqStepSCH2  int    // frame-sequencer SCH2
	lengthCounterSCH2 int    // length counter SCH2
	disableSCH2       bool   // disable SCH2
	volumeSCH2        int    // volume for SCH2 (controlled by envelope)
	volumeTimerSCH2   int    // volume timer for SCH2

	// channel 3 - wave table
	sCH3Buf           []float32
	lengthCounterSCH3 int    // length counter SCH3
	timerSCH3         uint16 // channel SCH3 timer (frequency)
	wavePositionSCH3  uint8  // channel SCH3 wave position
	frameSeqStepSCH3  int    // frame-sequencer SCH3
	disableSCH3       bool   // disable SCH3

	// channel 4 - noise
	sCH4Buf           []float32
	lengthCounterSCH4 int    // length counter SCH4
	timerSCH4         uint16 // channel SCH4 timer (frequency)
	frameSeqStepSCH4  int    // frame-sequencer SCH4
	disableSCH4       bool   // disable SCH4
	volumeSCH4        int    // volume for SCH4 (controlled by envelope)
	volumeTimerSCH4   int    // volume timer for SCH4
	lsfrSCH4          uint16 //  SCH4 LSFR register

	// toggle channels (usefull for troubleshooting)
	useSCH1  bool
	useSCH2  bool
	useSCH3  bool
	useSCH4  bool
	channels int
}

func NewSound(mem memoryArea, channels int) *Sound {
	return &Sound{
		mem:      mem,
		channels: channels,
	}
}

const (
	// number of samples to buffer before enqueuing on the audio device
	maxSamplesBufferSize = 4096
	sampleRate           = 44100
	bufferSize           = maxSamplesBufferSize
)

func (s *Sound) init() error {

	rl.InitAudioDevice()
	rl.SetAudioStreamBufferSizeDefault(bufferSize)
	s.stream = rl.LoadAudioStream(sampleRate, 32, 1)
	rl.PlayAudioStream(s.stream)

	if s.channels&0x1 > 0 {
		s.useSCH1 = true
	}

	if s.channels&0x2 > 0 {
		s.useSCH2 = true
	}

	if s.channels&0x4 > 0 {
		s.useSCH3 = true
	}

	if s.channels&0x8 > 0 {
		s.useSCH4 = true
	}

	return nil
}

/*
 * #############################################
 * # SOUND CHANNELS CH1 and CH2 (square waves) #
 * #############################################
 */

// for square wave channels (CH1 and CH2)
func (s *Sound) loadTimerPeriodSquareWaveChannel(frequencyMsbAddr, frequencyLsbAddr int) uint16 {
	// load frequency from register (11-bit)
	return (uint16(s.mem[frequencyMsbAddr]&0x7)<<8 | uint16(s.mem[frequencyLsbAddr])) & 0x7FF
}

// for square wave channels (CH1 and CH2)
func (s *Sound) loadTimerFrequencySquareWaveChannel(frequencyMsbAddr, frequencyLsbAddr int) uint16 {
	// load frequency from register (11-bit)
	freq := s.loadTimerPeriodSquareWaveChannel(frequencyMsbAddr, frequencyLsbAddr)
	// calculate the new frequency, since it is an eleven-bit two-compliment number, subtracting
	// 2048 from it wraps around and gives the equivalent negative number (without signal).
	return (2048 - freq) * 4
}

/*
 * #####################################################
 * # SOUND CHANNELS CH1, CH2 and CH4 RELATED FUNCTIONS #
 * #####################################################
 */

// for CH1, CH2 and CH4 (that have volume envelope)
func (s *Sound) loadVolumeEnvelopeTimer(volumePeriodAddress int) int {
	// reload volume timer
	period := int(s.mem[volumePeriodAddress] & 0x7)
	// see below (item 2)
	// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware#Obscure_Behavior
	if period == 0 {
		return 8
	}
	return period
}

// applies to channels CH1, CH2 and CH4
// https://nightshade256.github.io/2021/03/27/gb-sound-emulation.html#envelope-function
func (*Sound) dac(dacInput float32) float32 {
	return dacInput / 100
}

/*
 * #####################################
 * # SOUND CHANNEL 1 RELATED FUNCTIONS #
 * #####################################
 */

func (s *Sound) loadCH1SweepPeriod() int {
	// load sweep timer
	return int(s.mem[NR10]&0x70) >> 4
}

// ONLY applies to SCH1 (square+sweep)
func (s *Sound) reloadCH1SweepTimer() {
	// reload sweep timer
	s.sweepTimerSCH1 = s.loadCH1SweepPeriod()
	// envelope and sweepers treat 0 as 8
	// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware#Obscure_Behavior
	if s.sweepTimerSCH1 == 0 {
		s.sweepTimerSCH1 = 8
	}
}

func (s *Sound) loadCH1SweepShift() uint8 {
	return s.mem[NR10] & 0x7
}

// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware#Frequency_Sweep
// https://nightshade256.github.io/2021/03/27/gb-sound-emulation.html#sweep-function
func (s *Sound) calculateCH1SweepNewFrequency() uint16 {
	newFrequency := s.sweepShadowRegSCH1 >> uint16(s.loadCH1SweepShift())

	sweepIncrease := (s.mem[NR10] & 0x8) == 0

	if !sweepIncrease {
		newFrequency = s.sweepShadowRegSCH1 - newFrequency
	} else {
		newFrequency = s.sweepShadowRegSCH1 + newFrequency
	}

	// 11-bit overflow
	if newFrequency > 2047 {
		s.disableSCH1 = true
	}

	return newFrequency
}

// tick SCH1 channel timer
func (s *Sound) stepSCH1(tCycle int, fsStep bool) {

	// decrement SCH1 timer every t-cycle
	s.timerSCH1--

	// frequency reached 0
	if s.timerSCH1 == 0 {
		// reload timer
		s.timerSCH1 = s.loadTimerFrequencySquareWaveChannel(NR14, NR13)

		// increment duty position (since values are 0-7, we wrap around by ANDing 0x7)
		s.dutyPositionSCH1 = (s.dutyPositionSCH1 + 1) & 0x7
	}

	// frame-sequencer (ticks every 8192 t-cycles = 512hz)
	if fsStep {

		// sweep step
		if s.frameSeqStepSCH1 == 2 || s.frameSeqStepSCH1 == 6 {

			if s.sweepTimerSCH1 > 0 {
				// decrement sweep timer
				s.sweepTimerSCH1--

				// timer reached 0
				if s.sweepTimerSCH1 == 0 {
					// reload sweep timer
					s.reloadCH1SweepTimer()

					// sweep enabled and period non-zero
					if s.sweepEnableSCH1 && s.loadCH1SweepPeriod() > 0 {

						newFrequency := s.calculateCH1SweepNewFrequency()

						if newFrequency <= 2047 && s.loadCH1SweepShift() > 0 {
							//s.timerSCH1 = newFrequency
							s.sweepShadowRegSCH1 = newFrequency

							// write back period LSB
							s.mem[NR13] = uint8(newFrequency & 0xFF)

							// write back period MSB
							s.mem[NR13] = (s.mem[NR13] & 0xF8) | uint8((newFrequency&0x700)>>8)

							// check overflow again
							s.calculateCH1SweepNewFrequency()
						}
					}
				}
			}
		}

		// length-counter step AND length-enable flag is set
		if (s.frameSeqStepSCH1%2) == 0 && (s.mem[NR14]&0x40) > 0 {

			// decrement
			s.lengthCounterSCH1--

			// disable the channel
			if s.lengthCounterSCH1 == 0 {
				s.disableSCH1 = true
			}
		}

		// volume envelope (64hz), frame sequencer step 7
		if s.frameSeqStepSCH1 == 0x7 {

			// volume envelop period
			volumeEnvelopePeriod := s.loadVolumeEnvelopeTimer(NR12)

			// every sweep-pace ticks
			if volumeEnvelopePeriod > 0 {

				// decrement internal timer
				if s.volumeTimerSCH1 > 0 {
					s.volumeTimerSCH1--
				}

				// turned zero
				if s.volumeTimerSCH1 == 0 {

					// reload volume timer
					s.volumeTimerSCH1 = s.loadVolumeEnvelopeTimer(NR12)

					// 0=decrease, 1=increase
					envelopeIncreaseVolume := (s.mem[NR12] & 0x8) > 0

					// increase
					if envelopeIncreaseVolume && s.volumeSCH1 < 0xF {
						s.volumeSCH1++
					}

					// decrease
					if !envelopeIncreaseVolume && s.volumeSCH1 > 0 {
						s.volumeSCH1--
					}
				}
			}
		}

		// current frame-sequencer step (0-7)
		s.frameSeqStepSCH1 = (s.frameSeqStepSCH1 + 1) & 0x7
	}

	if !s.disableSCH1 {
		// channel SCH1 enabled
		s.mem[NR52] |= 0x1
	} else {
		// channel SCH1 disabled
		s.mem[NR52] &= ^uint8(0x1)
	}

	if tCycle%95 != 0 {
		return
	}

	var (
		output float32
	)

	// DAC is on and channel is enabled
	if (s.mem[NR12]&0xF8) > 0 && !s.disableSCH1 {

		// duty
		dutyPattern := (s.mem[NR21] & 0xC0) >> 6

		// amplitude (lookup table and get the bit pointed by dutyPosition)
		amplitude := int((dutyWaveformTable[dutyPattern] & (1 << s.dutyPositionSCH1)) >> s.dutyPositionSCH1)

		// calculate DAC
		output = s.dac(float32(amplitude * s.volumeSCH1))

	} else if (s.mem[NR12]&0xF8) > 0 && s.disableSCH1 {
		// channel is disabled but DAC is on
		// https://gbdev.io/pandocs/Audio_details.html#channels
		output = 1

	} else if (s.mem[NR12] & 0xF8) == 0 {
		// if DAC is off, disable the channel
		// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware#Channel_DAC
		s.disableSCH1 = true
	}

	// write samples (PCM)
	s.sCH1Buf = append(s.sCH1Buf, output) // left
	s.sCH1Buf = append(s.sCH1Buf, output) // right
}

// reset SCH1 (trigger event)
func (s *Sound) resetSCH1() {

	if s.lengthCounterSCH1 == 0 {
		// reset length timer
		s.lengthCounterSCH1 = 64
	}

	// reset FS
	// s.frameSeqStepSCH1 = 0

	// enable
	s.disableSCH1 = false

	// load initial volume (for envelope)
	s.volumeSCH1 = int(s.mem[NR12]&0xF0) >> 4

	// reload timer
	s.timerSCH1 = s.loadTimerFrequencySquareWaveChannel(NR14, NR13)

	// reload volume timer
	s.volumeTimerSCH1 = s.loadVolumeEnvelopeTimer(NR12)

	// load frequency to shadow
	s.sweepShadowRegSCH1 = s.loadTimerPeriodSquareWaveChannel(NR14, NR13)

	// reload sweep timer
	s.reloadCH1SweepTimer()

	// sweep enabled
	s.sweepEnableSCH1 = s.sweepTimerSCH1 > 0 || s.loadCH1SweepShift() > 0

	// if sweep enabled AND shift is non-zero we only check overflow (by throwing away the result)
	if s.sweepEnableSCH1 && s.loadCH1SweepShift() > 0 {
		s.calculateCH1SweepNewFrequency()
	}
}

/*
 * #####################################
 * # SOUND CHANNEL 2 RELATED FUNCTIONS #
 * #####################################
 */

// tick SCH2 channel timer
func (s *Sound) stepSCH2(tCycle int, fsStep bool) {

	// decrement SCH2 timer every t-cycle
	s.timerSCH2--

	// frequency reached 0
	if s.timerSCH2 == 0 {
		// reload timer
		s.timerSCH2 = s.loadTimerFrequencySquareWaveChannel(NR24, NR23)

		// increment duty position (since values are 0-7, we wrap around by ANDing 0x7)
		s.dutyPositionSCH2 = (s.dutyPositionSCH2 + 1) & 0x7
	}

	// frame-sequencer (ticks every 8192 t-cycles = 512hz)
	if fsStep {

		// length-counter step AND length-enable flag is set
		if (s.frameSeqStepSCH2%2) == 0 && (s.mem[NR24]&0x40) > 0 {

			// decrement
			s.lengthCounterSCH2--

			// disable the channel
			if s.lengthCounterSCH2 == 0 {
				s.disableSCH2 = true
			}
		}

		// volume envelop period
		volumeEnvelopePeriod := int(s.mem[NR22] & 0x7)

		// volume envelope (64hz), frame sequencer step 7
		if s.frameSeqStepSCH2 == 0x7 {

			// every sweep-pace ticks
			if volumeEnvelopePeriod > 0 {

				// decrement internal timer
				if s.volumeTimerSCH2 > 0 {
					s.volumeTimerSCH2--
				}

				// turned zero
				if s.volumeTimerSCH2 == 0 {

					// reload volume timer
					s.volumeTimerSCH2 = s.loadVolumeEnvelopeTimer(NR22)

					// 0=decrease, 1=increase
					envelopeIncreaseVolume := (s.mem[NR22] & 0x8) > 0

					// increase
					if envelopeIncreaseVolume && s.volumeSCH2 < 0xF {
						s.volumeSCH2++
					}

					// decrease
					if !envelopeIncreaseVolume && s.volumeSCH2 > 0 {
						s.volumeSCH2--
					}
				}
			}
		}

		// current frame-sequencer step (0-7)
		s.frameSeqStepSCH2 = (s.frameSeqStepSCH2 + 1) & 0x7
	}

	if !s.disableSCH2 {
		// channel SCH2 enabled
		s.mem[NR52] |= 0x2
	} else {
		// channel SCH2 disabled
		s.mem[NR52] &= ^uint8(0x2)
	}

	if tCycle%95 != 0 {
		return
	}

	var (
		output float32
	)

	// DAC is on and channel is enabled
	if (s.mem[NR22]&0xF8) > 0 && !s.disableSCH2 {

		// duty
		dutyPattern := (s.mem[NR21] & 0xC0) >> 6

		// amplitude (lookup table and get the bit pointed by dutyPosition)
		amplitude := int((dutyWaveformTable[dutyPattern] & (1 << s.dutyPositionSCH2)) >> s.dutyPositionSCH2)

		// calculate DAC
		output = s.dac(float32(amplitude * s.volumeSCH2))

	} else if (s.mem[NR22]&0xF8) > 0 && s.disableSCH2 {
		// channel is disabled but DAC is on
		// https://gbdev.io/pandocs/Audio_details.html#channels
		output = 1
	} else if (s.mem[NR22] & 0xF8) == 0 {
		// if DAC is off, disable the channel
		s.disableSCH2 = true
	}

	// write samples (PCM)
	s.sCH2Buf = append(s.sCH2Buf, output) // left
	s.sCH2Buf = append(s.sCH2Buf, output) // right
}

// reset SCH2 (trigger event)
func (s *Sound) resetSCH2() {

	if s.lengthCounterSCH2 == 0 {
		// reset length timer
		s.lengthCounterSCH2 = 64
	}

	// reset FS
	// s.frameSeqStepSCH2 = 0

	// enable
	s.disableSCH2 = false

	// load initial volume (for envelope)
	s.volumeSCH2 = int(s.mem[NR22]&0xF0) >> 4

	// reload timer
	s.timerSCH2 = s.loadTimerFrequencySquareWaveChannel(NR24, NR23)

	// reload volume timer
	s.volumeTimerSCH2 = s.loadVolumeEnvelopeTimer(NR22)

	// reload timer
	s.timerSCH2 = s.loadTimerFrequencySquareWaveChannel(NR24, NR23)
}

/*
 * #####################################
 * # SOUND CHANNEL 3 RELATED FUNCTIONS #
 * #####################################
 */

// for wave channel (CH3)
func (s *Sound) loadTimerFrequencyWaveChannel(frequencyMsbAddr, frequencyLsbAddr int) uint16 {
	// load frequency from register (11-bit)
	freq := s.loadTimerPeriodSquareWaveChannel(frequencyMsbAddr, frequencyLsbAddr)
	// calculate the new frequency, since it is an eleven-bit two-compliment number, subtracting
	// 2048 from it wraps around and gives the equivalent negative number (without signal).
	return (2048 - freq) * 2
}

// tick SCH3 channel timer
func (s *Sound) stepSCH3(tCycle int, fsStep bool) {

	// decrement SCH3 timer every t-cycle
	s.timerSCH3--

	// frequency reached 0
	if s.timerSCH3 == 0 {
		// reload timer
		s.timerSCH3 = s.loadTimerFrequencyWaveChannel(NR34, NR33)

		// increment duty position (since values are 0-31, we wrap around by mod 32/0x20)
		s.wavePositionSCH3 = (s.wavePositionSCH3 + 1) % 32
	}

	// frame-sequencer (ticks every 8192 t-cycles = 512hz)
	if fsStep {

		// length-counter step AND length-enable flag is set
		if (s.frameSeqStepSCH3%2) == 0 && (s.mem[NR34]&0x40) > 0 {

			// decrement
			s.lengthCounterSCH3--

			// disable the channel
			if s.lengthCounterSCH3 == 0 {
				s.disableSCH3 = true
			}
		}

		// current frame-sequencer step (0-7)
		s.frameSeqStepSCH3 = (s.frameSeqStepSCH3 + 1) & 0x7
	}

	if !s.disableSCH3 {
		// channel SCH3 enabled
		s.mem[NR52] |= 0x4
	} else {
		// channel SCH3 disabled
		s.mem[NR52] &= ^uint8(0x4)
	}

	if tCycle%95 != 0 {
		return
	}

	var (
		output float32
	)

	// DAC is on and channel is enabled
	if (s.mem[NR30]&0x80) > 0 && !s.disableSCH3 {

		// get volume
		volume := (s.mem[NR32] & 0x60) >> 5

		// calculate offsets
		position := int(s.wavePositionSCH3 / 2)
		nibble := int(s.wavePositionSCH3 % 2)
		var sample int

		// store sample buffer
		if nibble == 0 {
			sample = int(s.mem[0xFF30+position]&0xF0) >> 4
		} else {
			sample = int(s.mem[0xFF30+position] & 0xF)
		}

		// samples
		output = s.dac(float32(sample >> int(volume)))

	} else if (s.mem[NR30]&0x80) > 0 && s.disableSCH2 {
		// channel is disabled but DAC is on
		// https://gbdev.io/pandocs/Audio_details.html#channels
		output = 1

	} else if (s.mem[NR30] & 0x80) == 0 {
		// if DAC is off, disable the channel
		s.disableSCH3 = true
	}

	// write samples (PCM)
	s.sCH3Buf = append(s.sCH3Buf, output) // left
	s.sCH3Buf = append(s.sCH3Buf, output) // right
}

// reset SCH3 (trigger event)
func (s *Sound) resetSCH3() {

	if s.lengthCounterSCH3 == 0 {
		// reset length timer
		s.lengthCounterSCH3 = 256
	}

	// reset FS
	// s.frameSeqStepSCH3 =3

	// enable
	s.disableSCH3 = false

	// reload index
	s.wavePositionSCH3 = 0

	// reload timer
	s.timerSCH3 = s.loadTimerFrequencyWaveChannel(NR34, NR33)
}

/*
 * #####################################
 * # SOUND CHANNEL 4 RELATED FUNCTIONS #
 * #####################################
 */

// calculate divisor
func (s *Sound) loadNoiseTimer() uint16 {
	freq := noiseChannelDivisorTable[s.mem[NR43]&0x7] << ((s.mem[NR43] & 0xF0) >> 4)
	return uint16(freq)
}

func (s *Sound) stepSCH4(tCycle int, fsStep bool) {

	// decrement timer
	s.timerSCH4--

	// frequency reached 0
	if s.timerSCH4 == 0 {

		// reload timer
		s.timerSCH4 = s.loadNoiseTimer()

		// low bits XOR
		lfsrLO := ((s.lsfrSCH4 & 0x1) ^ ((s.lsfrSCH4 & 0x2) >> 1))

		// shift right by one and set the 14th bit with the XOR result
		s.lsfrSCH4 = (s.lsfrSCH4 >> 1) | (lfsrLO << 14)

		// if width mode is 1
		if (s.mem[NR43] & 0x8) > 0 {

			// also put the XOR result into bit 6
			s.lsfrSCH4 = (s.lsfrSCH4 & 0xBF) | (lfsrLO << 6)
		}

		// ensure LSFR is 15 bits long
		s.lsfrSCH4 &= 0x7FFF
	}

	// frame-sequencer (ticks every 8192 t-cycles = 512hz)
	if fsStep {

		// length-counter step AND length-enable flag is set
		if (s.frameSeqStepSCH4%2) == 0 && (s.mem[NR44]&0x40) > 0 {

			// decrement
			s.lengthCounterSCH4--

			// disable the channel
			if s.lengthCounterSCH4 == 0 {
				s.disableSCH4 = true
			}
		}

		// volume envelop period
		volumeEnvelopePeriod := int(s.mem[NR42] & 0x7)

		// volume envelope (64hz), frame sequencer step 7
		if s.frameSeqStepSCH4 == 0x7 {

			// every sweep-pace ticks
			if volumeEnvelopePeriod > 0 {

				// decrement internal timer
				if s.volumeTimerSCH4 > 0 {
					s.volumeTimerSCH4--
				}

				// turned zero
				if s.volumeTimerSCH4 == 0 {

					// reload volume timer
					s.volumeTimerSCH4 = s.loadVolumeEnvelopeTimer(NR42)

					// 0=decrease, 1=increase
					envelopeIncreaseVolume := (s.mem[NR42] & 0x8) > 0

					// increase
					if envelopeIncreaseVolume && s.volumeSCH4 < 0xF {
						s.volumeSCH4++
					}

					// decrease
					if !envelopeIncreaseVolume && s.volumeSCH4 > 0 {
						s.volumeSCH4--
					}
				}
			}
		}

		// current frame-sequencer step (0-7)
		s.frameSeqStepSCH4 = (s.frameSeqStepSCH4 + 1) & 0x7
	}

	if !s.disableSCH4 {
		// channel SCH4 enabled
		s.mem[NR52] |= 0x8
	} else {
		// channel SCH4 disabled
		s.mem[NR52] &= ^uint8(0x8)
	}

	if tCycle%95 != 0 {
		return
	}

	var (
		output float32
	)

	// DAC is on and channel is enabled
	if (s.mem[NR42]&0xF8) > 0 && !s.disableSCH4 {

		// amplitude (bit 0 of LFSR inverted)
		amplitude := ^(s.lsfrSCH4 & 0x1)

		// calculate DAC
		output = s.dac(float32(int(amplitude) * s.volumeSCH4))

	} else if (s.mem[NR42]&0xF8) > 0 && s.disableSCH4 {
		// channel is disabled but DAC is on
		// https://gbdev.io/pandocs/Audio_details.html#channels
		output = 1

	} else if (s.mem[NR42] & 0xF8) == 0 {
		// if DAC is off, disable the channel
		s.disableSCH4 = true
	}

	// write samples (PCM)
	s.sCH4Buf = append(s.sCH4Buf, output) // left
	s.sCH4Buf = append(s.sCH4Buf, output) // right
}

// reset SCH4 (trigger event)
func (s *Sound) resetSCH4() {

	if s.lengthCounterSCH4 == 0 {
		// reset length timer
		s.lengthCounterSCH4 = 64
	}

	// enable
	s.disableSCH4 = false

	// load initial volume (for envelope)
	s.volumeSCH4 = int(s.mem[NR42]&0xF0) >> 4

	// reload volume timer
	s.volumeTimerSCH4 = s.loadVolumeEnvelopeTimer(NR42)

	// reload timer
	s.timerSCH4 = s.loadNoiseTimer()

	// reset LFSR
	s.lsfrSCH4 = 0x7FFF
}

/*
 * #####################
 * # GENERAL FUNCTIONS #
 * #####################
 */

func (s *Sound) stop() {
	rl.UnloadAudioStream(s.stream)
	rl.CloseAudioDevice()
}

func (s *Sound) sync(tCycle int) {

	// falling edge of bit 5 from DIV (512hz for frame sequencers)
	divCurrentValue := (s.mem[PORT_DIV] & 0x10) >> 4

	// step FS
	fsStep := s.divLastValue == 1 && divCurrentValue == 0

	// store last value for div (falling edge)
	s.divLastValue = divCurrentValue

	// step each channel individually, to fill their respective sample buffers
	s.stepSCH1(tCycle, fsStep)
	s.stepSCH2(tCycle, fsStep)
	s.stepSCH3(tCycle, fsStep)
	s.stepSCH4(tCycle, fsStep)

	if len(s.sCH1Buf) >= maxSamplesBufferSize && len(s.sCH2Buf) >= maxSamplesBufferSize &&
		len(s.sCH3Buf) >= maxSamplesBufferSize && len(s.sCH4Buf) >= maxSamplesBufferSize {

		mixbuf := make([]float32, maxSamplesBufferSize)

		// read maxSamplesBufferSize samples
		for i := 0; i < maxSamplesBufferSize; i++ {
			var (
				res     float32 // mixed result (avg)
				volume  float32 // left or right volume
				samples float32 // to calculate the average properly
			)
			// left samples are added first, thus, even indexes are left samples and
			// odd indexes are right samples
			if i%2 == 0 {
				// left volume
				volume = float32(s.mem[NR50] & 0x70 >> 4)
			} else {
				// right volume
				volume = float32(s.mem[NR50] & 0x7)
			}

			if s.useSCH1 {
				samples++
				if i%2 == 0 {
					// CH1 left paning
					if s.mem[NR51]&0x10 > 0 {
						res += s.sCH1Buf[i]
					}
				} else {
					// CH1 right paning
					if s.mem[NR51]&0x1 > 0 {
						res += s.sCH1Buf[i]
					}
				}
			}

			if s.useSCH2 {
				samples++
				if i%2 == 0 {
					// CH2 left paning
					if s.mem[NR51]&0x20 > 0 {
						res += s.sCH2Buf[i]
					}
				} else {
					// CH2 right paning
					if s.mem[NR51]&0x2 > 0 {
						res += s.sCH2Buf[i]
					}
				}
			}
			if s.useSCH3 {
				samples++
				if i%2 == 0 {
					// CH3 left paning
					if s.mem[NR51]&0x40 > 0 {
						res += s.sCH3Buf[i]
					}
				} else {
					// CH3 right paning
					if s.mem[NR51]&0x4 > 0 {
						res += s.sCH3Buf[i]
					}
				}
			}
			if s.useSCH4 {
				samples++
				if i%2 == 0 {
					// CH4 left paning
					if s.mem[NR51]&0x80 > 0 {
						res += s.sCH4Buf[i]
					}
				} else {
					// CH4 right paning
					if s.mem[NR51]&0x8 > 0 {
						res += s.sCH4Buf[i]
					}
				}
			}

			// mixing
			f := (res * (volume + 1)) / samples

			// apply volume for sample and mix by averaging the channel's amplitude
			mixbuf[i] = f
		}

		if rl.IsAudioStreamProcessed(s.stream) {
			rl.UpdateAudioStream(s.stream, mixbuf)
			s.sCH1Buf = s.sCH1Buf[:0]
			s.sCH2Buf = s.sCH2Buf[:0]
			s.sCH3Buf = s.sCH3Buf[:0]
			s.sCH4Buf = s.sCH4Buf[:0]
		}
	}
}

func (s *Sound) powerOff() {

	// clears all APU registers
	// https://gbdev.io/pandocs/Audio_Registers.html#ff26--nr52-audio-master-control
	// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware#Power_Control
	for reg := NR10; reg <= NR51; reg++ {
		s.mem[reg] = 0
	}

	// frame sequencer step will begin from 0
	s.frameSeqStepSCH1 = 0
	s.frameSeqStepSCH2 = 0
	s.frameSeqStepSCH3 = 0
	s.frameSeqStepSCH4 = 0

	// reset timer
	s.timerSCH1 = 0
	s.timerSCH2 = 0
	s.timerSCH3 = 0
	s.timerSCH4 = 0

	// duty units are reset
	s.dutyPositionSCH2 = 0
}
