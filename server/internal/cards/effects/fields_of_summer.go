package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fields of Summer — Plane — Moag (#4870):
//
//	"Whenever a player casts a spell, that player may gain 2 life.
//	 Whenever chaos ensues, you may gain 10 life."
//
// The first ability triggers for every spell cast by anyone; the planar
// controller controls it but the caster answers the "may" and gains the life.
// The chaos ability's "may" is the planar controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81a0d96a-7ac8-4947-8779-2102c9b33d38",
		Name:         "Fields of Summer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			b6WheneverAPlayerCastsTheyMay("Fields of Summer — that player may gain 2 life", "Fields of Summer — gain 2 life?",
				func(g *game.Game, item *game.StackItem, caster uuid.UUID) error {
					return GainLife{Player: caster, Amount: 2}.Apply(NewContext(g, item))
				}),
			Optional(WheneverChaosEnsues("Fields of Summer — you may gain 10 life", Do(GainLife{Amount: 10})),
				"Fields of Summer — gain 10 life?"),
		},
	})
}
