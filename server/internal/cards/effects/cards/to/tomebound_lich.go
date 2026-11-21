package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tomebound Lich — Creature — Zombie Wizard {1}{U}{B}, 1/3 (Core Set 2020, #5518):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to
//	 destroy it.)
//	 Lifelink (Damage dealt by this creature also causes you to gain that
//	 much life.)
//	 Whenever this creature enters or deals combat damage to a player, draw
//	 a card, then discard a card."
//
// One ability over two events (Aerial Extortionist's condition), with the
// shared loot body: the draw lands first, then the discard is chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6d243164-e21d-46c5-943e-9a2eb759c742",
		Name:            "Tomebound Lich",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventDealDamage},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B26SelfEnteredOrDealtCombatDamageToPlayer(ev, source, g)
				},
				"Tomebound Lich — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) }),
		},
	})
}
