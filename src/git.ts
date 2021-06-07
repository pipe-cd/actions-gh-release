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
