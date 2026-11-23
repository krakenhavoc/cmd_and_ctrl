package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// threeBowlsOfPorridgeLabel is the activated ability's label, and with
// it the key its "hasn't been chosen" memory is kept under.
const threeBowlsOfPorridgeLabel = "{2}, {T}: Choose one that hasn't been chosen"

// Three Bowls of Porridge — Artifact — Food {2}:
//
//	"{2}, {T}: Choose one that hasn't been chosen —
//	 • This artifact deals 2 damage to target creature.
//	 • Tap target creature.
//	 • Sacrifice this artifact. You gain 3 life."
//
// An activated ability with ChooseOneNotChosen (ADR 0097): each bullet
// once for this object, announced with the activation (CR 602.2b) and
// recorded once it succeeds, so a used bullet is refused with nothing
// paid and the bot's enumerator never offers one. It is a Food by type
// only — it has no Food ability of its own, just the subtype, which
// "sacrifice a Food" costs can use.
//
// The damage comes from this artifact. "Sacrifice this artifact. You
// gain 3 life." gains the life whether or not the sacrifice happened,
// because the sentence is not "if you do"; a Bowls that left the
// battlefield in response is simply not sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5369e40a-5fd6-4a9d-bdc9-4af510449649",
		Name:         "Three Bowls of Porridge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   threeBowlsOfPorridgeLabel,
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Modes: effects.ChooseOneNotChosen(
				effects.ModeDoing("This artifact deals 2 damage to target creature.", effects.TargetCreature("target creature"),
					func(item *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}.Apply(ctx)
					}),
				effects.ModeDoing("Tap target creature.", effects.TargetCreature("target creature"), effects.TapTheModesTarget),
				effects.ModeDoing("Sacrifice this artifact. You gain 3 life.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						if effects.OnBattlefield(ctx.Game, item.SourceCardID) {
							if err := (effects.SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
								return err
							}
						}
						return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
					}),
			),
		}},
	})
}
