package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type signPayload struct {
	Content string `json:"content"`
	Token   string `json:"token,omitempty"`
}

type validatePayload struct {
	Content   string `json:"content"`
	Signature string `json:"signature"`
}

// SignatureResponse espelha o JSON devolvido pelos endpoints /sign e /validate.
type SignatureResponse struct {
	Signature string `json:"signature"`
	Valid      bool   `json:"valid"`
	Message    string `json:"message"`
}

// Sign chama POST /sign no servidor em execução.
func Sign(port int, content, token string) (*SignatureResponse, error) {
	return post(fmt.Sprintf("http://localhost:%d/sign", port),
		signPayload{Content: content, Token: token})
}

// Validate chama POST /validate no servidor em execução.
func Validate(port int, content, signature string) (*SignatureResponse, error) {
	return post(fmt.Sprintf("http://localhost:%d/validate", port),
		validatePayload{Content: content, Signature: signature})
}

func post(url string, payload any) (*SignatureResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("erro na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	var result SignatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("resposta inválida do servidor: %w", err)
	}
	return &result, nil
}
