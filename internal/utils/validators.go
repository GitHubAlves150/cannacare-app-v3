// ================================================================
// PACOTE UTILS - VALIDATORS
// ================================================================
// Funções de validação compartilhadas entre todos os serviços
// ================================================================

package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// IsValidEmail - Valida formato do email
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(regex, email)
	return matched
}

// IsValidPhone - Valida formato do telefone
func IsValidPhone(phone string) bool {
	if phone == "" {
		return true
	}
	clean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(phone, "")
	if len(clean) < 10 || len(clean) > 11 {
		return false
	}
	if clean[0] < '1' || clean[0] > '9' {
		return false
	}
	return true
}

// IsValidState - Valida se é uma UF válida
func IsValidState(state string) bool {
	states := map[string]bool{
		"AC": true, "AL": true, "AP": true, "AM": true, "BA": true,
		"CE": true, "DF": true, "ES": true, "GO": true, "MA": true,
		"MT": true, "MS": true, "MG": true, "PA": true, "PB": true,
		"PR": true, "PE": true, "PI": true, "RJ": true, "RN": true,
		"RS": true, "RO": true, "RR": true, "SC": true, "SP": true,
		"SE": true, "TO": true,
	}
	return states[strings.ToUpper(state)]
}

// ================================================================
// NOVA FUNÇÃO: VALIDAÇÃO DE CPF
// ================================================================
// IsValidCPF - Valida se um CPF é válido (algoritmo de validação)
func IsValidCPF(cpf string) bool {
	// Remove caracteres não numéricos
	cpf = regexp.MustCompile(`[^0-9]`).ReplaceAllString(cpf, "")

	// CPF deve ter 11 dígitos
	if len(cpf) != 11 {
		return false
	}

	// Verifica se todos os dígitos são iguais (ex: 111.111.111-11)
	allEqual := true
	for i := 1; i < 11; i++ {
		if cpf[i] != cpf[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		return false
	}

	// Calcula o primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		num, _ := strconv.Atoi(string(cpf[i]))
		sum += num * (10 - i)
	}
	firstDigit := 11 - (sum % 11)
	if firstDigit >= 10 {
		firstDigit = 0
	}

	// Calcula o segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		num, _ := strconv.Atoi(string(cpf[i]))
		sum += num * (11 - i)
	}
	secondDigit := 11 - (sum % 11)
	if secondDigit >= 10 {
		secondDigit = 0
	}

	// Verifica se os dígitos calculados são iguais aos do CPF
	firstDigitStr := strconv.Itoa(firstDigit)
	secondDigitStr := strconv.Itoa(secondDigit)
	return string(cpf[9]) == firstDigitStr && string(cpf[10]) == secondDigitStr
}