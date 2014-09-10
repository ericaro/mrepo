[![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=master)](https://travis-ci.org/ericaro/mrepo) TravisCI


mrepo - multi repo tool
=====

basically *mrepo* run commands on every repository in your workspace.

mrepo is a library but mrepo/az is a main that run a command on every repository found in the working directory.

    $ az git status -s

leads to the following output:

    subrepo: src/github.com/Rafflecopter/golang-relyq$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/Rafflecopter/golang-simpleq$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/Rafflecopter/golang-messageq$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/satori/go.uuid$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/garyburd/redigo$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/kr/hk$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/kr/binarydist$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/ericaro/mrepo$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/yanatan16/errorcaller$ git status -bs
    ## master...origin/master
    subrepo: src/github.com/yanatan16/gowaiter$ git status -bs
    ## master...origin/master

it has a "concurrent" option that run each command concurrently.

    $az -a git fetch

Some command really benefit from concurrency (fetch in particular).






# License
-------

mrepo is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

# Dev branches

dev: [![Build Status](https://travis-ci.org/ericaro/mrepo.png?branch=dev)](https://travis-ci.org/ericaro/mrepo) TravisCI
