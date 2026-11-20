package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gloom Sower — Creature — Horror {5}{B}{B}, 8/6 (Core Set 2021, #5511):
//
//	"Whenever this creature becomes blocked by a creature, that creature's
//	 controller loses 2 life and you gain 2 life."
//
// Vicious Battlerager's trigger (viciousBattleragerBlocked), once per
// blocker (CR 509.3d). "That creature's controller" is read as it resolves
// from the blocking object, or as it last existed on the battlefield (CR
// 608.2h). The gain is its own instruction and happens either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5cc1492-c9cc-48f0-8c11-02ead883675c",
		Name:         "Gloom Sower",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, viciousBattleragerBlocked,
				"Gloom Sower — that creature's controller loses 2 life and you gain 2 life", gloomSowerDrain),
		},
	})
}

// gloomSowerDrain costs the blocking creature's controller 2 life, then
// gives Gloom Sower's controller 2.
func gloomSowerDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if item.Trigger != nil && item.Trigger.Object != nil {
		o := item.Trigger.Object
		if info, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch}); ok && info.Controller != uuid.Nil {
			if err := (GainLife{Player: info.Controller, Amount: -2}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return GainLife{Amount: 2}.Apply(ctx)
}
