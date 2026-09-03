package shellquest

// Quest is one puzzle: a filesystem to explore and a secret word hidden in it.
type Quest struct {
	Title string
	// Brief is what the player is told when the quest starts.
	Brief string
	// Hint is the one-line nudge shown under the prompt.
	Hint string
	// Secret is the word that wins, found by reading the right file.
	Secret string
	// Par is the number of commands a confident player would need. Coming in
	// under it pays a bonus; going over never costs anything.
	Par int
	// Build makes a fresh filesystem. It is a function so replaying a quest
	// starts from a clean tree.
	Build func() *Node
}

// Quests is the curriculum. Each one introduces exactly one new idea.
var Quests = []Quest{
	{
		Title:  "Look Around",
		Brief:  "Somewhere here is a file with a secret word in it. Find it, read it, and say it.",
		Hint:   "ls lists what is here. cat reads a file. say tells me the word.",
		Secret: "porridge",
		Par:    4,
		Build: func() *Node {
			return NewRoot(
				File("hello.txt", "Hello! There is nothing useful in this file.\nKeep looking."),
				File("shopping.txt", "bread\nmilk\nsix eggs\napples"),
				File("secret.txt", "The secret word is: porridge\n\nNow type:  say porridge"),
				File("boring.txt", "This file is extremely boring. Sorry."),
			)
		},
	},
	{
		Title:  "Go Deeper",
		Brief:  "The word is in a file, but not in this directory. You will have to go in.",
		Hint:   "cd goes into a directory. cd .. comes back out. pwd says where you are.",
		Secret: "lighthouse",
		Par:    7,
		Build: func() *Node {
			return NewRoot(
				File("readme.txt", "Three doors. Only one of them goes anywhere.\nTry:  cd attic"),
				Dir("cellar",
					File("spiders.txt", "Just spiders down here. Go back up with:  cd .."),
				),
				Dir("garden",
					File("mud.txt", "Mud. A worm. No secrets."),
				),
				Dir("attic",
					File("dust.txt", "Dusty boxes."),
					Dir("box",
						File("note.txt", "The secret word is: lighthouse"),
					),
				),
			)
		},
	},
	{
		Title:  "Hidden Things",
		Brief:  "This one is hidden. Plain ls will not show it to you.",
		Hint:   "A name starting with a dot is hidden. ls -a shows everything.",
		Secret: "moonlight",
		Par:    6,
		Build: func() *Node {
			return NewRoot(
				File("nothing.txt", "Nothing here.\n\nSome files are hidden. Their names start with a dot.\nTry:  ls -a"),
				Dir("empty",
					File(".whisper", "Not this one. Look in the room next door."),
				),
				Dir("room",
					File(".secret", "The secret word is: moonlight"),
					File("lamp.txt", "A lamp, switched off."),
				),
			)
		},
	},
	{
		Title:  "Follow the Trail",
		Brief:  "Each note tells you where the next one is. Read carefully and keep going.",
		Hint:   "Everything you have learned, all at once. Read every note properly.",
		Secret: "thunderstorm",
		Par:    12,
		Build: func() *Node {
			return NewRoot(
				File("start.txt", "Begin in the library. Look for a book about birds."),
				Dir("library",
					File("birds.txt", "Page 41: the key is kept in the kitchen, in a jar."),
					File("maps.txt", "Old maps of nowhere in particular."),
				),
				Dir("kitchen",
					File("jar.txt", "Inside the jar: a note.\n'Under the stairs. The name starts with a dot.'"),
					File("bread.txt", "Half a loaf."),
				),
				Dir("stairs",
					File(".under", "The secret word is: thunderstorm"),
					File("shoes.txt", "Four shoes. None of them match."),
				),
			)
		},
	},
}

// QuestUnlocked reports whether quest i is open given how many are cleared.
func QuestUnlocked(i, cleared int) bool { return i <= cleared }
