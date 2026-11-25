package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lullmage's Domination — Sorcery {X}{U}{U}{U} (Zendikar Rising, #5508):
//
//	"This spell costs {3} less to cast if it targets a creature whose
//	 controller has eight or more cards in their graveyard.
//	 Gain control of target creature with mana value X."
//
// Price of Fame's target-reading discount (CR 601.2f, after the target is
// chosen): the creature's controller is read as the cost is determined, and
// the {3} comes off the generic part, which here is X. The target must have
// mana value exactly X (a creature's X is 0 off the stack, CR 202.3e). The
// control change has no duration (CR 611.2a), so it lasts until the game
// ends or another effect takes the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdac480b-03b5-4ece-b42a-e6928b609419",
		Name:         "Lullmage's Domination",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with mana value X").WithManaValueEqualsX(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a creature whose controller has eight or more cards in their graveyard.",
				func(g *game.Game, _ uuid.UUID, c game.Card) bool {
					p := g.PlayerByIDForEffect(c.Controller)
					return p != nil && p.Graveyard != nil && p.Graveyard.Size() >= 8
				}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GainControl{Target: id, Duration: game.IndefiniteDuration(), Label: "Lullmage's Domination — gain control"}.Apply(ctx)
		},
	})
}
