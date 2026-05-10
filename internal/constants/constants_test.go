package constants

import (
	"fmt"
	"strings"
	"testing"
)

func TestAccountDetailsFallbackWithCode_FormatsCode(t *testing.T) {
	out := fmt.Sprintf(AccountDetailsFallbackWithCode, "ABC-123")
	if !strings.Contains(out, `registration_code="ABC-123"`) {
		t.Errorf("expected output to embed registration_code=\"ABC-123\", got:\n%s", out)
	}
	if !strings.Contains(out, "register(") {
		t.Errorf("expected output to mention register() tool, got:\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "never invent") {
		t.Errorf("expected anti-hallucination guardrail, got:\n%s", out)
	}
}

func TestAccountDetailsFallbackNoCode_PointsAtCredentials(t *testing.T) {
	if !strings.Contains(AccountDetailsFallbackNoCode, "credentials.json") {
		t.Errorf("expected message to point at credentials.json, got:\n%s", AccountDetailsFallbackNoCode)
	}
	if !strings.Contains(AccountDetailsFallbackNoCode, "registration_code") {
		t.Errorf("expected message to name the field, got:\n%s", AccountDetailsFallbackNoCode)
	}
	if !strings.Contains(strings.ToLower(AccountDetailsFallbackNoCode), "never invent") {
		t.Errorf("expected anti-hallucination guardrail, got:\n%s", AccountDetailsFallbackNoCode)
	}
}
