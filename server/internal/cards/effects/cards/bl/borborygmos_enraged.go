package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borborygmos Enraged — Legendary Creature — Cyclops {4}{R}{R}{G}{G}, 7/6
// (Ravnica Remastered, #5948):
//
//	"Trample
//	 Whenever Borborygmos Enraged deals combat damage to a player, reveal the
//	 top three cards of your library. Put all land cards revealed this way
//	 into your hand and the rest into your graveyard.
//	 Discard a land card: Borborygmos Enraged deals 3 damage to any target."
//
// The trigger is Mulch's reveal-and-take (no prompt: every land goes to the
// hand and the rest are put, not milled). The ability's cost is a land card
// discarded as it is activated (CR 602.2b), so it works at instant speed with
// no mana; the damage is dealt by Borborygmos as it last existed if it has
// gone (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7505737-1fc8-4ee8-8b1d-5dde73b6a76d",
		Name:            "Borborygmos Enraged",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Borborygmos Enraged — reveal the top three cards of your library; put all land cards revealed this way into your hand and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return effects.Cmdr2B15RevealTopTakeAllMatching(g, item.Controller, item.SourceCardID, 3,
						"Borborygmos Enraged — reveal the top three cards of your library", effects.Land())
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a land card: Borborygmos Enraged deals 3 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Cost:    effects.DiscardCardsMatching(1, "a land card", game.Card.IsLand),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
