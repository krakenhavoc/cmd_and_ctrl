package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lotusguard Disciple — Creature — Bird Cleric {2}{W}, 2/2 (Aetherdrift,
// #4464):
//
//	"Flying
//	 When this creature enters, target creature or Vehicle gains lifelink
//	 and indestructible until end of turn."
//
// Fang Guardian's entry trigger with a keyword grant in place of the pump,
// and the target any creature or Vehicle on the battlefield, whoever
// controls it, the Disciple itself included. An uncrewed Vehicle keeps
// both keywords for the turn: indestructible works on any permanent
// (CR 702.12b), and lifelink once it deals damage as a creature. The
// target is re-checked as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1139ee31-cb6e-4ae6-a08b-252c46b2db97",
		Name:            "Lotusguard Disciple",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters(
				"Lotusguard Disciple — target creature or Vehicle gains lifelink and indestructible until end of turn",
				lotusguardDiscipleGrant),
				TargetPermanent("target creature or Vehicle", Or(Creature(), Subtype("Vehicle")))),
				ForTargets(BenefitUntilEOT(0, 0, 0, "lifelink", "indestructible"))),
		},
	})
}

// lotusguardDiscipleGrant gives the still-legal target lifelink and
// indestructible until end of turn.
func lotusguardDiscipleGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	return GrantKeywordUntilEOT{Target: target, Keywords: []string{"lifelink", "indestructible"},
		Label: "Lotusguard Disciple — lifelink and indestructible"}.Apply(ctx)
}
