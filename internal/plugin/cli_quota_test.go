package plugin

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

// This is a synthetic subprocess boundary, not the pending official-CLI helper.
// It returns the unchanged CodingPlanResult fixture and never reads an account.
func TestCLIQuotaFixtureProcess(t *testing.T) {
	for i, arg := range os.Args {
		if arg != "--cli-quota-fixture" || i+1 >= len(os.Args) {
			continue
		}
		var expected []string
		switch os.Args[i+1] {
		case "cli":
			expected = []string{
				"--source", "coding-plan", "--protocol", "qwen-coding-plan-v1",
				"--bl-path", "C:/synthetic CLI/bl & literal.cmd",
				"--cache-file", "C:/synthetic cache/coding quota.json", "--max-age=15m",
			}
		case "cache":
			expected = []string{
				"--source=coding-plan", "--protocol=qwen-coding-plan-v1", "--cache-only=true",
				"--cache-file=C:/synthetic cache/coding quota.json", "--max-age", "45m",
			}
		case "stale-cache":
			fmt.Print(`{"error":"stale_cache","message":"quota cache exceeds --max-age; refresh it before reading"}`)
			os.Exit(1)
		case "must-not-execute":
			fmt.Print(`{"error":"synthetic_helper_executed","message":"invalid product or mixed transport reached the subprocess"}`)
			os.Exit(2)
		}
		if expected == nil || !slices.Equal(os.Args[i+2:], expected) {
			fmt.Print(`{"error":"synthetic_argument_failure","message":"arguments did not reach the subprocess unchanged"}`)
			os.Exit(2)
		}
		fmt.Print(codingFixture)
		os.Exit(0)
	}
}

func TestCLIQuotaCacheOnlyOptionsStayUnbound(t *testing.T) {
	m := managementCommandManager(t, "coding")
	m.cfg.CommandArgs = []string{
		"-test.run=^TestCLIQuotaFixtureProcess$", "--", "--cli-quota-fixture", "cache",
		"--source=coding-plan", "--protocol=qwen-coding-plan-v1", "--cache-only=true",
		"--cache-file=C:/synthetic cache/coding quota.json", "--max-age", "45m",
	}
	resp, cards := managementUsage(t, m, `{}`)
	if resp.StatusCode != 200 || len(cards) != 1 || cards[0].QueryStatus != "success" || cards[0].AuthorityStatus != "needs_configuration" || cards[0].Error == nil || !strings.Contains(*cards[0].Error, "account_authority_unverified") || len(cards[0].Windows) != 0 || len(cards[0].Metrics) != 0 {
		t.Fatalf("explicit cached query was rejected or became bound credential quota: %s", resp.Body)
	}
}

func TestCLIQuotaOptionsRejectWrongProductAndMixedTransportBeforeExecution(t *testing.T) {
	for name, args := range map[string][]string{
		"wrong-product":         {"--source=bl", "--protocol=qwen-coding-plan-v1", "--bl-path=synthetic"},
		"wrong-protocol":        {"--source=coding-plan", "--protocol=legacy", "--cache-only", "--cache-file=synthetic"},
		"mixed-token":           {"--source=coding-plan", "--protocol=qwen-coding-plan-v1", "--bl-path=synthetic", "--token-file=synthetic"},
		"mixed-site":            {"--source=coding-plan", "--protocol=qwen-coding-plan-v1", "--bl-path=synthetic", "--console-site=domestic"},
		"empty-cli-path":        {"--source=coding-plan", "--protocol=qwen-coding-plan-v1", "--bl-path="},
		"nonpositive-cache-age": {"--source=coding-plan", "--protocol=qwen-coding-plan-v1", "--cache-only", "--cache-file=synthetic", "--max-age=0s"},
	} {
		t.Run(name, func(t *testing.T) {
			m := managementCommandManager(t, "coding")
			m.cfg.CommandArgs = append([]string{"-test.run=^TestCLIQuotaFixtureProcess$", "--", "--cli-quota-fixture", "must-not-execute"}, args...)
			resp, cards := managementUsage(t, m, `{}`)
			if resp.StatusCode != 200 || len(cards) != 1 || cards[0].QueryStatus != "failed" || cards[0].Error == nil || !strings.Contains(*cards[0].Error, "coding_plan_readonly_configuration_required") || len(cards[0].Windows) != 0 {
				t.Fatalf("invalid configuration reached the helper or hid its failure: %s", resp.Body)
			}
		})
	}
}

