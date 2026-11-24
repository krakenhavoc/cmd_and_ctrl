package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sonic the Hedgehog — Legendary Creature — Hedgehog Warrior {1}{U}{R}{W}, 2/4
// (Secret Lair Drop):
//
//	"Haste
//	 Gotta Go Fast — Whenever Sonic the Hedgehog attacks, put a +1/+1 counter
//	 on each creature you control with flash or haste.
//	 Whenever a creature you control with flash or haste is dealt damage,
//	 create a tapped Treasure token."
//
// Flash and haste are read through the layered view, Sonic himself included
// (he has haste). The set of creatures is fixed as the ability resolves
// (CR 611.2c). The damage trigger fires once per creature dealt damage, not
// once per point of damage: one event, one token.
func init() {
	hasFlashOrHaste := func(c game.Card) bool { return game.HasKeyword(&c, "flash") || game.HasKeyword(&c, "haste") }
	effects.Register(effects.Spec{
		OracleID:        "b07a7b00-9394-48a9-a597-17f391495909",
		Name:            "Sonic the Hedgehog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Sonic the Hedgehog — put a +1/+1 counter on each creature you control with flash or haste",
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && hasFlashOrHaste(c) {
							ids = append(ids, c.InstanceID)
						}
					}
					return effects.B13PutCounterOnEach(effects.NewContext(g, item), ids)
				}),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				return ok && c.IsCreature() && c.Controller == source.Controller && hasFlashOrHaste(c)
			}, "Sonic the Hedgehog — create a tapped Treasure token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TreasureToken()).EntersTapped(), N: 1})),
		},
	})
}
