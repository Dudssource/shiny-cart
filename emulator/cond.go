package emulator

func eval(flags flag, opcode uint8) bool {

	const (
		cond_nz = 0x0
		cond_z  = 0x1
		cond_nc = 0x2
		cond_c  = 0x3
	)

	// get the condition
	cc := (opcode & 0x18) >> 3

	var match bool

	switch cc {
	case cond_nz:
		match = (flags & z_flag) == 0
	case cond_z:
		match = (flags & z_flag) > 0
	case cond_c:
		match = (flags & c_flag) > 0
	case cond_nc:
		match = (flags & c_flag) == 0
	}

	return match
}
