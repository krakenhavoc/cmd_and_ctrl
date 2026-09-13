package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electrostatic Field — Creature — Wall {1}{R}, 0/4 (EDHREC rank
// 1593):
//
//	"Defender
//	 Whenever you cast an instant or sorcery spell, this creature
//	 deals 1 damage to each opponent."
//
// The spellslinger Wall. The trigger reads the spell off the stack,
// where its type line is intact, and the damage is from the Field
// itself — a Torbran or a prevention shield sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2fa94a07-c932-4f85-b6e0-a97d2b29eb52",
		Name:            "Electrostatic Field",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12InstantOrSorceryCastByYou(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Electrostatic Field — 1 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						return damageToEachOpponent(g, item, 1)
					})
			},
		}},
	})
}
