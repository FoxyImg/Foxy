package metadata

import (
	"errors"
	"strings"
)

var errInvalidFormat = errors.New("invalid format")

func ParseHexColor(s string) (c ColorRGBA, err error) {
	if s[0] == '#' {
		s = strings.TrimLeft(s, "#")
	}

	c = ColorRGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 255,
	}

	hexToByte := func(b uint8) uint8 {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 8:
		c.R = hexToByte(s[0])<<4 + hexToByte(s[1])
		c.G = hexToByte(s[2])<<4 + hexToByte(s[3])
		c.B = hexToByte(s[4])<<4 + hexToByte(s[5])
		c.A = hexToByte(s[6])<<4 + hexToByte(s[7])
	case 6:
		c.R = hexToByte(s[0])<<4 + hexToByte(s[1])
		c.G = hexToByte(s[2])<<4 + hexToByte(s[3])
		c.B = hexToByte(s[4])<<4 + hexToByte(s[5])
	case 4:
		c.R = hexToByte(s[0]) * 17
		c.G = hexToByte(s[1]) * 17
		c.B = hexToByte(s[2]) * 17
	default:
		err = errInvalidFormat
	}
	return
}
