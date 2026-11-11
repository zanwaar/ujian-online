package utils_test

import (
	"strings"
	"testing"

	"golang-backend/utils"
)

func TestSanitizeHTMLKeepsFormattingAndRemovesActiveContent(t *testing.T) {
	input := `<p>Halo <strong>dunia</strong></p><script>alert(1)</script><img src="x" onerror="alert(2)"><p onclick="alert(3)">Aman</p>`

	output := utils.SanitizeHTML(input)

	for _, fragment := range []string{"<p>Halo <strong>dunia</strong></p>", "<p>Aman</p>"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("SanitizeHTML() output %q does not contain %q", output, fragment)
		}
	}
	for _, fragment := range []string{"<script", "onerror", "onclick", "<img"} {
		if strings.Contains(strings.ToLower(output), fragment) {
			t.Fatalf("SanitizeHTML() output %q still contains %q", output, fragment)
		}
	}
}
