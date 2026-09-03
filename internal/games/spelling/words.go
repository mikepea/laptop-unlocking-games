package spelling

// Word is one spelling and a sentence that uses it.
//
// The sentence is not decoration. Half of spelling a word is knowing which
// word you are being asked for, and "Sound Alikes" is unanswerable without it:
// shown "weight" on its own there is no way to tell it from "wait". The
// sentence appears with the word picked out, and both vanish together.
type Word struct {
	Text string
	// Sentence contains Text exactly once. A test enforces that, because a
	// sentence that had drifted from its word would highlight nothing.
	Sentence string
}

// Level is a themed set of words. A round draws from the set at random without
// repeating, so playing the same level twice is not the same round twice.
type Level struct {
	Title string
	Hint  string
	// Words is the pool. It should hold comfortably more than Count.
	Words []Word
	// Count is how many words are asked in a round.
	Count int
	// MinAccuracy is the share that must be spelled right to pass.
	MinAccuracy float64
	// ShowFor is how long the word is visible before it is hidden, in
	// milliseconds. Longer words get longer, handled at round time, as does
	// the sentence that now comes with them.
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
		Words: []Word{
			{"because", "I stayed inside because it was raining."},
			{"friend", "My best friend lives on the next street."},
			{"people", "Too many people were waiting for the bus."},
			{"school", "We walk to school together every morning."},
			{"another", "He asked for another slice of toast."},
			{"today", "Today is the first day of the holidays."},
			{"always", "She always forgets her keys."},
			{"father", "His father taught him to ride a bike."},
			{"mother", "My mother grew up in Scotland."},
			{"little", "There is a little milk left in the bottle."},
			{"every", "He checks his pockets every time he leaves."},
			{"there", "Put the box down over there."},
			{"their", "The children collected their coats."},
			{"where", "Nobody knew where the cat had gone."},
			{"which", "Choose which one you want."},
			{"would", "I would help you if I had time."},
			{"could", "She could hear music through the wall."},
			{"should", "You should tell someone about it."},
			{"before", "Finish your homework before tea."},
			{"after", "We went to the park after lunch."},
		},
	},
	{
		Title:         "Double Trouble",
		Hint:          "Words with a doubled letter hiding in them.",
		Count:         10,
		MinAccuracy:   0.70,
		ShowForMillis: 2800,
		Words: []Word{
			{"different", "Every snowflake is different."},
			{"beginning", "The beginning of the book is slow."},
			{"address", "Write your address on the envelope."},
			{"happened", "Nobody saw what happened."},
			{"suddenly", "Suddenly the lights went out."},
			{"across", "The dog swam across the river."},
			{"success", "The concert was a huge success."},
			{"possible", "It is possible to finish this today."},
			{"difficult", "That last question was difficult."},
			{"arrive", "The train will arrive at six."},
			{"village", "The village has one shop and a postbox."},
			{"middle", "He sat in the middle of the row."},
			{"bottom", "The coin sank to the bottom of the pond."},
			{"swimming", "She goes swimming on Tuesdays."},
			{"running", "He came running down the hill."},
			{"better", "This one tastes better than the last."},
			{"letter", "A letter arrived for you this morning."},
			{"pretty", "The garden looked pretty in the spring."},
			{"rabbit", "A rabbit bolted into the hedge."},
			{"summer", "Last summer we camped by the sea."},
		},
	},
	{
		Title:         "Silent Letters",
		Hint:          "There is a letter in each of these you cannot hear.",
		Count:         10,
		MinAccuracy:   0.65,
		ShowForMillis: 2800,
		Words: []Word{
			{"knife", "Use a knife to cut the bread."},
			{"knock", "Knock twice and then wait."},
			{"know", "I know the answer to that one."},
			{"knee", "He grazed his knee on the path."},
			{"wrist", "The watch was loose on her wrist."},
			{"write", "Write your name at the top."},
			{"wrong", "That answer is wrong, but only just."},
			{"wrap", "Wrap the present in blue paper."},
			{"castle", "The castle stands above the town."},
			{"listen", "Listen carefully to the question."},
			{"island", "The island is an hour away by boat."},
			{"whistle", "The referee blew his whistle."},
			{"climb", "They will climb the hill at dawn."},
			{"thumb", "He hit his thumb with the hammer."},
			{"lamb", "A lamb followed its mother across the field."},
			{"science", "Science is his favourite lesson."},
			{"scissors", "Cut the card with scissors."},
			{"guess", "Have a guess before you look."},
			{"guard", "A guard stood at the gate."},
			{"honest", "Be honest about what happened."},
		},
	},
	{
		Title:         "Tricky Endings",
		Hint:          "Listen for the sound, then remember which ending it takes.",
		Count:         10,
		MinAccuracy:   0.65,
		ShowForMillis: 3000,
		Words: []Word{
			{"question", "Put your hand up if you have a question."},
			{"station", "The station was closed for repairs."},
			{"attention", "Pay attention to the road."},
			{"position", "He moved the sofa into a better position."},
			{"action", "The film is full of action."},
			{"decision", "That was a hard decision to make."},
			{"division", "We did long division in maths today."},
			{"television", "Turn the television off and come and eat."},
			{"occasion", "A birthday is a special occasion."},
			{"confusion", "There was some confusion about the time."},
			{"adventure", "The holiday turned into an adventure."},
			{"picture", "She drew a picture of the dog."},
			{"creature", "A strange creature lived in the cave."},
			{"furniture", "The furniture arrived in a van."},
			{"temperature", "The temperature dropped overnight."},
			{"special", "Today is a special day."},
			{"official", "The official rules are pinned to the wall."},
			{"musical", "He plays three musical instruments."},
			{"natural", "Honey is a natural sweetener."},
			{"central", "They live in central London."},
		},
	},
	{
		Title:         "Sound Alikes",
		Hint:          "These sound like other words. The sentence tells you which one.",
		Count:         10,
		MinAccuracy:   0.65,
		ShowForMillis: 2800,
		Words: []Word{
			{"weight", "The weight of the bag made his shoulder ache."},
			{"wait", "You will have to wait until the film starts."},
			{"break", "Be careful, or you will break the window."},
			{"brake", "She pulled the brake and the bike stopped."},
			{"piece", "He ate the last piece of cake."},
			{"peace", "After the argument they made peace."},
			{"whether", "Nobody knew whether the train would come."},
			{"weather", "The weather turned cold in October."},
			{"through", "The fox squeezed through a gap in the fence."},
			{"threw", "He threw the ball over the wall."},
			{"allowed", "We are not allowed to run in the corridor."},
			{"aloud", "Read the poem aloud so we can all hear."},
			{"great", "That was a great film."},
			{"grate", "Sparks fell through the grate of the fire."},
			{"heard", "I heard a noise upstairs."},
			{"herd", "A herd of cows blocked the lane."},
			{"night", "The owl hunts at night."},
			{"knight", "The knight wore heavy armour."},
			{"scene", "The last scene of the play was the best."},
			{"seen", "I have seen that trick before."},
		},
	},
	{
		Title:         "Long Words",
		Hint:          "Break each one into chunks and spell the chunks.",
		Count:         10,
		MinAccuracy:   0.60,
		ShowForMillis: 3500,
		Words: []Word{
			{"remember", "Try to remember where you left it."},
			{"important", "This is an important message."},
			{"difficult", "The puzzle was difficult to finish."},
			{"dangerous", "Thin ice is dangerous."},
			{"favourite", "Blue is her favourite colour."},
			{"beautiful", "The sunset over the hills was beautiful."},
			{"interesting", "That was an interesting idea."},
			{"immediately", "Come inside immediately."},
			{"particular", "He is very particular about his tools."},
			{"separate", "Keep the two piles separate."},
			{"necessary", "A coat is necessary today."},
			{"experience", "She has experience with horses."},
			{"experiment", "The experiment went wrong twice."},
			{"equipment", "The climbing equipment is in the shed."},
			{"government", "The government changed the law."},
			{"neighbour", "Our neighbour feeds the cat when we are away."},
			{"colourful", "The market was loud and colourful."},
			{"generous", "That was a generous gift."},
			{"ordinary", "It started as an ordinary Tuesday."},
			{"opposite", "The shop is opposite the library."},
		},
	},
}

// LevelUnlocked reports whether level i is open given how many are cleared.
func LevelUnlocked(i, cleared int) bool { return i <= cleared }
