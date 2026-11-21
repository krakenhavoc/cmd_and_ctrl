package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mutinous Massacre — Sorcery {3}{B}{B}{R}{R} (Edge of Eternities, #4461):
//
//	"Choose odd or even. Destroy each creature with mana value of the
//	 chosen quality. Then gain control of all creatures until end of turn.
//	 Untap them. They gain haste until end of turn. (Zero is even.)"
//
// Extinction Event's odd-or-even pick, then a sweep whose continuation is
// the rest of the spell: DestroyAllMatching hands its tail control, so
// the steal waits for a commander's replacement question, as the printed
// "Then" does. Mana value is read on the battlefield, so {X} is 0 and a
// token is 0 (even). The tail takes every creature then on the
// battlefield, the caster's own and the survivors of the destruction, by
// Act of Treason's layer-2 until-end-of-turn control, untaps all of them
// (the ones the caster already controls included) and gives them all haste
// until end of turn, each set fixed as it is taken (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fdd445a-c8e1-4309-9d74-374676fcac1d",
		Name:         "Mutinous Massacre",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PickOption{
				Question: "Mutinous Massacre — choose odd or even",
				Options:  []game.ChoiceOption{{Label: "Odd"}, {Label: "Even"}},
				ThenKey:  mutinousMassacreParity,
			}.Apply(ctx)
		},
	})
}

// mutinousMassacreParity destroys the chosen parity, then runs the steal.
var mutinousMassacreParity = effects.OptionPickThen("option-pick/mutinous-massacre-parity", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 {
		return nil
	}
	odd := r.Index == 0
	return effects.DestroyAllMatching{
		Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			if !c.IsCreature() {
				return false
			}
			mv, ok := g.ManaValueForEffect(c)
			return ok && (mv%2 != 0) == odd
		},
		ThenTail: ctx.Tail(mutinousMassacreSteal),
	}.Apply(ctx)
})

// mutinousMassacreSteal is "Then gain control of all creatures until end of
// turn. Untap them. They gain haste until end of turn." An on-disk key.
var mutinousMassacreSteal = effects.CardTail("mutinous-massacre/steal-everything", func(ctx *effects.Context, _ game.TailOutcome) error {
	creatures := effects.MatchingBattlefield(ctx, effects.Creature())
	for _, c := range creatures {
		if err := (effects.GainControl{
			Target:   c.InstanceID,
			Duration: effects.DurationUntilEndOfTurn(ctx),
			Label:    "Mutinous Massacre — gain control until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, c := range creatures {
		if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, c := range creatures {
		if err := (effects.GrantKeywordUntilEOT{
			Target:   c.InstanceID,
			Keywords: []string{"haste"},
			Label:    "Mutinous Massacre — haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
