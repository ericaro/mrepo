[![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=master)](https://travis-ci.org/ericaro/mrepo)


#mrepo - multi repo toolbox

  - `mrepo` is a programming library to deal with 'workspaces' that contains several git repositories, called 'subrepository'
  - `az` is a command line tool, to run arbitrary command, on each subrepository.
  - `git-deps` is a command line tool, to read all subrepository path, remote, branch. Enough to recreate them, in fact.


## `az`

`az`  is a command line utility to execute arbitrary commands on each repository. Typically you want to do:

    $ az git fetch

to `git fetch` every subrepository.

Even better, because `fetch` actually spend a lot of time waiting for server response: 

    $ az -a git fetch

that will `git fetch`  on each repository, but in parallel. It's 10x faster.


Of course you can run any command you want

    $ az mkdir -p src/main/java


### sync or async ?

The `-a` option selects the `asynchronous` mode.

  - In `sync` mode:
    + Commands are executed *sequentially*. 
    + Commands have direct access to stdin, stdout stderr, and therefore, can prompt for questions, and print using color
  - In `async` mode:
    + Commands are executed *asynchronously*.
    + Commands cannot run in interactive mode, but they usually operates faster.

The `async` mode is activate by `-a` option, or if you use a *statistics aggregator* like ( `-cat`, or `-count`)


### statistics

Sometimes you need to run some basic statistics on those results:

    git rev-parse --abbrev-ref HEAD

will give you the current branch

But what about all branches, in the workspace ?


    $ az -count git rev-parse --abbrev-ref HEAD
      24 : dev
      12 : master
      ___________
      36

Commands results can be analyzed through a few simple "aggregators"

 - *cat* : outputs are just `cat` to the output.
 - *sum* : outputs are considered as numbers, and they are sum up.
 - *count*: count differents outputs, and print the result and the count.
 - *digest*: the sha1 of all outputs, in the aphabetical order of project names, is computed and printed.


This set of "outputs" is very suitable with some git command:
The sha1 of all sha1: just print the each repository sha1, and compute the resulting one:

    $ az -digest git rev-parse --verify HEAD

Number of commit ahead or behind

    $ az -sum git  rev-list --count  dev...origin/dev

See all untracked file in all the repositories (to see if one need to be committed)

    $ az -cat git ls-files --exclude-standard --others

  
    $ az -cat git status --porcelain
      M  README.md
       M az/main.go
       M runner.go
      ?? DesignNotes.md


Check the current repartition of branches

    $ az -count git rev-parse --abbrev-ref HEAD
      24 : dev
      12 : master
      ___________
      36


# Installation

If you have [Go](http://golang.org) installed 

    go get github.com/ericaro/mrepo

you will get in `$GOPATH/bin` the `az`, and `git-deps` commands. try them with `az -h` or `git deps -h`

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


