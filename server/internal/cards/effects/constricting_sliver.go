package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const constrictingSliverGrant = "constricting-sliver/exile"

// Constricting Sliver — Creature — Sliver {5}{W}, 3/3 (Commander
// Masters, #5948):
//
//	"Sliver creatures you control have 'When this creature enters, you
//	 may exile target creature an opponent controls until this creature
//	 leaves the battlefield.'"
//
// Each Sliver (this one included) carries its own trigger, and the exile
// lasts only as long as that Sliver stays (CR 610.3). Two Slivers
// entering together each exile their own target and each holds it.
//
// No simplification.
func init() {
	const label = "Constricting Sliver — you may exile target creature an opponent controls until this creature leaves the battlefield"
	t := Targeting(WhenThisEnters(label,
		exileChosenTargetUntilThisLeaves("Constricting Sliver — the exiled creature returns when this creature leaves the battlefield")),
		TargetCreature("target creature an opponent controls", OpponentControls()))
	t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Exile target creature an opponent controls until this creature leaves the battlefield?"}
	Register(Spec{
		OracleID:     "fd26f893-6f45-44df-aa4e-c919fc118aff",
		Name:         "Constricting Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       constrictingSliverGrant,
			Triggered: []game.TriggeredAbility{t},
			Text:      "When this creature enters, you may exile target creature an opponent controls until this creature leaves the battlefield.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, constrictingSliverGrant),
		},
	})
}
