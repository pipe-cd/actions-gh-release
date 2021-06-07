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

import YAML from 'yaml'
import {Git} from './git'

export interface Config {
  tag: string
  title: string
  commitish: string
  body: string
  prerelease: boolean
}

export async function loadConfig(
  git: Git,
  commit: string,
  filePath: string
): Promise<Config> {
  const data = await git.readFile(commit, filePath)
  const cfg: Config = YAML.parse(data)
  return cfg
}
