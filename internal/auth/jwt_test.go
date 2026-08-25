package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParse(t *testing.T) {
	m := NewManager("secret", time.Hour)
	tok, exp, err := m.Generate("user-1")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !exp.After(time.Now()) {
		t.Fatal("expiry should be in the future")
	}
	uid, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if uid != "user-1" {
		t.Fatalf("subject = %q, want user-1", uid)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	tok, _, _ := NewManager("secret", time.Hour).Generate("u")
	if _, err := NewManager("other-secret", time.Hour).Parse(tok); err == nil {
		t.Fatal("expected error for token signed with a different secret")
	}
}

func TestParseRejectsExpired(t *testing.T) {
	tok, _, _ := NewManager("secret", -time.Hour).Generate("u")
	if _, err := NewManager("secret", time.Hour).Parse(tok); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if _, err := NewManager("secret", time.Hour).Parse("not-a-jwt"); err == nil {
		t.Fatal("expected error for malformed token")
	}
}
