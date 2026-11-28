package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander Greven il-Vec — Legendary Creature — Phyrexian Human Warrior
// {3}{B}{B}{B}, 7/5 (Tempest, S82 OLD batch 768, #5628):
//
//	"Fear (This creature can't be blocked except by artifact creatures and/or
//	 black creatures.)
//	 When Commander Greven il-Vec enters, sacrifice a creature."
//
// Fear is a printed keyword. The trigger is Accursed Centaur's: its controller
// sacrifices a creature they control as it resolves, and may choose Greven
// himself, the only creature they have if it is alone (CR 701.21a).
//
// No simplification.
func init() {
	const label = "Commander Greven il-Vec — sacrifice a creature"
	effects.Register(effects.Spec{
		OracleID:        "08fa1bee-cedf-4a29-9b02-cbdcc0926f8f",
		Name:            "Commander Greven il-Vec",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller, effects.SacrificeSpec("a creature", effects.Creature()), label)
				return nil
			}),
		},
	})
}
