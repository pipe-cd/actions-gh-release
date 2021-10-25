// Copyright 2021 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/creasty/defaults"
	"sigs.k8s.io/yaml"
)

var (
	releaseNoteBlockRegex = regexp.MustCompile(`(?s)(?:Release note\*\*:\s*(?:<!--[^<>]*-->\s*)?` + "```(?:release-note)?|```release-note)(.+?)```")
)

type ReleaseConfig struct {
	Tag             string `json:"tag,omitempty"`
	Title           string `json:"title,omitempty"`
	TargetCommitish string `json:"targetCommitish,omitempty"`
	ReleaseNote     string `json:"releaseNote,omitempty"`
	Prerelease      bool   `json:"prerelease,omitempty"`

	CommitInclude ReleaseCommitMatcherConfig `json:"commitInclude,omitempty"`
	CommitExclude ReleaseCommitMatcherConfig `json:"commitExclude,omitempty"`

	CommitCategories     []ReleaseCommitCategoryConfig `json:"commitCategories,omitempty"`
	ReleaseNoteGenerator ReleaseNoteGeneratorConfig    `json:"releaseNoteGenerator,omitempty"`
}

type ReleaseCommitCategoryConfig struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	ReleaseCommitMatcherConfig
}

type ReleaseNoteGeneratorConfig struct {
	ShowAbbrevHash      bool                       `json:"showAbbrevHash,omitempty" default:"false"`
	ShowCommitter       bool                       `json:"showCommitter,omitempty" default:"true"`
	UseReleaseNoteBlock bool                       `json:"useReleaseNoteBlock,omitempty" default:"false"`
	CommitInclude       ReleaseCommitMatcherConfig `json:"commitInclude,omitempty"`
	CommitExclude       ReleaseCommitMatcherConfig `json:"commitExclude,omitempty"`
}

type ReleaseCommitMatcherConfig struct {
	Prefixes []string `json:"prefixes,omitemtpy"`
	Contains []string `json:"contains,omitempty"`
}

func (c ReleaseCommitMatcherConfig) Empty() bool {
	return len(c.Prefixes)+len(c.Contains) == 0
}

func (c ReleaseCommitMatcherConfig) Match(commit Commit) bool {
	for _, s := range c.Prefixes {
		if strings.HasPrefix(commit.Subject, s) {
			return true
		}
	}
	for _, s := range c.Contains {
		if strings.Contains(commit.Body, s) {
			return true
		}
	}
	return false
}

func (c *ReleaseConfig) Validate() error {
	if c.Tag == "" {
		return fmt.Errorf("tag must be specified")
	}
	return nil
}

