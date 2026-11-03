package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9c0f7cdc-9095-40ae-bf83-75a5ffda72b9",
		Name:         "Quandrix Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Return target creature or planeswalker to its owner's hand.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				BounceTheModesTarget),
			ModeDoing("Counter target artifact or enchantment spell.",
				TargetSpell("target artifact or enchantment spell", Or(Artifact(), Enchantment())),
				CounterTheModesTarget),
			ModeDoing("Put two +1/+1 counters on target creature.",
				TargetCreature("target creature"),
				PlusOneCountersOnTheModesTarget(2)),
			ModeDoing("Target player shuffles up to three target cards from their graveyard into their library.",
				Clauses(
					TargetPlayer("target player"),
					FromTargetPlayersGraveyard(
						TargetCardInGraveyard("up to three target cards from their graveyard").WithCount(0, 3), 0),
				),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					return TargetPlayerShufflesTheirCards(ctx, ctx.ModeTargets(occ), 0, 1)
				}),
		),
	})
}
