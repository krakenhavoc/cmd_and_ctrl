package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Simian Sling — Artifact Creature — Equipment Monkey {R}, 1/1:
//
//	"Equipped creature gets +1/+1.
//	 Whenever this creature or equipped creature becomes blocked, it
//	 deals 1 damage to defending player.
//	 Reconfigure {2}"
//
// "It" is the creature that became blocked, so the damage comes from
// that creature as an object (its lifelink and deathtouch apply, and
// one that left deals it with what it last had). The defending player
// is the event's Actor, the player whose creatures blocked it.
// Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e49c4d9a-6413-440f-ba5b-396fea6c03d9",
		Name:         "Simian Sling",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.ThisOrEquippedCreatureBecomesBlocked,
				"Simian Sling — it deals 1 damage to defending player", simianSlingDamage),
		},
		Activated: effects.Reconfigure("{2}"),
	})
}

func simianSlingDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	tc := ctx.Trigger()
	if tc.Object == nil || tc.Event.Actor == uuid.Nil {
		return nil
	}
	ref := tc.Object.Ref()
	return effects.DealDamage{SourceObject: &ref, Target: tc.Event.Actor, Amount: 1}.Apply(ctx)
}
