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
import {Commit} from './git'

const mergeCommitPrefix = 'Merge pull request #'

export interface Options {
  maxCommitsNumber: number
}

export interface RenderOptions {
  showAbbrevHash: boolean
  showCommitter: boolean
  ignoreMergeCommit: boolean
  onlyUseMergeCommit: boolean
}

export function getCommits(
  repoDir: string,
  fromTag: string,
  toSHA: string,
  options: Options
): Commit[] {

  const commits = gitlog({
    repo: repoDir,
    branch: `${toSHA}...${fromTag}`,
    number: options.maxCommitsNumber,
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

  return commits
    .map(c => {
      return {
        author: c.authorName,
        committer: c.committerName,
        hash: c.hash,
        abbrevHash: c.abbrevHash,
        subject: c.subject,
        body: c.body,
      }
    })
}

export function renderChangeLog(
  commits: Commit[],
  options: RenderOptions
): string {
  const logs = commits
    .filter(c => {
      if (options.ignoreMergeCommit) {
        return !c.subject.startsWith(mergeCommitPrefix)
      }
      if (options.onlyUseMergeCommit) {
        return c.subject.startsWith(mergeCommitPrefix)
      }
      return true
    }).map(c => {
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
        fields.push(`- by ${c.committer}`)
      }

      return fields.join(' ')
    })

  return logs.join(`\n`)
}

export function renderChangeJSON(
  fromTag: string,
  toTag: string,
  commits: Commit[],
  options: RenderOptions
): string {
  const changes = {
    fromTag: fromTag,
    toTag: toTag,
    commits: commits.filter(c => {
      if (options.ignoreMergeCommit) {
        return !c.subject.startsWith(mergeCommitPrefix)
      }
      if (options.onlyUseMergeCommit) {
        return c.subject.startsWith(mergeCommitPrefix)
      }
      return true
    }),
  }

  return JSON.stringify(changes, null, 4);
}
