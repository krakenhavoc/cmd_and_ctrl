package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Varragoth, Bloodsky Sire — Legendary Creature — Demon Rogue {2}{B}, 2/3:
//
//	"Deathtouch
//	 Boast — {1}{B}: Target player searches their library for a card, then shuffles
//	 and puts that card on top. (Activate only if this creature attacked this turn
//	 and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "68b41a04-8cb0-4edf-b488-a219494453ae",
		Name:            "Varragoth, Bloodsky Sire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []ActivatedAbility{
			BoastTargeting(
				"{1}{B}: Target player searches their library for a card, then shuffles and puts that card on top.",
				ManaCost("{1}{B}"), TargetPlayer("target player"), varragothSearch),
		},
	})
}

// varragothSearch is "Target player searches their library for a card,
// then shuffles and puts that card on top": the TARGET is the searcher
// (and chooses the card), not the boaster. A target that left the game
// has already countered the ability (CR 608.2b), so nothing happens.
func varragothSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return SearchLibrary{
			Player:  t.ID,
			Dest:    game.ZoneLibrary,
			ToTop:   true,
			Limit:   1,
			Shuffle: true,
			Reason:  "Varragoth, Bloodsky Sire — search your library for a card",
		}.Apply(ctx)
	}
	return nil
}
