package validators

import (
	"regexp"
	"strings"
)

var nonDigit = regexp.MustCompile(`\D`)

func IsValidCpfCnpj(cpfCnpj string) bool {
	cpfCnpj = sanitize(cpfCnpj)

	if cpfCnpj == "" {
		return false
	}

	if len(cpfCnpj) == 11 {
		return isValidCpf(cpfCnpj)
	} else if len(cpfCnpj) == 14 {
		return isValidCnpj(cpfCnpj)
	}

	return false
}

func sanitize(value string) string {
	value = strings.TrimSpace(value)
	value = nonDigit.ReplaceAllString(value, "")
	return value
}

func isValidCpf(cpf string) bool {
	return checkCpfDigits(cpf, 10) && checkCpfDigits(cpf, 11)
}

func checkCpfDigits(digits string, checkDigitPos int) bool {
	if allDigitsEqual(digits) {
		return false
	}

	sum := 0
	for i := 0; i < checkDigitPos-1; i++ {
		digit := int(digits[i] - '0')
		sum += digit * (checkDigitPos - i)
	}

	return checkDigit(sum) == int(digits[checkDigitPos-1]-'0')
}

func isValidCnpj(cnpj string) bool {
	firstWeights := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	secondWeights := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	return checkCnpjDigits(cnpj, firstWeights, 12) && checkCnpjDigits(cnpj, secondWeights, 13)
}

func checkCnpjDigits(digits string, weights []int, checkDigitPos int) bool {
	if allDigitsEqual(digits) {
		return false
	}

	sum := 0
	for i := 0; i < checkDigitPos; i++ {
		digit := int(digits[i] - '0')
		sum += digit * weights[i]
	}

	return checkDigit(sum) == int(digits[checkDigitPos]-'0')
}

func allDigitsEqual(digits string) bool {
	for _, d := range digits {
		if d != rune(digits[0]) {
			return false
		}
	}
	return true
}

func checkDigit(sum int) int {
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
