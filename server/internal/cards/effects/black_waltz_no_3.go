package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Waltz No. 3 — Legendary Creature — Wizard {2}{B}{R}, 2/2
// (EDHREC rank 3042):
//
//	"Flying, deathtouch
//	 Whenever you cast a noncreature spell, Black Waltz No. 3 deals
//	 2 damage to each opponent."
//
// The Rakdos spellslinger's ping: every cantrip, every ritual, two
// to the table. Both keywords ride PrintedKeywords; the trigger is
// Firebrand Archer's condition (b10NoncreatureSpellCastByYou — the
// spell is read off the stack) with the pirates' damage-to-each-
// opponent body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c84c681-40c1-4ff3-86e6-48b4ab7491e5",
		Name:            "Black Waltz No. 3",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: b10NoncreatureSpellCastByYou,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Black Waltz No. 3 — 2 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						return damageToEachOpponent(g, item, 2)
					})
			},
		}},
	})
}
