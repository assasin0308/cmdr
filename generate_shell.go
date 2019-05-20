/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cmdr

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
)

func genShell(cmd *Command, args []string) (err error) {
	// logrus.Infof("OK gen shell. %v", *cmd)

	if GetBool("app.generate.shell.zsh") {
		// if !GetBool("app.quiet") {
		// 	logrus.Debugf("zsh-dump")
		// }
		// printHelpZsh(command, justFlags)

	} else if GetBool("app.generate.shell.bash") {
		err = genShellBash(cmd, args)
	} else {
		// auto
		shell := os.Getenv("SHELL")
		if strings.HasSuffix(shell, "/zsh") {
			//
		} else if strings.HasSuffix(shell, "/bash") {
			err = genShellBash(cmd, args)
		} else {
			_, _ = fmt.Fprint(os.Stderr, "Unknown shell. ignored.")
		}
		// err = genShellB(cmd, args)
	}
	return
}

func findLvl(cmd *Command, lvl int) (lvlMax int) {
	lvlMax = lvl + 1
	for _, cc := range cmd.SubCommands {
		l := findLvl(cc, lvl+1)
		if l > lvlMax {
			lvlMax = l
		}
	}
	return
}

func genShellBash(cmd *Command, args []string) (err error) {
	tmpl := template.New("bash.completion")
	tmpl.Parse(`

#

# bash completion wrapper for {{.AppName}}
# version: {{.Version}}
#
# Copyright (c) 2019-2025 Hedzr Yeh <hedzrz@gmail.com>
#

_cmdr_cmd_help_events () {
  $* --help|grep "^  [^ \[\$\#\!/\\@\"']"|awk -F'   ' '{print $1}'|awk -F',' '{for (i=1;i<=NF;i++) print $i}'
}


_cmdr_cmd_{{.AppName}}() {
  local cmd="{{.AppName}}" cur prev words
  _get_comp_words_by_ref cur prev words
  if [ "$prev" != "" ]; then
    unset 'words[${#words[@]}-1]'
  fi

  COMPREPLY=()
  #pre=${COMP_WORDS[COMP_CWORD-1]}
  #cur=${COMP_WORDS[COMP_CWORD]}

  case "$prev" in
    --help|--version)
      COMPREPLY=()
      return 0
      ;;
    $cmd)
      COMPREPLY=( $(compgen -W "$(_cmdr_cmd_help_events $cmd)" -- ${cur}) )
      return 0
      ;;
    *)
      COMPREPLY=( $(compgen -W "$(_cmdr_cmd_help_events ${words[@]})" -- ${cur}) )
      return 0
      ;;
  esac

  #opts="--help --version -q --quiet -v --verbose --system --dest="
  #opts="--help upgrade version deploy undeploy log ls ps start stop restart"
  opts="--help"
  cmds=$($cmd --help|grep "^  [^ \[\$\#\!/\\@\"']"|awk -F'   ' '{print $1}'|awk -F',' '{for (i=1;i<=NF;i++) print $i}')

  COMPREPLY=( $(compgen -W "${opts} ${cmds}" -- ${cur}) )

} # && complete -F _cmdr_cmd_{{.AppName}} {{.AppName}}

if type complete >/dev/null 2>&1; then
	# bash
	complete -F _cmdr_cmd_{{.AppName}} {{.AppName}}
else if type compdef >/dev/null 2>&1; then
	# zsh
	_cmdr_cmd_{{.AppName}}_zsh() { compadd $(_cmdr_cmd_{{.AppName}}); }
	compdef _cmdr_cmd_{{.AppName}}_zsh {{.AppName}}
fi; fi
`)

	for _, s := range []string{"/etc/bash_completion.d", "/usr/local/etc/bash_completion.d"} {
		if FileExists(s) {
			file := path.Join(s, cmd.root.AppName)
			var f *os.File
			if f, err = os.Create(file); e != nil {
				return
			}
			
			err = tmpl.Execute(f, cmd.root)
			if err == nil {
				fmt.Printf(`''%v generated.
Re-login to enable the new bash completion script.
`, file)
			}
			return

		}
	}

	err = tmpl.Execute(os.Stdout, cmd.root)
	return
}

func genShellB(cmd *Command, args []string) (err error) {
	// var sb strings.Builder
	// var sbca []strings.Builder

	// cx := &cmd.GetRoot().Command
	// lvl := findLvl(cx, 0)
	// sbca = make([]strings.Builder, lvl+1)

	return
}

