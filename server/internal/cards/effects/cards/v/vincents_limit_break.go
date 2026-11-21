package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vincent's Limit Break — Instant {1}{B} (Final Fantasy, #4462):
//
//	"Tiered (Choose one additional cost.)
//	 Until end of turn, target creature you control gains "When this
//	 creature dies, return it to the battlefield tapped under its owner's
//	 control" and has the chosen base power and toughness.
//	 • Galian Beast — {0} — 3/2.
//	 • Death Gigas — {1} — 5/2.
//	 • Hellmasker — {3} — 7/2."
//
// Tiered proof card shape (Cloud's Limit Break, CR 702.183a): the three
// bullets share one body, each with its own base power and toughness. The
// dies trigger is Malakir Rebirth's duration grant (the same registered
// bundle, so the same trigger text), and the base power and toughness is a
// layer 7b set until end of turn, which later pumps and counters still
// modify. Both are locked in as the spell resolves (CR 611.2c); a target
// that is no longer legal gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "814679b4-51bd-4b07-a7b8-d3375d36ff45",
		Name:         "Vincent's Limit Break",
		Completeness: effects.CompletenessFull,
		Modes: effects.Tiered(
			vincentsTier("Galian Beast — 3/2.", "{0}", 3, 2),
			vincentsTier("Death Gigas — 5/2.", "{1}", 5, 2),
			vincentsTier("Hellmasker — 7/2.", "{3}", 7, 2),
		),
	})
}

// vincentsTier is one bullet: the target creature you control gains the
// dies-and-returns trigger and has base power and toughness p/t until end
// of turn.
func vincentsTier(text, cost string, p, t int) game.ModeOption {
	return effects.ModeWithPurpose(effects.SpreeModeDoing(text, cost,
		effects.PermanentYouControl("target creature you control", effects.Creature()),
		func(_ *game.StackItem, ctx *effects.Context, occ int) error {
			target, ok := effects.ModeTarget(ctx, occ)
			if !ok {
				return nil
			}
			if err := (effects.GrantAbilitiesFor{
				Target: target.ID,
				Keys:   []string{effects.MalakirRebirthReturn},
				Label:  "Vincent's Limit Break — until end of turn, it returns when it dies",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntilEndOfTurn(ctx, target.ID, nil,
				"Vincent's Limit Break — base power and toughness "+text,
				game.SetBasePTMods(p, t)...)
		}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)))
}
