package bot

import (
	"testing"
	"unicode/utf8"
)

func TestShortenNickname(t *testing.T) {
	tests := []struct{ name, suffix, want string }{
		{"Deeptanshu Siddalingappa Umesh Babu", " | ZAE", "Deeptanshu Babu | ZAE"},
		{"Pongpanlop Charoenpacharaporn", " | C1", "Pongpanlop C | C1"},
		{"Pongpanlop Charoenpacharaporn Extra Long", "", "Pongpanlop Long"},
		{"Supercalifragilisticexpialidocious Longname", " | ZLA C1", "Supercalifragilisticexp | ZLA C1"},
		{"Zoë Ünlüerdemirlioglu Sekerciogluüü", " | ZLA C1", "Zoë Sekerciogluüü | ZLA C1"},
	}
	for _, tt := range tests {
		got := shortenNickname(tt.name, tt.suffix)
		if got != tt.want {
			t.Errorf("shortenNickname(%q, %q) = %q, want %q", tt.name, tt.suffix, got, tt.want)
		}
		if n := utf8.RuneCountInString(got); n > maxNicknameLength {
			t.Errorf("%q is %d chars", got, n)
		}
	}
}
