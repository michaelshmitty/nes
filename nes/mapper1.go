package nes



func (m *Mapper1) loadRegister(address uint16, value byte) {
	if value&0x80 == 0x80 {
		m.shiftRegister = 0x10
		m.writeControl(m.control | 0x0C)
	} else {
		complete := m.shiftRegister&1 == 1
		m.shiftRegister >>= 1
		m.shiftRegister |= (value & 1) << 4
		if complete {
			m.writeRegister(address, m.shiftRegister)
			m.shiftRegister = 0x10
		}
	}
}

func (m *Mapper1) writeRegister(address uint16, value byte) {
	switch {
	case address <= 0x9FFF:
		m.writeControl(value)
	case address <= 0xBFFF:
		m.writeCHRBank0(value)
	case address <= 0xDFFF:
		m.writeCHRBank1(value)
	case address <= 0xFFFF:
		m.writePRGBank(value)
	}
}

// Control (internal, $8000-$9FFF)
func (m *Mapper1) writeControl(value byte) {
	m.control = value
	m.chrMode = (value >> 4) & 1
	m.prgMode = (value >> 2) & 3
	mirror := value & 3
	switch mirror {
	case 0:
		m.Cartridge.Mirror = MirrorSingle0
	case 1:
		m.Cartridge.Mirror = MirrorSingle1
	case 2:
		m.Cartridge.Mirror = MirrorVertical
	case 3:
		m.Cartridge.Mirror = MirrorHorizontal
	}
	m.updateOffsets()
}

// CHR bank 0 (internal, $A000-$BFFF)
func (m *Mapper1) writeCHRBank0(value byte) {
	m.chrBank0 = value
	m.updateOffsets()
}

// CHR bank 1 (internal, $C000-$DFFF)
func (m *Mapper1) writeCHRBank1(value byte) {
	m.chrBank1 = value
	m.updateOffsets()
}

// PRG bank (internal, $E000-$FFFF)
func (m *Mapper1) writePRGBank(value byte) {
	m.prgBank = value & 0x0F
	m.updateOffsets()
}

func (m *Mapper1) prgBankOffset(index int) int {
	if index >= 0x80 {
		index -= 0x100
	}
	index %= len(m.PRG) / 0x4000
	offset := index * 0x4000
	if offset < 0 {
		offset += len(m.PRG)
	}
	return offset
}

func (m *Mapper1) chrBankOffset(index int) int {
	if index >= 0x80 {
		index -= 0x100
	}
	index %= len(m.CHR) / 0x1000
	offset := index * 0x1000
	if offset < 0 {
		offset += len(m.CHR)
	}
	return offset
}

// PRG ROM bank mode (0, 1: switch 32 KB at $8000, ignoring low bit of bank number;
//                    2: fix first bank at $8000 and switch 16 KB bank at $C000;
//                    3: fix last bank at $C000 and switch 16 KB bank at $8000)
// CHR ROM bank mode (0: switch 8 KB at a time; 1: switch two separate 4 KB banks)
func (m *Mapper1) updateOffsets() {
	switch m.prgMode {
	case 0, 1:
		m.prgOffsets[0] = m.prgBankOffset(int(m.prgBank & 0xFE))
		m.prgOffsets[1] = m.prgBankOffset(int(m.prgBank | 0x01))
	case 2:
		m.prgOffsets[0] = 0
		m.prgOffsets[1] = m.prgBankOffset(int(m.prgBank))
	case 3:
		m.prgOffsets[0] = m.prgBankOffset(int(m.prgBank))
		m.prgOffsets[1] = m.prgBankOffset(-1)
	}
	switch m.chrMode {
	case 0:
		m.chrOffsets[0] = m.chrBankOffset(int(m.chrBank0 & 0xFE))
		m.chrOffsets[1] = m.chrBankOffset(int(m.chrBank0 | 0x01))
	case 1:
		m.chrOffsets[0] = m.chrBankOffset(int(m.chrBank0))
		m.chrOffsets[1] = m.chrBankOffset(int(m.chrBank1))
	}
}
