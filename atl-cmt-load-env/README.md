atl-cmt-load-env
================

``atl-cmt-load-env`` is a simple program to load a ``CMT`` environment
previously saved with ``atl-cmt-save-env``.

## Installation

``sh
$ go get github.com/atlas-org/scripts/atl-cmt-load-env
```

## Examples

### Shell family
``sh
$ atl-cmt-load-env -f store.cmt -o setup.sh
$ . ./setup.sh
``

``sh
$ eval `atl-cmt-load-env -f store.cmt`
``

### C-Shell family

``sh
$ atl-cmt-load-env -sh=csh -f store.cmt -o setup.csh
$ source ./setup.csh
``

``sh
$ eval `atl-cmt-load-env -sh=csh -f store.cmt`
$ source ./setup.csh
``

