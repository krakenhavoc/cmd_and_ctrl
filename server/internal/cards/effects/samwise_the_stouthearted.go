package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Samwise the Stouthearted — Legendary Creature — Halfling Peasant
// {1}{W}, 2/1:
//
//	"Flash
//	 When Samwise enters, choose up to one target permanent card in
//	 your graveyard that was put there from the battlefield this turn.
//	 Return it to your hand. Then the Ring tempts you."
//
// "Put there from the battlefield this turn" is Cry of the Carnarium's
// reading of the turn's events (putIntoGraveyardFromBattlefieldThisTurn):
// the card's newest arrival in the graveyard happened this turn and
// came from the battlefield. With no target chosen the Ring still
// tempts; with a chosen target gone, the ability does nothing (CR
// 608.2b).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Samwise the Stouthearted — return a permanent card that died this turn, then the Ring tempts you", samwiseReturn)
	enters.Targets = TargetCardInGraveyard("up to one target permanent card in your graveyard that was put there from the battlefield this turn",
		YouOwn(), Permanent(), putThereFromTheBattlefieldThisTurn).WithCount(0, 1)
	Register(Spec{
		OracleID:        "0e26b429-0469-474a-818a-2d4696c38a05",
		Name:            "Samwise the Stouthearted",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

// putThereFromTheBattlefieldThisTurn is the clause's predicate.
func putThereFromTheBattlefieldThisTurn(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return putIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID)
}

// samwiseReturn is the enters ability's body.
func samwiseReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := returnLegalGraveyardTargetsToHand(ctx); err != nil {
		return err
	}
	return TheRingTemptsYou{}.Apply(ctx)
}
