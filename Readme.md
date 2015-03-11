[![Build Status](https://travis-ci.org/ericaro/sbr.png?branch=master)](https://travis-ci.org/ericaro/sbr)

sbr is a command line tool to manage a git repository as a workspace of git subrepositories (sbr for short).

Each subrepository goes into a dedicated subdirectory, pointing at **no particular* commit, but to a branch instead, so you can edit, commit, push and pull them. They are plain git repositories.

Subrepositories are composed from an 'sbr' definition made of three simple attibutes:

 - subdirectory path (rel)
 - git remote URL (remote)
 - git branch (branch)

All sbr definitions are collected into a '.sbr' file.

## benefits

**collaboration**. Having a '.sbr' file that describe your workspace it's the basis for collaboration. You, and your teammate can keep your workspace in sync. (`sbr checkout` it's all it takes to clone new repositories.)

**continuous integration**. 'sbr' comes with a CI agent, that keep in sync a collection of workspaces, and launch a build command.

Sync are triggered by git hosting POST hooks.





Then sbr offers a few utilities:

**sbr version** will compute the sha1 of all sha1 (self, and each subrepository), this sbr-version can be used to identify the project version.

**sbr checkout** will keep in sync all subrepositories from the '.sbr' file. Cloning new subrepositories, pruning (optional) deleted one, and pulling ( optionally ff-only, or --rebase) all the others.

**sbr diff** will compare the '.sbr' content with the actual subrepositories that can be found on the disk. Optionally, you can apply differences back to the '.sbr' file, or use meld to compare the two

**sbr format** rewrite the .sbr file in a cannonical format, avoiding useless conflicts
**sbr x** run any command on each subrepository. `sbr x git fetch` will fetch every surepository. `sbr x git status` will print a status of each subrepository

**sbr status** displays the number of commits to be pushed or pulled between the current branch and the remote. First column is for the number of commits to be pushed, the second for the number of commits to be pulled.

    2   1   src/github.com/ericaro/frontmatter        
    -   -   src/github.com/ericaro/help               
    5   -   src/github.com/ericaro/mrepo 


**sbr clone** all in one command, it clones a git repository, and also clones all of its subrepositories

**sbr ci** is a set of commands. One of the benefit of using sbr is that the whole workspace can be reproduced elsewhere, teammates or CI agent. **sbr ci** can launch a CI agent, an interact with it (add/remove workspaces, see logs, display a simple html dashboard)




# installation

get [go](http://golang.org) then 

~~~ sh
go get github.com/ericaro/sbr
~~~


# License

sbr is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

# Branches

master: [![Build Status](https://travis-ci.org/ericaro/sbr.png?branch=master)](https://travis-ci.org/ericaro/sbr) against go versions:

  - 1.1
  - 1.2
  - 1.3
  - tip

dev: [![Build Status](https://travis-ci.org/ericaro/sbr.png?branch=dev)](https://travis-ci.org/ericaro/dev) against go versions:

  - 1.2
  - 1.3
  - tip


