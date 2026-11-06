package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kyoshi Battle Fan — Artifact — Equipment {2} (Avatar: The Last
// Airbender, #4459):
//
//	"When this Equipment enters, create a 1/1 white Ally creature token,
//	 then attach this Equipment to it.
//	 Equipped creature gets +1/+0.
//	 Equip {2}"
//
// Dwarven Shortsword's enter trigger with an Ally (createTokenAndAttachSource,
// described): the token enters as a 1/1 and is then equipped (the
// ruling). If the Fan has left in response, the Ally is made and nothing
// is attached (CR 701.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "849e54f9-7662-4186-859c-e4f52afbccac",
		Name:         "Kyoshi Battle Fan",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kyoshi Battle Fan — create a 1/1 white Ally creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "1/1 white Ally", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
