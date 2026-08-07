// ================================================================
// PACOTE SERVICES - STRIPE SERVICE
// ================================================================
// Substitui o payment_service.go (Mercado Pago). Usa a API REST do
// Stripe direto via net/http — sem SDK, pra não depender de `go get`.
//
// VARIÁVEIS DE AMBIENTE NECESSÁRIAS:
//   STRIPE_SECRET_KEY     - chave secreta (sk_test_... ou sk_live_...)
//   STRIPE_WEBHOOK_SECRET - segredo do endpoint de webhook (whsec_...)
//   SITE_URL              - ex: https://cannamanager.com.br (back_urls)
//   BACKEND_URL            - ex: https://api.cannamanager.com.br (webhook)
// ================================================================

package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StripeService struct {
	secretKey     string
	webhookSecret string
	siteURL       string
	backendURL    string
	client        *http.Client
}

func NewStripeService() *StripeService {
	return &StripeService{
		secretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		webhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		siteURL:       os.Getenv("SITE_URL"),
		backendURL:    os.Getenv("BACKEND_URL"),
		client:        &http.Client{Timeout: 10 * time.Second},
	}
}

// ================================================================
// CREATECHECKOUTSESSION - cria a sessão de checkout hospedada
// ================================================================
// priceInReais vem no formato normal (197.00) — a conversão pra
// centavos (formato que o Stripe exige) é feita aqui dentro.
func (s *StripeService) CreateCheckoutSession(associationID uuid.UUID, payerEmail string, priceInReais float64) (checkoutURL string, sessionID string, err error) {
	if s.secretKey == "" {
		return "", "", fmt.Errorf("STRIPE_SECRET_KEY não configurado")
	}

	amountInCents := int64(priceInReais*100 + 0.5) // arredonda pra evitar erro de ponto flutuante

	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("line_items[0][price_data][currency]", "brl")
	form.Set("line_items[0][price_data][product_data][name]", "CannaCare — Plano Premium (12 meses)")
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(amountInCents, 10))
	form.Set("line_items[0][quantity]", "1")
	form.Set("success_url", s.siteURL+"/cadastro/sucesso")
	form.Set("cancel_url", s.siteURL+"/cadastro/erro")
	form.Set("customer_email", payerEmail)
	form.Set("metadata[association_id]", associationID.String())
	form.Set("payment_intent_data[metadata][association_id]", associationID.String())

	req, err := http.NewRequest("POST", "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.secretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("erro ao conectar com o Stripe: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("stripe retornou status %d ao criar checkout: %s", resp.StatusCode, string(errorBody))
	}

	var result struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("resposta inválida do Stripe")
	}

	return result.URL, result.ID, nil
}

// ================================================================
// WEBHOOK EVENT - o que a gente precisa do evento assinado do Stripe
// ================================================================
type StripeWebhookEvent struct {
	Type string `json:"type"`
	Data struct {
		Object struct {
			ID            string `json:"id"`
			PaymentStatus string `json:"payment_status"` // "paid", "unpaid"
			Status        string `json:"status"`
			Metadata      struct {
				AssociationID string `json:"association_id"`
			} `json:"metadata"`
		} `json:"object"`
	} `json:"data"`
}

// ================================================================
// VERIFYWEBHOOKSIGNATURE - confirma que o webhook é de verdade do Stripe
// ================================================================
// ⚠️ Igual fizemos com o Mercado Pago (buscar o pagamento de volta na
// API deles), aqui o equivalente é verificar a assinatura HMAC — sem
// isso, qualquer um poderia forjar uma notificação de "pagamento
// aprovado" só mandando um POST pro nosso endpoint.
func (s *StripeService) VerifyWebhookSignature(payload []byte, sigHeader string) (*StripeWebhookEvent, error) {
	if s.webhookSecret == "" {
		return nil, fmt.Errorf("STRIPE_WEBHOOK_SECRET não configurado")
	}

	// O header vem no formato: t=1234567890,v1=abcdef...
	var timestamp, signature string
	for _, part := range strings.Split(sigHeader, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signature = kv[1]
		}
	}
	if timestamp == "" || signature == "" {
		return nil, fmt.Errorf("assinatura do webhook em formato inválido")
	}

	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return nil, fmt.Errorf("assinatura do webhook inválida — evento pode não ser do Stripe de verdade")
	}

	var event StripeWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("corpo do webhook inválido")
	}

	return &event, nil
}