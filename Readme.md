[![Build Status](https://travis-ci.org/ericaro/sbr.png?branch=master)](https://travis-ci.org/ericaro/sbr)

`sbr` is a command line tool to solve the problem of:

  - sharing workspace between teammates
  - reproducible builds

A **workspace** is a top **git** repository that contains other git *subrepositories* ('sbr' for short).

The **workspace** layout is described in a `.sbr` files in its root.

    <relative path1> <remote url1>
    <relative path2> <remote url2>

`sbr checkout` will clone all the missing subrepository.

# Sharing workspace between teammates

A **workspace** is a plain git repo, you can have it on any git hosting server to share.

    sbr clone <remoteURL>

Is enough to clone the workspace, and to checkout all subrepositories.

*tip* : you would probably want to add the subrepositories to the `.gitignore` file, or better checkout all subrepositories to `src` and add `src/` to `.gitignore`

# reproducible workspace

The goal of having a reproducible workspace is not to be able to reproduce every single possible configuration, but to reproduce the one you want, 'branches' and 'tags'.

## branching

By default, `sbr` follows branches. All workspaces are at the 'head' of their branch.
Just like with git, there is a 'version' to identify this state

`sbr version` will compute the sha1 of all subrepositories sha1. providing a unique version number. You can check that you, and your teammate share the same workspace version.

You can have several workspace branch. For instance a 'master' branch where every sub-repository is in the 'master' branch. etc. But you can be more picky, and have all sub-repository in the 'master' branch but `src/github.com/mine/a`.


## tagging

In git, to freeze a revision you just 'tag' it. Nothing new here, just tag every sub-repository.

You have a fully reproducible workspace.

Usually, use the same tag name, that contains the product name, and the version, so the sub-repository module or library is annotated with the top product (or integration project) versions. You will be able to compare which library revision is present in which product version. That's a huge benefit.


## comparing with git-submodule

As *git-submodule* has a `.gitmodules`, 'sbr' has a `.sbr` file, describing the layout of this workspace.

*Git-submodule* checkouts modules right to a specific commit, in *headless* mode. It is very suitable to depend on another module, but not really to collaborate on it. 

*sbr* checkouts subrepositories to a specific **branch**, or **tag**. Subrepositories are plain repositories ready for collaboration.

*Git-submodule* manages version top/down: the top sha1 enforces the sha1 for each *module*. This is handy to checkout a workspace at a given point in time. But this is too restrictive to do collaborate with each other (which branch? how can I edit a submodule? )

*sbr* computes the top version ( `sbr version` ) from subrepositories versions (sha1 of all sha1). You don't always "control" the workspace version, you compute it. This is exactly like 'go get'. It is better for *continuous integration*, and it is the only viable solution to the dependency hell.


## What is reproductible ?

Everybody in the team can checkout the current *head* of each repository, checkouting new one, and pruning old ones., you can compute the workspace version, and compare with each other.

Everybody works on the "head" of their branch.

When you want to mark a version as *important* you would `git tag` every repositories with the same "tag" ( containing, for instance the product name, and probably a changelog). Then you'll have a huge benefit to be able to track, on each repository, the commits that were used



A big benefit of a computed version, is that, to get control over it, you would have a "production" branch, or "release" tag on each subrepository, related to the "workspace" using it.
For instance, all the subrepositories I use, have the tag "sbr-1.0" and "sbr-1.1" used to have a fully reproductible sbr 1.0 version.





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


