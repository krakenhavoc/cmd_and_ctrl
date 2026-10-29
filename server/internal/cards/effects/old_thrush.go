package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Old Thrush — Creature — Bird {2}, 1/2 (The Hobbit, HOB):
//
//	"Flying
//	 When this creature enters, you gain 2 life. You may search your
//	 library for a basic land card, reveal it, then shuffle and put that
//	 card on top."
//
// The life is the first sentence of the trigger and is gained whatever the
// search finds. The search is Moon-Blessed Cleric's shape: optional (a
// player may always decline), the card is revealed, and ToTop puts it on
// top AFTER the shuffle, so it is a known card on an unknown library. A
// declined search is not a search, so nothing is shuffled (CR 701.19).
// Flying rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8259b0e-131f-44f0-86d4-7c4d711b328d",
		Name:            "Old Thrush",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Old Thrush — gain 2 life, then you may search for a basic land and put it on top",
				Do(GainLife{Amount: 2}, Instruction(oldThrushSearch))),
		},
	})
}

// oldThrushSearch is "You may search your library for a basic land card,
// reveal it, then shuffle and put that card on top."
func oldThrushSearch(ctx *Context) error {
	return SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: IsBasicLand,
		Dest:      game.ZoneLibrary,
		ToTop:     true,
		Limit:     1,
		Optional:  true,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Old Thrush — a basic land card",
	}.Apply(ctx)
}
