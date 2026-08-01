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

func checkCpfDigits(digits string, position int) bool {
	sum := 0
	for i := 0; i < position-1; i++ {
		digit := int(digits[i] - '0')
		sum += digit * (position - i)
	}

	remainder := sum % 11
	expected := 0
	if remainder < 2 {
		expected = 0
	} else {
		expected = 11 - remainder
	}
	return expected == int(digits[position-1]-'0')
}

func isValidCnpj(cnpj string) bool {
	firstWeights := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	secondWeights := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	return checkCnpjDigits(cnpj, firstWeights, 12) && checkCnpjDigits(cnpj, secondWeights, 13)
}

func checkCnpjDigits(digits string, weights []int, position int) bool {
	sum := 0
	for i := 0; i < len(weights); i++ {
		digit := int(digits[i] - '0')
		sum += digit * weights[i]
	}
	remainder := sum % 11
	expected := 0
	if remainder < 2 {
		expected = 0
	} else {
		expected = 11 - remainder
	}
	return expected == int(digits[position]-'0')
}
