package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Tribute — Sorcery {4}{B}{B} (Commander 2017, #5533):
//
//	"Kicker—Tap an untapped Vampire you control. (You may tap a Vampire you
//	 control in addition to any other costs as you cast this spell.)
//	 Target opponent loses half their life, rounded up. If this spell was
//	 kicked, you gain life equal to the life lost this way."
//
// The kicker is New Blood's tap (KickerTapUntapped): a Vampire that came
// under your control this turn may pay it, since it is not the {T} symbol
// (CR 302.6). The loss is Rush of Dread's half, rounded up, read as the spell
// resolves. The life gained is what the opponent actually lost, read off the
// change in their life total, so a prevented or replaced loss gains less
// (CR 119.7). A target that is no longer legal makes the spell do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "aa2da79c-b36c-4120-8d7d-f5571e87e45e",
		Name:          "Blood Tribute",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.KickerTapUntapped("an untapped Vampire you control", effects.HasSubtype("Vampire"))},
		Targets:       effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			p := ctx.PlayerByID(t.ID)
			if p == nil || p.Life <= 0 {
				return nil
			}
			lose := -((p.Life + 1) / 2)
			if !ctx.WasKicked() {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, lose)
			}
			// The kicked state is read here, where the spell's paid costs
			// are in hand; the tail is told only the life that moved.
			return ctx.Game.ChangePlayerLifeKeyedForEffect(ctx.Source(), t.ID, lose, ctx.ItemTail(bloodTributeGain))
		},
	})
}

// bloodTributeGain is the kicked spell's "you gain life equal to the
// life lost this way", told the life that moved (negative when lost). An
// on-disk key: never renamed, never reused.
var bloodTributeGain = effects.CardTail("blood-tribute/gain-the-life-lost", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Amount >= 0 {
		return nil
	}
	return effects.GainLife{Amount: -o.Amount}.Apply(ctx)
})
