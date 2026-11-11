package decks

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// standard.go is the curated Standard decks (ADR 0222 slice 1, #2352):
// two 60-card decks, so a person can sit down at a Standard table
// against a bot without owning a list.
//
// They are held to the Commander decks' bar and one more:
//
//   - every non-basic card is in the effect catalog and graded full
//     (standard_test.go), so neither a bot nor a new player is handed a
//     card that does nothing;
//   - the list is legal by the numbers offline (standard_test.go: sixty
//     cards, at most four of a name, CR 100.2a), and by the card pool
//     only against a real Scryfall dump: deck.Validate reads
//     legalities.standard, and this package has no dump in CI. Both
//     decks are built from Foundations cards alone, the set Wizards
//     keeps in Standard longest, and mono-coloured, so the mana is
//     basic lands. realdump_standard_manual_test.go is the check
//     against the dump; run it when the dump or the lists change.
//
// A deck that rots (a card rotates out or is banned) fails at Load,
// exactly where a player's upload would, and the pickers then refuse
// it; the fix is a new list here.

// constructed is the registry of every deck that is not a Commander
// deck, in picker order.
var constructed = []Deck{
	standardGoblinRush,
	standardOvergrowth,
}

// Goblin Rush: mono-red aggro. Cheap Goblins that hit from turn one,
// Krenko to double them, and burn to finish.
var standardGoblinRush = Deck{
	Format:    game.FormatStandard,
	ID:        "standard-goblin-rush",
	Name:      "Goblin Rush",
	Archetype: "aggro",
	Identity:  "R",
	Summary:   "Cheap Goblins from turn one, Krenko to double them, and burn for the last points of damage.",
	Mainboard: []Card{
		{Name: "Fanatical Firebrand", OracleID: "d36e11f1-6ab3-4273-8114-a8fbbe21c1c3", Identity: "R", Count: 4},
		{Name: "Courageous Goblin", OracleID: "c107e271-5741-4a84-bfc4-3bc322972d0d", Identity: "R", Count: 4},
		{Name: "Viashino Pyromancer", OracleID: "5e78e6e0-60f0-4c3c-b8dd-bd673f5152b8", Identity: "R", Count: 4},
		{Name: "Searslicer Goblin", OracleID: "86ec0b56-497f-481c-a03a-c4e64e8e7404", Identity: "R", Count: 3},
		{Name: "Raging Redcap", OracleID: "679841ad-f3ac-48e2-ba11-2cf6d78642b4", Identity: "R", Count: 4},
		{Name: "Goblin Boarders", OracleID: "4c64feae-95d7-426d-9e42-d3b0841d2054", Identity: "R", Count: 3},
		{Name: "Krenko, Mob Boss", OracleID: "68418069-f615-40ef-ae0d-764192acae00", Identity: "R", Count: 3},
		{Name: "Gorehorn Raider", OracleID: "41699d92-c209-4fcb-bbd6-685b3437d0c2", Identity: "R", Count: 2},
		{Name: "Shivan Dragon", OracleID: "711eea87-0fa3-46e0-a42b-fa5a86455f04", Identity: "R", Count: 2},

		{Name: "Burst Lightning", OracleID: "ac2086fe-98ee-4280-9c7c-c5c2d6548a8b", Identity: "R", Count: 4},
		{Name: "Boltwave", OracleID: "bb8cccc8-e56f-42fa-a0d7-fc889b5e1c7b", Identity: "R", Count: 4},
		{Name: "Goblin Surprise", OracleID: "79cd09e8-d738-47c0-9473-b6281bec6935", Identity: "R", Count: 2},

		{Name: "Mountain", Identity: "R", Count: 21, Basic: true},
	},
}

// Overgrowth: mono-green stompy. Mana creatures on turns one and two,
// then creatures bigger than anything the other side has, and pump
// spells to win the fights they pick.
var standardOvergrowth = Deck{
	Format:    game.FormatStandard,
	ID:        "standard-overgrowth",
	Name:      "Overgrowth",
	Archetype: "stompy",
	Identity:  "G",
	Summary:   "Mana Elves into creatures bigger than the other side's, with pump spells and Overrun to end it.",
	Mainboard: []Card{
		{Name: "Llanowar Elves", OracleID: "68954295-54e3-4303-a6bc-fc4547a4e3a3", Identity: "G", Count: 4},
		{Name: "Druid of the Cowl", OracleID: "b1793c3b-25d6-4fae-a99d-cfdd2210ca67", Identity: "G", Count: 4},
		{Name: "Dwynen's Elite", OracleID: "3d4a7dd3-9258-4bd1-adc4-08f0205e196a", Identity: "G", Count: 3},
		{Name: "Cackling Prowler", OracleID: "898da62e-d29a-4ddc-be69-2ddec54614f5", Identity: "G", Count: 4},
		{Name: "Beast-Kin Ranger", OracleID: "db794af5-ba2e-4b10-b213-e481e8a50432", Identity: "G", Count: 3},
		{Name: "Wildwood Scourge", OracleID: "b9dec104-c636-4770-a7fc-7a3331face15", Identity: "G", Count: 3},
		{Name: "Mossborn Hydra", OracleID: "84122d81-9634-4d3f-85a5-f6cf4303691a", Identity: "G", Count: 3},
		{Name: "Elfsworn Giant", OracleID: "f6abdf34-cfdc-465c-b067-cdebb659406e", Identity: "G", Count: 2},
		{Name: "Quakestrider Ceratops", OracleID: "e56629ba-82d4-470f-9a6b-40c29d835af2", Identity: "G", Count: 2},
		{Name: "Ghalta, Primal Hunger", OracleID: "b0b6be0c-41cf-4757-9f0e-87227b6ba6b3", Identity: "G", Count: 1},

		{Name: "Giant Growth", OracleID: "5748ebf1-24e3-499d-ab7c-c2cebd462a24", Identity: "G", Count: 4},
		{Name: "Garruk's Uprising", OracleID: "3127ae9b-a7a7-43ec-89d7-688f8445b33d", Identity: "G", Count: 3},
		{Name: "Overrun", OracleID: "204f9afe-c20b-4933-b5cd-aa572784762a", Identity: "G", Count: 2},

		{Name: "Forest", Identity: "G", Count: 22, Basic: true},
	},
}
