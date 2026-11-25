package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roost of Drakes — Enchantment {U} (Zendikar Rising, S82 OLD #5508):
//
//	"Kicker {2}{U}
//	 When this enchantment enters, if it was kicked, create a 2/2 blue Drake
//	 creature token with flying.
//	 Whenever you cast a kicked spell, create a 2/2 blue Drake creature
//	 token with flying."
//
// The first trigger reads the kicker the entering permanent carries
// (CR 702.33a, 400.7d); the second reads the spell on the stack as it is
// cast (CR 601.2i), so the Roost's own cast does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "05e73886-8311-4a0d-9c8d-4a03c7f20ea2",
		Name:          "Roost of Drakes",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 2, CreatureTokens: 2},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Roost of Drakes — kicked: create a 2/2 blue Drake creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 blue Drake with flying"), N: 1})),
			effects.On(game.EventCast, effects.S82OldB089YouCastAKickedSpell,
				"Roost of Drakes — create a 2/2 blue Drake creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 blue Drake with flying"), N: 1})),
		},
	})
}
