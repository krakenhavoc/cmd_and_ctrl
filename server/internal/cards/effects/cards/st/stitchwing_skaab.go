package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stitchwing Skaab — Creature — Zombie Horror {3}{U}, 3/1 (Shadows over
// Innistrad, #5543):
//
//	"Flying
//	 {1}{U}, Discard two cards: Return this card from your graveyard to the
//	 battlefield tapped."
//
// Advanced Stitchwing's graveyard activation (ActivatedAbilityShape.Zones)
// with a discard-two cost, at a lower price. The discard is a cost
// (CR 602.2c) so the player picks the cards as the ability is announced;
// the card returns tapped, and does nothing if it has left the graveyard
// by resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b40e537-d385-450d-93a9-b60a25bfe508",
		Name:            "Stitchwing Skaab",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.S82OldB285ReturnThisDiscardingTwo("{1}{U}"),
		},
	})
}
