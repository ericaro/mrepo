package main

var sbrFormatMd = `

# sbr File Format

'.sbr' file is a simple easy to read, easy to write format for both human and computers.

'sbr' program reads the format to get the following information:

  - list of subrepositories
  - subrepository path
  - subrepository remote url
  - subrepository branch

There is a one line per subrepository (no extra, or blank lines)

Lines are sorted alphabetically by "path"

each line is made of:
 
   - 'git'
   - '%q', path
   - '%q', remote
   - '%q', branch (optional)

e.g.

    git "src/github.com/ericaro/sbr" "git@github.com/ericaro/sbr" "dev"


the "branch" parameter beeing optional, the previous value is used. If none are given, "master" is used.


e.g.

    git "src/github.com/ericaro/sbr" "git@github.com/ericaro/sbr" "dev"
    git "src/github.com/ericaro/help" "git@github.com/ericaro/help" 
    git "src/github.com/ericaro/ansifmt" "git@github.com/ericaro/ansifmt" 

the second, and third repos "help" and "ansifmt" are declared in branch "dev" too.

Everytime sbr rewrite the file, it sort the lines.

# why such a format ?

*json* does not play nice with humans, in particular list

*tabular* is much readable.

There are the three main uses cases driving the format.

  - *a human need to add/remove a repository*: the format need to be human friendly
  - *a program need to add/remove a repository*: the format need to be parseable
  - *a human will merge .sbr files*: the format need to be repeatable and sparse


`
