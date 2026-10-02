package decks

// tutorial.go is the fixed deck pair the tutorial's practice table
// seats (ADR 0076 §2.2, #1078): one deck for the player and one for
// the `random`-tier practice bot.
//
// # Why they are not in `all`
//
// The four decks in `all` are what the deck pickers offer, human and
// bot alike. These two are not a choice anybody makes: the practice
// table picks them, every time, and a player browsing for a real deck
// should not find a deck of seventy-two Forests among them. So they
// live beside the registry rather than in it — Lookup, IDs and All do
// not see them, and GET /decks and GET /bot/options do not list them.
// decks_test.go runs every per-deck invariant over both sets (see
// everyDeck there), so being outside `all` exempts them from nothing.
//
// # What each deck is for
//
// The player's deck has to put the tutorial's lessons in the player's
// hand on the first three turns: a land to play, a cheap creature to
// cast, a permanent with an ability behind right-click, and a body to
// attack with. So it is half Forests, and the other half is creatures
// at one to three mana, most of which tap for mana — every mana
// creature is a right-click ability (ADR 0076 §2.1 step 7). Nothing in
// it asks the player a question when it resolves: no targets, no
// searches, no "you may". The tutorial teaches the client, not the
// rules, and a prompt it has not explained is a lesson it did not
// plan.
//
// The bot's deck is deliberately slow (ADR 0076 §2.2): lands and small
// bodies, no removal, no evasion. A practice bot that kills the player
// during the tutorial is a bug, and the decklist is the cheapest place
// to fix it — cheaper than a dedicated bot tier, which the ADR
// rejected. It is seventy-two Forests and twenty-seven creatures of
// power three or less, none of which flies, tramples, or targets
// anything. decks_test.go and the dump-gated realdump test hold it to
// that.
//
// Both are mono-green so the board reads as two piles of Forests, and
// both are catalog-only, every card graded CompletenessFull: the
// tutorial is a new player's first game, and the one place a card that
// plays differently from what it prints would cost the most.

// TutorialPlayerDeckID and TutorialBotDeckID are the stable IDs of the
// pair. They are recorded on the seat (SeatInfo.BotDeck for the bot)
// and never renamed, like every deck ID.
const (
	TutorialPlayerDeckID = "tutorial-player"
	TutorialBotDeckID    = "tutorial-bot"
)

// TutorialPlayer is the deck the practice table seats the player with.
func TutorialPlayer() Deck { return tutorialPlayer }

// TutorialBot is the deck the practice table seats its bot with.
func TutorialBot() Deck { return tutorialBot }

// tutorialDecks is the pair, for the tests that walk every deck.
func tutorialDecks() []Deck { return []Deck{tutorialPlayer, tutorialBot} }

