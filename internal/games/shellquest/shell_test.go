package shellquest

import (
	"strings"
	"testing"
)

func testTree() *Node {
	return NewRoot(
		File("readme.txt", "line one\nline two"),
		File(".hidden", "shh"),
		Dir("attic",
			File("dust.txt", "dusty"),
			Dir("box", File("note.txt", "in the box")),
		),
		Dir("empty"),
	)
}

// out joins a command's output so a test can look for a substring in it.
func out(lines []string) string { return strings.Join(lines, "\n") }

func TestResolvePaths(t *testing.T) {
	root := testTree()
	attic, _ := root.child("attic")

	tests := []struct {
		name     string
		cwd      *Node
		path     string
		wantPath string
		wantErr  bool
	}{
		{"empty path is the working directory", root, "", "/", false},
		{"dot is the working directory", attic, ".", "/attic", false},
		{"child by name", root, "attic", "/attic", false},
		{"nested relative", root, "attic/box", "/attic/box", false},
		{"absolute from anywhere", attic, "/empty", "/empty", false},
		{"dot dot goes up", attic, "..", "/", false},
		{"dot dot at the root stays put", root, "..", "/", false},
		{"dot dot then down", attic, "../empty", "/empty", false},
		{"trailing slash is fine", root, "attic/", "/attic", false},
		{"a file resolves to itself", root, "readme.txt", "/readme.txt", false},
		{"missing name", root, "nowhere", "", true},
		{"through a file", root, "readme.txt/deeper", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolve(root, tc.cwd, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolve(%q) succeeded, want an error", tc.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve(%q): %v", tc.path, err)
			}
			if got.Path() != tc.wantPath {
				t.Fatalf("resolve(%q) = %q, want %q", tc.path, got.Path(), tc.wantPath)
			}
		})
	}
}

func TestLsHidesDotFilesUnlessAsked(t *testing.T) {
	sh := newShell(Quest{Title: "T", Secret: "banana", Par: 3, Build: testTree})

	plain := out(sh.ls(nil))
	if strings.Contains(plain, ".hidden") {
		t.Errorf("plain ls showed a dot file:\n%s", plain)
	}
	if !strings.Contains(plain, "readme.txt") {
		t.Errorf("plain ls did not show readme.txt:\n%s", plain)
	}

	all := out(sh.ls([]string{"-a"}))
	if !strings.Contains(all, ".hidden") {
		t.Errorf("ls -a did not show the dot file:\n%s", all)
	}
}

func TestLsOnAnEmptyDirectorySaysSo(t *testing.T) {
	sh := newShell(Quest{Title: "T", Secret: "banana", Par: 3, Build: testTree})
	if got := out(sh.ls([]string{"empty"})); !strings.Contains(got, "nothing here") {
		t.Fatalf("ls of an empty directory said %q", got)
	}
}

func TestCdMovesAndRefusesFiles(t *testing.T) {
	sh := newShell(Quest{Title: "T", Secret: "banana", Par: 3, Build: testTree})

	sh.run("cd attic")
	if sh.cwd.Path() != "/attic" {
		t.Fatalf("after cd attic, cwd = %q", sh.cwd.Path())
	}
	sh.run("cd box")
	if sh.cwd.Path() != "/attic/box" {
		t.Fatalf("after cd box, cwd = %q", sh.cwd.Path())
	}
	sh.run("cd ..")
	if sh.cwd.Path() != "/attic" {
		t.Fatalf("after cd .., cwd = %q", sh.cwd.Path())
	}

	got := out(sh.run("cd dust.txt"))
	if !strings.Contains(got, "Not a directory") {
		t.Fatalf("cd into a file said %q", got)
	}
	if sh.cwd.Path() != "/attic" {
		t.Fatalf("a failed cd moved the working directory to %q", sh.cwd.Path())
	}

	// Bare cd goes home, which here is the root.
	sh.run("cd")
	if sh.cwd.Path() != "/" {
		t.Fatalf("bare cd left cwd at %q", sh.cwd.Path())
	}
}

func TestCatReadsFilesAndRefusesDirectories(t *testing.T) {
	sh := newShell(Quest{Title: "T", Secret: "banana", Par: 3, Build: testTree})

	if got := out(sh.run("cat readme.txt")); !strings.Contains(got, "line two") {
		t.Fatalf("cat readme.txt gave %q", got)
	}
	if got := out(sh.run("cat attic")); !strings.Contains(got, "Is a directory") {
		t.Fatalf("cat of a directory gave %q", got)
	}
	if got := out(sh.run("cat nope.txt")); !strings.Contains(got, "No such file") {
		t.Fatalf("cat of a missing file gave %q", got)
	}
	if got := out(sh.run("cat")); !strings.Contains(got, "which file") {
		t.Fatalf("bare cat gave %q", got)
	}
}

func TestSayWinsOnlyWithTheRightWord(t *testing.T) {
	sh := newShell(Quest{Title: "T", Secret: "banana", Par: 3, Build: testTree})

	if got := out(sh.run("say apple")); !strings.Contains(got, "not the word") {
		t.Fatalf("a wrong word gave %q", got)
	}
	if sh.solved {
		t.Fatal("a wrong word solved the quest")
	}

	// Case does not matter: the word was read off a screen, not typed from
	// dictation.
	if got := out(sh.run("say BANANA")); !strings.Contains(got, "That's it") {
		t.Fatalf("the right word in capitals gave %q", got)
	}
	if !sh.solved {
		t.Fatal("the right word did not solve the quest")
	}
}

