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
