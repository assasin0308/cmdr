/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cmdr

import (
	"bytes"
	"fmt"
	"github.com/hedzr/cmdr/tool"
	"io"
	"strings"
	"time"
)

type (
	markdownPainter struct {
		writer io.Writer
		// buffer bufio.Writer
	}
)

func newMarkdownPainter() *markdownPainter {
	return &markdownPainter{
		writer: new(bytes.Buffer),
		// buffer:bufio.NewWriterSize(),
	}
}

func (s *markdownPainter) Results() (res []byte) {
	if bb, ok := s.writer.(*bytes.Buffer); ok {
		res = bb.Bytes()
	}
	return
}

func (s *markdownPainter) Reset() {
	s.writer = nil
	s.writer = new(bytes.Buffer)
}

func (s *markdownPainter) Flush() {
	// if bb, ok := s.writer.(*bytes.Buffer); ok {
	// 	_, _ = fmt.Fprintf(os.Stdout, "%v\n", bb.String())
	// }
}

func (s *markdownPainter) Printf(fmtStr string, args ...interface{}) {
	str := fmt.Sprintf(fmtStr, args...)
	str = replaceAll(str, "-", `\-`)
	str = replaceAll(str, "`cmdr`", `\fBcmdr\fP`)
	_, _ = s.writer.Write([]byte(str))
}

type mkdHdrData struct {
	RootCommand
	TimeMY      string
	ManExamples string
}

func (s *markdownPainter) FpPrintHeader(command *Command) {
	root := command.root
	a := &mkdHdrData{
		*root,
		time.Now().Format("Jan 2006"),
		// manExamples(root.Examples, root),
		fmt.Sprintf("\n```bash\n%v\n```\n", tplApply(root.Examples, root)),
	}

	s.Printf("%v", tplApply(`
## {{.AppName}} v{{.Version}}

{{.AppName}} v{{.Version}} - {{.Copyright}}

`, a))

	if command.IsRoot() {
		s.Printf("%v", tplApply(`

### SYNOPSIS

**{{.AppName}} generate manual [flags]**

### DESCRIPTIONS

{{.LongDescription}}

### EXAMPLES

{{.ManExamples}}

`, a))
	}
}

func (s *markdownPainter) FpPrintHelpTailLine(command *Command) {
	// root := command.root
	s.Printf(`
### SEE ALSO

%v

### HISTORY

[^1]: %v Auto generated by [hedzr/cmdr](https://github.com/hedzr/cmdr)

`,
		strings.Join(mkdSubCommands(command), "\n"),
		time.Now().Format("02-Jan-2006")) // , time.RFC822Z
}

func mkdSubCommands(command *Command) (ret []string) {
	for _, sc := range command.SubCommands {
		title := replaceAll(internalGetWorker().backtraceCmdNames(sc), ".", "-")
		// if len(title) == 0 {
		// 	title = command.root.AppName
		// } else {
		title = command.root.AppName + "-" + title
		// }
		var wrapChars, tail string
		if len(sc.Deprecated) > 0 {
			wrapChars = "~~"
			tail = fmt.Sprintf(" (deprecated since %v)", sc.Deprecated)
		}
		ret = append(ret, fmt.Sprintf("* [%s**%v**%s](%v.md) - *%v*%v", wrapChars, title, wrapChars, title, sc.Description, tail))
	}
	return
}

func (s *markdownPainter) FpUsagesTitle(command *Command, title string) {
	if !command.IsRoot() {
		s.Printf("\n### %s\n", "SYNOPSIS")
	}
	// s.Printf("\n\x1b[%dm\x1b[%dm%s\x1b[0m", bgNormal, darkColor, title)
	// fp("  [\x1b[%dm\x1b[%dm%s\x1b[0m]", bgDim, darkColor, normalize(group))
}

func (s *markdownPainter) FpUsagesLine(command *Command, fmt, appName, cmdList, cmdsTitle, tailPlaceHolder string) {
	if !command.IsRoot() {
		if len(tailPlaceHolder) > 0 {
			tailPlaceHolder = command.TailPlaceHolder
		} else {
			tailPlaceHolder = "[tail args...]"
		}
		s.Printf("```bash\n%s %v%s%s [Options] [Parent/Global Options]"+fmt+"\n```\n",
			appName, cmdList, cmdsTitle, tailPlaceHolder)
	}
}

func (s *markdownPainter) FpDescTitle(command *Command, title string) {
	if !command.IsRoot() {
		if len(command.LongDescription) > 0 {
			s.Printf("\n### %s\n", title)
		} else if len(command.Description) > 0 {
			s.Printf("\n### %s\n", title)
		}
	}
}

