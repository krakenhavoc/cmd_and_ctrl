package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glaring Fleshraker — Creature — Eldrazi Drone {2}{C}, 2/2 (EDHREC
// rank 1222):
//
//	"Whenever you cast a colorless spell, create a 0/1 colorless
//	 Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
//	 Whenever another colorless creature you control enters, this
//	 creature deals 1 damage to each opponent."
//
// The two abilities feed each other: every colorless spell makes a
// Spawn, and every Spawn is a colorless creature entering, so the
// table takes a point per spell. Both are ordinary triggers — the
// cast one reads the spell's colours off the stack, the enters one is
// the Impact Tremors shape narrowed to colorless and "another" — and
// the Spawn is the token Awakening Zone already makes, mana ability
// and all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ffce5e0-6b1d-4d1a-9318-1a1a219df532",
		Name:         "Glaring Fleshraker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B11ColorlessSpellCastByYou(ev, source, g)
			}, "Glaring Fleshraker — create an Eldrazi Spawn", effects.Do(effects.CreateToken{Template: effects.EldraziSpawnToken(), N: 1})),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.IsColorless()
			}, "Glaring Fleshraker — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			}),
		},
	})
}
