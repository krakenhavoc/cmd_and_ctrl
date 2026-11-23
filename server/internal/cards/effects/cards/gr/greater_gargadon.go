package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greater Gargadon — Creature — Beast {9}{R}, 9/7 (Time Spiral Remastered /
// Planar Chaos):
//
//	"Suspend 10—{R}
//	 Sacrifice an artifact, creature, or land: Remove a time counter from
//	 this card. Activate only if this card is suspended."
//
// Suspend 10—{R} is the keyword's two numbers (CR 702.62a): the special
// action, the time counters, the upkeep countdown, the free cast and the haste
// all come with it. The second ability works from exile, where a suspended
// card sits, and only while the card is suspended (CR 702.62b: exiled with a
// time counter on it), so a Gargadon exiled any other way, or in a hand or
// on the battlefield, can't use it. The sacrifice is a cost paid at
// announcement and may be as many creatures, artifacts or lands as you have
// priority for; removing the last counter casts it without paying its mana
// cost through the keyword's own trigger (CR 702.62a), exactly as an upkeep
// removal does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "406b4b1d-756b-4062-897d-3693832086a0",
		Name:           "Greater Gargadon",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Suspend(10, "{R}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an artifact, creature, or land: Remove a time counter from this card. Activate only if this card is suspended.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeN(1, "an artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Zones:   []game.ZoneKind{game.ZoneExile},
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				c, ok := g.LookupCardForEffect(source)
				return ok && game.CardIsSuspended(c)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				// A card that has left exile and come back is a new object (CR 400.7).
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !game.CardIsSuspended(c) {
					return nil
				}
				return g.AddCounterForEffect(item.SourceCardID, game.CounterTime, -1)
			},
		}},
	})
}
