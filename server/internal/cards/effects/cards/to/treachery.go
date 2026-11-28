package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treachery — {3}{U}{U} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 When this Aura enters, untap up to five lands.
//	 You control enchanted creature."
//
// Mind Control's control effect and a free untap of up to five of your lands
// (UntapUpToLands).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ed57194-7508-4aef-9373-64f7e80612d8",
		Name:         "Treachery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Treachery — untap up to five lands",
				effects.Do(effects.UntapUpToLands{N: 5, Question: "Treachery — untap up to five lands"})),
		},
	})
}
