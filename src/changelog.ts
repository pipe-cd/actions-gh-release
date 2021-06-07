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

import gitlog from 'gitlog'

const mergeCommitPrefix = 'Merge pull request #'

export interface Options {
  showAbbrevHash: boolean
  showCommitter: boolean
  onlyUseMergeCommit: boolean
  ignoreMergeCommit: boolean
}

export function generateChangeLog(
  repoDir: string,
  fromTag: string,
  toSHA: string,
  options: Options
): string {
  const commits = gitlog({
    repo: repoDir,
    branch: `${toSHA}...${fromTag}`,
    fields: [
      'subject',
      'body',
      'hash',
      'abbrevHash',
      'authorName',
      'committerName',
      'committerDate',
    ] as const,
  })
  const logs = commits
    .filter(c => {
      if (options.ignoreMergeCommit) {
        return !c.subject.startsWith(mergeCommitPrefix)
      }
      if (options.onlyUseMergeCommit) {
        return c.subject.startsWith(mergeCommitPrefix)
      }
      return true
    })
    .map(c => {
      let fields: string[] = ['*']
      if (options.showAbbrevHash) {
        fields.push(c.abbrevHash)
      }
      if (!options.onlyUseMergeCommit) {
        fields.push(c.subject)
      } else {
        const message = c.body.split('\n', 1)[0]
        const subject = c.subject.replace(mergeCommitPrefix, '')
        const pr = subject.split(' ', 1)[0]
        fields.push(`${message} #${pr}`)
      }
      if (options.showCommitter) {
        fields.push(`- by @${c.committerName}`)
      }
      return fields.join(' ')
    })

  return logs.join(`\n`)
}
