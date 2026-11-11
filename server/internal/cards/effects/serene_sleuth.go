package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serene Sleuth — Creature — Human Detective {1}{W}, 2/2 (Murders at
// Karlov Manor Commander):
//
//	"When this creature enters, investigate.
//	 At the beginning of combat on your turn, investigate for each goaded
//	 creature you control. Then each creature you control is no longer
//	 goaded."
//
// The count is taken as the second ability resolves, over the creatures its
// controller controls that are goaded by anyone, a continuous goad
// included (Card.Goaded); the Clues come first and the goads end after
// them. Only a goad marker can end (game.ClearGoadsForEffect): a goad a
// static ability applies stays for as long as that ability does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0d639cb-71d4-454f-9dd6-89cd7a59b702",
		Name:         "Serene Sleuth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Serene Sleuth — investigate", Do(Investigate{N: 1})),
			AtBeginningOfYourCombat("Serene Sleuth — investigate for each goaded creature you control, then each creature you control is no longer goaded", sereneSleuthCombat),
		},
	})
}

func sereneSleuthCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var mine []uuid.UUID
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		mine = append(mine, c.InstanceID)
		if c.Goaded() {
			n++
		}
	}
	if n > 0 {
		if err := (Investigate{N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range mine {
		g.ClearGoadsForEffect(id)
	}
	return nil
}
