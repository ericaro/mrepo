package cmd

import "errors"

const (
	CodeNoWorkingDir        = -2
	CodeMissingServerConfig = -3
	CodeMissingJobConfig    = -4
	CodeCannotDelete        = -5
	CodeMissingBranch       = -6
	CodeMissingRemoteOrigin = -7
	CodeCannotAddJob        = -8
)

var (
	ErrNoSbrfile = errors.New("Not in an 'sbr' workspace")
	ErrNoWd      = errors.New("Cannot find out the working dir")
)
