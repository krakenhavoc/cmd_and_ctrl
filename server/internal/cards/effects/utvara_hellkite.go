package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Utvara Hellkite — Creature — Dragon {6}{R}{R}, 6/6 (EDHREC rank
// 1533):
//
//	"Flying
//	 Whenever a Dragon you control attacks, create a 6/6 red Dragon
//	 creature token with flying."
//
// The Dragon that makes more Dragons. "A Dragon you control", not
// "another", so the Hellkite's own attack triggers it; the trigger is
// per Dragon, so an alpha strike with three Dragons makes three
// tokens — which is the printed outcome, not a batching gap, because
// the printed text is not "one or more". The tokens arrive after the
// attack is declared and are not attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "05a6a571-643e-429e-8e5c-1c3f8b0dc746",
		Name:            "Utvara Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b14DragonYouControlAttacked(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Utvara Hellkite — create a 6/6 red Dragon with flying",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: b14RedDragonToken(6), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
