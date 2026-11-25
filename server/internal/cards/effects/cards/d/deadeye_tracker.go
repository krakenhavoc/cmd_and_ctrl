package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadeye Tracker — Creature — Human Pirate {B}, 1/1 (Ixalan, #5532):
//
//	"{1}{B}, {T}: Exile two target cards from an opponent's graveyard. This
//	 creature explores."
//
// Exactly two targets, both from one opponent's graveyard (the clause's
// single-graveyard rule judges them at announce and again at resolution),
// so an opponent with fewer than two cards leaves the ability unactivatable.
// The two sentences are independent: the Tracker explores even if the
// exile did nothing, unless every target became illegal (CR 608.2b), when
// the ability does nothing at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26184baf-5d48-4919-8df2-c98ac460b78f",
		Name:         "Deadeye Tracker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Exile two target cards from an opponent's graveyard. This creature explores.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("two target cards from an opponent's graveyard", effects.Not(effects.YouOwn())).
				WithCount(2, 2).AllShare(effects.FromASingleGraveyard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				if err := effects.ExileTargetCards(g, item); err != nil {
					return err
				}
				return effects.Explores{Explorer: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
