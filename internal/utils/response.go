// ================================================================
// PACOTE UTILS - RESPONSE
// ================================================================
// Funções auxiliares para padronizar as respostas HTTP
//
// TODAS AS RESPOSTAS SÃO EM JSON COM O FORMATO:
//   {
//     "success": true/false,
//     "data": {...},    // Em caso de sucesso
//     "error": "..."    // Em caso de erro
//   }
// ================================================================

package utils

import (
	"encoding/json"
	"net/http"
)

// ================================================================
// FUNÇÃO SENDSUCCESS()
// ================================================================
// Envia uma resposta de sucesso em JSON
//
// PARÂMETROS:
//   w      - ResponseWriter
//   status - Código HTTP (ex: 200, 201)
//   data   - Dados a serem retornados (pode ser struct, slice, map)
//
// EXEMPLO:
//   sendSuccess(w, 200, user)
//   // Retorna: {"success": true, "data": {...}}
func SendSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// ================================================================
// FUNÇÃO SENDERROR()
// ================================================================
// Envia uma resposta de erro em JSON
//
// PARÂMETROS:
//   w      - ResponseWriter
//   status - Código HTTP (ex: 400, 401, 404, 500)
//   message - Mensagem de erro
//
// EXEMPLO:
//   sendError(w, 400, "email já cadastrado")
//   // Retorna: {"success": false, "error": "email já cadastrado"}
func SendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}