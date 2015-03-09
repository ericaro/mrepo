package makefile

import (
	"fmt"
	"io"
	"os/exec"
)

//Make invoke make on the prj with the target as argument.
func Run(prj, target string, buf io.Writer) (err error) {

	//because, and I don't know why, $(PWD) will return "os.Getwd" instead of prj
	// I need to go through a bash
	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && make %s", prj, target))
	cmd.Dir = prj
	cmd.Stdout = buf
	cmd.Stderr = buf
	return cmd.Run()
}
