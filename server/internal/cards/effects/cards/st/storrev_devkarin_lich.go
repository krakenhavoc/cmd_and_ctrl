package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storrev, Devkarin Lich — Legendary Creature — Zombie Elf Wizard {1}{B}{B}{G},
// 5/4 (War of the Spark, #5520):
//
//	"Trample
//	 Whenever Storrev deals combat damage to a player or planeswalker, return
//	 to your hand target creature or planeswalker card in your graveyard that
//	 wasn't put there this combat."
//
// The target is chosen as the trigger goes on the stack, so a card Storrev's
// own combat damage killed (put there this combat) is not offered; the clause
// is read again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "159098f1-172f-408e-a962-86763a29a8e1",
		Name:            "Storrev, Devkarin Lich",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrPlaneswalker,
					"Storrev — return target creature or planeswalker card in your graveyard that wasn't put there this combat to your hand",
					effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target creature or planeswalker card in your graveyard that wasn't put there this combat",
					effects.YouOwn(), effects.Or(effects.Creature(), effects.Planeswalker()), effects.S82OldB169NotPutIntoGraveyardThisCombat)),
		},
	})
}
