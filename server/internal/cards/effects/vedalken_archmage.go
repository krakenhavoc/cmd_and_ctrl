package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vedalken Archmage — Creature — Vedalken Wizard {2}{U}{U}, 0/2
// (EDHREC rank 3113):
//
//	"Whenever you cast an artifact spell, draw a card."
//
// The artifact deck's Beast Whisperer. Sai's condition
// (b29ArtifactSpellCastByYou — the spell is read off the stack,
// where its type line is intact) and a draw of one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "568cf486-0261-4634-ac36-a6507101b2d0",
		Name:         "Vedalken Archmage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b29ArtifactSpellCastByYou(ev, source, g)
			}, "Vedalken Archmage — draw a card", b27DrawOne),
		},
	})
}
