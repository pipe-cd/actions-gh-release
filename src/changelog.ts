export function generateChangeLog(fromTag: string, toSHA: string): string {
  return `- Added foo ${fromTag}\n- Deleted xyz ${toSHA}`
}
