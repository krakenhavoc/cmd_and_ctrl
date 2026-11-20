package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tip the Scales — Sorcery {2}{B} (Tarkir: Dragonstorm Commander):
//
//	"Sacrifice a creature. When you do, all creatures get -X/-X until end of
//	 turn, where X is the sacrificed creature's toughness."
//
// The sacrifice is chosen as the spell resolves, over what you control then
// (mustSacrificeOne: exactly one if you control a creature, the ruling), not
// as it is cast. "When you do" is a CR 603.12 reflexive trigger raised once a
// creature has really left (Bolg of the North's shape), so players may respond
// to it knowing X. X is the sacrificed creature's toughness as it last existed
// on the battlefield (CR 608.2h, the ruling), fixed as the reflexive trigger
// is created. The -X/-X reaches every creature on the battlefield when the
// trigger resolves, the sacrificed one is already gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7cac691-960f-4e36-8eae-f24bcdc1269b",
		Name:         "Tip the Scales",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.MustSacrificeOne(ctx.Game, item, "Tip the Scales — sacrifice a creature", effects.Creature(), tipTheScalesSacrificed)
		},
	})
}

// tipTheScalesSacrificed raises the reflexive trigger once a creature has
// really gone, carrying its last-known toughness as X.
var tipTheScalesSacrificed = effects.RegisterSacrificedOne("prompt-run/tip-the-scales-sacrifice", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	return effects.ReflexiveTrigger{
		Label:  "Tip the Scales — all creatures get -X/-X until end of turn, X the sacrificed creature's toughness",
		Cards:  []uuid.UUID{sacrificed},
		Body:   tipTheScalesBody,
		Params: game.EffectParams{Amount: effects.DepartedCreatureToughness(ctx.Game, sacrificed)},
	}.Apply(ctx)
})

var tipTheScalesBody = game.ReflexiveBody("tip-the-scales/minus", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	return effects.BoostUntilEOT{Match: effects.Creature(), Power: -p.Amount, Toughness: -p.Amount, Label: "Tip the Scales — -X/-X"}.Apply(effects.NewContext(g, item))
}, nil)
