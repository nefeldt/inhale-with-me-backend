package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("hunter2hunter", 4) // low cost keeps the test fast
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "hunter2hunter" {
		t.Fatal("hash must not equal the plaintext")
	}
	if !CheckPassword(hash, "hunter2hunter") {
		t.Fatal("correct password should verify")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("incorrect password must not verify")
	}
}
