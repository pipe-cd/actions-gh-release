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

import * as core from '@actions/core'
import * as io from '@actions/io'
import * as github from '@actions/github'
import * as path from 'path'
import {loadConfig} from './config'
import {Git} from './git'
import {getCommits, renderChangeLog, renderChangeJSON} from './changelog'
import {
  Commenter,
  release,
  Releaser,
  getBaseCommitSHA,
  getHeadCommitSHA,
} from './github'

async function run(): Promise<void> {
  try {
    const event = github.context.eventName
    if (event !== 'push' && event !== 'pull_request') {
      throw Error(`This action does not support ${event} event`)
    }

    let workingDir = process.env['GITHUB_WORKSPACE']
    if (!workingDir) {
      throw new Error('GITHUB_WORKSPACE was not defined')
    }
    workingDir = path.resolve(workingDir)

    core.info(`Start handling for ${event} event`)
    const releaseFile: string = core.getInput('release_file')

    const gitPath = await io.which('git', true)
    const git = new Git(gitPath, workingDir)

    // Load release configuration data at the head commit.
    const headSHA = getHeadCommitSHA()
    const headCfg = await loadConfig(git, headSHA, releaseFile)
    core.info(
      `Loaded release config from ${releaseFile} at the HEAD commit (${headSHA})`
    )

    // Load release configuration data at the base commit.
    const baseSHA = getBaseCommitSHA()
    const baseCfg = await loadConfig(git, baseSHA, releaseFile)
    core.info(
      `Loaded release config from ${releaseFile} file at the BASE commit (${baseSHA})`
    )

    const token = core.getInput('token')
    const owner = github.context.repo.owner
    const repo = github.context.repo.repo

    const onlyUseMergeCommit = core.getInput('changelog_only_use_merge_commit').toLowerCase() === 'true'
    const ignoreMergeCommit = core.getInput('changelog_ignore_merge_commit').toLowerCase() === 'true'
    const maxCommitsNumber = Number(core.getInput('changelog_max_commits_number')) || 100

    let commits = getCommits(workingDir, baseCfg.tag, headSHA, {
      onlyUseMergeCommit: onlyUseMergeCommit,
      ignoreMergeCommit: ignoreMergeCommit,
      maxCommitsNumber: maxCommitsNumber,
    })
    let changeJSON = renderChangeJSON(baseCfg.tag, headCfg.tag, commits)
    core.info(
      `Successfully generated change list \n${changeJSON}`
    )

    // Determine the release body.
    let body = headCfg.body
    if (!body) {
      body = core.getInput('body')
    }
    if (!body) {
      const showAbbrevHash = core.getInput('changelog_show_abbrev_hash').toLowerCase() === 'true'
      const showCommitter = core.getInput('changelog_show_committer').toLowerCase() === 'true'

      body = renderChangeLog(commits, {
        showAbbrevHash: showAbbrevHash,
        showCommitter: showCommitter,
        onlyUseMergeCommit: onlyUseMergeCommit,
      })
      core.info(
        `Successfully generated changelog \n${body}`
      )
    }

    const octokit = github.getOctokit(token)

    // Make a new release or update the existing one.
    if (event === 'push') {
      const releaser = new Releaser(octokit, owner, repo)
      const title = headCfg.title ?? `Release ${headCfg.tag}`
      const r = await release(releaser, {
        tagName: headCfg.tag,
        name: title,
        target_commitish: headSHA,
        body: body,
        draft: false,
        prerelease: headCfg.prerelease,
      })

      core.setOutput('id', r.id)
      core.setOutput('tag', r.tagName)
      core.setOutput('html_url', r.html_url)
      core.setOutput('upload_url', r.upload_url)
      core.setOutput('changelog', body)
      core.setOutput('change_json', changeJSON)
      core.info(`Successfully released ${headCfg.tag}. See ${r.html_url}`)
      return
    }

    // Leave a comment to show changelog on the pull request.
    const pull_number = github.context.payload.pull_request!.number
    const commenter = new Commenter(octokit, owner, repo)
    const message = `A GitHub release with \`${headCfg.tag}\` tag will be created once this pull request got merged.\n\n## Changelog since ${baseCfg.tag}\n${body}`
    const c = await commenter.comment(pull_number, message)

    core.setOutput('changelog', body)
    core.setOutput('change_json', changeJSON)
    core.info(
      `Successfully commented the changelog to pull request ${pull_number}`
    )
    return
  } catch (error) {
    core.setFailed(error.message)
  }
}

run()
