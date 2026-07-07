// PAIMOS Claude Code adapter.
// Copyright (C) 2026 Markus Barta <markus@barta.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, version 3.

package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	adapterName   = "claude-code"
	adapterBinary = "paimos-adapter-claude-code"
)

//go:embed paimos-adapter.json
var manifestJSON []byte

type renderResult struct {
	Content       string
	SuggestedPath string
}

type canonicalArtifact struct {
	Project struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Key  string `json:"key"`
	} `json:"project"`
	Agent struct {
		Name             string   `json:"name"`
		Description      string   `json:"description"`
		SlashCommandName string   `json:"slash_command_name"`
		LaneTags         []string `json:"lane_tags"`
		Body             string   `json:"body"`
		BootstrapSteps   []struct {
			Title     string `json:"title"`
			Command   string `json:"command"`
			Rationale string `json:"rationale"`
		} `json:"bootstrap_steps"`
		NonNegotiableRules []struct {
			Title     string `json:"title"`
			Body      string `json:"body"`
			MemoryRef string `json:"memory_ref"`
		} `json:"non_negotiable_rules"`
		Metadata map[string]any `json:"metadata"`
	} `json:"agent"`
	Repos []struct {
		Label         string `json:"label"`
		URL           string `json:"url"`
		DefaultBranch string `json:"default_branch"`
	} `json:"repos"`
	Environments []struct {
		Name      string `json:"name"`
		URL       string `json:"url"`
		HostAlias string `json:"host_alias"`
		HostIP    string `json:"host_ip"`
	} `json:"environments"`
	DeployRecipes []struct {
		Name    string `json:"name"`
		Command string `json:"command"`
		Summary string `json:"summary"`
	} `json:"deploy_recipes"`
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return usageError("missing verb")
	}
	switch args[0] {
	case "describe":
		_, err := stdout.Write(manifestJSON)
		return err
	case "validate":
		canonical, err := readInput(args[1:], stdin)
		if err != nil {
			return err
		}
		_, err = decodeCanonical(canonical)
		return err
	case "render":
		canonical, err := readInput(args[1:], stdin)
		if err != nil {
			return err
		}
		result, err := render(canonical)
		if err != nil {
			return err
		}
		_, err = io.WriteString(stdout, result.Content)
		return err
	default:
		return usageError(fmt.Sprintf("unknown verb %q", args[0]))
	}
}

func readInput(args []string, stdin io.Reader) ([]byte, error) {
	if len(args) == 0 {
		return io.ReadAll(stdin)
	}
	if len(args) != 2 || args[0] != "--input" {
		return nil, usageError("expected --input - or --input <path>")
	}
	if args[1] == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(args[1])
}

func render(canonical []byte) (renderResult, error) {
	art, err := decodeCanonical(canonical)
	if err != nil {
		return renderResult{}, err
	}
	return renderResult{
		Content:       renderBody(&art),
		SuggestedPath: suggestedPath(&art),
	}, nil
}

func decodeCanonical(canonical []byte) (canonicalArtifact, error) {
	var art canonicalArtifact
	if err := json.Unmarshal(canonical, &art); err != nil {
		return canonicalArtifact{}, fmt.Errorf("decode canonical artifact: %w", err)
	}
	if strings.TrimSpace(art.Agent.Name) == "" {
		return canonicalArtifact{}, errors.New("canonical artifact missing agent.name")
	}
	return art, nil
}

