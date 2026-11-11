package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// End-Raze Forerunners — Creature — Boar {5}{G}{G}{G}, 7/7 (Bloomburrow
// Commander, #5368):
//
//	"Vigilance, trample, haste
//	 When this creature enters, other creatures you control get +2/+2 and
//	 gain vigilance and trample until end of turn."
//
// Pyrewood Gearhulk's enters trigger: the set is fixed as the ability
// resolves (CR 611.2c), so a creature that arrives later is not pumped, and
// the Forerunners themselves are left out ("other").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69b52907-8aff-4f2a-a391-7d1ab2669b5c",
		Name:            "End-Raze Forerunners",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("End-Raze Forerunners — other creatures you control get +2/+2 and gain vigilance and trample until end of turn",
				cmdr2S12EndRazePump),
		},
	})
}