func parseReleaseConfig(data []byte) (*ReleaseConfig, error) {
	js, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}

	c := &ReleaseConfig{}
	if err := json.Unmarshal(js, c); err != nil {
		return nil, err
	}

	if err := defaults.Set(c); err != nil {
		return nil, err
	}
	for i := range c.CommitCategories {
		if c.CommitCategories[i].Id == "" {
			c.CommitCategories[i].Id = fmt.Sprintf("_category_%d", i)
		}
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

type ReleaseProposal struct {
	Tag             string `json:"tag,omitempty"`
	Title           string `json:"title,omitempty"`
	TargetCommitish string `json:"targetCommitish,omitempty"`
	ReleaseNote     string `json:"releaseNote,omitempty"`
	Prerelease      bool   `json:"prerelease,omitempty"`

	Owner      string          `json:"owner,omitempty"`
	Repo       string          `json:"repo,omitempty"`
	PreTag     string          `json:"preTag,omitempty"`
	BaseCommit string          `json:"baseCommit,omitempty"`
	HeadCommit string          `json:"headCommit,omitempty"`
	Commits    []ReleaseCommit `json:"commits,omitempty"`
}

type ReleaseCommit struct {
	Commit
	ReleaseNote  string `json:"releaseNote,omitempty"`
	CategoryName string `json:"categoryName,omitempty"`
}

func buildReleaseProposal(ctx context.Context, releaseFile string, gitExecPath, repoDir string, event *githubEvent) (*ReleaseProposal, error) {
	configLoader := func(commit string) (*ReleaseConfig, error) {
		data, err := readFileAtCommit(ctx, gitExecPath, repoDir, releaseFile, commit)
		if err != nil {
			return nil, err
		}
		return parseReleaseConfig(data)
	}

	baseCfg, err := configLoader(event.BaseCommit)
	if err != nil {
		return nil, err
	}

	headCfg, err := configLoader(event.HeadCommit)
	if err != nil {
		return nil, err
	}

	// List all commits from the last release until now.
	revisions := fmt.Sprintf("%s...%s", baseCfg.Tag, event.HeadCommit)
	commits, err := listCommits(ctx, gitExecPath, repoDir, revisions)
	if err != nil {
		return nil, err
	}

	releaseCommits := buildReleaseCommits(commits, *headCfg)
	p := ReleaseProposal{
		Tag:             headCfg.Tag,
		Title:           headCfg.Title,
		TargetCommitish: headCfg.TargetCommitish,
		ReleaseNote:     headCfg.ReleaseNote,
		Prerelease:      headCfg.Prerelease,
		Owner:           event.Owner,
		Repo:            event.Repo,
		PreTag:          baseCfg.Tag,
		BaseCommit:      event.BaseCommit,
		HeadCommit:      event.HeadCommit,
		Commits:         releaseCommits,
	}

	if p.Title == "" {
		p.Title = fmt.Sprintf("Release %s", p.Tag)
	}
	if p.TargetCommitish == "" {
		p.TargetCommitish = event.HeadCommit
	}
	if p.ReleaseNote == "" {
		ln := renderReleaseNote(p, *headCfg)
		p.ReleaseNote = string(ln)
	}

	return &p, nil
}

func buildReleaseCommits(commits []Commit, cfg ReleaseConfig) []ReleaseCommit {
	out := make([]ReleaseCommit, 0, len(commits))

	for _, commit := range commits {
		// Exclude was specified and matched.
		if !cfg.CommitExclude.Empty() && cfg.CommitExclude.Match(commit) {
			continue
		}
		// Include was specified and not matched.
		if !cfg.CommitInclude.Empty() && !cfg.CommitInclude.Match(commit) {
			continue
		}

		c := ReleaseCommit{
			Commit:       commit,
			ReleaseNote:  determineCommitReleaseNote(commit, cfg.ReleaseNoteGenerator.UseReleaseNoteBlock),
			CategoryName: determineCommitCategory(commit, cfg.CommitCategories),
		}
		out = append(out, c)
	}

	return out
}

func determineCommitReleaseNote(c Commit, useReleaseNoteBlock bool) string {
	if !useReleaseNoteBlock {
		return c.Subject
	}

	subs := releaseNoteBlockRegex.FindStringSubmatch(c.Body)
	if len(subs) != 2 {
		return c.Subject
	}
	if rn := strings.TrimSpace(subs[1]); rn != "" {
		return rn
	}
	return c.Subject
}

func determineCommitCategory(commit Commit, categories []ReleaseCommitCategoryConfig) string {
	for _, c := range categories {
		if c.ReleaseCommitMatcherConfig.Empty() {
			return c.Id
		}
		if c.ReleaseCommitMatcherConfig.Match(commit) {
			return c.Id
		}
	}
	return ""
}

func renderReleaseNote(p ReleaseProposal, cfg ReleaseConfig) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Release %s with changes since %s\n\n", p.Tag, p.PreTag))

	renderCommit := func(c ReleaseCommit) {
		b.WriteString(fmt.Sprintf("* %s", c.ReleaseNote))
		if cfg.ReleaseNoteGenerator.ShowAbbrevHash {
			b.WriteString(fmt.Sprintf(" [%s](https://github.com/%s/%s/commit/%s)", c.AbbreviatedHash, p.Owner, p.Repo, c.Hash))
		}
		if cfg.ReleaseNoteGenerator.ShowCommitter {
			b.WriteString(fmt.Sprintf(" - by %s", c.Committer))
		}
		b.WriteString("\n")
	}

	filteredCommits := make([]ReleaseCommit, 0, len(p.Commits))
	for _, c := range p.Commits {
		// Exclude was specified and matched.
		if !cfg.ReleaseNoteGenerator.CommitExclude.Empty() && cfg.ReleaseNoteGenerator.CommitExclude.Match(c.Commit) {
			continue
		}
		// Include was specified and not matched.
		if !cfg.ReleaseNoteGenerator.CommitInclude.Empty() && !cfg.ReleaseNoteGenerator.CommitInclude.Match(c.Commit) {
			continue
		}
		filteredCommits = append(filteredCommits, c)
	}

	for _, ctg := range cfg.CommitCategories {
		commits := make([]ReleaseCommit, 0, 0)
		for _, c := range filteredCommits {
			if c.CategoryName == ctg.Id {
				commits = append(commits, c)
			}
		}
		if len(commits) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s\n\n", ctg.Title))
		for _, c := range commits {
			renderCommit(c)
		}
		b.WriteString("\n")
	}

	for _, c := range filteredCommits {
		if c.CategoryName == "" {
			renderCommit(c)
		}
	}

	return []byte(b.String())
}