func TestCLIQuotaStaleCacheFailureIsNotQuerySuccess(t *testing.T) {
	m := managementCommandManager(t, "coding")
	m.cfg.CommandArgs = []string{
		"-test.run=^TestCLIQuotaFixtureProcess$", "--", "--cli-quota-fixture", "stale-cache",
		"--source=coding-plan", "--protocol=qwen-coding-plan-v1",
		"--cache-only", "--cache-file=synthetic", "--max-age=15m",
	}
	resp, cards := managementUsage(t, m, `{}`)
	if resp.StatusCode != 200 || len(cards) != 1 || cards[0].QueryStatus != "failed" || cards[0].Error == nil || *cards[0].Error != "stale_cache: quota cache exceeds --max-age; refresh it before reading" || len(cards[0].Windows) != 0 || len(cards[0].Metrics) != 0 {
		t.Fatalf("stale cache was accepted or its own helper error was lost: %s", resp.Body)
	}
}

func TestCLIQuotaHelperDiagnosticRedactionStillApplies(t *testing.T) {
	m := managementCommandManager(t, "secret")
	m.cfg.CommandArgs = append(m.cfg.CommandArgs, "--bl-path=synthetic", "--cache-file=synthetic", "--max-age=15m")
	resp, cards := managementUsage(t, m, `{}`)
	if resp.StatusCode != 200 || len(cards) != 1 || cards[0].QueryStatus != "failed" || cards[0].Error == nil || !strings.Contains(*cards[0].Error, "login") {
		t.Fatalf("helper failure was hidden: %s", resp.Body)
	}
	for _, secret := range []string{"dummy-config-key", "private-cookie", "private-token", "sk-sp-private-key"} {
		if strings.Contains(string(resp.Body), secret) {
			t.Fatalf("helper secret leaked: %s", resp.Body)
		}
	}
}

func TestCLIQuotaExplicitOptionsReachManagementWithoutKeyAuthority(t *testing.T) {
	m := managementCommandManager(t, "coding")
	m.cfg.CommandArgs = []string{
		"-test.run=^TestCLIQuotaFixtureProcess$", "--", "--cli-quota-fixture", "cli",
		"--source", "coding-plan", "--protocol", "qwen-coding-plan-v1",
		"--bl-path", "C:/synthetic CLI/bl & literal.cmd",
		"--cache-file", "C:/synthetic cache/coding quota.json", "--max-age=15m",
	}
	resp, cards := managementUsage(t, m, `{"auth_index":"host-index-0"}`)
	if resp.StatusCode != 200 || len(cards) != 1 || cards[0].QueryStatus != "success" {
		t.Fatalf("explicit Coding Plan CLI options did not reach a successful query: %s", resp.Body)
	}
	card := cards[0]
	if card.AuthIndex != "host-index-0" || card.AuthorityStatus != "needs_configuration" || card.Error == nil || !strings.Contains(*card.Error, "account_authority_unverified") || len(card.Windows) != 0 || len(card.Metrics) != 0 || card.Plan != "" || card.ObservedAt != "" {
		t.Fatalf("CLI-account query was falsely attributed to a credential: %s", resp.Body)
	}
	result, err := m.FetchQuota(context.Background(), quotaRequest())
	if err == nil || !strings.Contains(err.Error(), "account_authority_unverified") || result.Subscription != nil || len(result.Groups) != 0 || len(result.Summary) != 0 {
		t.Fatalf("unbound CLI-account query became policy quota: %#v %v", result, err)
	}
}
