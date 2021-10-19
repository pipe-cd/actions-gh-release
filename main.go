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
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

const (
	gitExecPath        = "git"
	defaultReleaseFile = "RELEASE"
)

func main() {
	log.Println("Start running actions-gh-release")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	args, err := parseArgs(os.Args)
	if err != nil {
		log.Fatalf("Failed to parse arguments: %v\n", err)
	}
	log.Println("Successfully parsed arguments")

	workspace := os.Getenv("GITHUB_WORKSPACE")
	if workspace == "" {
		log.Fatal("GITHUB_WORKSPACE was not defined")
	}

	var ghClient *github.Client
	if args.Token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: args.Token},
		)
		tc := oauth2.NewClient(ctx, ts)
		ghClient = github.NewClient(tc)
	} else {
		ghClient = github.NewClient(nil)
	}

	event, err := parseGitHubEvent(ctx, ghClient)
	if err != nil {
		log.Fatalf("Failed to parse GitHub event: %v\n", err)
	}
	log.Printf("Successfully parsed GitHub event %s\n\tbase-commit %s\n\thead-commit %s\n",
		event.Name,
		event.BaseCommit,
		event.HeadCommit,
	)

	// Find all changed RELEASE files.
	changedFiles, err := changedFiles(ctx, gitExecPath, workspace, event.BaseCommit, event.HeadCommit)
	if err != nil {
		log.Fatalf("Failed to list changed files: %v\n", err)
	}

	changedReleaseFiles := make([]string, 0, 0)
	matcher, err := NewPatternMatcher([]string{args.ReleaseFile})
	if err != nil {
		log.Fatalf("Failed to create pattern matcher for release file: %v\n", err)
	}
	for _, f := range changedFiles {
		if matcher.Matches(f) {
			changedReleaseFiles = append(changedReleaseFiles, f)
		}
	}

	if len(changedReleaseFiles) == 0 {
		log.Println("Nothing to do since there were no modified release files")
		return
	}

	proposals := make([]ReleaseProposal, 0, len(changedReleaseFiles))
	for _, f := range changedReleaseFiles {
		p, err := buildReleaseProposal(ctx, f, gitExecPath, workspace, event)
		if err != nil {
			log.Fatalf("Failed to build release for %s: %v\n", f, err)
		}
		proposals = append(proposals, *p)
	}

	releasesJSON, err := json.Marshal(proposals)
	if err != nil {
		log.Fatalf("Failed to marshal releases: %v\n", err)
	}
	fmt.Printf("::set-output name=releases::%s\n", string(releasesJSON))

	// Create GitHub releases if the event was push.
	if event.Name == eventPush {
		log.Printf("Will create %d GitHub releases\n", len(proposals))
		for _, p := range proposals {
			r, err := createRelease(ctx, ghClient, event.Owner, event.Repo, p)
			if err != nil {
				log.Fatalf("Failed to create release %s: %v\n", p.Tag, err)
			}
			log.Printf("Successfully created a new GitHub release %s\n%s\n", r.GetTagName(), r.GetHTMLURL())
		}

		log.Printf("Successfully created all %d GitHub releases\n", len(proposals))
		return
	}

	// Otherwise, just leave a comment to show the changelogs.
	body := makeCommentBody(proposals)
	comment, err := sendComment(ctx, ghClient, event.Owner, event.Repo, event.PRNumber, body)
	if err != nil {
		log.Fatalf("Failed to send comment: %v\n", err)
	}

	log.Printf("Successfully commented actions-gh-release result on pull request\n%s\n", *comment.HTMLURL)
}

type arguments struct {
	ReleaseFile string
	Token       string
}

func parseArgs(args []string) (arguments, error) {
	var out arguments

	for _, arg := range args {
		ps := strings.SplitN(arg, "=", 2)
		if len(ps) != 2 {
			continue
		}
		switch ps[0] {
		case "release-file":
			out.ReleaseFile = ps[1]
		case "token":
			out.Token = ps[1]
		}
	}

	if out.ReleaseFile == "" {
		out.ReleaseFile = defaultReleaseFile
	}
	if out.Token == "" {
		return out, fmt.Errorf("token argument must be set")
	}
	return out, nil
}
