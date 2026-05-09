package index

import "testing"

func TestSafeFTSQueryQuotesSpecialCharacterTokens(t *testing.T) {
	got := SafeFTSQuery(`2001:db8::1 colon:and-json`)
	want := `"2001:db8::1" AND "colon:and-json"`
	if got != want {
		t.Fatalf("SafeFTSQuery() = %q, want %q", got, want)
	}
}
