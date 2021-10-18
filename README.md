# actions-gh-release

An action that enables operating GitHub release via pull request. You send a pull request to update a RELEASE file then release note will be generated and commented in that pull request for reviewing. Once merged, a new GitHub release will be created with that release note.

![](https://github.com/pipe-cd/actions-gh-release/blob/main/assets/changelog-comment.png)

## Usage

- Adding a RELEASE file to the repository. You can also have multiple RELEASE files in case of monorepo style. Its content looks like this:

``` yaml
tag: v0.1.0                       # The tag number will be created. Required.

# # Optional fields:
#
# title: string                   # The release tile. Default is "Release ${tag}".
# targetCommitish: string         # The release commitish. Default is the merged commit.
# releaseNote: string             # The release body. Default is the auto-generated release note.
# prerelease: bool                # True if this is a prerelease. Default is false.
#
#
# # If specified, all matching commits will be excluded from release. Empty means excluding nothing.
#
# commitExclude:
#   prefixes: []string            # Matches if commit's subject is prefixed by one of the given values. Default is emtpy.
#   contains: []string            # Matches if commit's body is containing one of the given values. Default is emtpy.
#
# # If specified, all matching commits will be included to release. Empty means including alls.
#
# commitInclude:
#   prefixes: []string            # Matches if commit's subject is prefixed by one of the given values. Default is emtpy.
#   contains: []string            # Matches if commit's body is containing one of the given values. Default is emtpy.
#
#
# # List of categories and how to decide which category a commit should belong to.
#
# commitCategories:
#   - title: string               # Category title.
#     contains: []string          # Matches if commit's subject is prefixed by one of the given values. Default is emtpy.
#     prefixes: []string          # Matches if commit's body is containing one of the given values. Default is emtpy.
#
#
# # Config used while generating release note.
#
# releaseNoteGenerator:
#   showAbbrevHash: boolean       # Whether to include abbreviated hash value in release note. Default is false.
#   showCommitter: boolean        # Whether to include committer in release note. Default is true.
#   useReleaseNoteBlock: boolean  # Whether to use release note block instead of commit message. Default is false.
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
    steps:
      - uses: actions/checkout@v1
      - uses: pipe-cd/actions-gh-release@v1.3.0
        with:
          release_file: '**/RELEASE'
          token: ${{ secrets.GITHUB_TOKEN }}
```

## Inputs

| Name                  | Description                                                                       | Required | Default Value |
|-----------------------|-----------------------------------------------------------------------------------|:--------:|:-------------:|
| token                 | The GITHUB_TOKEN secret.                                                          |    yes   |               |
| release_file          | The path to the RELEASE file or pattern to match one or multiple RELEASE files.   |    no    |    RELEASE    |

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

commitInclude:
  contains:
    - application/foo
```

``` yaml
tag: bar-v1.0.0

commitInclude:
  contains:
    - application/bar
```
