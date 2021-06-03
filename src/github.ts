import * as core from '@actions/core'
import * as github from '@actions/github'
import {GitHub} from '@actions/github/lib/utils'

export interface ReleaseInput {
  tag_name: string
  name: string
  target_commitish: string
  body: string
  draft: boolean
  prerelease: boolean
}

export interface Release {
  id: number
  upload_url: string
  html_url: string
  tag_name: string
  body: string | null | undefined
  target_commitish: string
}

export class Releaser {
  gh: InstanceType<typeof GitHub>
  owner: string
  repo: string

  constructor(gh: InstanceType<typeof GitHub>, owner: string, repo: string) {
    this.gh = gh
    this.owner = owner
    this.repo = repo
  }

  async getReleaseByTag(tag: string): Promise<Release> {
    const r = await this.gh.rest.repos.getReleaseByTag({
      owner: this.owner,
      repo: this.repo,
      tag: tag,
    })

    return {
      id: r.data.id,
      upload_url: r.data.upload_url,
      html_url: r.data.html_url,
      tag_name: r.data.tag_name,
      body: r.data.body,
      target_commitish: r.data.target_commitish,
    }
  }

  async createRelease(input: ReleaseInput): Promise<Release> {
    const r = await this.gh.rest.repos.createRelease({
      owner: this.owner,
      repo: this.repo,
      tag_name: input.tag_name,
      name: input.name,
      body: input.body,
      draft: input.draft,
      prerelease: input.prerelease,
      target_commitish: input.target_commitish,
    })

    return {
      id: r.data.id,
      upload_url: r.data.upload_url,
      html_url: r.data.html_url,
      tag_name: r.data.tag_name,
      body: r.data.body,
      target_commitish: r.data.target_commitish,
    }
  }

  async updateRelease(id: number, input: ReleaseInput): Promise<Release> {
    const r = await this.gh.rest.repos.updateRelease({
      owner: this.owner,
      repo: this.repo,
      release_id: id,
      tag_name: input.tag_name,
      name: input.name,
      body: input.body,
      draft: input.draft,
      prerelease: input.prerelease,
      target_commitish: input.target_commitish,
    })

    return {
      id: r.data.id,
      upload_url: r.data.upload_url,
      html_url: r.data.html_url,
      tag_name: r.data.tag_name,
      body: r.data.body,
      target_commitish: r.data.target_commitish,
    }
  }
}

export const release = async (
  releaser: Releaser,
  input: ReleaseInput
): Promise<Release> => {
  const tag_name = input.tag_name

  try {
    const cur = await releaser.getReleaseByTag(tag_name)
    const release = await releaser.updateRelease(cur.id, input)
    return release
  } catch (error) {
    if (error.status !== 404) {
      core.error(
        `Unexpected error while fetching GitHub release for tag ${tag_name}: ${error}`
      )
      throw error
    }

    core.info(`Creating new GitHub release for tag ${tag_name}...`)
    try {
      const release = await releaser.createRelease(input)
      return release
    } catch (error) {
      core.error(
        `Failed to create GitHub release for tag ${tag_name}: ${error}`
      )
      throw error
    }
  }
}

export interface Comment {
  id: number
  html_url: string
}

export class Commenter {
  gh: InstanceType<typeof GitHub>
  owner: string
  repo: string

  constructor(gh: InstanceType<typeof GitHub>, owner: string, repo: string) {
    this.gh = gh
    this.owner = owner
    this.repo = repo
  }

  async comment(issue_number: number, body: string): Promise<Comment> {
    const c = await this.gh.rest.issues.createComment({
      owner: this.owner,
      repo: this.repo,
      issue_number: issue_number,
      body: body,
    })
    return {
      id: c.data.id,
      html_url: c.data.html_url,
    }
  }
}

export function getHeadCommitSHA(): string {
  if (github.context.eventName === 'push') {
    return github.context.payload['after']
  }

  const pr = github.context.payload.pull_request
  if (pr === undefined) {
    throw Error('Missing pull request data in webhook event')
  }
  const head: {sha: string} = pr['head']
  return head.sha
}

export function getBaseCommitSHA(): string {
  if (github.context.eventName === 'push') {
    return github.context.payload['before']
  }

  const pr = github.context.payload.pull_request
  if (pr === undefined) {
    throw Error('Missing pull request data in webhook event')
  }
  const base: {sha: string} = pr['base']
  return base.sha
}
