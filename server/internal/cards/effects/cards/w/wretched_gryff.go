package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wretched Gryff — Creature — Eldrazi Hippogriff {7}, 3/4:
//
//	"Emerge {5}{U} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, draw a card.
//	 Flying"
//
// Emerge is the shared alternative cost (ADR 0135 §4). The draw is a cast
// trigger, so it resolves above the spell and happens even if the Gryff
// is countered, however it was cast. Flying is the canonical keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "83d0c4bc-aab9-4c14-a7dc-a344fcce4fea",
		Name:             "Wretched Gryff",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Emerge("{5}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Wretched Gryff — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
