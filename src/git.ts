import * as exec from '@actions/exec'

export interface Commit {
  author: string
  committer: string
  hash: string
  abbrevHash: string
  message: string
  body: string
}

export interface GitFileReader {
  readFile(commit: string, path: string): Promise<string>
}

export class Git implements GitFileReader {
  gitPath: string
  workingDir: string

  constructor(gitPath: string, workingDir: string) {
    this.workingDir = workingDir
    this.gitPath = gitPath
  }

  async readFile(commit: string, path: string): Promise<string> {
    const out = await this.execute(['show', `${commit}:${path}`])
    return out.stdout.trim()
  }

  private async execute(
    args: string[],
    allowAllExitCodes = false,
    silent = false
  ): Promise<cmdOut> {
    const result = new cmdOut()
    const stdout: string[] = []
    const options = {
      cwd: this.workingDir,
      silent,
      ignoreReturnCode: allowAllExitCodes,
      listeners: {
        stdout: (data: Buffer) => {
          stdout.push(data.toString())
        },
      },
    }

    result.exitCode = await exec.exec(`"${this.gitPath}"`, args, options)
    result.stdout = stdout.join('')
    return result
  }
}

class cmdOut {
  stdout = ''
  exitCode = 0
}
