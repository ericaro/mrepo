package cmd

var SbrFormatMd = `

# sbr File

'.sbr' file is a simple easy to read, easy to write format for both human and
 computers.

  - 'sbr' commands are permissive when reading '.sbr' files, allowing human-ish edition
  - 'sbr' commands are strict while writing '.sbr' files, they always generate the same output.

the 'sbr format' command can be use to simply reformat the '.sbr'.

    $ sbr format


'.sbr' content is basically a list of subrepositories information. 

A subrepository is made of:

  - a path: relative to the .sbr's directory, that shall contain a git repository
  - remote: the git remote address (e.g. "git@github.com:ericaro/mrepo.git")
  - branch: the git branch that shall be checkouted ( e.g. "master" )


## Record and Fields

An '.sbr' file contains zero or more *records* of one or more *fields* per record. 

Each record is separated by the newline character. The final record may optionally be followed by a newline character.

Fields which start and stop with the quote character " are called quoted-fields. 
The beginning and ending quote are not part of the field.

Within a quoted-field a quote character followed by a second quote character is considered a single quote.

## Readable Format

'sbr' parser is more permissive than the 'sbr' writer, here is the format it can read.


There are 4 kinds of record based on the number of fields:

  - *branch* declaration          1-field: "branch"
  - *subrepository* declaration: 2-fields: "remote" "path"
  - *subrepository* declaration: 3-fields: "remote" "path" "branch"
  - *subrepository* declaration: 4-fields: git "remote" "path" "branch"

The [reading algorithm](https://github.com/ericaro/mrepo/blob/master/workspace.go#L269) is simple:

    currentBranch := "master"
    for _, record := range records{
        switch len(record){
          case 1:
            currentBranch= record[0]
          case 2:
            newSubrepository(record[0], record[1], currentBranch)
          case 3:
            newSubrepository(record[0], record[1], record[2]) 
          case 4: //legacy
            newSubrepository(record[1], record[2],record[3]) 
        }
    }



## Normalized Format

The normalized format applies the following rules:

  - The default branch is always "master"
  - In the normalized format subrepositories are sorted:
    - by branch
    - then by path
  - always uses quoted fields.
  - make use of 1,2-fields records.


# why such a format ?

I've tried to comply with [those hints](http://monkey.org/~marius/unix-tools-hints.html). *json* does not play nice with humans, *tabular* is much readable.

Driving uses cases:

  - **a human need to add/remove/update a repository**: the format need to be human friendly. Hence the 3-fields format to "override" normalized format.
  - **a human will merge .sbr files using git**: the format need to be repeatable and sparse. Useless repetion, or random order will result in painful and meaningless merge conflicts.
  - **programs need to add/remove/edit repository**: it should be easy to parse.


You can just move a subrepository record to change it's branch

from:

    1: "dev"
    2: "src/github.com/ericaro/ansifmt" "git@github.com:ericaro/ansifmt.git"
    3: "src/github.com/ericaro/mrepo" "git@github.com:ericaro/mrepo.git"
    4: "master"
    5: "src/github.com/ericaro/frontmatter" "git@github.com:ericaro/frontmatter.git"
    6: "src/github.com/ericaro/ringbuffer" "git@github.com:ericaro/ringbuffer.git"

To move 'mrepo' to branch master just move line '3:' to line '5:'

    1: "dev"
    2: "src/github.com/ericaro/ansifmt" "git@github.com:ericaro/ansifmt.git"
    4: "master"
    3: "src/github.com/ericaro/mrepo" "git@github.com:ericaro/mrepo.git"
    5: "src/github.com/ericaro/frontmatter" "git@github.com:ericaro/frontmatter.git"
    6: "src/github.com/ericaro/ringbuffer" "git@github.com:ericaro/ringbuffer.git"

You could also have use the 3-field record


    1: "dev"
    2: "src/github.com/ericaro/ansifmt" "git@github.com:ericaro/ansifmt.git"
    3: "src/github.com/ericaro/mrepo" "git@github.com:ericaro/mrepo.git" "master"
    4: "master"
    5: "src/github.com/ericaro/frontmatter" "git@github.com:ericaro/frontmatter.git"
    6: "src/github.com/ericaro/ringbuffer" "git@github.com:ericaro/ringbuffer.git"

In both cases 'sbr format' will rewrite it correctly, and cannonically.

`
