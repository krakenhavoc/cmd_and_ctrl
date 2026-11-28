package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crimson Roc — Creature — Bird {4}{R}, 2/2 (Mirage, S82 OLD b797, #5633):
//
//	"Flying
//	 Whenever this creature blocks a creature without flying, this creature
//	 gets +1/+0 and gains first strike until end of turn."
//
// One trigger per attacker blocked (CR 509.3b), judged against the attacker's
// flying as the block is declared (CR 509.3f), as Frostweb Spider reads it.
// The boost and the keyword are two layered effects that end at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "acc7d3b8-60f2-4f3b-9cf9-b6b39627d8f6",
		Name:            "Crimson Roc",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B28SelfBlocked(ev, source) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.Target)
				return ok && !game.HasKeyword(&attacker, "flying")
			}, "Crimson Roc — it gets +1/+0 and gains first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.ThisGetsUntilEndOfTurn(1, 0, "Crimson Roc — +1/+0 until end of turn")(g, item); err != nil {
						return err
					}
					return effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Crimson Roc — first strike until end of turn")(g, item)
				}),
		},
	})
}
