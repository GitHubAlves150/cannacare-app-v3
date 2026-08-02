// ================================================================
// PACOTE SERVICES - PAYMENT SERVICE (MERCADO PAGO)
// ================================================================
// Cria uma "preference" (Checkout Pro) via API REST direta — sem
// SDK, pra não depender de `go get` de pacote novo.
//
// VARIÁVEIS DE AMBIENTE NECESSÁRIAS:
//   MERCADOPAGO_ACCESS_TOKEN - token de acesso (sandbox ou produção)
//   SITE_URL                 - ex: https://cannacare.com.br (back_urls)
//   BACKEND_URL               - ex: https://api.cannacare.com.br (webhook)
// ================================================================

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PaymentService struct {
	accessToken string
	siteURL     string
	backendURL  string
	client      *http.Client
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		accessToken: os.Getenv("MERCADOPAGO_ACCESS_TOKEN"),
		siteURL:     os.Getenv("SITE_URL"),
		backendURL:  os.Getenv("BACKEND_URL"),
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

type mpItem struct {
	Title      string  `json:"title"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	CurrencyID string  `json:"currency_id"`
}

type mpBackURLs struct {
	Success string `json:"success"`
	Failure string `json:"failure"`
	Pending string `json:"pending"`
}

type mpPayer struct {
	Email string `json:"email,omitempty"`
}

type mpPreferenceRequest struct {
	Items             []mpItem   `json:"items"`
	Payer             *mpPayer   `json:"payer,omitempty"`
	BackURLs          mpBackURLs `json:"back_urls"`
	AutoReturn        string     `json:"auto_return,omitempty"`
	NotificationURL   string     `json:"notification_url"`
	ExternalReference string     `json:"external_reference"`
}

type mpPreferenceResponse struct {
	ID        string `json:"id"`
	InitPoint string `json:"init_point"`
}

// CreateCheckoutPreference cria a preferência de pagamento e devolve
// a URL do checkout hospedado do Mercado Pago (init_point).
// associationID vai como external_reference — é assim que o webhook
// vai saber qual associação ativar quando o pagamento for confirmado.
func (p *PaymentService) CreateCheckoutPreference(associationID uuid.UUID, payerEmail string, price float64) (checkoutURL string, preferenceID string, err error) {
	if p.accessToken == "" {
		return "", "", fmt.Errorf("MERCADOPAGO_ACCESS_TOKEN não configurado")
	}

	reqBody := mpPreferenceRequest{
		Items: []mpItem{
			{
				Title:      "CannaCare — Plano Premium (12 meses)",
				Quantity:   1,
				UnitPrice:  price,
				CurrencyID: "BRL",
			},
		},
		Payer: &mpPayer{
			Email: payerEmail,
		},
		BackURLs: mpBackURLs{
			Success: p.siteURL + "/cadastro/sucesso",
			Failure: p.siteURL + "/cadastro/erro",
			Pending: p.siteURL + "/cadastro/pendente",
		},
		NotificationURL:   p.backendURL + "/api/public/billing/webhook/mercadopago",
		ExternalReference: associationID.String(),
	}

	// ⚠️ O Mercado Pago só aceita "auto_return" quando back_urls.success é
	// uma URL https válida — em desenvolvimento local (http://localhost)
	// ele rejeita com "auto_return invalid". Por isso só ativamos quando
	// o SITE_URL já for https (produção). Em dev, o cliente só vê um
	// botão "Voltar ao site" em vez de ser redirecionado automaticamente.
	if strings.HasPrefix(p.siteURL, "https://") {
		reqBody.AutoReturn = "approved"
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest("POST", "https://api.mercadopago.com/checkout/preferences", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("erro ao conectar com o Mercado Pago: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("mercado pago retornou status %d ao criar preferência: %s", resp.StatusCode, string(errorBody))
	}

	var mpResp mpPreferenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&mpResp); err != nil {
		return "", "", fmt.Errorf("resposta inválida do Mercado Pago")
	}

	return mpResp.InitPoint, mpResp.ID, nil
}

// ================================================================
// MPPayment - dados do pagamento retornados pela API do Mercado Pago
// ================================================================
type MPPayment struct {
	ID                string
	Status            string // approved, pending, rejected, cancelled, refunded...
	ExternalReference string // o associationID que mandamos na criação da preferência
	PayerEmail        string
}

type mpPaymentResponse struct {
	ID                int64  `json:"id"`
	Status            string `json:"status"`
	ExternalReference string `json:"external_reference"`
	Payer             struct {
		Email string `json:"email"`
	} `json:"payer"`
}

// ================================================================
// GetPayment - busca o pagamento DIRETO na API do Mercado Pago
// ================================================================
// ⚠️ NUNCA confiar no corpo do webhook sozinho — qualquer um pode
// forjar uma notificação. O webhook só diz "algo mudou no pagamento
// X"; é essa função que confirma de verdade o que aconteceu.
func (p *PaymentService) GetPayment(paymentID string) (*MPPayment, error) {
	if p.accessToken == "" {
		return nil, fmt.Errorf("MERCADOPAGO_ACCESS_TOKEN não configurado")
	}

	req, err := http.NewRequest("GET", "https://api.mercadopago.com/v1/payments/"+paymentID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com o Mercado Pago: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mercado pago retornou status %d ao buscar pagamento: %s", resp.StatusCode, string(errorBody))
	}

	var mpResp mpPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&mpResp); err != nil {
		return nil, fmt.Errorf("resposta inválida do Mercado Pago")
	}

	return &MPPayment{
		ID:                fmt.Sprintf("%d", mpResp.ID),
		Status:            mpResp.Status,
		ExternalReference: mpResp.ExternalReference,
		PayerEmail:        mpResp.Payer.Email,
	}, nil
}