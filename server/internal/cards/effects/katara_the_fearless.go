package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Katara, the Fearless — Legendary Creature — Human Warrior Ally
// {G}{W}{U}, 3/3:
//
//	"If a triggered ability of an Ally you control triggers, that
//	 ability triggers an additional time."
//
// Annie Joins Up's trigger doubler (CR 603.2d) over Allies you control,
// judged on the source as it was when the ability triggered, so an Ally
// entering beside Katara (or Katara herself) is covered (the ruling).
// Each instance is its own trigger: modes, targets and resolution
// choices are made separately. Replacement effects are untouched, so
// "enters with a counter" still gives one. Two Kataras are two doublers,
// one more trigger each, for three in all (the ruling).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	doubler := DoublesAbilitiesOf(HasSubtype("Ally"), DoublesAbilitiesOfOptions{})
	doubler.Label = "Katara, the Fearless"
	doubler.RowLabel = "Triggered abilities of your Allies trigger an additional time."
	Register(Spec{
		OracleID:        "0972d46e-423b-454e-87c7-a2d40fb6fb6d",
		Name:            "Katara, the Fearless",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