func suggestedPath(art *canonicalArtifact) string {
	slug := strings.TrimSpace(art.Agent.SlashCommandName)
	if slug == "" {
		slug = strings.TrimSpace(art.Agent.Name)
	}
	slug = strings.ReplaceAll(slug, string(filepath.Separator), "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	return filepath.Join(".claude", "commands", slug+".md")
}

func renderBody(art *canonicalArtifact) string {
	var b strings.Builder

	projectName := strOrFallback(art.Project.Name, art.Project.Key)
	projectKey := strOrFallback(art.Project.Key, fmt.Sprintf("id=%d", art.Project.ID))
	fmt.Fprintf(&b, "You are operating as the **%s session** for %s (PMO project **%s**).\n",
		art.Agent.Name, projectName, projectKey)

	if hasLaneContent(art) {
		b.WriteString("\n## Your lane\n\n")
		writeLane(&b, art)
	}
	if len(art.Agent.BootstrapSteps) > 0 {
		b.WriteString("\n## Bootstrap\n\n")
		writeBootstrap(&b, art)
	}
	if len(art.Agent.NonNegotiableRules) > 0 {
		b.WriteString("\n## Non-negotiable rules\n\n")
		writeRules(&b, art)
	}
	if len(art.DeployRecipes) > 0 {
		b.WriteString("\n## Deploy cheat sheet\n\n")
		writeDeployRecipes(&b, art)
	}
	if body := strings.TrimSpace(art.Agent.Body); body != "" {
		b.WriteString("\n## Free body\n\n")
		b.WriteString(body)
		b.WriteString("\n")
	}

	return b.String()
}

func hasLaneContent(art *canonicalArtifact) bool {
	return strings.TrimSpace(art.Agent.Description) != "" ||
		len(art.Repos) > 0 ||
		len(art.Environments) > 0 ||
		len(art.Agent.LaneTags) > 0
}

func writeLane(b *strings.Builder, art *canonicalArtifact) {
	if desc := strings.TrimSpace(art.Agent.Description); desc != "" {
		b.WriteString(desc)
		b.WriteString("\n")
	}
	if len(art.Agent.LaneTags) > 0 {
		fmt.Fprintf(b, "\n**Lane tags:** %s\n", strings.Join(art.Agent.LaneTags, ", "))
	}
	if len(art.Repos) > 0 {
		b.WriteString("\n**Repos:**\n")
		for _, r := range art.Repos {
			label := strOrFallback(r.Label, r.URL)
			if r.DefaultBranch != "" {
				fmt.Fprintf(b, "- %s - %s (`%s`)\n", label, r.URL, r.DefaultBranch)
			} else {
				fmt.Fprintf(b, "- %s - %s\n", label, r.URL)
			}
		}
	}
	if len(art.Environments) > 0 {
		b.WriteString("\n**Environments:**\n")
		for _, e := range art.Environments {
			fmt.Fprintf(b, "- **%s**", e.Name)
			if e.URL != "" {
				fmt.Fprintf(b, " - %s", e.URL)
			}
			if host := formatHost(e.HostAlias, e.HostIP); host != "" {
				fmt.Fprintf(b, " (host: %s)", host)
			}
			b.WriteString("\n")
		}
	}
}

func writeBootstrap(b *strings.Builder, art *canonicalArtifact) {
	for i, s := range art.Agent.BootstrapSteps {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			title = fmt.Sprintf("Step %d", i+1)
		}
		fmt.Fprintf(b, "%d. **%s**\n", i+1, title)
		if cmd := strings.TrimSpace(s.Command); cmd != "" {
			fmt.Fprintf(b, "   ```sh\n   %s\n   ```\n", cmd)
		}
		if rat := strings.TrimSpace(s.Rationale); rat != "" {
			fmt.Fprintf(b, "   _%s_\n", rat)
		}
	}
}

func writeRules(b *strings.Builder, art *canonicalArtifact) {
	for i, r := range art.Agent.NonNegotiableRules {
		title := strings.TrimSpace(r.Title)
		if title == "" {
			title = fmt.Sprintf("Rule %d", i+1)
		}
		fmt.Fprintf(b, "- **%s**", title)
		if ref := strings.TrimSpace(r.MemoryRef); ref != "" {
			fmt.Fprintf(b, " _(memory: `%s`)_", ref)
		}
		b.WriteString("\n")
		if body := strings.TrimSpace(r.Body); body != "" {
			for _, line := range strings.Split(body, "\n") {
				fmt.Fprintf(b, "  %s\n", line)
			}
		}
	}
}

func writeDeployRecipes(b *strings.Builder, art *canonicalArtifact) {
	for _, rec := range art.DeployRecipes {
		fmt.Fprintf(b, "### %s\n\n", rec.Name)
		if s := strings.TrimSpace(rec.Summary); s != "" {
			fmt.Fprintf(b, "%s\n\n", s)
		}
		if c := strings.TrimSpace(rec.Command); c != "" {
			fmt.Fprintf(b, "```sh\n%s\n```\n\n", c)
		}
	}
}

func formatHost(alias, ip string) string {
	switch {
	case alias != "" && ip != "":
		return alias + " (" + ip + ")"
	case alias != "":
		return alias
	case ip != "":
		return ip
	default:
		return ""
	}
}

func strOrFallback(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

type usageErr string

func usageError(s string) error { return usageErr(s) }

func (e usageErr) Error() string {
	return fmt.Sprintf("%s: %s [describe | validate --input - | render --input -]", adapterBinary, string(e))
}

func exitCode(err error) int {
	var u usageErr
	if errors.As(err, &u) {
		return 64
	}
	return 1
}
