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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-github/v39/github"
)

type githubEvent struct {
	Name string

	Owner string
	Repo  string

	HeadCommit string
	BaseCommit string

	PRNumber   int
	IsComment  bool
	CommentURL string
}

const (
	eventPush         = "push"
	eventPullRequest  = "pull_request"
	eventIssueComment = "issue_comment"
)

// parsePullRequestEvent uses the given environment variables
// to parse and build githubEvent struct.
func parseGitHubEvent(ctx context.Context, client *github.Client) (*githubEvent, error) {
	var parseEvents = map[string]struct{}{
		eventPush:         {},
		eventPullRequest:  {},
		eventIssueComment: {},
	}

	eventName := os.Getenv("GITHUB_EVENT_NAME")
	if _, ok := parseEvents[eventName]; !ok {
		return &githubEvent{
			Name: eventName,
		}, nil
	}

	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	payload, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read event payload: %v", err)
	}

	event, err := github.ParseWebHook(eventName, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event payload: %v", err)
	}

	switch e := event.(type) {
	case *github.PushEvent:
		return &githubEvent{
			Name:       eventPush,
			Owner:      e.Repo.Owner.GetLogin(),
			Repo:       e.Repo.GetName(),
			HeadCommit: e.GetAfter(),
			BaseCommit: e.GetBefore(),
		}, nil

	case *github.PullRequestEvent:
		return &githubEvent{
			Name:       eventPullRequest,
			Owner:      e.Repo.Owner.GetLogin(),
			Repo:       e.Repo.GetName(),
			HeadCommit: e.PullRequest.Head.GetSHA(),
			BaseCommit: e.PullRequest.Base.GetSHA(),
			PRNumber:   e.GetNumber(),
		}, nil

	case *github.IssueCommentEvent:
		var (
			owner = e.Repo.Owner.GetLogin()
			repo  = e.Repo.GetName()
			prNum = e.Issue.GetNumber()
		)

		pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNum)
		if err != nil {
			return nil, err
		}

		return &githubEvent{
			Name:       eventIssueComment,
			Owner:      owner,
			Repo:       repo,
			HeadCommit: pr.Head.GetSHA(),
			BaseCommit: pr.Base.GetSHA(),
			PRNumber:   prNum,
			IsComment:  true,
			CommentURL: e.Comment.GetHTMLURL(),
		}, nil

	default:
		return nil, fmt.Errorf("got an unexpected event type, got: %t", e)
	}
}

func sendComment(ctx context.Context, client *github.Client, owner, repo string, prNum int, body string) (*github.IssueComment, error) {
	c, _, err := client.Issues.CreateComment(ctx, owner, repo, prNum, &github.IssueComment{
		Body: &body,
	})
	return c, err
}

func createRelease(ctx context.Context, client *github.Client, owner, repo string, p ReleaseProposal) (*github.RepositoryRelease, error) {
	release, _, err := client.Repositories.CreateRelease(ctx, owner, repo, &github.RepositoryRelease{
		TagName:         &p.Tag,
		Name:            &p.Title,
		TargetCommitish: &p.TargetCommitish,
		Body:            &p.ReleaseNote,
		Prerelease:      &p.Prerelease,
	})
	return release, err
}
