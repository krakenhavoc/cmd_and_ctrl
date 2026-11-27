package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golgari Decoy — Creature — Elf Rogue {3}{G}, 2/2 (Return to Ravnica, #5566):
//
//	"All creatures able to block this creature do so.
//	 Scavenge {3}{G}{G} ({3}{G}{G}, Exile this card from your graveyard: Put a
//	 number of +1/+1 counters equal to this card's power on target creature.
//	 Scavenge only as a sorcery.)"
//
// Prized Unicorn's lure (CR 509.1c) and the engine's scavenge (scavenge.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fb2eb66-d33b-40e8-9d0d-06d73650a9ac",
		Name:         "Golgari Decoy",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllAbleToBlockDoSo()},
		Activated:    []effects.ActivatedAbility{effects.Scavenge("{3}{G}{G}")},
	})
}
