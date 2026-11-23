package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Velukan Dragon — Creature — Dragon {5}{R}{R}, 5/5 (Mystery Booster 2):
//
//	"Flying
//	 Whenever this creature attacks or blocks, roll a six-sided die. This
//	 creature gets +X/+0 until end of turn, where X is the result minus 1."
//
// One trigger for either event (Carrion Rats' shape: it fires once however many
// attackers it blocks, CR 509.3a). The roll goes through the engine's one dice
// door (game.RollDiceThenForEffect), so a replacement of the roll (Pixie
// Guide, a sticker) applies and may pause it; the bonus is read from the
// settled result, minus one, so a 1 gives +0/+0. A Dragon that has left by
// then gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "667aef37-a47a-4593-a252-807c02416863",
		Name:            "Velukan Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
				},
				"Velukan Dragon — roll a d6; it gets +X/+0 until end of turn, where X is the result minus 1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.RollDiceThenForEffect(game.DiceRoll{
						Player: item.Controller, Source: ctx.Source(), Sides: 6, N: 1,
						Then: ctx.ItemTail(effects.S82GapB23VelukanRolled),
					})
				}),
		},
	})
}
