atl-cmt-save-env
================

``atl-cmt-save-env`` is a simple program to save a ``CMT`` environment
into a ``JSON`` file.

## Installation

``sh
$ go get github.com/atlas-org/scripts/atl-cmt-save-env
```

## Examples

``sh
$ atl-cmt-save-env rel1,devval
$ atl-cmt-save-env -f store.cmt rel1,devval
``

The resulting file can then be used by ``atl-cmt-load-env``.

