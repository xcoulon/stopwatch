package shell

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vatriathlon/stopwatch/pkg/model"
	"github.com/vatriathlon/stopwatch/pkg/service"

	"github.com/peterh/liner"
	"github.com/sirupsen/logrus"
)

var (
	history   = filepath.Join("./tmp/liner_example_history")
	cmds      = []string{"races", "start", "exit", "cancel", "time"}
	emptyRace = model.Race{ID: -1}
)

// Execute runs the "command liner"
func Execute(svc service.ApplicationService) {
	l := liner.NewLiner()
	defer l.Close()

	l.SetCtrlCAborts(false)
	l.SetCompleter(func(line string) (c []string) {
		for _, n := range cmds {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})
	if f, err := os.Open(history); err == nil {
		l.ReadHistory(f)
		f.Close()
	}

	fmt.Println("********************************************************")
	fmt.Println("type 'start' to begin the stopwatch for the current race")
	fmt.Println("********************************************************")

	var race = emptyRace
loop:
	for {
		prompt := ""
		if race.ID != -1 {
			d := time.Since(race.StartTime)
			d = d.Truncate(time.Second)
			prompt = fmt.Sprintf("%s ⏱ %s > ", race.Name, d.String())
		} else {
			prompt = "> "
		}
		cmd, err := l.Prompt(prompt)
		if err == liner.ErrPromptAborted {
			log.Print("Aborted")
		} else if err != nil {
			log.Print("Error reading line: ", err)
		}
		cmd = strings.TrimSpace(cmd)
		l.AppendHistory(cmd)

		switch {
		case cmd == "exit":
			break loop
		case cmd == "help":
			// TODO: print help?
		case cmd == "races":
			listRaces(os.Stdout, svc, cmd)
		case strings.HasPrefix(cmd, "start "):
			race, err = startRace(os.Stdout, svc, cmd)
		case strings.HasPrefix(cmd, "join "):
			race, err = joinRace(os.Stdout, svc, cmd)
		case cmd == "teams" && race.ID != -1:
			err = listTeams(os.Stdout, svc, race)
		}
		if err != nil {
			logrus.Error(err.Error())
		}
	}

	if f, err := os.Create(history); err != nil {
		log.Print("Error opening/creating history file: ", err)
	} else {
		l.WriteHistory(f)
		f.Close()
	}
	log.Printf("bye 👋")
}

func listRaces(out io.Writer, svc service.ApplicationService, cmd string) error {
	races, err := svc.ListRaces()
	if err != nil {
		return errors.Errorf("Failed to list races")
	}
	for _, r := range races {
		switch {
		case r.EndTimeStr() != "":
			fmt.Fprintf(out, "%d\t%s (ended at %s)\n", r.ID, r.Name, r.EndTimeStr())
		case r.StartTimeStr() != "":
			fmt.Fprintf(out, "%d\t%s (started at %s)\n", r.ID, r.Name, r.StartTimeStr())
		default:
			fmt.Fprintf(out, "%d\t%s\n", r.ID, r.Name)
		}
	}
	return nil
}

func startRace(out io.Writer, svc service.ApplicationService, cmd string) (model.Race, error) {
	raceIDStr := cmd[len("start "):]
	raceID, err := strconv.Atoi(raceIDStr)
	if err != nil {
		return emptyRace, errors.Errorf("Invalid race ID: '%s'", raceIDStr)
	}
	race, err := svc.StartRace(raceID)
	if err != nil {
		return emptyRace, errors.Errorf("Error while starting race '%s'", raceIDStr)
	}
	return race, nil
}
func joinRace(out io.Writer, svc service.ApplicationService, cmd string) (model.Race, error) {
	raceIDStr := cmd[len("join "):]
	id, err := strconv.Atoi(raceIDStr)
	if err != nil {
		return emptyRace, errors.Errorf("Invalid race ID: '%s'", raceIDStr)
	}
	race, err := svc.GetRace(id)
	if err != nil {
		return emptyRace, errors.Errorf("Error while joining race '%d'", raceIDStr)
	}
	return race, nil
}

func listTeams(out io.Writer, svc service.ApplicationService, race model.Race) error {
	teams, err := svc.ListTeams(race.ID)
	if err != nil {
		return errors.Errorf("Error while listing teams in race '%d'", race.ID)
	}
	for _, t := range teams {
		fmt.Fprintf(out, "%d\t%s\n", t.BibNumber, t.Name)
	}
	return nil
}
