package spelling

// Level is a themed set of words. A round draws from the set at random without
// repeating, so playing the same level twice is not the same round twice.
type Level struct {
	Title string
	Hint  string
	// Words is the pool. It should hold comfortably more than Count.
	Words []string
	// Count is how many words are asked in a round.
	Count int
	// MinAccuracy is the share that must be spelled right to pass.
	MinAccuracy float64
	// ShowFor is how long the word is visible before it is hidden, in
	// milliseconds. Longer words get longer, handled at round time.
	ShowForMillis int
}

// Levels is the curriculum. Spellings are British: the laptop is in Britain.
var Levels = []Level{
	{
		Title:         "Words We Use A Lot",
		Hint:          "The ones that turn up in every single sentence.",
		Count:         10,
		MinAccuracy:   0.70,
		ShowForMillis: 2500,
		Words: []string{
			"because", "friend", "people", "school", "another",
			"today", "always", "father", "mother", "little",
			"every", "there", "their", "where", "which",
			"would", "could", "should", "before", "after",
		},
	},
	{
		Title:         "Double Trouble",
		Hint:          "Words with a doubled letter hiding in them.",
		Count:         10,
		MinAccuracy:   0.70,
		ShowForMillis: 2800,
		Words: []string{
			"different", "beginning", "address", "happened", "suddenly",
			"across", "success", "possible", "difficult", "arrive",
			"village", "middle", "bottom", "swimming", "running",
			"better", "letter", "pretty", "rabbit", "summer",
		},
	},
	{
		Title:         "Silent Letters",
		Hint:          "There is a letter in each of these you cannot hear.",
		Count:         10,
		MinAccuracy:   0.65,
		ShowForMillis: 2800,
		Words: []string{
			"knife", "knock", "know", "knee", "wrist",
			"write", "wrong", "wrap", "castle", "listen",
			"island", "whistle", "climb", "thumb", "lamb",
			"science", "scissors", "guess", "guard", "honest",
		},
	},
	{
		Title:         "Tricky Endings",
		Hint:          "Listen for the sound, then remember which ending it takes.",
		Count:         10,
		MinAccuracy:   0.65,
		ShowForMillis: 3000,
		Words: []string{
			"question", "station", "attention", "position", "action",
			"decision", "division", "television", "occasion", "confusion",
			"adventure", "picture", "creature", "furniture", "temperature",
			"special", "official", "musical", "natural", "central",
		},
	},
	{
		Title:         "Sound Alikes",
		Hint:          "These sound like other words. Spell the one you are shown.",
		Count:         10,
		MinAccuracy:   0.65,
		ShowForMillis: 2800,
		Words: []string{
			"weight", "wait", "break", "brake", "piece",
			"peace", "whether", "weather", "through", "threw",
			"allowed", "aloud", "great", "grate", "heard",
			"herd", "night", "knight", "scene", "seen",
		},
	},
	{
		Title:         "Long Words",
		Hint:          "Break each one into chunks and spell the chunks.",
		Count:         10,
		MinAccuracy:   0.60,
		ShowForMillis: 3500,
		Words: []string{
			"remember", "important", "difficult", "dangerous", "favourite",
			"beautiful", "interesting", "immediately", "particular", "separate",
			"necessary", "experience", "experiment", "equipment", "government",
			"neighbour", "colourful", "generous", "ordinary", "opposite",
		},
	},
}

// LevelUnlocked reports whether level i is open given how many are cleared.
func LevelUnlocked(i, cleared int) bool { return i <= cleared }
