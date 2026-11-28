package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reign of Terror — Sorcery {3}{B}{B} (Mirage):
//
//	"Destroy all green creatures or all white creatures. They can't be
//	 regenerated. You lose 2 life for each creature that died this way."
//
// Catastrophe's shape: not a modal spell, its controller chooses between the
// two as it resolves (CR 608.2d). The creatures are destroyed together
// ignoring regeneration (CR 701.19c), and the life loss counts the creatures
// that actually died (the mass tail's last-known copies), so an indestructible
// or a replaced one costs nothing.
//
// No simplification.
func init() {
	reignOfTerrorChosen = effects.OptionPickThen("option-pick/reign-of-terror-green-or-white", func(ctx *effects.Context, r game.OptionPicked) error {
		var colour string
		switch r.Index {
		case 0:
			colour = "G"
		case 1:
			colour = "W"
		default:
			return nil
		}
		return effects.DestroyAllMatching{
			Match:             effects.And(effects.Creature(), effects.OfColor(colour)),
			CantBeRegenerated: true,
			ThenTail:          ctx.ItemTail(reignOfTerrorLoss),
		}.Apply(ctx)
	})
	effects.Register(effects.Spec{
		OracleID:     "67d12a5a-9e0c-4550-86ad-cd1e173fea8f",
		Name:         "Reign of Terror",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PickOption{
				Question: "Reign of Terror — destroy all green creatures or all white creatures?",
				Options: []game.ChoiceOption{
					{Label: "Destroy all green creatures"},
					{Label: "Destroy all white creatures"},
				},
				ThenKey: reignOfTerrorChosen,
				Carry:   []uuid.UUID(nil),
			}.Apply(ctx)
		},
	})
}

// reignOfTerrorChosen is the controller's answer, keyed. Set in init, with the
// card. An on-disk key.
var reignOfTerrorChosen game.OptionPickThen

// reignOfTerrorLoss is "you lose 2 life for each creature that died this
// way", told how many were destroyed. An on-disk key.
var reignOfTerrorLoss = effects.MassTail("mass/reign-of-terror-life-loss", func(ctx *effects.Context, _ []game.Card, destroyed int) error {
	if destroyed <= 0 {
		return nil
	}
	return effects.YouLoseLife(2 * destroyed)(ctx)
})
