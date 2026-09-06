package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	m "fastcup/_pkg/models"
)

const endpoint = "https://hasura.fastcup.net/v1/graphql"

// endpointOverride подменяется в тестах; в проде всегда пустая.
var endpointOverride string

func targetURL() string {
	if endpointOverride != "" {
		return endpointOverride
	}
	return endpoint
}

// httpClient с явным таймаутом: http.DefaultClient ждёт бесконечно,
// и зависший запрос держал бы serverless-функцию до лимита платформы.
var httpClient = &http.Client{Timeout: 20 * time.Second}

// graphQLError — ошибка, пришедшая в теле ответа GraphQL.
// Hasura отдаёт такие с HTTP 200, поэтому их нужно разбирать отдельно.
type graphQLError struct {
	Message    string `json:"message"`
	Extensions struct {
		Code string `json:"code"`
		Path string `json:"path"`
	} `json:"extensions"`
}

func (e graphQLError) String() string {
	if e.Extensions.Code != "" {
		return fmt.Sprintf("%s (code: %s)", e.Message, e.Extensions.Code)
	}
	return e.Message
}

// SendGraphQLRequest выполняет запрос к Hasura и раскладывает ответ в responseBody.
//
// Возвращает конкретную причину сбоя: раньше функция отдавала голый bool,
// из-за чего сетевая ошибка, HTTP 4xx/5xx и GraphQL-ошибка в теле ответа
// были неотличимы друг от друга.
func SendGraphQLRequest(ctx context.Context, query string, variables map[string]int, responseBody interface{}) error {
	requestBody := m.GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL(), bytes.NewReader(requestBodyJSON))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to hasura failed: %w", err)
	}
	defer resp.Body.Close()

	// Тело читаем целиком: оно нужно и для проверки ошибок, и для разбора данных,
	// а при сбое — для сообщения с реальным ответом сервера
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return fmt.Errorf("failed to read hasura response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hasura returned HTTP %d: %s", resp.StatusCode, snippet(raw))
	}

	// Hasura отдаёт ошибки (allow-list, невалидный запрос) с кодом 200,
	// поэтому поле errors проверяем до разбора данных
	var errEnvelope struct {
		Errors []graphQLError `json:"errors"`
	}
	if err := json.Unmarshal(raw, &errEnvelope); err == nil && len(errEnvelope.Errors) > 0 {
		messages := make([]string, 0, len(errEnvelope.Errors))
		for _, e := range errEnvelope.Errors {
			messages = append(messages, e.String())
		}
		return fmt.Errorf("hasura returned errors: %s", strings.Join(messages, "; "))
	}

	if err := json.Unmarshal(raw, responseBody); err != nil {
		return fmt.Errorf("failed to decode hasura response: %w (body: %s)", err, snippet(raw))
	}

	return nil
}

// snippet обрезает тело ответа для сообщения об ошибке.
func snippet(b []byte) string {
	const limit = 300
	s := strings.TrimSpace(string(b))
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
