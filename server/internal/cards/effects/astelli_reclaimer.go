package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Astelli Reclaimer — Creature — Angel Warrior {3}{W}{W}, 5/4 (Edge of
// Eternities, #4461):
//
//	"Flying
//	 When this creature enters, return target noncreature, nonland
//	 permanent card with mana value X or less from your graveyard to the
//	 battlefield, where X is the amount of mana spent to cast this
//	 creature.
//	 Warp {2}{W}"
//
// X is the permanent's record of the spell it came from (the cast
// provenance, CR 400.7d), read as the trigger's target clause is built
// (CR 603.3d), as Sundering Archaic reads its colours: five for the
// printed cost, three for the warp cost, more with a tax or a raised
// cost. Only a noncreature, nonland permanent card you own in your
// graveyard with that mana value or less is offered, and the clause is
// checked again as the trigger resolves (CR 608.2b). A Reclaimer that
// was not cast spent no mana, so it reaches only a card with mana value
// 0. With no legal target the trigger is removed. ReadsOwnPayment makes
// a waived payment ask what was paid.
//
// The return is Carmen's (b26ReturnFirstLegalGraveyardTargetToBattlefield,
// Sun Titan's shape): the card comes back under its owner's control,
// and its own enters triggers fire. Warp is the shared constructor.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Astelli Reclaimer — return target noncreature, nonland permanent card with mana value X or less from your graveyard to the battlefield",
		b26ReturnFirstLegalGraveyardTargetToBattlefield)
	enters.TargetsFrom = eoeABReclaimerClause
	Register(Spec{
		OracleID:         "bb956492-e3ed-483a-9169-5b795a8e619e",
		Name:             "Astelli Reclaimer",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		ReadsManaSpent:   game.ReadsOwnPayment,
		AlternativeCosts: []game.AlternativeCost{Warp("{2}{W}")},
		Triggered:        []game.TriggeredAbility{enters},
	})
}

// eoeABReclaimerClause is "target noncreature, nonland permanent card
// with mana value X or less from your graveyard, where X is the amount
// of mana spent to cast this creature".
func eoeABReclaimerClause(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	spent := 0
	if source != nil {
		spent = g.CastProvenanceForEffect(source.InstanceID).Spent().Total()
	}
	return TargetCardInGraveyard("target noncreature, nonland permanent card with mana value X or less from your graveyard",
		YouOwn(), Permanent(), Noncreature(), Nonland(), ManaValueLE(spent))
}
