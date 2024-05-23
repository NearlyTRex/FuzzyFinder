//go:build windows

package fzf

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/junegunn/fzf/src/util"
)

func isMintty345() bool {
	return util.CompareVersions(os.Getenv("TERM_PROGRAM_VERSION"), "3.4.5") >= 0
}

func needWinpty(opts *Options) bool {
	if os.Getenv("TERM_PROGRAM") != "mintty" {
		return false
	}
	if isMintty345() {
		/*
		 See: https://github.com/junegunn/fzf/issues/3809

		 "MSYS=enable_pcon" allows fzf to run properly on mintty 3.4.5 or later,
		 however `--height` option still doesn't work, so let's just disable it.

		 We're not going to worry too much about restoring the original value.
		*/
		if os.Getenv("MSYS") == "enable_pcon" {
			opts.Height = heightSpec{}
			return false
		}

		// Setting the environment variable here unfortunately doesn't help,
		// so we need to start a child process with "MSYS=enable_pcon"
		//   os.Setenv("MSYS", "enable_pcon")
		return true
	}
	if _, err := exec.LookPath("winpty"); err != nil {
		return false
	}
	if opts.NoWinpty {
		return false
	}
	return true
}

func runWinpty(args []string, opts *Options) (int, error) {
	sh, err := sh()
	if err != nil {
		return ExitError, err
	}

	argStr := escapeSingleQuote(args[0])
	for _, arg := range args[1:] {
		argStr += " " + escapeSingleQuote(arg)
	}
	argStr += ` --no-winpty --no-height`

	if isMintty345() {
		return runProxy(argStr, func(temp string) *exec.Cmd {
			cmd := exec.Command(sh, temp)
			cmd.Env = append(os.Environ(), "MSYS=enable_pcon")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd
		}, opts, false)
	}

	return runProxy(argStr, func(temp string) *exec.Cmd {
		cmd := exec.Command(sh, "-c", fmt.Sprintf(`winpty < /dev/tty > /dev/tty -- sh %q`, temp))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd
	}, opts, false)
}