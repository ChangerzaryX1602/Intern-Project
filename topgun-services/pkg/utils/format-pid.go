package utils

import "unicode"

func CheckThaiPID(pid string) bool {
	if len(pid) != 13 {
		return false
	}

	sum := 0
	for i := range 12 {
		if !unicode.IsDigit(rune(pid[i])) {
			return false
		}
		digit := int(pid[i] - '0')
		sum += digit * (13 - i)
	}

	checkDigit := int(pid[12] - '0')
	calculated := (11 - (sum % 11)) % 10

	return checkDigit == calculated
}
