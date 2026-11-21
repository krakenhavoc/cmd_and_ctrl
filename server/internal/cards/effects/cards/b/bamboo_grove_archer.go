package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bamboo Grove Archer — Enchantment Creature — Snake Archer {1}{G}, 3/3 (Neon
// Dynasty, #5493):
//
//	"Defender, reach
//	 Channel — {4}{G}, Discard this card: Destroy target creature with
//	 flying."
//
// The target is read through the layered view (HasKeyword), so a creature
// that gains or loses flying is judged as it is now, and re-checked as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "657f6318-aa09-46f1-ab69-896b17a48950",
		Name:            "Bamboo Grove Archer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {4}{G}, Discard this card: Destroy target creature with flying.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{4}{G}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
