package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMemoryIssueValidate(t *testing.T) {
	a := NewMemoryAuthenticator()
	p := Principal{Role: RolePlayer, PlayerID: uuid.New(), GameID: uuid.New(), Name: "Alice"}

	tok, issued, err := a.Issue(context.Background(), p, time.Minute)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if tok == "" {
		t.Error("Issue returned empty token")
	}
	if issued.IssuedAt.IsZero() || issued.ExpiresAt.IsZero() {
		t.Error("Issue did not stamp IssuedAt / ExpiresAt")
	}

	back, err := a.Validate(context.Background(), tok)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if back.PlayerID != p.PlayerID || back.GameID != p.GameID || back.Name != p.Name {
		t.Errorf("Validate returned mismatched principal: %+v", back)
	}
}

func TestMemoryValidateRejectsUnknown(t *testing.T) {
	a := NewMemoryAuthenticator()
	_, err := a.Validate(context.Background(), "bogus-token")
	if err != ErrInvalidCredential {
		t.Errorf("Validate bogus: got %v, want ErrInvalidCredential", err)
	}
}

func TestMemoryValidateRejectsEmpty(t *testing.T) {
	a := NewMemoryAuthenticator()
	_, err := a.Validate(context.Background(), "")
	if err != ErrInvalidCredential {
		t.Errorf("Validate empty: got %v, want ErrInvalidCredential", err)
	}
}

func TestMemoryValidateRejectsExpired(t *testing.T) {
	a := NewMemoryAuthenticator()
	// Pin "now" so we can advance time without sleeping.
	start := time.Unix(1_700_000_000, 0).UTC()
	now := start
	a.now = func() time.Time { return now }

	tok, _, err := a.Issue(context.Background(), Principal{Role: RoleAdmin}, time.Minute)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	// Walk past the expiry.
	now = start.Add(2 * time.Minute)
	if _, err := a.Validate(context.Background(), tok); err != ErrExpiredCredential {
		t.Errorf("Validate expired: got %v, want ErrExpiredCredential", err)
	}
	// Expired token is purged — second validate sees "invalid" not "expired".
	if _, err := a.Validate(context.Background(), tok); err != ErrInvalidCredential {
		t.Errorf("Validate after purge: got %v, want ErrInvalidCredential", err)
	}
	if got := a.Count(); got != 0 {
		t.Errorf("Count after expiry: got %d, want 0", got)
	}
}

func TestMemoryIssueRejectsNonPositiveTTL(t *testing.T) {
	a := NewMemoryAuthenticator()
	if _, _, err := a.Issue(context.Background(), Principal{}, 0); err == nil {
		t.Error("Issue with ttl=0: expected error, got nil")
	}
	if _, _, err := a.Issue(context.Background(), Principal{}, -time.Second); err == nil {
		t.Error("Issue with negative ttl: expected error, got nil")
	}
}

func TestMemoryRevoke(t *testing.T) {
	a := NewMemoryAuthenticator()
	tok, _, err := a.Issue(context.Background(), Principal{Role: RolePlayer}, time.Minute)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if err := a.Revoke(context.Background(), tok); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if _, err := a.Validate(context.Background(), tok); err != ErrInvalidCredential {
		t.Errorf("Validate after Revoke: got %v, want ErrInvalidCredential", err)
	}
	// Revoking an unknown token is a no-op.
	if err := a.Revoke(context.Background(), "unknown"); err != nil {
		t.Errorf("Revoke unknown: got %v, want nil", err)
	}
}

func TestCredentialFromRequestPrecedence(t *testing.T) {
	// Cookie wins over header wins over query param.
	req := httptest.NewRequest(http.MethodGet, "/x?token=fromquery", nil)
	req.Header.Set("Authorization", "Bearer fromheader")
	req.AddCookie(&http.Cookie{Name: SessionCookie, Value: "fromcookie"})
	if got := CredentialFromRequest(req); got != "fromcookie" {
		t.Errorf("precedence: got %q, want %q", got, "fromcookie")
	}

	req = httptest.NewRequest(http.MethodGet, "/x?token=fromquery", nil)
	req.Header.Set("Authorization", "Bearer fromheader")
	if got := CredentialFromRequest(req); got != "fromheader" {
		t.Errorf("header precedence: got %q, want %q", got, "fromheader")
	}

	req = httptest.NewRequest(http.MethodGet, "/x?token=fromquery", nil)
	if got := CredentialFromRequest(req); got != "fromquery" {
		t.Errorf("query fallback: got %q, want %q", got, "fromquery")
	}

	req = httptest.NewRequest(http.MethodGet, "/x", nil)
	if got := CredentialFromRequest(req); got != "" {
		t.Errorf("empty: got %q, want empty", got)
	}
}

func TestMiddlewareRejectsMissingCredential(t *testing.T) {
	a := NewMemoryAuthenticator()
	h := Middleware(a)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run without credential")
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestMiddlewareAttachesPrincipal(t *testing.T) {
	a := NewMemoryAuthenticator()
	want := Principal{Role: RolePlayer, PlayerID: uuid.New(), Name: "Alice"}
	tok, _, err := a.Issue(context.Background(), want, time.Minute)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	var gotPrincipal Principal
	var gotOK bool
	h := Middleware(a)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPrincipal, gotOK = PrincipalFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookie, Value: tok})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	if !gotOK {
		t.Fatal("PrincipalFromContext returned !ok")
	}
	if gotPrincipal.PlayerID != want.PlayerID {
		t.Errorf("principal: got %+v, want %+v", gotPrincipal, want)
	}
}

func TestMiddlewareEnforcesRole(t *testing.T) {
	a := NewMemoryAuthenticator()
	tok, _, _ := a.Issue(context.Background(), Principal{Role: RolePlayer}, time.Minute)

	h := Middleware(a, RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run for wrong role")
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusForbidden)
	}
	var body map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["error"] == "" {
		t.Error("response body missing error message")
	}
}
