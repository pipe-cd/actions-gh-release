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
	"fmt"
	"strings"
)

const (
	successBadgeURL = `<!-- RELEASE -->
[![RELEASE](https://img.shields.io/static/v1?label=GitHub&message=RELEASE&color=success&style=flat)](https://github.com/pipe-cd/actions-gh-release)

`
	noReleaseTitleFormat = "This pull request does not touch any RELEASE files. It means no GitHub releases will be created once this pull request got merged.\n"
	titleFormat          = "The following %d GitHub releases will be created once this pull request got merged.\n"
)

func makeCommentBody(proposals []ReleaseProposal) string {
	var b strings.Builder
	b.WriteString(successBadgeURL)

	if len(proposals) == 0 {
		fmt.Fprintf(&b, noReleaseTitleFormat)
		return b.String()
	}

	b.WriteString(fmt.Sprintf(titleFormat, len(proposals)))
	for _, p := range proposals {
		fmt.Fprintf(&b, "\n")
		fmt.Fprintf(&b, p.ReleaseNote)
		fmt.Fprintf(&b, "\n")
	}

	return b.String()
}
