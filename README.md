[![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=master)](https://travis-ci.org/ericaro/mrepo) [![GoDoc](https://godoc.org/github.com/ericaro/mrepo?status.svg)](https://godoc.org/github.com/ericaro/mrepo)

What is mrepo for ?

  - `mrepo` is a programming library to deal with 'workspaces' that contains several git repositories, called 'subrepository'
  - `a` is a command line tool, to run arbitrary command, on each subrepository.
  - `sbr` is a command line tool, to synchronize a whole workspace between the working directory and a local .sbr file.


# `a` run on `a`ll subrepositories

Run a simple git command on each subrepository. 

    a <command>

For instance:

    $ a git status -s
    .$ git status -s
     M mkfiles/redis.mk
     M mkfiles/sql.mk
    ?? toto.txt
    src/github/mrepo$ git status -s
    ?? editor.go.orig


### Concurrent mode

Run each command in parallel:

    a -a <command>

For instance:

    $ a -a git fetch

It's 10x faster, because `git fetch` spend lot of time waiting.

*caveat*:
Command executed in *async* mode cannot be interactive, and cannot print in coloring mode neither.

*tip*:
Set an alias to run commands in async mode:

    alias af='a -a'


### summary command

`a` is built in with the ability to run a few post processing operations on command results:

 - *cat*   : outputs are just con`cat`enated together.
 - *sum*   : outputs are interpreted as numbers, and they are added.
 - *count* : count different outputs
 - *digest*: compute the sha1 of all outputs. subrepositories are sorted in aphabetical order of project names

*caveat*: like with `-a` option, when using summary options, commands cannot be interactive.


# `sbr`, workspace subrepositories manager

It manages two sets of subrepositories:
  
  - ".sbr": is the set made of subrepositories declarations in a '.sbr' file
  - "disk": is the set made of actual subrepositories in the current directory hierarchy

## Usage

    sbr [-options] <command> <args...>

  <command> can be:

  - *init*    : copy "disk" set into `.sbr` file
  - *describe*: print the "disk" set
  - *apply*   : apply ".sbr" dependencies to the current working dir (prune and clone where necessary)
  - *refresh* : `git-pull`, `apply`, and `git-pull` on each subrepository.
  - *compare* : diff ".sbr" and "disk" sets. In the form of operations to apply to ".sbr" set
  - *merge*   : edit the two sets in meld
  - *version* : compute the sha1 of each "disk" dependencies

# Examples

## branch distribution

The git command:

    $ git rev-parse --abbrev-ref HEAD

will give you the current branch.

But what is the branches distribution in the workspace ?

    $ a -count git rev-parse --abbrev-ref HEAD
      24   dev
      12   master
    
      36   Total


## sha1 of all sha1

The command:

    $ git rev-parse HEAD

will return HEAD's sha1.

How can I compute a new sha1 that depends on each subrepository ?

    $ a -digest git rev-parse HEAD
    bb502cc5594cf1dd2f175942dfe2cdfea4961048


Explanation:

`a` will execute `git rev-parse HEAD` on each subrepository, in a deterministic order (alphabetically by path).
a new message is build by concatenating all outputs together, and its sha1 is computed.

You have a version number for the workspace that depends on each subrepository version.

`sbr` has it built in:

    $ sbr digest
    bb502cc5594cf1dd2f175942dfe2cdfea4961048

Leading to the same result.


## counting commits

This git command:

    $ git  rev-list --count  HEAD...origin/master

count the number of commit between HEAD and origin/master (telling you how much behind you are).

What about all repositories ?

    $ a -sum git  rev-list --count  dev...origin/dev
        0     foo
        4     bar
        2     ba
        __________
        30  

## working with a CI

Generate a ".sbr" file

    $ sbr reflect
    $ git add .sbr
    $ git commit -m "Added .sbr for my CI"

On the CI side, you don't just need a pull on the main repository, you also need to clone all dependencies

    $ sbr apply
    $ a -a git fetch
You have fully copied your workspace.

Whenever something new happens ?

    $ git pull
    $ sbr apply
    $ a -a git fetch
    $ a git merge --ff-only

First statement is to pull the top repository



# Installation

If you have [Go](http://golang.org) installed 

    go get github.com/ericaro/mrepo/{a,sbr}

you will get in `$GOPATH/bin` the `a`, and `sbr` commands. try them with `a -h` or `sbr -h`

# License

mrepo is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

# Branches


master: [![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=master)](https://travis-ci.org/ericaro/mrepo) against go versions:

  - 1.0
  - 1.1
  - 1.2
  - 1.3
  - tip

dev: [![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=dev)](https://travis-ci.org/ericaro/mrepo) against go versions:

  - 1.0
  - 1.1
  - 1.2
  - 1.3
  - tip


