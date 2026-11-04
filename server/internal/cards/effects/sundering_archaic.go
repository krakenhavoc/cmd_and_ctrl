package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sundering Archaic — Creature — Avatar {6}, 3/3 (Secrets of Strixhaven,
// #4456):
//
//	"Converge — When this creature enters, exile target nonland permanent an
//	 opponent controls with mana value less than or equal to the number of
//	 colors of mana spent to cast this creature.
//	 {2}: Put target card from a graveyard on the bottom of its owner's
//	 library."
//
// Converge is an ability word (CR 207.2c). The colours of mana spent are the
// permanent's CR 400.7d record of the spell it came from (the cast
// provenance, stamped as the spell became the permanent), read as the
// trigger's target clause is built (CR 603.3d): only a nonland permanent an
// opponent controls whose mana value is at most that many is offered, and
// the clause is checked again as the trigger resolves (CR 608.2b). A
// Sundering Archaic that was not cast spent no mana, so it reaches only a
// permanent with mana value 0 (a token). With no legal target the trigger is
// removed from the stack. WantsDistinctColors pays the generic {6} with
// colours not yet spent when it can, and ReadsOwnPayment makes a waived
// payment ask what was paid.
//
// The activated ability is Grazing Kelpie's: any graveyard's card goes to the
// bottom of its owner's library.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Sundering Archaic — Converge: exile target nonland permanent an opponent controls with mana value at most the colors of mana spent",
		socB37ExileTheLegalTarget)
	enters.TargetsFrom = socB37SunderingClause
	Register(Spec{
		OracleID:            "b095526e-94a4-416b-83de-d6271804ccf3",
		Name:                "Sundering Archaic",
		Completeness:        CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Triggered:           []game.TriggeredAbility{enters},
		Activated: []ActivatedAbility{{
			Label:   "{2}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    ManaCost("{2}"),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  socB37PutTargetGraveyardCardOnBottom,
		}},
	})
}

// socB37SunderingClause is "target nonland permanent an opponent controls
// with mana value less than or equal to the number of colors of mana spent to
// cast this creature".
func socB37SunderingClause(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	colors := 0
	if source != nil {
		colors = g.CastProvenanceForEffect(source.InstanceID).Spent().ColorCount()
	}
	return TargetPermanent("target nonland permanent an opponent controls with mana value at most the colors of mana spent",
		Nonland(), OpponentControls(), ManaValueLE(colors))
}
