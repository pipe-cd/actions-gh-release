# actions-gh-release

An action that enables operating GitHub release via pull request. You send a pull request to update a RELEASE file then release note will be generated and commented in that pull request for reviewing. Once merged, a new GitHub release will be created with that release note.

![](https://github.com/pipe-cd/actions-gh-release/blob/main/assets/changelog-comment.png)

**NOTE**: The source code of this GitHub Action is placing under the tool directory of of [pipe-cd/pipecd](https://github.com/pipe-cd/pipecd/tree/master/tool) repository. If you want to make a pull request or raise an issue, please send it to [pipe-cd/pipecd](https://github.com/pipe-cd/pipecd) repository.

## Usage

- Adding a RELEASE file to the repository. You can also have multiple RELEASE files in case of monorepo style. Its content looks like this:

``` yaml
tag: v0.1.0                        # The tag number will be created. Required.

# # Optional fields:
#
# name: string                     # The release name. Default is empty.
# title: string                    # The release title. Default is "Release ${tag}".
# targetCommitish: string          # The release commitish. Default is the merged commit.
# releaseNote: string              # The release body. Default is the auto-generated release note.
# prerelease: bool                 # True if this is a prerelease. Default is false.
#
#
# # If specified, all matching commits will be excluded from release. Empty means excluding nothing.
#
# commitExclude:
#   parentOfMergeCommit: bool      # True is whether the commit is the parent commit of the matching merge commit. Default is false.
#   prefixes: []string             # Matches if commit's subject is prefixed by one of the given values. Default is emtpy.
#   contains: []string             # Matches if commit's body is containing one of the given values. Default is emtpy.
#
#
# # If specified, all matching commits will be included to release. Empty means including alls.
#
# commitInclude:
#   parentOfMergeCommit: bool      # True is whether the commit is the parent commit of the matching merge commit. Default is false.
#   prefixes: []string             # Matches if commit's subject is prefixed by one of the given values. Default is emtpy.
#   contains: []string             # Matches if commit's body is containing one of the given values. Default is emtpy.
#
#
# # List of categories and how to decide which category a commit should belong to.
#
# commitCategories:
#   - title: string                # Category title.
#     parentOfMergeCommit: bool    # True is whether the commit is the parent commit of the matching merge commit. Default is false.
#     contains: []string           # Matches if commit's subject is prefixed by one of the given values. Default is emtpy.
#     prefixes: []string           # Matches if commit's body is containing one of the given values. Default is emtpy.
#
#
# # Config used while generating release note.
#
# releaseNoteGenerator:
#   showAbbrevHash: bool           # Whether to include abbreviated hash value in release note. Default is false.
#   showCommitter: bool            # Whether to include committer in release note. Default is true.
#   useReleaseNoteBlock: bool      # Whether to use release note block instead of commit message. Default is false.
#   usePullRequestMetadata: bool   # Whether to use pull request metadata instead of commit message when using merge-commit. If useReleaseNoteBlock is also true, release note block of pull request is used. Otherwise pull request title is used. If this option is set, showAbbrevHash and showCommitter is ignored. Default is false.
#   usePullRequestLink: bool       # Whether to use the pull request links in the release note. Default is false.
#   commitExclude:                 # Additional excludes applied while generating release note.
#     parentOfMergeCommit: bool    # True is whether the commit is the parent commit of the matching merge commit. Default is false.
#     prefixes: []string           # Matches if commit's subject is prefixed by one of the given values. Default is emtpy.
#     contains: []string           # Matches if commit's body is containing one of the given values. Default is emtpy.
#   commitInclude:                 # Additional includes applied while generating release note.
#     parentOfMergeCommit: bool    # True is whether the commit is the parent commit of the matching merge commit. Default is false.
#     prefixes: []string           # Matches if commit's subject is prefixed by one of the given values. Default is emtpy.
#     contains: []string           # Matches if commit's body is containing one of the given values. Default is emtpy.
```

- Adding a new workflow (eg: `.github/workflows/gh-release.yaml`) with the content as below:

```yaml
on:
  push:
    branches:    
      - main
    paths:
      - '**/RELEASE'
  pull_request:
    types: [opened, synchronize]
    branches:
      - main
    paths:
      - '**/RELEASE'

jobs:
  gh-release:
    runs-on: ubuntu-latest
    permissions:
      contents: write      # Required to create a GitHub release and tag, as GITHUB_TOKEN is read-only by default.
      pull-requests: write # Required to comment the release note on the pull request.
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      - uses: pipe-cd/actions-gh-release@v2.3.4
        with:
          release_file: '**/RELEASE'
          token: ${{ secrets.GITHUB_TOKEN }}
```

## Inputs

| Name                  | Description                                                                       | Required | Default Value |
|-----------------------|-----------------------------------------------------------------------------------|:--------:|:-------------:|
| token                 | The GITHUB_TOKEN secret.                                                          |    yes   |               |
| release_file          | The path to the RELEASE file or pattern to match one or multiple RELEASE files.   |    no    |    RELEASE    |
| output_releases_file  | The path to output the list of releases formatted in JSON.                        |    no    |               |

## Outputs

| Name            | Description                                          |
|-----------------|------------------------------------------------------|
| releases        | The list of releases formatted in JSON.              |

## RELEASE examples

- Excluding merge pull request from release note

``` yaml
tag: v1.1.0

commitExclude:
  prefixes:
    - "Merge pull request #"
```

- Grouping commits by category in release note

``` yaml
tag: v1.1.0

commitCategories:
  - title: "Breaking Changes"
    contains:
      - change-category/breaking-change
  - title: "New Features"
    contains:
      - change-category/new-feature
  - title: "Notable Changes"
    contains:
      - change-category/notable-change
  - title: "Internal Changes"

releaseNoteGenerator:
  showCommitter: true
  useReleaseNoteBlock: true
```

- Multiple RELEASE files for mono-repo style

``` yaml
tag: foo-v0.1.0
name: foo

commitInclude:
  contains:
    - "application/foo"

releaseNoteGenerator:
  showAbbrevHash: true
  showCommitter: true
  useReleaseNoteBlock: true
```

- Multiple RELEASE files for mono-repo style (Include parent commits of the matching merge commit in `outoputs.release`)


``` yaml
tag: bar-v1.0.0
name: bar

commitInclude:
  parentOfMergeCommit: true
  prefixes:
    - "bar:"

releaseNoteGenerator:
  showAbbrevHash: true
  showCommitter: true
  useReleaseNoteBlock: true
  commitInclude:
    prefixes:
      - "bar:"
```
