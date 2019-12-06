package shell

// func Execute() {
// 	l, err := readline.NewEx(&readline.Config{
// 		Prompt:              "\033[31m»\033[0m ",
// 		HistoryFile:         "/tmp/stopwatch-readline.tmp",
// 		AutoComplete:        completer,
// 		InterruptPrompt:     "^C",
// 		EOFPrompt:           "exit",
// 		HistorySearchFold:   true,
// 		FuncFilterInputRune: filterInput,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer l.Close()

// 	for {
// 		line, err := l.Readline()
// 		if err == readline.ErrInterrupt {
// 			if len(line) == 0 {
// 				break
// 			} else {
// 				continue
// 			}
// 		} else if err == io.EOF {
// 			break
// 		}

// 		line = strings.TrimSpace(line)

// 		if line == "help" {
// 			usage(l.Stderr())
// 		} else if line == "bye" {
// 			return
// 		} else if bib, err := strconv.Atoi(line); err != nil {
// 			io.WriteString(l.Stdout(), fmt.Sprintf("adding lap to %d", bib))
// 		} else {
// 			io.WriteString(l.Stderr(), "invalid command")
// 		}
// 	}
// }

// func usage(w io.Writer) {
// 	io.WriteString(w, "commands:\n")
// 	io.WriteString(w, completer.Tree("    "))
// }

// var completer = readline.NewPrefixCompleter(
// 	readline.PcItem("bye"),
// 	readline.PcItem("help"),
// 	// readline.PcItem("cancel"),
// )

// func filterInput(r rune) (rune, bool) {
// 	switch r {
// 	// block CtrlZ feature
// 	case readline.CharCtrlZ:
// 		return r, false
// 	}
// 	return r, true
// }
