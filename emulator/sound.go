package emulator

import (
	"fmt"

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

type Sound struct {
	mem    memoryArea
	stream rl.AudioStream
	volume float32

	// DIV lastCycleValue, used for falling edge and ticking the frame-sequencer
	divLastValue uint8

	// channel 1 - square wave with sweep
	sC1Buf            []float32
	timerSC1          uint16 // channel SC1 timer (frequency)
	dutyPositionSC1   uint8  // channel SC1 wave duty position
	frameSeqStepSC1   int    // frame-sequencer SC1
	lengthCounterSC1  int    // length counter SC1
	disableSC1        bool   // disable SC1
	volumeSC1         int    // volume for SC1 (controlled by envelope)
	volumeTimerSC1    int    // volume timer for SC1
	sweepTimerSC1     int    // frequency sweep timer for SC1
	sweepEnableSC1    bool   // sweep enabled for SC1
	sweepShadowRegSC1 uint16 // shadow sweep register for SC1

	// channel 2 - square wave
	sC2Buf           []float32
	timerSC2         uint16 // channel SC2 timer (frequency)
	dutyPositionSC2  uint8  // channel SC2 wave duty position
	frameSeqStepSC2  int    // frame-sequencer SC2
	lengthCounterSC2 int    // length counter SC2
	disableSC2       bool   // disable SC2
	volumeSC2        int    // volume for SC2 (controlled by envelope)
	volumeTimerSC2   int    // volume timer for SC2

	// channel 3 - wave table
	sC3Buf           []float32
	lengthCounterSC3 int    // length counter SC3
	timerSC3         uint16 // channel SC3 timer (frequency)
	wavePositionSC3  uint8  // channel SC3 wave position
	frameSeqStepSC3  int    // frame-sequencer SC3
	disableSC3       bool   // disable SC3

	// channel 4 - noise
	lengthCounterSC4 int // length counter SC4

	sC4    []float32
	useSC1 bool
	useSC2 bool
	useSC3 bool
	useSC4 bool
}

func NewSound(mem memoryArea) *Sound {
	return &Sound{
		mem:    mem,
		volume: 0.5,
	}
}

const (
	// number of samples to buffer before enqueuing on the audio device
	maxSamplesBufferSize = 100
	sampleRate           = 44100
	bufferSize           = 4096
)

func (s *Sound) init() {
	rl.InitAudioDevice()
	rl.SetAudioStreamBufferSizeDefault(bufferSize * 10)
	s.stream = rl.LoadAudioStream(sampleRate, 32, 2)
	rl.PlayAudioStream(s.stream)

	s.useSC1 = true
	s.useSC2 = true
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
func (*Sound) dac(amplitude, volume int) float32 {
	dacInput := float32(amplitude * volume)
	dacOutput := (dacInput / 7.5) - 1.0

	// digital 0 maps to analog 1, not -1
	// https://gbdev.io/pandocs/Audio_details.html#dacs
	if dacInput == 0 {
		dacOutput = 1
	}
	return dacOutput
}

/*
 * #####################################
 * # SOUND CHANNEL 1 RELATED FUNCTIONS #
 * #####################################
 */

func (s *Sound) loadCH1SweepPeriod() int {
	// load sweep timer
	return int(s.mem[NR10] & 0x70)
}

// ONLY applies to SCH1 (square+sweep)
func (s *Sound) reloadCH1SweepTimer() {
	// reload sweep timer
	s.sweepTimerSC1 = s.loadCH1SweepPeriod()
	// envelope and sweepers treat 0 as 8
	// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware#Obscure_Behavior
	if s.sweepTimerSC1 == 0 {
		s.sweepTimerSC1 = 8
	}
}

func (s *Sound) loadCH1SweepShift() uint8 {
	return s.mem[NR10] & 0x7
}

// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware#Frequency_Sweep
// https://nightshade256.github.io/2021/03/27/gb-sound-emulation.html#sweep-function
func (s *Sound) calculateCH1SweepNewFrequency() uint16 {
	newFrequency := s.sweepShadowRegSC1 >> uint16(s.loadCH1SweepShift())

	sweepIncrease := (s.mem[NR10] & 0x8) == 0

	if !sweepIncrease {
		newFrequency = s.sweepShadowRegSC1 - newFrequency
	} else {
		newFrequency = s.sweepShadowRegSC1 + newFrequency
	}

	// 11-bit overflow
	if newFrequency > 2047 {
		s.disableSC1 = true
	}

	return newFrequency
}

// tick SCH1 channel timer
func (s *Sound) stepSC1(divCurrentValue uint8) {

	// decrement SC1 timer every t-cycle
	s.timerSC1--

	// frequency reached 0
	if s.timerSC1 == 0 {
		// reload timer
		s.timerSC1 = s.loadTimerFrequencySquareWaveChannel(NR14, NR13)

		// increment duty position (since values are 0-7, we wrap around by ANDing 0x7)
		s.dutyPositionSC1 = (s.dutyPositionSC1 + 1) & 0x7
	}

	// frame-sequencer (ticks every 8192 t-cycles = 512hz)
	if s.divLastValue == 1 && divCurrentValue == 0 {

		// sweep step
		if s.frameSeqStepSC1 == 2 || s.frameSeqStepSC1 == 6 {

			if s.sweepTimerSC1 > 0 {
				// decrement sweep timer
				s.sweepTimerSC1--

				// timer reached 0
				if s.sweepTimerSC1 == 0 {
					// reload sweep timer
					s.reloadCH1SweepTimer()

					// sweep enabled and period non-zero
					if s.sweepEnableSC1 && s.loadCH1SweepPeriod() > 0 {

						newFrequency := s.calculateCH1SweepNewFrequency()

						if newFrequency <= 2047 && s.loadCH1SweepShift() > 0 {
							//s.timerSC1 = newFrequency
							s.sweepShadowRegSC1 = newFrequency

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
		if (s.frameSeqStepSC1%2) == 0 && (s.mem[NR14]&0x40) > 0 {

			// decrement
			s.lengthCounterSC1--

			// disable the channel
			if s.lengthCounterSC1 == 0 {
				s.disableSC1 = true
			}
		}

		// volume envelope (64hz), frame sequencer step 7
		if s.frameSeqStepSC1 == 0x7 {

			// volume envelop period
			volumeEnvelopePeriod := s.loadVolumeEnvelopeTimer(NR12)

			// every sweep-pace ticks
			if volumeEnvelopePeriod > 0 {

				// decrement internal timer
				if s.volumeTimerSC1 > 0 {
					s.volumeTimerSC1--
				}

				// turned zero
				if s.volumeTimerSC1 == 0 {

					// reload volume timer
					s.volumeTimerSC1 = s.loadVolumeEnvelopeTimer(NR12)

					// 0=decrease, 1=increase
					envelopeIncreaseVolume := (s.mem[NR12] & 0x8) > 0

					// increase
					if envelopeIncreaseVolume && s.volumeSC1 < 0xF {
						s.volumeSC1++
					}

					// decrease
					if !envelopeIncreaseVolume && s.volumeSC1 > 0 {
						s.volumeSC1--
					}
				}
			}
		}

		// current frame-sequencer step (0-7)
		s.frameSeqStepSC1 = (s.frameSeqStepSC1 + 1) & 0x7
	}

	if !s.disableSC1 {
		// channel SC1 enabled
		s.mem[NR52] |= 0x1
	} else {
		// channel SC1 disabled
		s.mem[NR52] &= ^uint8(0x1)
	}

	var (
		leftOutput  float32
		rightOutput float32
	)

	// DAC is on and channel is enabled
	if (s.mem[NR12]&0xF8) > 0 && !s.disableSC1 {

		// duty
		dutyPattern := s.mem[NR21] & 0xC0

		// amplitude (lookup table and get the bit pointed by dutyPosition)
		amplitude := int((dutyWaveformTable[dutyPattern] & (1 << s.dutyPositionSC1)) >> s.dutyPositionSC1)

		// calculate DAC
		dacOutput := s.dac(amplitude, s.volumeSC1)

		// CH1 left paning
		if s.mem[NR51]&0x10 > 0 {
			leftOutput = dacOutput
		}

		// CH1 right paning
		if s.mem[NR51]&0x1 > 0 {
			rightOutput = dacOutput
		}

	} else if (s.mem[NR12]&0xF8) > 0 && s.disableSC1 {
		// channel is disabled but DAC is on
		// https://gbdev.io/pandocs/Audio_details.html#channels

		// CH1 left paning
		if s.mem[NR51]&0x10 > 0 {
			leftOutput = 1
		}

		// CH1 right paning
		if s.mem[NR51]&0x1 > 0 {
			rightOutput = 1
		}
	}

	// write samples (PCM)
	s.sC1Buf = append(s.sC1Buf, leftOutput)  // left
	s.sC1Buf = append(s.sC1Buf, rightOutput) // right
}

// reset SCH1 (trigger event)
func (s *Sound) resetSC1() {

	if s.lengthCounterSC1 > 0 {
		// reset length timer
		s.lengthCounterSC1 = 64 - int(s.mem[NR11]&0x3F)
	} else {
		// reset length timer
		s.lengthCounterSC1 = 64
	}

	// enable
	s.disableSC1 = false

	// load initial volume (for envelope)
	s.volumeSC1 = int(s.mem[NR12]&0xF0) >> 4

	// reload volume timer
	s.volumeTimerSC1 = s.loadVolumeEnvelopeTimer(NR12)

	// load frequency to shadow
	s.sweepShadowRegSC1 = s.loadTimerPeriodSquareWaveChannel(NR14, NR13)

	// reload sweep timer
	s.reloadCH1SweepTimer()

	// sweep enabled
	s.sweepEnableSC1 = s.sweepTimerSC1 > 0 || s.loadCH1SweepShift() > 0

	// if sweep enabled AND shift is non-zero we only check overflow (by throwing away the result)
	if s.sweepEnableSC1 && s.loadCH1SweepShift() > 0 {
		s.calculateCH1SweepNewFrequency()
	}
}

/*
 * #####################################
 * # SOUND CHANNEL 2 RELATED FUNCTIONS #
 * #####################################
 */

// tick SCH2 channel timer
func (s *Sound) stepSC2(divCurrentValue uint8) {

	// decrement SC2 timer every t-cycle
	s.timerSC2--

	// frequency reached 0
	if s.timerSC2 == 0 {
		// reload timer
		s.timerSC2 = s.loadTimerFrequencySquareWaveChannel(NR24, NR23)

		// increment duty position (since values are 0-7, we wrap around by ANDing 0x7)
		s.dutyPositionSC2 = (s.dutyPositionSC2 + 1) & 0x7
	}

	// frame-sequencer (ticks every 8192 t-cycles = 512hz)
	if s.divLastValue == 1 && divCurrentValue == 0 {

		// length-counter step AND length-enable flag is set
		if (s.frameSeqStepSC2%2) == 0 && (s.mem[NR24]&0x40) > 0 {

			// decrement
			s.lengthCounterSC2--

			// disable the channel
			if s.lengthCounterSC2 == 0 {
				s.disableSC2 = true
			}
		}

		// volume envelop period
		volumeEnvelopePeriod := int(s.mem[NR22] & 0x7)

		// volume envelope (64hz), frame sequencer step 7
		if s.frameSeqStepSC2 == 0x7 {

			// every sweep-pace ticks
			if volumeEnvelopePeriod > 0 {

				// decrement internal timer
				if s.volumeTimerSC2 > 0 {
					s.volumeTimerSC2--
				}

				// turned zero
				if s.volumeTimerSC2 == 0 {

					// reload volume timer
					s.volumeTimerSC2 = s.loadVolumeEnvelopeTimer(NR22)

					// 0=decrease, 1=increase
					envelopeIncreaseVolume := (s.mem[NR22] & 0x8) > 0

					// increase
					if envelopeIncreaseVolume && s.volumeSC2 < 0xF {
						s.volumeSC2++
					}

					// decrease
					if !envelopeIncreaseVolume && s.volumeSC2 > 0 {
						s.volumeSC2--
					}
				}
			}
		}

		// current frame-sequencer step (0-7)
		s.frameSeqStepSC2 = (s.frameSeqStepSC2 + 1) & 0x7
	}

	if !s.disableSC2 {
		// channel SC2 enabled
		s.mem[NR52] |= 0x2
	} else {
		// channel SC2 disabled
		s.mem[NR52] &= ^uint8(0x2)
	}

	var (
		leftOutput  float32
		rightOutput float32
	)

	// DAC is on and channel is enabled
	if (s.mem[NR22]&0xF8) > 0 && !s.disableSC2 {

		// duty
		dutyPattern := s.mem[NR21] & 0xC0

		// amplitude (lookup table and get the bit pointed by dutyPosition)
		amplitude := int((dutyWaveformTable[dutyPattern] & (1 << s.dutyPositionSC2)) >> s.dutyPositionSC2)

		// calculate DAC
		dacOutput := s.dac(amplitude, s.volumeSC2)

		// CH2 left paning
		if s.mem[NR51]&0x20 > 0 {
			leftOutput = dacOutput
		}

		// CH2 right paning
		if s.mem[NR51]&0x2 > 0 {
			rightOutput = dacOutput
		}

	} else if (s.mem[NR22]&0xF8) > 0 && s.disableSC2 {
		// channel is disabled but DAC is on
		// https://gbdev.io/pandocs/Audio_details.html#channels

		// CH2 left paning
		if s.mem[NR51]&0x20 > 0 {
			leftOutput = 1
		}

		// CH2 right paning
		if s.mem[NR51]&0x2 > 0 {
			rightOutput = 1
		}
	}

	// write samples (PCM)
	s.sC2Buf = append(s.sC2Buf, leftOutput)  // left
	s.sC2Buf = append(s.sC2Buf, rightOutput) // right
}

// reset SCH2 (trigger event)
func (s *Sound) resetSC2() {

	if s.lengthCounterSC2 == 0 {
		// reset length timer
		s.lengthCounterSC2 = 64
	}

	// enable
	s.disableSC2 = false

	// load initial volume (for envelope)
	s.volumeSC2 = int(s.mem[NR22]&0xF0) >> 4

	// reload volume timer
	s.volumeTimerSC2 = s.loadVolumeEnvelopeTimer(NR22)

	// reload timer
	s.timerSC2 = s.loadTimerFrequencySquareWaveChannel(NR24, NR23)
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
func (s *Sound) stepSC3(divCurrentValue uint8) {

	// decrement SC3 timer every t-cycle
	s.timerSC3--

	// frequency reached 0
	if s.timerSC3 == 0 {
		// reload timer
		s.timerSC3 = s.loadTimerFrequencyWaveChannel(NR34, NR33)

		// increment duty position (since values are 0-31, we wrap around by mod 32/0x20)
		s.wavePositionSC3 = (s.wavePositionSC3 + 1) % 32
	}

	// frame-sequencer (ticks every 8192 t-cycles = 512hz)
	if s.divLastValue == 1 && divCurrentValue == 0 {

		// length-counter step AND length-enable flag is set
		if (s.frameSeqStepSC3%2) == 0 && (s.mem[NR34]&0x40) > 0 {

			// decrement
			s.lengthCounterSC3--

			// disable the channel
			if s.lengthCounterSC3 == 0 {
				s.disableSC3 = true
			}
		}

		// current frame-sequencer step (0-7)
		s.frameSeqStepSC3 = (s.frameSeqStepSC3 + 1) & 0x7
	}

	if !s.disableSC3 {
		// channel SC3 enabled
		s.mem[NR52] |= 0x4
	} else {
		// channel SC3 disabled
		s.mem[NR52] &= ^uint8(0x4)
	}

	var (
		leftOutput  float32
		rightOutput float32
	)

	// DAC is on and channel is enabled
	if (s.mem[NR32]&0xF8) > 0 && !s.disableSC3 {

		// duty
		dutyPattern := s.mem[NR21] & 0xC0

		// amplitude (lookup table and get the bit pointed by dutyPosition)
		amplitude := int((dutyWaveformTable[dutyPattern] & (1 << s.dutyPositionSC2)) >> s.dutyPositionSC2)

		// calculate DAC
		dacOutput := s.dac(amplitude, s.volumeSC2)

		// CH2 left paning
		if s.mem[NR51]&0x20 > 0 {
			leftOutput = dacOutput
		}

		// CH2 right paning
		if s.mem[NR51]&0x2 > 0 {
			rightOutput = dacOutput
		}

	} else if (s.mem[NR22]&0xF8) > 0 && s.disableSC2 {
		// channel is disabled but DAC is on
		// https://gbdev.io/pandocs/Audio_details.html#channels

		// CH2 left paning
		if s.mem[NR51]&0x20 > 0 {
			leftOutput = 1
		}

		// CH2 right paning
		if s.mem[NR51]&0x2 > 0 {
			rightOutput = 1
		}
	}

	// write samples (PCM)
	s.sC2Buf = append(s.sC2Buf, leftOutput)  // left
	s.sC2Buf = append(s.sC2Buf, rightOutput) // right
}

// reset SCH3 (trigger event)
func (s *Sound) resetSC3() {

	if s.lengthCounterSC3 == 0 {
		// reset length timer
		s.lengthCounterSC3 = 256
	}

	// enable
	s.disableSC3 = false

	// reload timer
	s.timerSC3 = s.loadTimerFrequencySquareWaveChannel(NR24, NR23)
}

func (s *Sound) stop() {
	rl.UnloadAudioStream(s.stream)
	rl.CloseAudioDevice()
}

func (s *Sound) sync(_ int) {

	// falling edge of bit 5 from DIV (512hz for frame sequencers)
	divCurrentValue := (s.mem[PORT_DIV] & 0x10) >> 4

	// step each channel individually, to fill their respective sample buffers
	s.stepSC1(divCurrentValue)
	s.stepSC2(divCurrentValue)

	// store last value for div (falling edge)
	s.divLastValue = divCurrentValue

	// truncate
	s.sC1Buf = s.sC1Buf[:0]
	s.sC2Buf = s.sC2Buf[:0]
	s.sC3 = s.sC3[:0]
	s.sC4 = s.sC4[:0]

	return

	if len(s.sC1Buf) >= maxSamplesBufferSize && len(s.sC2Buf) >= maxSamplesBufferSize {

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

			if s.useSC1 {
				samples++
				res += s.sC1Buf[i]
			}
			if s.useSC2 {
				samples++
				res += s.sC2Buf[i]
			}
			if s.useSC3 {
				samples++
				res += s.sC3[i]
			}
			if s.useSC4 {
				samples++
				res += s.sC4[i]
			}

			// apply volume for sample and mix by averaging the channel's amplitude
			mixbuf[i] = (res * volume) / samples
		}

		if rl.IsAudioStreamProcessed(s.stream) {

			fmt.Println(len(mixbuf))

			// update audio stream
			rl.UpdateAudioStream(s.stream, mixbuf)

			// truncate
			s.sC1Buf = s.sC1Buf[:0]
			s.sC2Buf = s.sC2Buf[:0]
			s.sC3 = s.sC3[:0]
			s.sC4 = s.sC4[:0]
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
	s.frameSeqStepSC1 = 0
	s.frameSeqStepSC2 = 0

	// reset timer
	s.timerSC2 = 0

	// duty units are reset
	s.dutyPositionSC2 = 0
}
