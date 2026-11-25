package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chamber Sentry — Artifact Creature — Construct {X}, 0/0 (Jumpstart, #5510):
//
//	"This creature enters with a +1/+1 counter on it for each color of mana
//	 spent to cast it.
//	 {X}, {T}, Remove X +1/+1 counters from this creature: It deals X damage
//	 to any target.
//	 {W}{U}{B}{R}{G}: Return this card from your graveyard to your hand."
//
// Wildgrowth Archaic's counters per color spent, Marath's {X} paid with the
// same number of counters removed, and Mother Bear's graveyard-only ability
// returning the card to hand instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "65f8e726-be50-49ee-96fc-d3a2080d7b8c",
		Name:                "Chamber Sentry",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		XMatters:            true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerColorSpent(game.CounterPlusOne),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{X}, {T}, Remove X +1/+1 counters from this creature: It deals X damage to any target.",
				Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
				Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost(), effects.RemoveCountersXFromThisSameX(game.CounterPlusOne)),
				Targets: effects.TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := ctx.X()
					if x <= 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{W}{U}{B}{R}{G}: Return this card from your graveyard to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{W}{U}{B}{R}{G}"),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  effects.ReturnThisCardFromYourGraveyardToHand,
			},
		},
	})
}
