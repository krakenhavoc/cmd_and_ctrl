package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tether Technician — Creature — Minotaur Artificer {4}{R}, 4/5:
//
//	"Reach
//	 When this creature enters, you may discard a card. When you do,
//	 this creature deals 2 damage to any target."
//
// "You may discard" is an up-to-one discard prompt; the reflexive
// trigger (CR 603.12) is created only once a card really was discarded,
// and chooses its target as it goes on the stack, so the table can
// answer the discard before the target is picked. With an empty hand
// nothing is discarded and no trigger is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9d0657f-2c5e-4b6f-b8ae-d657cc0f7aef",
		Name:            "Tether Technician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tether Technician — you may discard a card; when you do, 2 damage to any target", tetherTechnicianETB),
		},
	})
}

var tetherTechnicianStrikeBody = game.ReflexiveBody("tether-technician/strike", breechesBlastEffect, constTargets(TargetAny))

func tetherTechnicianETB(g *game.Game, item *game.StackItem) error {
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        1,
		UpTo:     true,
		Question: "Tether Technician — you may discard a card to deal 2 damage to any target",
		Then: func(g *game.Game, _ uuid.UUID, discarded []uuid.UUID) error {
			if len(discarded) == 0 {
				return nil
			}
			t := WhenYouDo("Tether Technician — 2 damage to any target", tetherTechnicianStrikeBody)
			t.Params = game.EffectParams{Amount: 2}
			return t.Apply(NewContext(g, item))
		},
	})
	return nil
}
