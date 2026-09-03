package shellquest

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mikepea/laptop-unlocking-games/internal/ui"
)

// shell runs commands against a quest's filesystem. It holds no terminal state
// beyond a scrollback of lines, so the whole command surface is testable by
// calling run and reading what comes back.
type shell struct {
	quest Quest
	root  *Node
	cwd   *Node

	// commands counts everything the player entered that was a real command,
	// which is what the score is measured against.
	commands int
	solved   bool

	started   bool
	startedAt time.Time
	endedAt   time.Time

	now func() time.Time
}

func newShell(q Quest) *shell {
	root := q.Build()
	return &shell{quest: q, root: root, cwd: root, now: time.Now}
}

func (s *shell) start() {
	if !s.started {
		s.started = true
		s.startedAt = s.now()
	}
}

func (s *shell) elapsed() time.Duration {
	if !s.started {
		return 0
	}
	end := s.endedAt
	if end.IsZero() {
		end = s.now()
	}
	return end.Sub(s.startedAt)
}

// prompt is the string shown before what the player types.
func (s *shell) prompt() string { return s.cwd.Path() + " $ " }

// run executes one line and returns the output. An empty line does nothing and
// is not counted against the player.
func (s *shell) run(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	s.start()
	s.commands++

	fields := strings.Fields(line)
	cmd, args := fields[0], fields[1:]

	switch cmd {
	case "help":
		return s.help()
	case "pwd":
		return []string{s.cwd.Path()}
	case "ls":
		return s.ls(args)
	case "cd":
		return s.cd(args)
	case "cat":
		return s.cat(args)
	case "say":
		return s.say(args)
	case "clear":
		return nil
	case "exit", "quit":
		return []string{ui.Dim.Render("Press esc to leave the quest.")}
	default:
		return []string{ui.Bad.Render(cmd + ": command not found"),
			ui.Dim.Render("Type help to see what you can do.")}
	}
}

func (s *shell) help() []string {
	return []string{
		ui.Selected.Render("pwd") + ui.Dim.Render("           where am I?"),
		ui.Selected.Render("ls") + ui.Dim.Render("            what is here?"),
		ui.Selected.Render("ls -a") + ui.Dim.Render("         what is here, including hidden things"),
		ui.Selected.Render("cd <name>") + ui.Dim.Render("     go into a directory"),
		ui.Selected.Render("cd ..") + ui.Dim.Render("         go back out"),
		ui.Selected.Render("cat <name>") + ui.Dim.Render("    read a file"),
		ui.Selected.Render("say <word>") + ui.Dim.Render("    say the secret word once you find it"),
		ui.Selected.Render("clear") + ui.Dim.Render("         tidy the screen"),
	}
}

// ls lists a directory, directories first and each group in name order, the
// way a real ls does.
func (s *shell) ls(args []string) []string {
	all := false
	var path string
	for _, a := range args {
		if a == "-a" || a == "-la" || a == "-al" {
			all = true
			continue
		}
		path = a
	}

	target, err := resolve(s.root, s.cwd, path)
	if err != nil {
		return []string{ui.Bad.Render("ls: " + err.Error())}
	}
	if !target.IsDir {
		return []string{target.Name}
	}

	var names []string
	for _, c := range target.Children {
		if c.Hidden() && !all {
			continue
		}
		if c.IsDir {
			names = append(names, ui.Selected.Render(c.Name+"/"))
			continue
		}
		names = append(names, c.Name)
	}
	if len(names) == 0 {
		return []string{ui.Dim.Render("(nothing here)")}
	}
	sort.Strings(names)
	return []string{strings.Join(names, "   ")}
}

func (s *shell) cd(args []string) []string {
	if len(args) == 0 {
		s.cwd = s.root
		return nil
	}
	target, err := resolve(s.root, s.cwd, args[0])
	if err != nil {
		return []string{ui.Bad.Render("cd: " + err.Error())}
	}
	if !target.IsDir {
		return []string{ui.Bad.Render("cd: " + target.Name + ": Not a directory")}
	}
	s.cwd = target
	return nil
}

func (s *shell) cat(args []string) []string {
	if len(args) == 0 {
		return []string{ui.Dim.Render("cat: which file? Try:  cat something.txt")}
	}
	target, err := resolve(s.root, s.cwd, args[0])
	if err != nil {
		return []string{ui.Bad.Render("cat: " + err.Error())}
	}
	if target.IsDir {
		return []string{ui.Bad.Render("cat: " + target.Name + ": Is a directory")}
	}
	return strings.Split(target.Content, "\n")
}

// say is the win condition. It is a made-up command, and the only one here that
// does not exist on the real machine.
func (s *shell) say(args []string) []string {
	if len(args) == 0 {
		return []string{ui.Dim.Render("say: say what? Find the word first.")}
	}
	if !strings.EqualFold(args[0], s.quest.Secret) {
		return []string{ui.Bad.Render("That is not the word. Keep looking.")}
	}
	s.solved = true
	s.endedAt = s.now()
	return []string{ui.Good.Render("That's it! The way opens.")}
}

// efficiency is how directly the quest was solved, 0..1, measured against the
// quest's par. Solving it in par or fewer commands scores 1.
func (s *shell) efficiency() float64 {
	if !s.solved || s.commands == 0 {
		return 0
	}
	if s.commands <= s.quest.Par {
		return 1
	}
	return float64(s.quest.Par) / float64(s.commands)
}

func (s *shell) score() int { return scoreOf(s.quest, s.commands, s.solved) }

func (s *shell) notes() []string {
	if s.solved {
		return nil
	}
	return []string{fmt.Sprintf("The word was hidden somewhere in %s.", s.quest.Title)}
}

// scoreOf turns a finished quest into points.
//
// Wandering around is not punished — exploring is the skill being built — but
// solving it close to par pays a real bonus.
func scoreOf(q Quest, commands int, solved bool) int {
	if !solved {
		return 0
	}
	points := 40 + 10*q.Par/2
	if spare := q.Par - commands; spare > 0 {
		points += 8 * spare
	}
	return points
}
