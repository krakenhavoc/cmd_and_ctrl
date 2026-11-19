package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Urchin — Creature — Elemental {1}{R}, 1/3 (Ravnica: Clue Edition):
//
//	"Trample
//	 Whenever you cast an instant or sorcery spell, this creature gets
//	 +1/+0 until end of turn."
//
// Elemental Mascot's trigger without the mana-spent rider.
//
// No simplification.
func init() {
	const label = "Fire Urchin — +1/+0 until end of turn"
	Register(Spec{
		OracleID:        "4000e119-3372-431c-bc3a-0db0639311ce",
		Name:            "Fire Urchin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastByYou(ev, source, g)
			}, label, Do(Instruction(func(ctx *Context) error {
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: label}.Apply(ctx)
			}))),
		},
	})
}
