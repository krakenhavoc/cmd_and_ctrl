package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Septic Rats — Creature — Phyrexian Rat {1}{B}{B}, 2/2 (Mirrodin Besieged,
// S82 OLD b466, #5576):
//
//	"Infect
//	 Whenever this creature attacks, if defending player is poisoned, it gets
//	 +1/+1 until end of turn."
//
// An intervening "if" (CR 603.4), Reaper of Night's shape: the defending player
// is fixed as the ability triggers (CR 506.2: the player attacked, or the
// controller of the planeswalker or protector of the battle attacked) and
// carried on the item, and poisoned (one or more poison counters, CR 122.1f)
// is checked as it triggers and again as it resolves. Infect is the engine's
// damage tail (ADR 0056).
//
// No simplification.
func init() {
	const label = "Septic Rats — it gets +1/+1 until end of turn"
	t := effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.AttackDeclared(ev, source) && oldB466Poisoned(g, g.DefendingPlayerForAttackForEffect(ev.Target))
	}, label, func(g *game.Game, item *game.StackItem) error {
		if !oldB466Poisoned(g, item.Params.Player) {
			return nil
		}
		return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: label}.Apply(effects.NewContext(g, item))
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Player = g.DefendingPlayerForAttackForEffect(ev.Target)
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "c7335845-ace9-402d-a5be-3bfe3fe9b089",
		Name:            "Septic Rats",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

// oldB466Poisoned reports whether `player` has one or more poison counters.
func oldB466Poisoned(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && p.Counters[game.CounterPoison] >= 1
}
