package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

var riddleGateGargoylePaid = effects.EnergyPaid("pay-unless/riddle-gate-gargoyle-lifelink",
	func(g *game.Game, item *game.StackItem) error {
		return effects.WhenYouDo("Riddle Gate Gargoyle — target creature you control gains lifelink",
			effects.RiddleGateLifelinkBody).Apply(effects.NewContext(g, item))
	})

// Riddle Gate Gargoyle — Artifact Creature — Gargoyle {W}{U}, 2/2:
//
//	"Flying
//	 When this creature enters, you get {E}{E}{E} (three energy counters).
//	 Whenever you attack, you may pay {E}{E}. When you do, target
//	 creature you control gains lifelink until end of turn."
//
// ADR 0129 §3 (#1995): "whenever you attack" is one trigger per attack
// declaration (OncePerBatch). The energy is paid as it resolves (CR
// 118.12), and "when you do" is a reflexive trigger (CR 603.12) whose
// target is chosen as it goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c5830515-9316-4b08-a990-3026ee15ffea",
		Name:            "Riddle Gate Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Riddle Gate Gargoyle", 3),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Riddle Gate Gargoyle — you may pay {E}{E}",
				effects.MayPayEnergyThen("Riddle Gate Gargoyle", 2, "give a creature you control lifelink", riddleGateGargoylePaid))),
		},
	})
}
