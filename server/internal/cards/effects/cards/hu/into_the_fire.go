package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Into the Fire — Sorcery {2}{R} (March of the Machine):
//
//	"Choose one —
//	 • Into the Fire deals 2 damage to each creature, planeswalker, and
//	   battle.
//	 • Put any number of cards from your hand on the bottom of your
//	   library, then draw that many cards plus one."
//
// The first mode is a sweep: one damage event to every creature,
// planeswalker and battle (damage to a battle removes defense counters,
// CR 120.3h). The second is Valakut Awakening's whole text, whose
// helper it calls: you choose the cards, put them on the bottom in the
// order you choose, then draw that many plus one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7f735ce-e53b-40c0-b449-1f28fdfb2484",
		Name:         "Into the Fire",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Into the Fire deals 2 damage to each creature, planeswalker, and battle.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.Or(effects.Creature(), effects.Planeswalker(), effects.Battle()), 2)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}}),
			effects.ModeWithPurpose(effects.ModeDoing("Put any number of cards from your hand on the bottom of your library, then draw that many cards plus one.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.ValakutAwakeningPutThenDraw(ctx, item.Controller)
				}),
				game.Purpose{Draws: 1}),
		),
	})
}
