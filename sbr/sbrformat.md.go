package main

var sbrFormatMd = `

# sbr File

'.sbr' file is a simple easy to read, easy to write format for both human and computers.

  - 'sbr' reading is permissive, allowing humanish edition
  - 'sbr' writing is normative, always generate the same output, to avoid useless conflict.

the 'sbr' command provide a "format" subcommand to reformat the '.sbr'.

    $ sbr format


'.sbr' describes a list of subrepositories. A subrepository is made of:

  - a path: where the subrepository lies, relative to the .sbr directory
  - remote: git remote address
  - branch: git branch to use

## Record and Fields

A '.sbr' file contains zero or more *records* of one or more *fields* per record. Each record is separated by the newline character. The final record may optionally be followed by a newline character.

Fields which start and stop with the quote character " are called quoted-fields. The beginning and ending quote are not part of the field.

Within a quoted-field a quote character followed by a second quote character is considered a single quote.

## Readable Format

There are 4 kinds of record based on the number of fields:

  - *branch* declaration          1-field: "branch"
  - *subrepository* declaration: 2-fields: "remote" "path"
  - *subrepository* declaration: 3-fields: "remote" "path" "branch"
  - *subrepository* declaration: 4-fields: git "remote" "path" "branch"

The reading algorithm is simple:

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
  - The normalized format sort the list of subrepository:
    - by branch
    - then by path
  - always uses quoted fields.
  - only uses branch declarations, and 2-fields subrepository declarations.


# why such a format ?

I've tried to comply with [those hints](http://monkey.org/~marius/unix-tools-hints.html). *json* does not play nice with humans, *tabular* is much readable.

Driving uses cases:

  - **a human need to add/remove/update a repository**: the format need to be human friendly. Hence the 3-fields format to "override" normalized format.
  - **a human will merge .sbr files using git**: the format need to be repeatable and sparse. Useless repetion, or random order will result in painful and meaningless merge conflicts.
  - **programs need to add/remove/edit repository**: it should be easy to parse.
`
