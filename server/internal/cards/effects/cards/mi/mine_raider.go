package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mine Raider — Creature — Human Rogue {2}{R}, 3/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Trample
//	 When this creature enters, if you control another outlaw, create a
//	 Treasure token. (Assassins, Mercenaries, Pirates, Rogues, and
//	 Warlocks are outlaws. …)"
//
// An intervening "if" (CR 603.4), Dundoolin Weaver's shape: with no other
// outlaw (CR 700.12) as it enters it does not trigger, and an outlaw lost
// in response stops the Treasure. "Another" is object identity, so the
// Raider, a Rogue itself, never counts toward its own condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d575b78-3d7b-4860-abc3-e99e429279a1",
		Name:            "Mine Raider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control another outlaw",
				Test: func(ctx *effects.Context) bool {
					return mineRaiderAnotherOutlaw(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID)
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && mineRaiderAnotherOutlaw(g, source.Controller, source.InstanceID)
			}, "Mine Raider — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}))),
		},
	})
}

// mineRaiderAnotherOutlaw is "you control another outlaw": an outlaw
// permanent `you` control other than the object `self`.
func mineRaiderAnotherOutlaw(g *game.Game, you, self uuid.UUID) bool {
	isOutlaw := effects.Outlaw()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != self && c.Controller == you && isOutlaw(g, you, c) {
			return true
		}
	}
	return false
}
