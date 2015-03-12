[![Build Status](https://travis-ci.org/ericaro/sbr.png?branch=master)](https://travis-ci.org/ericaro/sbr)

sbr is a command line tool to manage a git repository as a workspace of other git subrepositories ('sbr' for short).

Each subrepository goes into a dedicated subdirectory, pointing at a branch, so you can edit, commit, push and pull them. They are plain git repositories.

Subrepositories are composed from an 'sbr' definition made of three simple attributes:

 - **rel**: subdirectory path
 - **remote**: git remote URL
 - **branch**: git branch

All sbr definitions are collected into a '.sbr' file.

# Benefits

## Collaboration

**Share your workspace with your teammate**

Collaborate as usual throught git, but when you need to add a new repo, or delete a deprecated one, just remove it from the **.sbr** file, and push it, so your team can:

- know that there is a change 'sbr diff'
- apply it 'sbr checkout'

When working with many git repositories, you are not always sure which one need to be pulled.

    $ sbr status origin/dev

    2   1   src/github.com/ericaro/frontmatter        
    -   -   src/github.com/ericaro/help               
    5   -   src/github.com/ericaro/mrepo 

first  col: commits to be pushed
second col: commits to be pulled


## Continuous Integration

**Run CI agents that can reproduce your workspace**

'sbr' CI agent listens to http POST hooks from the git hosting provider (*github*, *bitbucket*, *gitlab*), and start a job.

Every change in any subrepository is likely to start a build. That's real continuous integration.

On every http POST call 'sbr' computes the *workspace version*, made of the sha1 of all git sha1s very quickly, and launch a build it has changed. 

## sub commands


Type `sbr -h ` or `sbr help` or `sbr <command> -h` for details

**sbr version** will compute the sha1 of all sha1 (self, and each subrepository), this sbr-version can be used to identify the project version.

**sbr checkout** will keep in sync all subrepositories from the '.sbr' file. Cloning new subrepositories, pruning (optional) deleted one, and pulling ( optionally ff-only, or --rebase) all the others.

**sbr diff** will compare the '.sbr' content with the actual subrepositories that can be found on the disk. Optionally, you can apply differences back to the '.sbr' file, or use meld to compare the two

**sbr format** rewrite the .sbr file in a cannonical format, avoiding useless conflicts

**sbr x** run any command on each subrepository. `sbr x git fetch` will fetch every surepository. `sbr x git status` will print a status of each subrepository, or `sbr x git push` to push all commits. Checkout the command 'a' also available as a standalone one.

**sbr status** displays the number of commits to be pushed or pulled between the current branch and the remote. First column is for the number of commits to be pushed, the second for the number of commits to be pulled.

    2   1   src/github.com/ericaro/frontmatter        
    -   -   src/github.com/ericaro/help               
    5   -   src/github.com/ericaro/mrepo 



**sbr clone** all in one command, it clones a git repository, and also clones all of its subrepositories

**sbr ci** set of commands start and control a CI agent, and a simple dashboard for it


# installation

**from source**

get [go](http://golang.org) then 

~~~ sh
go get github.com/ericaro/sbr
~~~

**download**

To be done.



# License

sbr is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

# Branches

master: [![Build Status](https://travis-ci.org/ericaro/sbr.png?branch=master)](https://travis-ci.org/ericaro/sbr) against go versions:

  - 1.2
  - 1.3
  - 1.4
  - tip

dev: [![Build Status](https://travis-ci.org/ericaro/sbr.png?branch=dev)](https://travis-ci.org/ericaro/sbr) against go versions:

  - 1.2
  - 1.3
  - 1.4
  - tip


