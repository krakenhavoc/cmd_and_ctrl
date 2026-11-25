package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scale Up — Sorcery {G} (Modern Horizons):
//
//	"Until end of turn, target creature you control becomes a green Wurm
//	 with base power and toughness 6/4.
//	 Overload {4}{G}{G} (You may cast this spell for its overload cost. If
//	 you do, change "target" in its text to "each.")"
//
// Polymorphist's Jest's type change: "becomes a green Wurm" replaces the
// creature's colours and creature types (CR 205.1a, layer 4), and its base
// power and toughness are set to 6/4 (layer 7b), so counters and pumps still
// apply on top. Its other card types, abilities and non-creature subtypes
// stay. The affected creatures are fixed as the spell resolves (CR 611.2c).
// Overloaded, every creature you control as it resolves is changed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fd025ef-14f5-4621-831d-dd12b7c604dd",
		Name:         "Scale Up",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerAnimate},
		AlternativeCosts: []game.AlternativeCost{
			effects.CostWithPurpose(effects.Overload("{4}{G}{G}"), game.Purpose{Answers: game.AnswerAnimate}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mods := []game.Mod{
				game.SetCreatureTypesMod(ctx.TextSubtype("Wurm")),
				game.SetColorsMod(ctx.TextColor("G")),
				game.SetBasePowerMod(6),
				game.SetBaseToughnessMod(4),
			}
			const label = "Scale Up — becomes a green Wurm with base power and toughness 6/4"
			if ctx.PaidAltCost("overload") {
				return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()), label, mods...)
			}
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, label, mods...)
		},
	})
}
