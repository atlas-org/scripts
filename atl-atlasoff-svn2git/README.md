atl-atlasoff-svn2git
====================

Converts the ``atlasoff`` SVN repository into a ``git`` one.

## Installation

```sh
$ go get atlas-org/scripts/atl-atlasoff-svn2git
```

## Usage

```sh
$ atl-atlasoff-svn2git atlasoff

# re-sync with SVN
$ cd atlasoff-git
$ atl-atlasoff-svn2git -sync
```
