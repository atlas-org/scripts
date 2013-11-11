atl-svn
=======

``atl-svn`` is a multi-command executable.

## Installation

```sh
$ go get github.com/atlas-org/scripts/atl-svn
```

## ``atl-svn diff``

``atl-svn diff`` runs ``svn diff`` between 2 packages tags.

```sh
$ atl-svn diff AthenaKernel-00-01-02 Control/AthenaKernel-00-02-02
$ atl-svn diff AthenaKernel-00-01-02 AthenaKernel-00-02-02
$ atl-svn diff AthenaKernel-00-01-02 AthenaKernel-trunk
$ atl-svn diff AthenaKernel-00-01-02 AthenaKernel-HEAD
```
