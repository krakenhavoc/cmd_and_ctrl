package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ceaseless Searblades — Creature — Elemental Warrior {3}{R}, 2/4 (Lorwyn,
// S82 OLD b545, #5593):
//
//	"Whenever you activate an ability of an Elemental, this creature gets
//	 +1/+0 until end of turn."
//
// One trigger that watches both kinds of activation, since "an ability"
// includes a mana ability (CR 605, as Kai Winn reads it). The activator
// must be the Searblades' controller and the ability's source an Elemental
// (a changeling is one), read as it last existed when the cost sacrificed
// it (CR 608.2h). The Searblades' own abilities count: it is an Elemental.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbed284c-987a-4765-b3a8-4f18a228ad0e",
		Name:         "Ceaseless Searblades",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventActivateAbility, game.EventManaAbilityActivated},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.S82OldB545ActivatedAbilityOfSubtype(ev, source, g, "Elemental")
				},
				"Ceaseless Searblades — this creature gets +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Ceaseless Searblades — +1/+0"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
