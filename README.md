[![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=master)](https://travis-ci.org/ericaro/mrepo)


#mrepo - multi repo toolbox

*mrepo* is a programming toolbox to deal with 'workspace' that contains several repositories (git, mercurial etc.)

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
if a directory's name is one of:
  - ".git"
  - ".hg"
  - ".bzr"
  - ".svn"
  - "CVS"
Then

  1 it is skipped
  2 it's parent dir is considered to be a 'repository'.

Command are executed in the context of each repository ( the current working dir is the repository)

Commands can be executed *sequentially* with direct access to stdin, stdout stderr. Command can ask questions, and print using color
Commands can be executed *asynchronously*  with buffered access to stdout and stderr.



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


