package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidCpfCnpj_CPF(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"cpf valido sem formatacao", "52998224725", true},
		{"cpf valido formatado", "529.982.247-25", true},
		{"cpf com digitos verificadores invalidos", "52998224700", false},
		{"cpf com todos os digitos iguais", "11111111111", false},
		{"cpf com menos digitos que o esperado", "5299822472", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidCpfCnpj(tt.input))
		})
	}
}

func TestIsValidCpfCnpj_CNPJ(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"cnpj valido sem formatacao", "11222333000181", true},
		{"cnpj valido formatado", "11.222.333/0001-81", true},
		{"cnpj com digitos verificadores invalidos", "11222333000100", false},
		{"cnpj com todos os digitos iguais", "11111111111111", false},
		{"cnpj com menos digitos que o esperado", "1122233300018", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidCpfCnpj(tt.input))
		})
	}
}

func TestIsValidCpfCnpj_InvalidInputs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"string vazia", ""},
		{"apenas espacos", "   "},
		{"tamanho invalido (nem cpf nem cnpj)", "123456789"},
		{"caracteres nao numericos", "abc.def.ghi-jk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, IsValidCpfCnpj(tt.input))
		})
	}
}
