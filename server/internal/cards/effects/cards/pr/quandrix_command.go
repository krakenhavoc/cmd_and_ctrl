package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quandrix Command — Instant {1}{G}{U}:
//
//	"Choose two —
//	 • Return target creature or planeswalker to its owner's hand.
//	 • Counter target artifact or enchantment spell.
//	 • Put two +1/+1 counters on target creature.
//	 • Target player shuffles up to three target cards from their
//	   graveyard into their library."
//
// Prismari Command's "choose two" shape (ADR 0065 §3): each chosen bullet
// has its own target group and is re-checked against its own clauses on
// resolution (CR 608.2b), and the bullets run in printed order.
//
// The fourth bullet is two clauses, the second tied to the first (#4862):
// up to three cards, each of which must be in the graveyard of the player
// picked by the first (FromTargetPlayersGraveyard). The rulings: if the
// player is still legal they shuffle their library even when every picked
// card is gone, and if every target of the spell is illegal it does
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c0f7cdc-9095-40ae-bf83-75a5ffda72b9",
		Name:         "Quandrix Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Return target creature or planeswalker to its owner's hand.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				effects.BounceTheModesTarget),
			effects.ModeDoing("Counter target artifact or enchantment spell.",
				effects.TargetSpell("target artifact or enchantment spell", effects.Or(effects.Artifact(), effects.Enchantment())),
				effects.CounterTheModesTarget),
			effects.ModeDoing("Put two +1/+1 counters on target creature.",
				effects.TargetCreature("target creature"),
				effects.PlusOneCountersOnTheModesTarget(2)),
			effects.ModeDoing("Target player shuffles up to three target cards from their graveyard into their library.",
				effects.Clauses(
					effects.TargetPlayer("target player"),
					effects.FromTargetPlayersGraveyard(
						effects.TargetCardInGraveyard("up to three target cards from their graveyard").WithCount(0, 3), 0),
				),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					return effects.TargetPlayerShufflesTheirCards(ctx, ctx.ModeTargets(occ), 0, 1)
				}),
		),
	})
}
