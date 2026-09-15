package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0ffce5e0-6b1d-4d1a-9318-1a1a219df532",
		Name:         "Glaring Fleshraker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b11ColorlessSpellCastByYou(ev, source, g)
			}, "Glaring Fleshraker — create an Eldrazi Spawn", Do(CreateToken{Template: EldraziSpawnToken(), N: 1})),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.IsColorless()
			}, "Glaring Fleshraker — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
