package store

import (
	"os"
	"testing"
)

// TestWriteDemoVault creates a vault with fake entries for screenshots.
// Only runs when TOKENFINDER_DEMO_VAULT points to the output path.
func TestWriteDemoVault(t *testing.T) {
	path := os.Getenv("TOKENFINDER_DEMO_VAULT")
	if path == "" {
		t.Skip("TOKENFINDER_DEMO_VAULT not set")
	}
	os.Remove(path)
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	demo := []Entry{
		{Name: "GITHUB_TOKEN", Kind: KindToken, Value: "ghp_4f2a9c1e7b3d8e6f0a1b2c3d4e5f6a7b8c9d", Note: "personal, repo scope"},
		{Name: "OPENAI_API_KEY", Kind: KindAPIKey, Value: "sk-proj-9Xk2mQ7vLp3nR8sT1uW4yZ6aB0cD5eF7gH", Note: "side project\nrotate monthly"},
		{Name: "AWS_SECRET_ACCESS_KEY", Kind: KindAPIKey, Value: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", Note: "prod account"},
		{Name: "CLOUDFLARE_API_TOKEN", Kind: KindToken, Value: "cfat_Q1w2E3r4T5y6U7i8O9p0A1s2D3f4G5h6J7k8", Note: "dns edit only"},
		{Name: "STRIPE_SECRET_KEY", Kind: KindAPIKey, Value: "sk_test_51Hq8f2KxYz9LmN3oP4qR5sT6uV7wX8yZ", Note: "test mode"},
		{Name: "postgres / app_user", Kind: KindPassword, Value: "correct-horse-battery-staple", Note: "staging db"},
		{Name: "SENDGRID_API_KEY", Kind: KindAPIKey, Value: "SG.k9Lm2Nq4Rt6Vw8Yz.aB1cD3eF5gH7iJ9kL2mN4oP6qR8sT0uV", Note: "transactional mail"},
		{Name: "VPS root", Kind: KindPassword, Value: "Tr0ub4dor&3-demo", Note: "hetzner"},
		{Name: "Telegram bot", Kind: KindToken, Value: "7123456789:AAF1demoTokenValue_abcDEFghiJKL", Note: "alerts bot"},
	}
	for _, e := range demo {
		if err := s.Upsert(e); err != nil {
			t.Fatal(err)
		}
	}
}
