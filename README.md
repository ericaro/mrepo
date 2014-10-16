[![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=master)](https://travis-ci.org/ericaro/mrepo)


#mrepo - multi repo toolbox

*mrepo* is a programming toolbox to deal with 'workspace' that contains several repositories

`az`  is a command line utility to execute arbitrary commands on each repository. Typically you want to do:

    $ az git fetch

to run `git fetch` on every repository in your workspace.

Even better, because fetch spent a lot of time waiting for server response, 

    $ az -a git fetch

to start a 'git fetch' command on each repository asynchronously. It's 10x faster.

Of course you can run any command you want

    $ az mkdir -p src/main/java

# Details

`az` scan recursively all directories in the current working directory hierarchy.
if a directory's name is ".git" then

  1 it is skipped
  2 it's parent dir is considered to be a 'repository'.

Command are executed in the context of each repository (the current working dir is the repository)


Commands can be executed *sequentially* with direct access to stdin, stdout stderr. Command can ask questions, and print using color
Commands can be executed *asynchronously*  with buffered access to stdout and stderr.

#statistics 

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

    go get github.com/ericaro/mrepo

you will get in `$GOPATH/bin` the 'az' command. try it with `az -h`


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


