package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Advanced Stitchwing — Creature — Zombie Horror {3}{U}{U}, 3/4 (S82 GAP, #5948):
//
//	"Flying
//	 {2}{U}, Discard two cards: Return this card from your graveyard to the
//	 battlefield tapped."
//
// Reassembling Skeleton's graveyard activation (ActivatedAbilityShape.Zones)
// with a discard-two cost.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e5d2801-873f-4318-9c23-9a73e00e31a4",
		Name:            "Advanced Stitchwing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, Discard two cards: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.DiscardN(2, "two cards")),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisFromGraveyardTapped,
		}},
	})
}
