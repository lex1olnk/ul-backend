package graphql

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// withServer подменяет endpoint и клиент на тестовый сервер.
func withServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	srv := httptest.NewServer(handler)
	origClient := httpClient

	httpClient = srv.Client()
	httpClient.Timeout = 5 * time.Second

	// endpoint — пакетная переменная только для теста подменяется через хук ниже
	origEndpoint := endpointOverride
	endpointOverride = srv.URL

	t.Cleanup(func() {
		srv.Close()
		httpClient = origClient
		endpointOverride = origEndpoint
	})
}

// Ошибка GraphQL приходит с HTTP 200 — раньше это молча считалось успехом.
func TestGraphQLErrorsWithStatus200(t *testing.T) {
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":[{"message":"query is not allowed","extensions":{"code":"validation-failed"}}]}`))
	})

	var out struct {
		Data struct{ Match struct{ ID int } }
	}
	err := SendGraphQLRequest(context.Background(), "{x}", nil, &out)

	if err == nil {
		t.Fatal("ожидалась ошибка для ответа с полем errors")
	}
	if !strings.Contains(err.Error(), "query is not allowed") {
		t.Errorf("сообщение Hasura должно попасть в ошибку, получено: %v", err)
	}
	if !strings.Contains(err.Error(), "validation-failed") {
		t.Errorf("код ошибки должен попасть в сообщение, получено: %v", err)
	}
}

// HTTP 4xx/5xx раньше тоже не отличался от успеха.
func TestNon200Status(t *testing.T) {
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden by upstream"))
	})

	var out struct{}
	err := SendGraphQLRequest(context.Background(), "{x}", nil, &out)

	if err == nil {
		t.Fatal("ожидалась ошибка для HTTP 403")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("код статуса должен быть в сообщении, получено: %v", err)
	}
}

// Успешный ответ должен разбираться как раньше.
func TestSuccessfulResponse(t *testing.T) {
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"match":{"id":42}}}`))
	})

	var out struct {
		Data struct {
			Match struct {
				ID int `json:"id"`
			} `json:"match"`
		} `json:"data"`
	}
	if err := SendGraphQLRequest(context.Background(), "{x}", nil, &out); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if out.Data.Match.ID != 42 {
		t.Errorf("ожидался id 42, получено %d", out.Data.Match.ID)
	}
}

// Таймаут не должен висеть до лимита платформы.
func TestTimeout(t *testing.T) {
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := SendGraphQLRequest(ctx, "{x}", nil, &struct{}{})
	if err == nil {
		t.Fatal("ожидалась ошибка по таймауту контекста")
	}
}
