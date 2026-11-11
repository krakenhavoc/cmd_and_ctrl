package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kopala, Warden of Waves — Legendary Creature — Merfolk Wizard {1}{U}{U}, 2/2:
//
//	"Spells your opponents cast that target a Merfolk you control cost {2}
//	 more to cast.
//	 Abilities your opponents activate that target a Merfolk you control
//	 cost {2} more to activate."
//
// Two cost increases (CR 601.2f, 602.2b) that read the announced targets
// (ReadsTargets), one scoped to casts and one to activations. Merfolk is
// read as an effective subtype on the target; Kopala itself is one.
func init() {
	spells := CostsMore(2,
		"Spells your opponents cast that target a Merfolk you control cost {2} more to cast.",
		OpponentsSpell(), cmdr2B05TargetsYourMerfolk())
	spells.ReadsTargets = true
	abilities := ActivationCostsMore(2,
		"Abilities your opponents activate that target a Merfolk you control cost {2} more to activate.",
		OpponentsSpell(), cmdr2B05TargetsYourMerfolk())
	abilities.ReadsTargets = true
	Register(Spec{
		OracleID:      "0237cccc-3161-4d91-8e69-e3d315dcb246",
		Name:          "Kopala, Warden of Waves",
		Completeness:  CompletenessFull,
		CostModifiers: []game.CostModifier{spells, abilities},
	})
}
