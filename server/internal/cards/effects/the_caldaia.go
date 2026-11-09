package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Caldaia — Plane — Capenna (Commander Masters, #4870):
//
//	"Creature spells you cast from your hand have blitz {3}. (If you
//	 cast a spell for its blitz cost, it gains haste and "When this
//	 creature dies, draw a card." Sacrifice it at the beginning of the next
//	 end step.)
//	 Whenever chaos ensues, return target creature card from your graveyard
//	 to your hand."
//
// A plane (ADR 0216 §6). The static is a granted alternative cost (CR
// 118.9, ADR 0118 §3) that reaches the planar controller's creature spells
// cast from their hand: blitz {3} is blitz's offer with the planar
// controller as the caster ("you"), and the haste, the draw when it dies and
// the sacrifice at the next end step are blitz's own clause (game.
// AltCostKeyGrantedBlitz). It is an alternative cost, so it can't be
// combined with another one (CR 118.9a), and the spell is still cast, so it
// can be responded to and countered (the ruling). The chaos ability (CR
// 311.7) targets a creature card in the planar controller's graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                "86534917-5d51-4eba-9bf5-f9fc93c83c80",
		Name:                    "The Caldaia",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{BlitzForCreatureSpellsFromHand("{3}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("The Caldaia — return target creature card from your graveyard to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
		},
	})
}
