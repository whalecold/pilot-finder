package ip

import "net"

type Tools struct {
	Prefix [2]byte
}

// Bigger than we need, not too big to worry about overflow
const big = 0xFFFFFF

// Decimal to integer.
// Returns number, characters consumed, success.
func dtoi(s string) (n int, i int, ok bool) {
	n = 0
	for i = 0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
		if n >= big {
			return big, i, false
		}
	}
	if i == 0 {
		return 0, 0, false
	}
	return n, i, true
}

func New(s string) *Tools {
	tool := &Tools{}
	for i := 0; i < len(tool.Prefix); i++ {
		if len(s) == 0 {
			// Missing octets.
			return nil
		}
		if i > 0 {
			if s[0] != '.' {
				return nil
			}
			s = s[1:]
		}
		n, c, ok := dtoi(s)
		if !ok || n > 0xFF {
			return nil
		}
		if c > 1 && s[0] == '0' {
			// Reject non-zero components with leading zeroes.
			return nil
		}
		s = s[c:]
		tool.Prefix[i] = byte(n)
	}
	if len(s) != 0 {
		return nil
	}
	return tool
}

func (t *Tools) IPv4(i, j int) net.IP {
	return net.IPv4(t.Prefix[0], t.Prefix[1], byte(i), byte(j))
}
