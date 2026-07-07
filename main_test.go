package main

import (
	"bytes"
	"strings"
	"testing"
)

const fixtureBon26 = `{
  "project": {"id": 6, "name": "Bonelio 2026", "key": "BON26"},
  "agent": {
    "name": "ops",
    "description": "Infra, deploys, runtime.",
    "slash_command_name": "ops",
    "lane_tags": ["ops", "infra"],
    "metadata": {"deploy_recipes_used": ["backend-staging"]},
    "body": "## What ops owns\n\nDeployments and runtime probes.",
    "bootstrap_steps": [
      {"title": "Probe staging", "command": "curl -sf https://stg.example.com/healthz", "rationale": "confirm reachability"}
    ],
    "non_negotiable_rules": [
      {"title": "No prod writes without PR", "body": "Always go through CI.", "memory_ref": "feedback_no_silent_prod_writes"}
    ]
  },
  "repos": [
    {"label": "bonelio26-backend", "url": "https://github.com/example/bonelio26-backend", "default_branch": "main"}
  ],
  "environments": [
    {"name": "staging", "url": "https://stg.example.com", "host_alias": "ops-staging", "host_ip": "10.0.0.5"}
  ],
  "deploy_recipes": [
    {"name": "backend-staging", "command": "ssh ops-staging 'docker pull img && systemctl reload stack'", "summary": "Reload backend on staging"}
  ]
}`

func TestRenderClaudeCodeSkill(t *testing.T) {
	res, err := render([]byte(fixtureBon26))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"You are operating as the **ops session** for Bonelio 2026 (PMO project **BON26**)",
		"## Your lane",
		"Infra, deploys, runtime.",
		"## Bootstrap",
		"Probe staging",
		"## Non-negotiable rules",
		"feedback_no_silent_prod_writes",
		"## Deploy cheat sheet",
		"backend-staging",
		"## Free body",
		"What ops owns",
		"bonelio26-backend",
		"ops-staging",
	} {
		if !strings.Contains(res.Content, want) {
			t.Fatalf("rendered output missing %q\n--- output ---\n%s", want, res.Content)
		}
	}
	if strings.HasPrefix(res.Content, "<!-- paimos: rendered from") {
		t.Fatal("adapter must not emit the paimos managed header")
	}
	if res.SuggestedPath != ".claude/commands/ops.md" {
		t.Fatalf("suggested path=%q, want .claude/commands/ops.md", res.SuggestedPath)
	}
}

func TestSparseAgentSkipsEmptySections(t *testing.T) {
	res, err := render([]byte(`{
		"project": {"id": 7, "name": "Tiny", "key": "TINY"},
		"agent": {"name": "s"}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content, "TINY") {
		t.Fatalf("preamble missing project key:\n%s", res.Content)
	}
	for _, heading := range []string{"## Bootstrap", "## Non-negotiable rules", "## Deploy cheat sheet", "## Free body", "## Your lane"} {
		if strings.Contains(res.Content, heading) {
			t.Fatalf("sparse agent should skip %s:\n%s", heading, res.Content)
		}
	}
}

func TestRunDescribeValidateRender(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"describe"}, nil, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"name": "claude-code"`) {
		t.Fatalf("describe output missing adapter name: %s", out.String())
	}

	out.Reset()
	if err := run([]string{"validate", "--input", "-"}, strings.NewReader(fixtureBon26), &out); err != nil {
		t.Fatalf("validate: %v", err)
	}

	out.Reset()
	if err := run([]string{"render", "--input", "-"}, strings.NewReader(fixtureBon26), &out); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out.String(), "Bonelio 2026") {
		t.Fatalf("render output unexpected: %s", out.String())
	}
}

func TestInvalidCanonicalRejected(t *testing.T) {
	if _, err := render([]byte(`{"project":{"key":"X"},"agent":{}}`)); err == nil {
		t.Fatal("expected missing agent.name to fail")
	}
	if _, err := render([]byte(`not json`)); err == nil {
		t.Fatal("expected malformed JSON to fail")
	}
}
