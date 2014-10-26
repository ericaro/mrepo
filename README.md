[![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=master)](https://travis-ci.org/ericaro/mrepo) [![GoDoc](https://godoc.org/github.com/ericaro/mrepo?status.svg)](https://godoc.org/github.com/ericaro/mrepo)


#mrepo - multi repo toolbox

What is mrepo for ?

  - `mrepo` is a programming library to deal with 'workspaces' that contains several git repositories, called 'subrepository'
  - `az` is a command line tool, to run arbitrary command, on each subrepository.
  - `git-deps` is a command line tool, to read all subrepositories path, remote and branch. Enough to recreate them, in fact.


## `az`

### simple command

Run a simple git command on each subrepository. 

    az <command>

to run `<command>` on each subrepository

For instance:

    $ az git status -s


### simple async command

Run each command in parallel:

    az -a <command>

For instance:

    $ az -a git fetch

It's 10x faster, because `git fetch` spend lot of time waiting.

*caveat*:
Command executed in *async* mode cannot be interactive, and cannot print in coloring mode neither.


### summary command


Collect each command response and print out a *summary* of those response.

 - *cat*   : outputs are just con`cat`enated together.
 - *sum*   : outputs are interpreted as numbers, and they are added.
 - *count* : count different outputs
 - *digest*: compute the sha1 of all outputs. subrepositories are sorted in aphabetical order of project names

*caveat*: like for `-a` option, when using summary options, commands cannot be interactive.

## git-deps

Find all subrepositories of the current working dir, and extract their:

  - path
  - remote: git remote url `git config --get remote.origin.url`
  - branch: the current branch : `git rev-parse --abbrev-ref HEAD`

By default, it prints this result.

With `-makefile` option, it prints this result in a `Makefile` format.

    <path>: ; git clone <remote> -b <branch> $@

It is then easy to reproduce your workspace elsewhere ( other developpers, CI, build machine)

# Examples

## branch distribution

The git command:

    $ git rev-parse --abbrev-ref HEAD

will give you the current branch.

But what the branches distribution in the workspace ?

    $ az -count git rev-parse --abbrev-ref HEAD
      24   dev
      12   master
    
      36   Total

Explanations:
In the current workspace, there are `36` subrepositories.
There have been `36`  responses
`24` were `dev` and `12` were `master`.

## sha1 of all sha1

The command:

    $ git rev-parse HEAD

will return HEAD's sha1.

How can I compute a new sha1 that depends on each subrepository ?

    $ az -digest git rev-parse HEAD
    bb502cc5594cf1dd2f175942dfe2cdfea4961048

Explanation:

`az` will execute `git rev-parse HEAD` on each subrepository, in a deterministic order (alphabetically by path).
a new message is build by concatenating all outputs together, and its sha1 is computed.

You have a version number for the workspace that depends on each subrepository version.

## counting commits

This git command:

    $ git  rev-list --count  HEAD...origin/master

count the number of commit between HEAD and origin/master (telling you how much behind you are).

What about all repositories ?

    $ az -sum git  rev-list --count  dev...origin/dev
        0     foo
        4     bar
        2     baz
        __________
        30  

## working with a CI

Generate a Dependencyfile

    $ git deps  > Dependencyfile
    $ git add Dependencyfile
    $ git commit -m "Added Dependencyfile for my CI"

On the CI side, you don't just need a pull on the main repository, you also need to clone new repositories:

    $ git deps -diff -clone -prune < Dependencyfile
    $ az -a git fetch
    $ az git merge --ff-only

The first statement will clone missing subrepositories and prune old ones.
The second will fetch all new stuff (asynchronously, so really fast, no possible conflict)
The third, will apply changes (fast forward only (this should be the case for a CI))


# Installation

If you have [Go](http://golang.org) installed 

    go get github.com/ericaro/mrepo

you will get in `$GOPATH/bin` the `az`, and `git-deps` commands. try them with `az -h` or `git deps -h`

# License

mrepo is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

# Branches


I've dropped support for 1.0, because I'm using bufio.NewScanner() that was not in 1.0.

master: [![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=master)](https://travis-ci.org/ericaro/mrepo) against go versions:

  - 1.1
  - 1.2
  - 1.3
  - tip

dev: [![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=dev)](https://travis-ci.org/ericaro/mrepo) against go versions:

  - 1.1
  - 1.2
  - 1.3
  - tip


