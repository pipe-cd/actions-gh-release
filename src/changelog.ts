import gitlog, {GitlogOptions} from 'gitlog'

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