func (s *markdownPainter) FpDescLine(command *Command) {
	if !command.IsRoot() {
		if len(command.LongDescription) > 0 {
			s.Printf("\n%v\n", command.LongDescription)
		} else if len(command.Description) > 0 {
			s.Printf("\n%v\n", command.Description)
		}
	}
}

func (s *markdownPainter) FpExamplesTitle(command *Command, title string) {
	if !command.IsRoot() {
		if len(command.Examples) > 0 && command.HasParent() {
			s.Printf("\n### %s", title)
		}
	}
}

func (s *markdownPainter) FpExamplesLine(command *Command) {
	if !command.IsRoot() {
		if len(command.Examples) > 0 && command.HasParent() {
			s.Printf("\n```bash\n%v\n```\n", tplApply(command.Examples, command.root))
		}
	}
}

func (s *markdownPainter) FpCommandsTitle(command *Command) {
	var title string
	title = "SUB-COMMANDS"
	// if command.HasParent() {
	// 	title = "Commands"
	// } else {
	// 	title = "Sub-Commands"
	// }
	s.Printf("\n### %s\n", title)
}

func (s *markdownPainter) FpCommandsGroupTitle(group string) {
	if group != UnsortedGroup {
		// fp("  [%s]:", normalize(group))
		s.Printf("#### %s\n", tool.StripOrderPrefix(group))
	} else {
		s.Printf("#### %s\n", "General")
	}
}

func (s *markdownPainter) FpCommandsLine(command *Command) {
	if !command.Hidden {
		// s.Printf("  %-48s%v", command.GetTitleNames(), command.Description)
		// s.Printf("\n\x1b[%dm\x1b[%dm%s\x1b[0m", bgNormal, darkColor, title)
		// s.Printf("  [\x1b[%dm\x1b[%dm%s\x1b[0m]", bgDim, darkColor, normalize(group))
		// s.Printf(".TP\n.BI %s\n%s\n", manWs(command.GetTitleNames()), command.Description)
		title := command.Full
		if len(title) == 0 {
			title = command.Short
		}

		var wrapChars, tail string
		if len(command.Deprecated) > 0 {
			wrapChars = "~~"
			tail = fmt.Sprintf("> deprecated since %v", command.Deprecated)
		}

		s.Printf("##### %s%s%s", wrapChars, title, wrapChars)
		if len(command.Short) > 0 && len(command.Full) > 0 {
			s.Printf(" (**Short**: %v) ", command.Short)
		}
		if len(command.Aliases) > 0 {
			s.Printf(" (**Aliases**: %v) ", command.Aliases)
		}
		s.Printf("\n\n%v\n\n", tail)

		if len(command.Description) > 0 {
			s.Printf("%v\n\n", command.Description)
		}
		if len(command.LongDescription) > 0 {
			s.Printf("%v\n\n", command.LongDescription)
		}
		if len(command.Examples) > 0 {
			s.Printf("```bash\n%v\n```\n", tplApply(command.Examples, command.root))
		}
	}
}

func (s *markdownPainter) FpFlagsTitle(command *Command, flag *Flag, title string) {
	s.Printf("\n### %s\n", title)
}

func (s *markdownPainter) FpFlagsGroupTitle(group string) {
	if group != UnsortedGroup {
		s.Printf("#### %s\n", tool.StripOrderPrefix(group))
	} else {
		s.Printf("#### %s\n", "General")
	}
}

func (s *markdownPainter) FpFlagsLine(command *Command, flag *Flag, maxShort int, defValStr string) {
	// s.Printf(".TP\n.BI %s\n%s\n%s\n", manWs(flag.GetTitleFlagNames()), flag.Description, defValStr)

	title := "--" + flag.Full
	if len(title) == 2 {
		title = "-" + flag.Short
	}

	var wrapChars, tail string
	if len(flag.Deprecated) > 0 {
		wrapChars = "~~"
		tail = fmt.Sprintf("> deprecated since %v", flag.Deprecated)
	}

	s.Printf("##### %s%s%s %s ", wrapChars, title, wrapChars, flag.DefaultValuePlaceholder)
	if len(flag.Short) > 0 && len(flag.Full) > 0 {
		s.Printf(" (**Short**: -%v) ", flag.Short)
	}
	if len(flag.Aliases) > 0 {
		tt := strings.Join(flag.Aliases, ", --")
		if len(tt) > 0 {
			tt = "--" + tt
		}
		s.Printf(" (**Aliases**: %v) ", tt)
	}
	s.Printf("\n\n%v\n\n%v\n\n", defValStr, tail)

	if len(flag.Description) > 0 {
		s.Printf("%v\n\n", flag.Description)
	}
	if len(flag.LongDescription) > 0 {
		s.Printf("%v\n\n", flag.LongDescription)
	}
	if len(flag.Examples) > 0 {
		s.Printf("```bash\n%v\n```\n", tplApply(flag.Examples, command.root))
	}
}

//
//
//
