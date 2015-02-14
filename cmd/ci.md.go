package cmd

var CIServerMd = `
# CI Server

A CI Server is basically a 'daemon' that can run continuous integration jobs.

## Job

A Job is made of:

  - *path*  : it's name (also a relative path on the disk)
  - *remote*: the remote url of the top project to clone
  - *branch*: the remote branch to checkout


A Job is executed in its *path* directory, in two phases:
  
  - *refresh*:
    - if the *path* does not exists, 'git clone' it otherwise 'git pull' it.
    - read it's root '.sbr' file and 'prune/clone/pull' dependencies
  - *build* : 
    - if the 'sbr version' has changed run 'make ci'

A Job execution is meant to be very quick if there was no changes.

git pull uses the --ff-only (fast forward only).


## Hooks

The daemon also starts a Hook server, and for each "POST" received, it will trigger a "Heartbeat"

The daemon does not "parse" the POST payload, it is therefore possible to trigger heartbeats from any git hosting service (bitbucket, github, gitlab ...) or from a post commit hook locally.

## Heartbeat

An Heartbeat can occur after a webhook POST, or at a regular pace. Both ways, 
the heartbeat will wait for a few seconds (10s) before launching all Jobs.

## Communication

There are three way to communicate with the CI server.


  - Hook Server: http://localhost:2121
    + POST:  triggers a heartbeat
    + GET :  return a status code in the http response body
      * 0 for "KO"
      * 1 for "Running"
      * 2 for "OK"
  - Daemon Server: http://localhost:2020
    + decodes the http request as a protobuf 'request' encodes back a protobuf 'response' into the http response. 

See [https://github.com/ericaro/mrepo/tree/master/format/ci.proto] for details.

There is a go client implementation for the Daemon Server.

`
