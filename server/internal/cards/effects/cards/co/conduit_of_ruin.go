package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conduit of Ruin — Creature — Eldrazi {6}, 5/5 (Battle for Zendikar, #5547):
//
//	"When you cast this spell, you may search your library for a colorless
//	 creature card with mana value 7 or greater, reveal it, then shuffle and
//	 put that card on top.
//	 The first creature spell you cast each turn costs {2} less to cast."
//
// The search is a cast trigger (CR 603.2): it resolves above the spell, and
// finding nothing is allowed (CR 701.19b). The discount applies while this is
// on the battlefield, so it never discounts the spell that is itself.
//
// No simplification.
func init() {
	const label = "Conduit of Ruin — you may search your library for a colorless creature card with mana value 7 or greater, reveal it, then shuffle and put that card on top"
	effects.Register(effects.Spec{
		OracleID:     "cd3d79e9-c72e-4321-bca1-54505d51331a",
		Name:         "Conduit of Ruin",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "The first creature spell you cast each turn costs {2} less to cast.",
				effects.YourSpell(), effects.CreatureSpell(), effects.NoCreatureSpellCastYetThisTurn()),
		},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: effects.Self,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Controller, item.Owner = ev.Actor, ev.Actor
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.S82OldB315ColorlessCreatureMV7,
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Source:    item.SourceCardID,
					Reason:    "Conduit of Ruin — a colorless creature card with mana value 7 or greater",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
