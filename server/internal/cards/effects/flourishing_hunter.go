package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flourishing Hunter — Creature — Wolf Spirit {4}{G}{G}, 6/6 (Innistrad:
// Crimson Vow, #5495):
//
//	"When this creature enters, you gain life equal to the greatest
//	 toughness among other creatures you control."
//
// Arbor Adherent's count (arborAdherentX): the greatest toughness among
// the OTHER creatures its controller controls as the trigger resolves,
// through the layers, so the Hunter's own 6 never counts and a negative or
// zero toughness gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d714dc46-2db2-439f-b558-2b69d5c36b46",
		Name:         "Flourishing Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Flourishing Hunter — you gain life equal to the greatest toughness among other creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					return GainLife{Amount: arborAdherentX(g, item.Controller, item.SourceCardID)}.Apply(NewContext(g, item))
				}),
		},
	})
}
