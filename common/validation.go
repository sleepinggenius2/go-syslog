package common

func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func IsAllDigits(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if !IsDigit(c) {
			return false
		}
	}
	return true
}
