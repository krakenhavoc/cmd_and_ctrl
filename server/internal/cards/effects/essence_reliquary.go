package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Essence Reliquary — Artifact {2}{W} (Modern Horizons 3, #5948):
//
//	"{T}: Return another target permanent you control and all Auras you
//	 control attached to it to their owner's hand. Activate only during your
//	 turn."
//
// The target is a permanent other than the Reliquary itself that its
// controller controls. As the ability resolves, the permanent and every Aura
// the controller controls that is attached to it are returned together (one
// batch, so no state-based action runs between them and an Aura is not put
// into the graveyard first): each goes to its owner's hand, so an Aura an
// opponent owns goes to that player's hand. A target that has left makes the
// ability do nothing (CR 608.2b), and an Aura that was attached to it is not
// returned then. The timing is a restriction on the activation, checked as it
// is announced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "528039f3-e8e6-499d-add4-43cf2d731b65",
		Name:         "Essence Reliquary",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Return another target permanent you control and all Auras you control attached to it to their owner's hand. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerSurvive},
			Cost:      TapCost(),
			Condition: DuringYourTurn(),
			Targets:   Another(TargetPermanent("another target permanent you control", YouControl())),
			Effect:    essenceReliquaryReturn,
		}},
	})
}

func essenceReliquaryReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
			break
		}
	}
	if target == uuid.Nil {
		return nil
	}
	ids := []uuid.UUID{target}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.HasSubtypeWord("Aura") &&
			c.AttachedTo.Kind == game.TargetCard && c.AttachedTo.ID == target {
			ids = append(ids, c.InstanceID)
		}
	}
	return g.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
}
