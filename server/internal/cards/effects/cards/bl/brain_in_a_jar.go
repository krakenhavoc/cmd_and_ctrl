package cardsbl

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brain in a Jar — Artifact {2} (Shadows over Innistrad, #5543):
//
//	"{1}, {T}: Put a charge counter on this artifact, then you may cast an
//	 instant or sorcery spell with mana value equal to the number of charge
//	 counters on this artifact from your hand without paying its mana cost.
//	 {3}, {T}, Remove X charge counters from this artifact: Scry X."
//
// The count is read after the counter lands, live while the artifact is the
// same object on the battlefield and as it last existed once it has left
// (CR 608.2h). The free cast is Baral's Expertise's offer with "mana value
// N or less" and "N or more" both set, so only an exact match is offered;
// the offer is the cast itself (CR 608.2g), at any timing. X in the second
// ability is the number removed as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "321dbd10-1d48-49fc-ba6a-1df241a53338",
		Name:         "Brain in a Jar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Put a charge counter on this artifact, then you may cast an instant or sorcery spell with mana value equal to the number of charge counters on this artifact from your hand without paying its mana cost.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Effect:  brainInAJarCharge,
			},
			{
				Label:   "{3}, {T}, Remove X charge counters from this artifact: Scry X.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.RemoveCountersXFromThis(game.CounterCharge, 0)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Scry{N: ctx.CountersRemoved()}.Apply(ctx)
				},
			},
		},
	})
}

var brainInAJarCast = effects.RegisterFreeHandCastOffer("choose-cards/s82-old-b292-brain-in-a-jar-cast", "Brain in a Jar")

func brainInAJarCharge(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.CounterOnThis{Kind: game.CounterCharge, N: 1}).Apply(ctx); err != nil {
		return err
	}
	n := effects.CountersOnThisAsItResolves(game.CounterCharge)(ctx)
	effects.OfferFreeHandCast(ctx, ctx.Controller(), brainInAJarCast,
		"an instant or sorcery spell with mana value "+strconv.Itoa(n),
		effects.FreeInstantOrSorcery().WithManaValueAtMost(n).WithManaValueAtLeast(n))
	return nil
}
