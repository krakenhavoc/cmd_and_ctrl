package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deepchannel Duelist — {W}{U} Creature — Merfolk Soldier 2/2 (Lorwyn
// Eclipsed, #4458):
//
//	"At the beginning of your end step, untap target Merfolk you control.
//	 Other Merfolk you control get +1/+1."
//
// The anthem is Merfolk Sovereign's (TribalAnthem over other Merfolk you
// control), so the Duelist does not pump itself and an opponent's Merfolk
// get nothing. The end-step untap is Voltaic Key's body on a trigger; the
// target is any Merfolk permanent you control, the Duelist included, chosen
// as the trigger goes on the stack. Merfolk is an effective subtype on both
// halves, so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c892d28c-4a78-4a7f-b8d5-c3663b819c5e",
		Name:         "Deepchannel Duelist",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Merfolk"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Deepchannel Duelist — untap target Merfolk you control", untapFirstLegalTarget),
				TargetPermanent("target Merfolk you control", HasSubtype("Merfolk"), YouControl())),
		},
	})
}
