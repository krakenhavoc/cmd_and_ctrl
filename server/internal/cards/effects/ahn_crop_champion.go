package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ahn-Crop Champion — Creature — Human Warrior {2}{G}{W}, 4/4:
//
//	"You may exert this creature as it attacks. When you do, untap all
//	 other creatures you control. (An exerted creature won't untap
//	 during your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h). The untap is Combat Celebrant's, minus the extra
// combat: every OTHER creature its controller controls untaps, the
// attackers among them (they stay attacking, CR 506.4), and the Champion
// itself stays tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c03d073e-ea59-4873-952a-bcf2730f1115",
		Name:          "Ahn-Crop Champion",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			WhenExerted("Ahn-Crop Champion — untap all other creatures you control", untapAllOtherCreaturesYouControl),
		},
	})
}