func TestUnknownCommandPointsAtHelp(t *testing.T) {
	sh := newShell(Quest{Title: "T", Secret: "banana", Par: 3, Build: testTree})
	got := out(sh.run("sudo rm -rf /"))
	if !strings.Contains(got, "command not found") {
		t.Fatalf("unknown command gave %q", got)
	}
	if !strings.Contains(got, "help") {
		t.Fatalf("unknown command did not mention help: %q", got)
	}
}

func TestBlankLinesAreNotCountedAsCommands(t *testing.T) {
	sh := newShell(Quest{Title: "T", Secret: "banana", Par: 3, Build: testTree})
	sh.run("")
	sh.run("   ")
	if sh.commands != 0 {
		t.Fatalf("commands = %d after two blank lines, want 0", sh.commands)
	}
	sh.run("pwd")
	if sh.commands != 1 {
		t.Fatalf("commands = %d after one real command, want 1", sh.commands)
	}
}

func TestPwdReportsTheWorkingDirectory(t *testing.T) {
	sh := newShell(Quest{Title: "T", Secret: "banana", Par: 3, Build: testTree})
	sh.run("cd attic/box")
	if got := out(sh.run("pwd")); got != "/attic/box" {
		t.Fatalf("pwd = %q, want /attic/box", got)
	}
}

func TestEfficiencyAndScoreRewardDirectness(t *testing.T) {
	q := Quest{Title: "T", Secret: "banana", Par: 4, Build: testTree}

	quick := newShell(q)
	quick.run("ls")
	quick.run("say banana")
	if got, want := quick.efficiency(), 1.0; got != want {
		t.Fatalf("efficiency under par = %v, want %v", got, want)
	}

	wandering := newShell(q)
	for i := 0; i < 12; i++ {
		wandering.run("ls")
	}
	wandering.run("say banana")
	if !(wandering.efficiency() < 1) {
		t.Fatalf("efficiency well over par = %v, want below 1", wandering.efficiency())
	}
	if !(quick.score() > wandering.score()) {
		t.Fatalf("quick scored %d, wandering scored %d; want quick higher", quick.score(), wandering.score())
	}
	if wandering.score() <= 0 {
		t.Fatalf("wandering scored %d; solving should always pay something", wandering.score())
	}
}

func TestUnsolvedQuestScoresNothingAndExplains(t *testing.T) {
	sh := newShell(Quest{Title: "The Test Quest", Secret: "banana", Par: 4, Build: testTree})
	sh.run("ls")

	if sh.score() != 0 {
		t.Fatalf("score = %d on an unsolved quest, want 0", sh.score())
	}
	notes := sh.notes()
	if len(notes) != 1 || !strings.Contains(notes[0], "The Test Quest") {
		t.Fatalf("notes = %v, want a line naming the quest", notes)
	}
}

// TestEverySecretIsActuallyFindable is the invariant that matters most: a quest
// whose secret is not written down anywhere is unwinnable, and a child would
// have no way to know that rather than assuming they were bad at it.
func TestEverySecretIsActuallyFindable(t *testing.T) {
	for _, q := range Quests {
		root := q.Build()
		if !mentions(root, q.Secret) {
			t.Errorf("quest %q hides the word %q in no file at all", q.Title, q.Secret)
		}
	}
}

// TestEveryQuestIsSolvableFromTheTree walks the whole tree reading everything,
// then says the word, and checks the quest completes. It proves the commands
// and the content agree.
func TestEveryQuestIsSolvableFromTheTree(t *testing.T) {
	for _, q := range Quests {
		sh := newShell(q)
		visitAll(sh.root, func(n *Node) {
			if !n.IsDir {
				sh.run("cat " + n.Path())
			}
		})
		sh.run("say " + q.Secret)
		if !sh.solved {
			t.Errorf("quest %q could not be solved by reading every file and saying the word", q.Title)
		}
	}
}

func TestCuratedQuestsAreWellFormed(t *testing.T) {
	for i, q := range Quests {
		if q.Title == "" {
			t.Errorf("quest %d has no title", i)
		}
		if q.Brief == "" || q.Hint == "" {
			t.Errorf("quest %q is missing a brief or a hint", q.Title)
		}
		if q.Secret == "" {
			t.Errorf("quest %q has no secret", q.Title)
		}
		if q.Par < 1 {
			t.Errorf("quest %q has par %d", q.Title, q.Par)
		}
		if q.Build == nil {
			t.Fatalf("quest %q has no filesystem", q.Title)
		}
		// Build must return a fresh tree each time, or replaying a quest would
		// start from wherever the last player left it.
		if a, b := q.Build(), q.Build(); a == b {
			t.Errorf("quest %q returns the same tree twice", q.Title)
		}
	}
}

func mentions(n *Node, word string) bool {
	if !n.IsDir && strings.Contains(strings.ToLower(n.Content), strings.ToLower(word)) {
		return true
	}
	for _, c := range n.Children {
		if mentions(c, word) {
			return true
		}
	}
	return false
}

func visitAll(n *Node, fn func(*Node)) {
	fn(n)
	for _, c := range n.Children {
		visitAll(c, fn)
	}
}
