package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amalia Benavides Aguirre — Legendary Creature — Vampire Scout {W}{B}, 2/2
// (The Lost Caverns of Ixalan, #4470):
//
//	"Ward—Pay 3 life.
//	 Whenever you gain life, Amalia Benavides Aguirre explores. Then
//	 destroy all other creatures if its power is exactly 20."
//
// Two steps (CR 608.2c): Amalia explores, and then, once the explore's
// graveyard question has been answered, the power is read and every other
// creature destroyed if it is exactly 20. Her power is read as she is on
// the battlefield after the counter (or as she last existed if she left,
// CR 608.2h), through the layers, so a pump counts. One trigger per life
// gain event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85eefda9-4aab-467d-a08c-be026cb8ff42",
		Name:         "Amalia Benavides Aguirre",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(3), "Amalia Benavides Aguirre — ward, pay 3 life"),
			WheneverYouGainLife("Amalia Benavides Aguirre — it explores, then destroy all other creatures if its power is exactly 20",
				Do(Instruction(func(ctx *Context) error {
					return Explores{Explorer: ctx.Source()}.Apply(ctx)
				}), Instruction(func(ctx *Context) error {
					info, ok := ctx.SourcePermanent()
					if !ok || info.Power != 20 {
						return nil
					}
					me := ctx.Source()
					return DestroyAllMatching{Match: And(Creature(), OtherThan(me))}.Apply(ctx)
				}))),
		},
	})
}