var tutorialPlayer = Deck{
	ID:        TutorialPlayerDeckID,
	Name:      "First Steps",
	Archetype: "practice",
	Identity:  "G",
	Summary:   "The tutorial's deck: Forests and cheap green creatures, most of which tap for mana.",
	Commander: Card{Name: "Marwyn, the Nurturer", OracleID: "ee35de1c-aef1-4bd4-85fd-fe77bc927790", Identity: "G"},
	Mainboard: []Card{
		// One-drops. The mana creatures are the right-click lesson.
		{Name: "Llanowar Elves", OracleID: "68954295-54e3-4303-a6bc-fc4547a4e3a3", Identity: "G"},
		{Name: "Elvish Mystic", OracleID: "3f3b2c10-21f8-4e13-be83-4ef3fa36e123", Identity: "G"},
		{Name: "Fyndhorn Elves", OracleID: "df317532-7d36-40fd-938f-e972749c8792", Identity: "G"},
		{Name: "Boreal Druid", OracleID: "2fcc69ff-8ab5-4e14-afe3-db892049a872", Identity: "G"},
		{Name: "Birds of Paradise", OracleID: "d3a0b660-358c-41bd-9cd2-41fbf3491b1a", Identity: "G"},
		{Name: "Essence Warden", OracleID: "6ca2a89e-7032-4864-b4e9-66f3178f90ab", Identity: "G"},
		{Name: "Experiment One", OracleID: "8520ee67-439b-4f15-838b-420bacbb8b13", Identity: "G"},
		{Name: "Sazh's Chocobo", OracleID: "c2b082b4-5d9c-4125-b353-d1cf4f334720", Identity: "G"},
		{Name: "Virulent Emissary", OracleID: "3d4bec90-7bbc-4385-a2b8-303c7a8d0a0f", Identity: "G"},
		{Name: "Dragon Sniper", OracleID: "913dbce2-6805-453f-a9c3-64f1cc658627", Identity: "G"},
		{Name: "Gladecover Scout", OracleID: "385c1208-bfea-44e2-b236-4f38bc90db9f", Identity: "G"},
		{Name: "Memnite", OracleID: "7663ac7c-1de3-4250-b96a-fae9dbd66a27"},
		{Name: "Ornithopter", OracleID: "a3a98bc9-caa0-49b7-951c-fe4e4f54e4ba"},
		{Name: "Phyrexian Walker", OracleID: "7af75024-6c9b-4844-aeb6-81de25464822"},

		// Two-drops.
		{Name: "Ambush Viper", OracleID: "8957f7c2-040c-4048-9f21-efa7c97682b7", Identity: "G"},
		{Name: "Elvish Visionary", OracleID: "c6a3a882-a127-4590-93d7-679ef4313efe", Identity: "G"},
		{Name: "Gemhide Sliver", OracleID: "2c09ca09-8e62-4fe3-9b3d-61573dd2ffbc", Identity: "G"},
		{Name: "Gyre Sage", OracleID: "e3aa16cf-079b-4737-9ffd-7bfdffef0cb2", Identity: "G"},
		{Name: "Hedron Crawler", OracleID: "0b9a4e06-b21d-4cbe-906f-9dbd08dbe5d3"},
		{Name: "Manakin", OracleID: "d2343af0-468b-42bc-8a0c-347c10f7e2f3"},
		{Name: "Manaweft Sliver", OracleID: "bd47398d-da35-4a09-8754-771af91b14f4", Identity: "G"},
		{Name: "Nightshade Dryad", OracleID: "f8263274-8cb4-4eab-b541-c7655b555830", Identity: "G"},
		{Name: "Ornithopter of Paradise", OracleID: "3a940dfa-a026-4969-8981-ef90cdfbe9ef"},
		{Name: "Priest of Titania", OracleID: "3a198a16-17b9-481e-b516-5bc945c7e247", Identity: "G"},
		{Name: "Prosperous Innkeeper", OracleID: "da785227-cf8a-4d44-9e7c-fc909ea868f2", Identity: "G"},
		{Name: "Squirrel Sovereign", OracleID: "f43d1ea5-8127-4b64-a87b-ee5cc9b9e9fa", Identity: "G"},
		{Name: "Wall of Blossoms", OracleID: "ef4d5fb3-70a3-433d-a9d3-18b2beb8d79f", Identity: "G"},
		{Name: "Wall of Roots", OracleID: "3a21a6ae-b2f2-4f0c-acfd-5f3e8d63fd2f", Identity: "G"},
		{Name: "Crashing Drawbridge", OracleID: "328d5ef0-b25e-4f2a-80fe-35c6ae419e8a"},

		// Three and four.
		{Name: "Alloy Myr", OracleID: "efb0394c-2a45-4dd8-bca3-08704056fa31"},
		{Name: "Anurid Swarmsnapper", OracleID: "fd275ea7-9ccc-4148-9b33-392a612486dd", Identity: "G"},
		{Name: "Circuit Mender", OracleID: "1665ca9f-176d-40f1-a4e9-42da4f1236e9"},
		{Name: "Imperious Perfect", OracleID: "3fa71348-fa4d-4f39-a451-cf1570591991", Identity: "G"},
		{Name: "Llanowar Tribe", OracleID: "cfd0f0e6-1bbf-4450-97d2-c54907abb7e4", Identity: "G"},
		{Name: "Llanowar Visionary", OracleID: "f75ed312-3a23-4624-80c5-03980aa22d0b", Identity: "G"},
		{Name: "Managorger Hydra", OracleID: "b3f2265b-dd65-4b74-8b74-35ee0b147617", Identity: "G"},
		{Name: "Armorcraft Judge", OracleID: "d7f49243-a96e-499f-b2d7-8e9842432420", Identity: "G"},
		{Name: "Beast Whisperer", OracleID: "5da7eea8-bb9e-47ce-a554-8a1ee058bd7a", Identity: "G"},
		{Name: "Canopy Tactician", OracleID: "8b20d6f4-6322-4435-92d5-acaae74774f4", Identity: "G"},
		{Name: "Giant Spider", OracleID: "e740ce2f-2134-473c-afa1-1b6d2d1e38ef", Identity: "G"},
		{Name: "Ouroboroid", OracleID: "50d6fd91-23d3-4d32-804f-6233e4386904", Identity: "G"},

		// Top end, so a player who keeps going after the tutorial has
		// something to play for.
		{Name: "Sporemound", OracleID: "1be56a3d-a6c0-4b65-ae71-3d90ceefc6c0", Identity: "G"},
		{Name: "Watcher in the Web", OracleID: "232a4128-133e-446d-9f1c-dd3875094473", Identity: "G"},
		{Name: "Gigantosaurus", OracleID: "e666bae7-dd51-4921-8b89-7e8d423caba0", Identity: "G"},
		{Name: "Colossal Dreadmaw", OracleID: "08c7db90-c0cf-4482-b7ee-bb033e5996d2", Identity: "G"},
		{Name: "Rampaging Baloths", OracleID: "2d3e6549-6cc6-434f-a189-ba3b55e64c34", Identity: "G"},
		{Name: "Wolverine Riders", OracleID: "09070db6-01f6-4a8a-b167-a7825e6959f4", Identity: "G"},

		// Half the deck, so the first lesson (play a land) is in the
		// opening hand: fewer than one opening seven in a hundred has
		// no Forest.
		{Name: "Forest", Count: 52, Basic: true},
	},
}

