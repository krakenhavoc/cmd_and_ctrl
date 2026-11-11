package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Laurine, the Diversion — Legendary Creature — Human Rogue {2}{R}, 3/3
// (Outlaws of Thunder Junction Commander, #5368):
//
//	"Partner with Kamber, the Plunderer (When this creature enters,
//	 target player may put Kamber into their hand from their library,
//	 then shuffle.)
//	 First strike
//	 {2}, Sacrifice an artifact or creature: Goad target creature. (Until
//	 your next turn, that creature attacks each combat if able and
//	 attacks a player other than you if able.)"
//
// Bloodboil Sorcerer's goad ability at {2}: the sacrifice is a cost
// (CR 602.2b), chosen as the ability is activated, so the Laurine can pay
// with herself; GoadTarget goads the target if it is still legal and the
// goad ends at the start of your next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f2f8847d-17a8-4c26-8a7a-1bb6c02f2b39",
		Name:            "Laurine, the Diversion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Laurine, the Diversion", "Kamber, the Plunderer"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice an artifact or creature: Goad target creature.",
			Cost:    Plus(ManaCost("{2}"), SacrificeN(1, "an artifact or creature", Or(Artifact(), Creature()))),
			Targets: TargetCreature("target creature"),
			Effect:  GoadTarget,
		}},
	})
}
