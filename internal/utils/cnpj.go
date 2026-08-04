// ================================================================
// UTILS - VALIDAÇÃO DE CNPJ (SERVER-SIDE)
// ================================================================
// Espelha a validação que já existe no front (lib/cnpj.ts), mas
// roda no backend — o front pode ser burlado, o servidor não pode
// confiar em "o front já validou".
// ================================================================

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var onlyDigits = regexp.MustCompile(`\D`)

func UnmaskCNPJ(cnpj string) string {
	return onlyDigits.ReplaceAllString(cnpj, "")
}

// IsValidCNPJDigits confere o dígito verificador (matemática do CNPJ)
func IsValidCNPJDigits(cnpjRaw string) bool {
	cnpj := UnmaskCNPJ(cnpjRaw)
	if len(cnpj) != 14 {
		return false
	}

	allSame := true
	for i := 1; i < len(cnpj); i++ {
		if cnpj[i] != cnpj[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	calcDigit := func(base string, weights []int) int {
		sum := 0
		for i, w := range weights {
			d, _ := strconv.Atoi(string(base[i]))
			sum += d * w
		}
		rest := sum % 11
		if rest < 2 {
			return 0
		}
		return 11 - rest
	}

	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	digit1 := calcDigit(cnpj[:12], weights1)
	digit2 := calcDigit(cnpj[:12]+strconv.Itoa(digit1), weights2)

	expected := fmt.Sprintf("%s%d%d", cnpj[:12], digit1, digit2)
	return cnpj == expected
}

// DadosCNPJ - retorno da consulta pública (BrasilAPI, espelha Receita Federal)
type DadosCNPJ struct {
	CNPJ         string
	RazaoSocial  string
	NomeFantasia string
	Situacao     string
	Endereco     string // já formatado, pronto pra salvar em associations.address
}

type brasilAPIResponse struct {
	RazaoSocial               string `json:"razao_social"`
	NomeFantasia              string `json:"nome_fantasia"`
	DescricaoSituacaoCadastral string `json:"descricao_situacao_cadastral"`
	DescricaoTipoLogradouro   string `json:"descricao_tipo_de_logradouro"`
	Logradouro                string `json:"logradouro"`
	Numero                    string `json:"numero"`
	Bairro                    string `json:"bairro"`
	Municipio                 string `json:"municipio"`
	UF                        string `json:"uf"`
	CEP                       string `json:"cep"`
}

// ConsultarCNPJ valida o dígito e confirma na Receita Federal que o
// CNPJ existe de verdade e está ATIVO.
func ConsultarCNPJ(cnpjRaw string) (*DadosCNPJ, error) {
	cnpj := UnmaskCNPJ(cnpjRaw)

	if !IsValidCNPJDigits(cnpj) {
		return nil, fmt.Errorf("CNPJ inválido")
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get("https://brasilapi.com.br/api/cnpj/v1/" + cnpj)
	if err != nil {
		return nil, fmt.Errorf("não foi possível consultar o CNPJ agora, tente novamente")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("CNPJ não encontrado na Receita Federal")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("não foi possível validar o CNPJ agora, tente novamente")
	}

	var data brasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("resposta inválida ao consultar o CNPJ")
	}

	if data.DescricaoSituacaoCadastral != "ATIVA" {
		return nil, fmt.Errorf("este CNPJ está com situação \"%s\" na Receita Federal e não pode ser cadastrado", data.DescricaoSituacaoCadastral)
	}

	endereco := strings.TrimSpace(fmt.Sprintf("%s %s, %s - %s - %s/%s - CEP %s",
		data.DescricaoTipoLogradouro, data.Logradouro, data.Numero, data.Bairro, data.Municipio, data.UF, data.CEP))

	razaoSocial := data.RazaoSocial
	nomeFantasia := data.NomeFantasia
	if nomeFantasia == "" {
		nomeFantasia = razaoSocial
	}

	return &DadosCNPJ{
		CNPJ:         cnpj,
		RazaoSocial:  razaoSocial,
		NomeFantasia: nomeFantasia,
		Situacao:     data.DescricaoSituacaoCadastral,
		Endereco:     endereco,
	}, nil
}