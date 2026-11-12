package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindlink Mech — {2}{U} Artifact — Vehicle, 4/3 (Kamigawa: Neon
// Dynasty, #5671):
//
//	"Flying
//	 Whenever this Vehicle becomes crewed for the first time each turn,
//	 until end of turn, this Vehicle becomes a copy of target nonlegendary
//	 creature that crewed it this turn, except it's 4/3, it's a Vehicle
//	 artifact in addition to its other types, and it has flying.
//	 Crew 1"
//
// The trigger is WhenBecomesCrewedFirstTimeEachTurn (CR 702.122e): only
// the first crew ability of the Mech to resolve in a turn triggers it,
// whether or not that trigger found a target. The target clause is
// CrewedThis: a creature tapped to pay a crew cost of this Mech this
// turn, still on the battlefield as the same object, and not legendary.
// The copy is a duration copy (BecomeCopy, CR 707.2) whose exceptions
// are copiable values (CR 707.9b): 4/3, Artifact and Vehicle added to the
// copied types, and flying. The copy has neither the crew ability nor
// this trigger, which is what the card says; it stays a creature for the
// turn because the crew effect that animated it still applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "009a4073-5946-4131-928c-c5c373745344",
		Name:            "Mindlink Mech",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Mindlink Mech"),
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenBecomesCrewedFirstTimeEachTurn("Mindlink Mech — becomes a copy of target nonlegendary creature that crewed it this turn",
				mindlinkMechCopy),
				CrewedThis(TargetCreature("target nonlegendary creature that crewed it this turn", Not(Legendary())))),
		},
	})
}

func mindlinkMechCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return BecomeCopy{
		Targets: []uuid.UUID{ctx.Source()},
		Of:      targets[0].ID,
		Except: func(v *game.PrintedValues) {
			v.SetPowerToughness(4, 3)
			v.AddCardType("Artifact")
			v.AddSubtype("Vehicle")
			v.AddKeyword("flying")
		},
		Label: "Mindlink Mech — becomes a copy until end of turn",
	}.Apply(ctx)
}