var tutorialBot = Deck{
	ID:        TutorialBotDeckID,
	Name:      "Practice Partner",
	Archetype: "practice",
	Identity:  "G",
	Summary:   "The practice bot's deck: mostly Forests, and small green creatures with no evasion and no removal.",
	Commander: Card{Name: "Azusa, Lost but Seeking", OracleID: "6c2c8bf3-9bf8-4a86-89d3-3bb36260dc51", Identity: "G"},
	Mainboard: []Card{
		// Small bodies, power three at most. Nothing flies, tramples,
		// has deathtouch or first strike, and nothing targets.
		{Name: "Memnite", OracleID: "7663ac7c-1de3-4250-b96a-fae9dbd66a27"},
		{Name: "Phyrexian Walker", OracleID: "7af75024-6c9b-4844-aeb6-81de25464822"},
		{Name: "Boreal Druid", OracleID: "2fcc69ff-8ab5-4e14-afe3-db892049a872", Identity: "G"},
		{Name: "Elvish Mystic", OracleID: "3f3b2c10-21f8-4e13-be83-4ef3fa36e123", Identity: "G"},
		{Name: "Essence Warden", OracleID: "6ca2a89e-7032-4864-b4e9-66f3178f90ab", Identity: "G"},
		{Name: "Fyndhorn Elves", OracleID: "df317532-7d36-40fd-938f-e972749c8792", Identity: "G"},
		{Name: "Llanowar Elves", OracleID: "68954295-54e3-4303-a6bc-fc4547a4e3a3", Identity: "G"},
		{Name: "Elvish Visionary", OracleID: "c6a3a882-a127-4590-93d7-679ef4313efe", Identity: "G"},
		{Name: "Gemhide Sliver", OracleID: "2c09ca09-8e62-4fe3-9b3d-61573dd2ffbc", Identity: "G"},
		{Name: "Great Divide Guide", OracleID: "79e69a91-d580-47fb-be76-1e32c50d2fa0", Identity: "G"},
		{Name: "Hedron Crawler", OracleID: "0b9a4e06-b21d-4cbe-906f-9dbd08dbe5d3"},
		{Name: "Lotus Cobra", OracleID: "8ad91f64-ccab-4edc-bd54-b2ee9267d614", Identity: "G"},
		{Name: "Manakin", OracleID: "d2343af0-468b-42bc-8a0c-347c10f7e2f3"},
		{Name: "Manaweft Sliver", OracleID: "bd47398d-da35-4a09-8754-771af91b14f4", Identity: "G"},
		{Name: "Priest of Titania", OracleID: "3a198a16-17b9-481e-b516-5bc945c7e247", Identity: "G"},
		{Name: "Prosperous Innkeeper", OracleID: "da785227-cf8a-4d44-9e7c-fc909ea868f2", Identity: "G"},
		{Name: "Wall of Blossoms", OracleID: "ef4d5fb3-70a3-433d-a9d3-18b2beb8d79f", Identity: "G"},
		{Name: "Wall of Roots", OracleID: "3a21a6ae-b2f2-4f0c-acfd-5f3e8d63fd2f", Identity: "G"},
		{Name: "Alloy Myr", OracleID: "efb0394c-2a45-4dd8-bca3-08704056fa31"},
		{Name: "Anurid Swarmsnapper", OracleID: "fd275ea7-9ccc-4148-9b33-392a612486dd", Identity: "G"},
		{Name: "Circuit Mender", OracleID: "1665ca9f-176d-40f1-a4e9-42da4f1236e9"},
		{Name: "Llanowar Tribe", OracleID: "cfd0f0e6-1bbf-4450-97d2-c54907abb7e4", Identity: "G"},
		{Name: "Llanowar Visionary", OracleID: "f75ed312-3a23-4624-80c5-03980aa22d0b", Identity: "G"},
		{Name: "Beast Whisperer", OracleID: "5da7eea8-bb9e-47ce-a554-8a1ee058bd7a", Identity: "G"},
		{Name: "Giant Spider", OracleID: "e740ce2f-2134-473c-afa1-1b6d2d1e38ef", Identity: "G"},
		{Name: "Watcher in the Web", OracleID: "232a4128-133e-446d-9f1c-dd3875094473", Identity: "G"},
		{Name: "Sporemound", OracleID: "1be56a3d-a6c0-4b65-ae71-3d90ceefc6c0", Identity: "G"},

		// The rest is land. Azusa plays up to three a turn, which is
		// most of what this deck does.
		{Name: "Forest", Count: 72, Basic: true},
	},
}
