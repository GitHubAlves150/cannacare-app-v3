// ================================================================
// PACOTE SERVICES - EMAIL SERVICE (RESEND)
// ================================================================
// Chama a API REST do Resend direto via net/http — sem SDK, pra não
// depender de `go get` de pacote novo.
//
// VARIÁVEIS DE AMBIENTE NECESSÁRIAS:
//   RESEND_API_KEY  - chave de API gerada em resend.com
//   EMAIL_FROM      - remetente, ex: "CannaCare <onboarding@seudominio.com.br>"
//                      precisa ser de um domínio verificado no Resend
// ================================================================

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type EmailService struct {
	apiKey string
	from   string
	client *http.Client
}

func NewEmailService() *EmailService {
	return &EmailService{
		apiKey: os.Getenv("RESEND_API_KEY"),
		from:   os.Getenv("EMAIL_FROM"),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type resendPayload struct {
	From    string `json:"from"`
	To      []string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

func (e *EmailService) send(to, subject, html string) error {
	if e.apiKey == "" || e.from == "" {
		return fmt.Errorf("RESEND_API_KEY ou EMAIL_FROM não configurados")
	}

	payload := resendPayload{
		From:    e.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("resend retornou status %d ao enviar email", resp.StatusCode)
	}
	return nil
}

// ================================================================
// SendInviteEmail - link de primeiro acesso
// ================================================================
// plano: "gratuito" ou "premium" — só muda o texto, o link é o mesmo.
func (e *EmailService) SendInviteEmail(toEmail, nomeResponsavel, nomeAssociacao, inviteLink, plano string) error {
	subject := "Bem-vindo ao CannaCare — acesse sua conta"

	planoTexto := "plano gratuito"
	if plano == "premium" {
		planoTexto = "plano premium"
	}

	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
			<h2 style="color:#16332a;">Olá, %s 👋</h2>
			<p>A associação <strong>%s</strong> já está cadastrada no CannaCare, no %s.</p>
			<p>Clique no botão abaixo para definir sua senha e acessar o sistema:</p>
			<p style="text-align:center; margin: 32px 0;">
				<a href="%s" style="background:#c9a227; color:#0e1f18; padding:14px 28px; border-radius:999px; text-decoration:none; font-weight:bold;">
					Definir minha senha e acessar
				</a>
			</p>
			<p style="color:#666; font-size:13px;">Este link expira em 7 dias e só pode ser usado uma vez.</p>
		</div>
	`, nomeResponsavel, nomeAssociacao, planoTexto, inviteLink)

	return e.send(toEmail, subject, html)
}

// ================================================================
// SendPlanActivatedEmail - usado depois, quando o webhook do
// Mercado Pago confirmar o pagamento do plano premium
// ================================================================
func (e *EmailService) SendPlanActivatedEmail(toEmail, nomeResponsavel, nomeAssociacao, inviteLink string) error {
	subject := "Pagamento confirmado — CannaCare Premium ativado"

	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
			<h2 style="color:#16332a;">Pagamento confirmado, %s! 🎉</h2>
			<p>O plano premium da associação <strong>%s</strong> foi ativado por 12 meses.</p>
			<p style="text-align:center; margin: 32px 0;">
				<a href="%s" style="background:#c9a227; color:#0e1f18; padding:14px 28px; border-radius:999px; text-decoration:none; font-weight:bold;">
					Definir minha senha e acessar
				</a>
			</p>
			<p style="color:#666; font-size:13px;">Este link expira em 7 dias e só pode ser usado uma vez.</p>
		</div>
	`, nomeResponsavel, nomeAssociacao, inviteLink)

	return e.send(toEmail, subject, html)
}