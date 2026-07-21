// ================================================================
// PACOTE UTILS - VALIDATORS
// ================================================================
// Funções de validação compartilhadas entre todos os serviços
// ================================================================

package utils

import (
	"regexp"
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
		return true // Telefone é opcional
	}
	// Remove espaços, parênteses, traços
	clean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(phone, "")
	
	// Telefone deve ter entre 10 e 11 dígitos (com DDD)
	if len(clean) < 10 || len(clean) > 11 {
		return false
	}
	
	// Verifica se começa com DDD válido (1-9)
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