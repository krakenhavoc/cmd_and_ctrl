package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of Torment — Enchantment Creature — Demon {1}{B}{B}, 3/3 (Born of the
// Gods, S82 OLD b377, #5559):
//
//	"Bestow {3}{B}{B}
//	 Flying
//	 At the beginning of your upkeep, you lose 1 life.
//	 Enchanted creature gets +3/+3 and has flying."
//
// Moroii's upkeep loss, which a bestowed Aura keeps paying (it keeps its
// abilities, CR 702.103), and Hopeful Eidolon's pump-and-grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "3c04aeda-ae09-4162-ab51-983e29d21225",
		Name:             "Herald of Torment",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{3}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Herald of Torment — you lose 1 life", effects.Do(effects.GainLife{Amount: -1})),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("flying"),
		},
	})
}