// not complete
func genShellA(cmd *Command, args []string) (err error) {
	var sb strings.Builder
	var sbca []strings.Builder

	cx := &cmd.GetRoot().Command
	lvl := findLvl(cx, 0)
	sbca = make([]strings.Builder, lvl+1)

	sb.WriteString(fmt.Sprintf(`#compdef _%v %v

# zsh completion wrapper for %v
# version: %v
# deep: %v
#
# Copyright (c) 2019-2025 Hedzr Yeh <hedzrz@gmail.com>
#

__ac() {
	local state
	typeset -A words
	_arguments \
`,
		cmd.GetRoot().AppName, cmd.GetRoot().AppName, cmd.GetRoot().AppName, cmd.GetRoot().Version, lvl))

	for i := 1; i < lvl; i++ {
		sb.WriteString(fmt.Sprintf("\t\t'%d: :->level%d' \\\n", i, i))
	}
	sb.WriteString(fmt.Sprintf("\t\t'%d: :_files'\n\n\tcase $state in\n", lvl))

	cx = &cmd.GetRoot().Command
	body1, body2 := genShellLoopCommands(cx, 1, sbca)
	// sb.WriteString(body1)
	// sb.WriteString(body2)
	logrus.Debugf("%v,%v", len(body1), len(body2))
	for i := 1; i <= lvl; i++ {
		sb.WriteString(fmt.Sprintf("\t\tlevel%d)\n\t\t\tcase $words[%d] in\n", i, i))
		sb.WriteString(sbca[i].String())
		sb.WriteString(fmt.Sprintf("\t\t\t\t*) _arguments '%d: :_files' ;;\n\t\t\tesac\n\t\t;;\n\n", i))
	}

	sb.WriteString(fmt.Sprintf(`
	esac
}


__ac "$@"


# Local Variables:
# mode: Shell-Script
# sh-indentation: 4
# indent-tabs-mode: nil
# sh-basic-offset: 4
# End:
# vim: ft=zsh sw=4 ts=4 et

`))

	err = ioutil.WriteFile("_"+cmd.GetRoot().AppName, []byte(sb.String()), 0644)
	if err == nil {
		logrus.Infof("_%v written.", cmd.GetRoot().AppName)
	}
	return
}

func genShellLoopCommands(cmd *Command, level int, sbca []strings.Builder) (scrFlg, scrCmd string) {
	var sbCmds, sbFlags strings.Builder

	sbca[level].WriteString(fmt.Sprintf("\t\t\t\t%v) _arguments '%d: :(%v)' ;;\n",
		cmd.GetName(), level, cmd.GetSubCommandNamesBy(" ")))

	for _, cc := range cmd.SubCommands {
		// sbCmds.WriteString(fmt.Sprintf(`%v:::`, cc.Name))

		// sbFlags.WriteString(fmt.Sprintf("\t\t\t\n"))

		// '(- *)'{--version,-V}'[display version info]' \
		// '(- *)'{--help,-h}'[display help]' \
		// '(--background -b)'{--background,-b}'[run in background]' \
		// 		if len(cc.Flags) > 0 {
		// 			for _, flg := range cc.Flags {
		// 				sbFlags.WriteString(fmt.Sprintf(`		'(%v)'{%v}'[%v]' \
		// `, eraseMultiWSs(flg.GetTitleFlagNamesBy(" ")), eraseMultiWSs(flg.GetTitleFlagNames()), flg.Description))
		// 			}
		// 		}

		if len(cc.SubCommands) > 0 {
			a, b := genShellLoopCommands(cc, level+1, sbca)
			// sbChild.WriteString(a)
			// sbca[level+1].WriteString(fmt.Sprintf("\t\tlevel%d)\n\t\t\tcase $words[%d] in\n", level+1, level+1))
			sbca[level+1].WriteString(a)
			// sbFlags.WriteString(fmt.Sprintf("\t\t\t\t*) _arguments '%d: :_files' ;;\n\t\t\tesac\n\t\t;;\n", level+1))
			logrus.Debugf("level %v \nflgs:\n%v\ncmds:\n%v", level, a, b)
		}
	}

	// sbFlags.WriteString(fmt.Sprintf("\t\tlevel%d)\n\t\t\tcase $words[%d] in\n", level+1, level+1))
	// sbFlags.WriteString(sbChild.String())
	// sbFlags.WriteString(fmt.Sprintf("\t\t\t\t*) _arguments '%d: :_files' ;;\n\t\t\tesac\n\t\t;;\n", level+1))

	if level == 0 {
		// 		scrFlg = fmt.Sprintf(`	_arguments -s -S \
		// %v && return 0
		//
		// `, sbFlags.String())
		// 		scrCmd = fmt.Sprintf(`	_alternative \
		// %v
		//
		// `, sbCmds.String())
	} else {
		scrFlg = sbFlags.String()
		scrCmd = sbCmds.String()
	}
	return
}
