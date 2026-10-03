package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dovin, Hand of Control — Legendary Planeswalker — Dovin {2}{W/U},
// loyalty 5:
//
//	"Artifact, instant, and sorcery spells your opponents cast cost {1}
//	 more to cast.
//	 −1: Until your next turn, prevent all damage that would be dealt to
//	 and dealt by target permanent an opponent controls."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the static is Aura of Silence's
// cost increase, narrowed to an opponent's artifact, instant or sorcery
// spell. The −1 is one to-and-by record (Mod.AndDealtBy) on the target,
// all damage rather than combat damage, lasting until your next turn
// (CR 611.2b). A permanent that leaves and returns is a new object it
// no longer covers (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "95bbe04c-bcea-461d-9946-6c1a35dfbeb1",
		Name:            "Dovin, Hand of Control",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		CostModifiers: []game.CostModifier{
			CostsMore(1, "Artifact, instant, and sorcery spells your opponents cast cost {1} more to cast.",
				OpponentsSpell(), func(q game.CostQuery) bool {
					return q.Card.IsArtifact() || q.Card.IsInstant() || q.Card.IsSorcery()
				}),
		},
		Activated: []ActivatedAbility{
			untilYourNextTurnToAndByRow("−1: Until your next turn, prevent all damage that would be dealt to and dealt by target permanent an opponent controls.",
				LoyaltyCost(-1)),
		},
	})
}
