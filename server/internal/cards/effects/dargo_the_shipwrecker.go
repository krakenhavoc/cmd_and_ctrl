package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dargo, the Shipwrecker — Legendary Creature — Giant Pirate {6}{R},
// 7/5:
//
//	"As an additional cost to cast this spell, you may sacrifice any
//	 number of artifacts and/or creatures. This spell costs {2} less to
//	 cast for each permanent sacrificed this way and {2} less to cast
//	 for each other artifact or creature you've sacrificed this turn.
//	 Trample
//	 Partner"
//
// The variable sacrifice (ADR 0100 §3) with two discounts, both read at
// CR 601.2f through the one pricer: the count this announcement names
// (CostQuery.Sacrificing), and the artifacts and creatures its caster
// already sacrificed this turn (PlayerTurnTally.
// ArtifactsOrCreaturesSacrificed). "Other" is the second count's word
// for "not the ones sacrificed for this spell", which the tally is by
// construction: the cost's own sacrifices happen at 601.2h, after the
// total is settled. Both discounts spend generic mana only, so Dargo
// never costs less than {R}.
//
// Trample is printed and read from the card. Partner is a deck-building
// rule the deck validator reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "a2414a2d-0fd8-4084-8762-ace48f70c853",
		Name:           "Dargo, the Shipwrecker",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAnyNumberCost("any number of artifacts and/or creatures", Or(Artifact(), Creature())),
		SelfCostModifiers: []game.CostModifier{
			CostsLessPerSacrificed("{2}", "This spell costs {2} less to cast for each permanent sacrificed this way"),
			CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return 2 * q.Game.TurnTallyFor(q.Controller).ArtifactsOrCreaturesSacrificed
			}, "This spell costs {2} less to cast for each other artifact or creature you've sacrificed this turn"),
		},
	})
}
