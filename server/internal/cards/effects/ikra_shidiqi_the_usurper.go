package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ikra Shidiqi, the Usurper — Legendary Creature — Snake Wizard
// {3}{B}{G}, 3/7 (EDHREC rank 3723):
//
//	"Menace
//	 Whenever a creature you control deals combat damage to a player,
//	 you gain life equal to that creature's toughness.
//	 Partner (You can have two commanders if both have partner.)"
//
// The big-butt partner. Menace rides PrintedKeywords. The trigger is
// Bident of Thassa's condition (combatDamageToPlayerBy — a creature
// the controller controls dealt combat damage to a player, one event
// per creature per player, so two connecting creatures are two
// triggers, as printed) and the life is that creature's toughness,
// read live at resolution when it is still on the battlefield — so a
// pump in response counts — and otherwise the toughness it had when
// the damage was dealt. That fallback is computed once, at trigger
// (Build) time, and carried on the item's Params rather than baked
// into a per-instance closure (ADR 0041 P9) — Righteous Valkyrie's
// fallback. Partner is a deck-construction rule (CR 702.124), the
// deck importer's business.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a1a1761c-88e5-40b4-ba4c-60735c054b09",
		Name:            "Ikra Shidiqi, the Usurper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Key: "Ikra Shidiqi, the Usurper — gain life equal to that creature's toughness",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				dealer := ev.Source
				fallback := 0
				if c, ok := g.LookupCardForEffect(dealer); ok {
					fallback = c.CurrentToughness()
				}
				item := game.NewTriggeredItem(source, "Ikra Shidiqi, the Usurper — gain life equal to that creature's toughness", nil)
				item.Params.Object = game.ObjectRef{ID: dealer}
				item.Params.Amount = fallback
				return item
			},
			Effect: b35GainLifeEqualToToughness,
		}},
	})
}
