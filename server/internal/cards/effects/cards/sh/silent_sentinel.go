package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silent Sentinel — {5}{W}{W} Creature — Archon 4/6:
//
//	"Flying
//	 Whenever this creature attacks, you may return target enchantment card
//	 from your graveyard to the battlefield."
//
// Sun Titan's attack half (the return is Sheoldred's: the target, if it is
// still legal as the ability resolves, CR 608.2b) with an enchantment-card clause. The card returns
// under its owner's control, which is "your graveyard" (CR 400.3).
func init() {
	t := effects.Targeting(
		effects.WheneverThisAttacks("Silent Sentinel — return an enchantment card to the battlefield",
			effects.SheoldredReanimate),
		effects.TargetCardInGraveyard("target enchantment card in your graveyard",
			effects.YouOwn(), effects.Enchantment()),
	)
	t = effects.Optional(t, "Silent Sentinel — return an enchantment card from your graveyard to the battlefield?")
	effects.Register(effects.Spec{
		OracleID:        "26396b80-e5b7-44f5-916c-56a9215a7a49",
		Name:            "Silent Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
