package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bishop of Rebirth — Creature — Vampire Cleric {3}{W}{W}, 3/4 (Ixalan,
// S82 older sets):
//
//	"Vigilance
//	 Whenever this creature attacks, you may return target creature card
//	 with mana value 3 or less from your graveyard to the battlefield."
//
// Silent Sentinel's shape: the target is chosen as the trigger goes on the
// stack (CR 603.3d), the "may" is asked as it resolves, and the card is
// checked again for being a creature card of mana value 3 or less in the
// graveyard (CR 608.2b).
//
// No simplification.
func init() {
	t := effects.Targeting(
		effects.WheneverThisAttacks("Bishop of Rebirth — return a creature card with mana value 3 or less to the battlefield",
			effects.HobB8ReturnTargetGraveyardCardToBattlefield),
		effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
			effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3)),
	)
	t = effects.Optional(t, "Bishop of Rebirth — return a creature card with mana value 3 or less from your graveyard to the battlefield?")
	effects.Register(effects.Spec{
		OracleID:        "05058594-608f-4046-bf44-b736a7072f0a",
		Name:            "Bishop of Rebirth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
