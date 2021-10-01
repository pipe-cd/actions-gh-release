# actions-gh-release

An action that enables operating GitHub release via pull request. You send a pull request to update the RELEASE file and the changelog will be generated and commented in that pull request for reviewing. And after merging it, a new GitHub release will be created with that changelog.

![](https://github.com/pipe-cd/actions-gh-release/blob/main/assets/changelog-comment.png)

## Usage

- Adding a RELEASE file to the repository. Its content looks like this:

``` yaml
tag: v0.1.0           # The tag number will be created. Required.

# Optional fields:

# title: ""           # The release tile. Default is "Release ${tag}".
# commitish: ""       # The release commitish. Default is the merged commit.
# body: ""            # The release body. Default is the auto-generated changelog.
# prerelease: false   # True if this is a prerelease.
```

- Adding a new workflow (eg: `.github/workflows/gh-release.yaml`) with the content as below:

```yaml
on:
  push:
    branches:    
      - main
    paths:
      - 'RELEASE'
  pull_request:
    types: [opened, synchronize]
    branches:
      - main
    paths:
      - 'RELEASE'

jobs:
  gh-release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
      - uses: pipe-cd/actions-gh-release@v1.1.0
        with:
          release_file: 'RELEASE'
          token: ${{ secrets.GITHUB_TOKEN }}
```

## Inputs

| Name                            | Description                                                                     | Required | Default Value |
|---------------------------------|---------------------------------------------------------------------------------|:--------:|:-------------:|
| token                           | The GITHUB_TOKEN secret.                                                        |    yes   |               |
| release_file                    | The path to the RELEASE file.                                                   |    no    |    RELEASE    |
| body                            | The release body should be used. Empty means using an auto-generated changelog. |    no    |       ""      |
| changelog_show_abbrev_hash      | Include abbreviated hash value in the changelog.                                |    no    |      true     |
| changelog_show_committer        | Include committer name in the changelog.                                        |    no    |      true     |
| changelog_only_use_merge_commit | Only use merge commits to generate changelog. Ignore all other commit kinds.    |    no    |     false     |
| changelog_ignore_merge_commit   | Ignore merge commits from the changelog.                                        |    no    |     false     |

## Outputs

| Name        | Description                                          |
|-------------|------------------------------------------------------|
| id          | The ID of the created release.                       |
| tag         | The tag name of the created release.                 |
| html_url    | The HTML URL to view the created release.            |
| upload_url  | The URL for uploading assets to the created release. |
| changelog   | The generated changelog for the release.             |
| change_json | The change list formatted in JSON.                   |
