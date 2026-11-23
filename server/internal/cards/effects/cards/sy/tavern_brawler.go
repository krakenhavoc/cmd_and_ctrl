package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tavern Brawler — Legendary Enchantment — Background {2}{R} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "At the beginning of your upkeep, exile
//	 the top card of your library. This creature gets +X/+0 until end of turn,
//	 where X is that card's mana value. You may play that card this turn.""
//
// A granted trigger (ADR 0093), the commander creature's own, on Kellan's
// impulse exile (ExileTopWithPermission: you may play the card this turn). The
// pump waits for the exile to settle and is told the card that really reached
// exile (B12ImpulseExileForTurnThen): X is its mana value (a land is 0), and
// the creature is the trigger's source, which gets nothing if it has left (CR
// 608.2h). Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const tavernBrawlerGrant = "tavern-brawler/impulse"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c5cd912-eddb-47c8-a31b-8ad2af46abd3",
		Name:         "Tavern Brawler",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: tavernBrawlerGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Tavern Brawler — exile the top card of your library; this creature gets +X/+0 until end of turn, X that card's mana value; you may play that card this turn",
					func(g *game.Game, item *game.StackItem) error {
						return effects.B12ImpulseExileForTurnThen(g, item, 1, tavernBrawlerPump)
					}),
			},
			Text: "At the beginning of your upkeep, exile the top card of your library. This creature gets +X/+0 until end of turn, where X is that card's mana value. You may play that card this turn.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(tavernBrawlerGrant)},
	})
}

// tavernBrawlerPump is "this creature gets +X/+0 until end of turn", once the
// card is in exile. An on-disk key: never renamed, never reused.
var tavernBrawlerPump = effects.CardTail("tavern-brawler/pump", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || c.ManaValue() <= 0 {
		return nil
	}
	return effects.BoostUntilEOT{Target: ctx.Source(), Power: c.ManaValue(), Label: "Tavern Brawler — +X/+0"}.Apply(ctx)
})
