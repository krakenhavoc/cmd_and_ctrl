package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cavalier of Dawn — Creature — Elemental Knight {2}{W}{W}{W}, 4/6:
//
//	"Vigilance
//	 When this creature enters, destroy up to one target nonland permanent.
//	 Its controller creates a 3/3 colorless Golem artifact creature token.
//	 When this creature dies, return target artifact or enchantment card from
//	 your graveyard to your hand."
//
// The enters trigger is Beast Within's destroy-then-give: the target's
// controller is read before the destroy, so the token goes to whoever
// controlled the permanent (last-known information, CR 608.2h) even though it
// is gone. The token does not depend on the destroy succeeding (an
// indestructible target still gives its controller the Golem); a target that
// became illegal, or no target chosen, makes none. The dies trigger targets an artifact or
// enchantment card in your graveyard and returns it to your hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7c520355-1ab8-4d6d-9a29-0f63d6b60024",
		Name:            "Cavalier of Dawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Cavalier of Dawn — destroy up to one target nonland permanent; its controller creates a 3/3 Golem", cavalierOfDawnDestroy),
				effects.TargetPermanent("up to one target nonland permanent", effects.Nonland()).WithCount(0, 1)),
				game.Purpose{
					Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})),
					Answers: game.AnswerRemove,
				}),
			effects.Targeting(
				effects.WhenThisDies("Cavalier of Dawn — return target artifact or enchantment card from your graveyard to your hand", effects.ReturnFirstLegalGraveyardTargetToHand),
				effects.TargetCardInGraveyard("target artifact or enchantment card from your graveyard", effects.YouOwn(), effects.Or(effects.Artifact(), effects.Enchantment()))),
		},
	})
}

// cavalierOfDawnDestroy destroys the chosen permanent, then its controller
// (read before the destroy) creates a 3/3 colorless Golem artifact creature
// token whether or not the destroy took.
func cavalierOfDawnDestroy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			target = t.ID
			break
		}
	}
	if target == uuid.Nil {
		return nil
	}
	controller, ok := effects.ControllerOfTarget(ctx, target)
	if !ok {
		return nil
	}
	if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return effects.CreateToken{
		Controller: controller,
		Template:   effects.TokenCard("3/3 colorless Golem artifact creature"),
		N:          1,
	}.Apply(ctx)
}
