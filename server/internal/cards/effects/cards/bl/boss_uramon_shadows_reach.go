package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boss Uramon, Shadow's Reach — Legendary Creature — Human Rogue {3}{B}, 2/2
// (Mystery Booster Commander, #5948):
//
//	"Shadow (This creature can block or be blocked by only creatures with
//	 shadow.)
//	 Attacking Ninjas and Rogues you control have deathtouch.
//	 Whenever one or more Ninjas and/or Rogues you control deal combat damage
//	 to a player, create a 1/1 black Rat Rogue creature token."
//
// The grant is Blade Historian's attacking-creatures static narrowed to the
// two types, so Uramon (a Rogue) gives it to himself while he attacks. The
// trigger is Zurgo and Ojutai's batch shape (CR 603.2c): one trigger per
// damaged player however many Ninjas and Rogues dealt it, and a combat step
// that hurts two players triggers twice.
//
// No simplification.
func init() {
	ninjaOrRogue := func(c game.Card) bool { return c.HasSubtypeWord("Ninja") || c.HasSubtypeWord("Rogue") }
	effects.Register(effects.Spec{
		OracleID:        "92eeedc6-a331-445f-af56-c91dac69412b",
		Name:            "Boss Uramon, Shadow's Reach",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Static: []game.StaticAbility{
			effects.AttackingCreaturesHave("Attacking Ninjas and Rogues you control have deathtouch.",
				func(c *game.Card) bool { return ninjaOrRogue(*c) }, "deathtouch"),
		},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerPlayer(effects.On(game.EventDealDamage,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return g.PlayerByIDForEffect(ev.Target) != nil &&
						effects.S82CombatDamageToPlayerOrBattleBy(ev, source.Controller, g, ninjaOrRogue)
				},
				"Boss Uramon, Shadow's Reach — create a 1/1 black Rat Rogue creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Rat Rogue"), N: 1}))),
		},
	})
}
