// Package infra proxies Terraform commands.
package infra

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apex/log"

	"github.com/apex/apex/project"
)

// Proxy is a wrapper around Terraform commands.
type Proxy struct {
	Project     *project.Project
	Environment string
}

// functionVars returns the function ARN's as terraform -var arguments.
func (p *Proxy) functionVars() (args []string) {
	for _, fn := range p.Project.Functions {
		// TODO: not ideal to do network I/O here
		config, err := fn.GetConfig()
		if err != nil {
			log.Debugf("can't fetch function config: %s", err.Error())
			continue
		}

		args = append(args, "-var")
		args = append(args, fmt.Sprintf("apex_function_%s=%s", fn.Name, *config.Configuration.FunctionArn))
	}

	return args
}

// shouldInjectVars checks if the command accepts -var flags.
func (p *Proxy) shouldInjectVars(args []string) bool {
	if len(args) == 0 {
		return false
	}

	return args[0] == "plan" || args[0] == "apply"
}

// Run terraform command in Environment.
func (p *Proxy) Run(args ...string) error {
	if p.shouldInjectVars(args) {
		args = append(args, p.functionVars()...)
	}

	log.WithFields(log.Fields{
		"env":  p.Environment,
		"args": args,
	}).Debug("terraform")

	dir := filepath.Join("infrastructure", p.Environment)
	cmd := exec.Command("terraform", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir

	return cmd.Run()
}
