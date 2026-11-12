package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spotter Thopter — Artifact Creature — Thopter {8}, 4/5 (The Brothers'
// War, #5656): "Prototype {3}{U} — 2/3", "Flying", "When this creature
// enters, scry X, where X is its power."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Flying is a printed keyword. X is the Thopter's power as the trigger
// resolves (CR 608.2h): 2 for a prototyped Thopter, 4 for one cast
// normally, and its last-known power if it has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ccf13531-60bd-4ebd-9a5b-68e2d05ce620",
		Name:            "Spotter Thopter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spotter Thopter — scry X, where X is its power", spotterThopterScry),
		},
	})
}

// spotterThopterScry reads the Thopter's power as the trigger resolves:
// the permanent as it is now, or as it last existed if it has left. A
// power of zero or less scries nothing.
func spotterThopterScry(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return Scry{Player: ctx.Controller(), N: info.Power}.Apply(ctx)
}
