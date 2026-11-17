package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tainted Observer — Creature — Phyrexian Bird {1}{G}{U}, 2/3 (Phyrexia: All
// Will Be One, S82):
//
//	"Flying
//	 Toxic 1
//	 Whenever another creature you control enters, you may pay {2}. If you
//	 do, proliferate."
//
// One trigger per creature entering under its controller's control; the {2}
// is MayPay's payment asked as the trigger resolves and the proliferate is
// the paid branch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "420aaa77-f429-4f3d-a73e-91dc25cb38db",
		Name:            "Tainted Observer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "toxic 1"},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Tainted Observer — you may pay {2}; if you do, proliferate",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Tainted Observer — pay {2} to proliferate?",
						Then:     s82OneB06ProliferatePaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
