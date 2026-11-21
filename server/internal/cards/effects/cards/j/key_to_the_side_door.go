package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Key to the Side-Door — Artifact {1} (The Hobbit):
//
//	"{2}, {T}: Target creature can't be blocked this turn.
//	 {1}, {T}, Discard a legendary card with the same name as a legendary
//	 permanent you control: Draw two cards."
//
// The first ability is Rogue's Passage's. The second is a rummage whose
// discard is narrowed by the board: the card must be legendary (the
// predicate) and share a name with a legendary permanent the activator
// controls (game.DiscardCost.SameNameAsLegendaryPermanent, #4487), so the
// legend rule's usual casualty, a second copy of a legend, pays for two
// cards. The board half is read by the engine when the cost is offered
// and when it is paid, so the enumerator and the bot only see the
// ability while it is payable.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76c036df-44f5-4021-b7fc-656147351341",
		Name:         "Key to the Side-Door",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Target creature can't be blocked this turn.",
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					// CR 608.2b: a target that left in response is
					// skipped rather than errored.
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Key to the Side-Door — can't be blocked",
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{1}, {T}, Discard a legendary card with the same name as a legendary permanent you control: Draw two cards.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.DiscardLegendaryNamedLikeYourPermanent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.DrawNForEffect(item.Controller, 2)
				},
			},
		},
	})
}
