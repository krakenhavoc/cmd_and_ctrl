package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sidar Jabari of Zhalfir — Legendary Creature — Human Knight
// {1}{W}{U}{B}, 4/3 (EDHREC rank 7470):
//
//	"Eminence — Whenever you attack with one or more Knights, if Sidar
//	 Jabari is in the command zone or on the battlefield, draw a card,
//	 then discard a card.
//	 Flying, first strike
//	 Whenever Sidar Jabari deals combat damage to a player, return
//	 target Knight creature card from your graveyard to the
//	 battlefield."
//
// The eminence line works from the command zone (#2802,
// EminenceTrigger) and is one trigger per attack declaration
// (OncePerBatch), whatever the number of Knights. Sidar Jabari is a
// Knight, so its own attack triggers it on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b229f314-a5e5-41a3-a8c6-217a3c5a61c3",
		Name:            "Sidar Jabari of Zhalfir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			effects.EminenceTrigger(effects.OncePerBatch(effects.On(game.EventAttack, youAttackWithAKnight,
				"Sidar Jabari of Zhalfir — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.LootOne(g, item, 1)
				}))),
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Sidar Jabari of Zhalfir — return target Knight creature card from your graveyard to the battlefield",
				effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("target Knight creature card from your graveyard", effects.OfCreatureType("Knight"), effects.YouOwn())),
		},
	})
}

// youAttackWithAKnight is "Whenever you attack with one or more
// Knights", one event of the batch.
func youAttackWithAKnight(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return effects.AttackDeclaredByYou(ev, source.Controller) && effects.AttackerHasSubtype(g, ev, source.TextSubtype("Knight"))
}
