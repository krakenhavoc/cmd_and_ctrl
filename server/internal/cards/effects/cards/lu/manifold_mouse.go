package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manifold Mouse — Creature — Mouse Soldier {1}{R}, 1/2 (Bloomburrow,
// #4467):
//
//	"Offspring {2}
//	 At the beginning of combat on your turn, target Mouse you control
//	 gains your choice of double strike or trample until end of turn."
//
// Offspring is the shared keyword (offspring.go): an optional
// additional cost and an entry trigger that makes a 1/1 token copy,
// which carries the combat trigger too. The trigger targets as it goes
// on the stack and the keyword is chosen as it resolves, as Practiced
// Offense's is: a registered option pick told the target by Carry, so a
// table waiting on the choice is a restore point. A target that has
// left by then gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b254ca74-2a96-4590-91bc-650f407daf5d",
		Name:          "Manifold Mouse",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			effects.OffspringToken("Manifold Mouse"),
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Manifold Mouse — target Mouse you control gains your choice of double strike or trample until end of turn",
					manifoldMouseChoose),
				effects.TargetCreature("target Mouse you control", effects.YouControl(), effects.OfCreatureType("Mouse"))),
		},
	})
}

// manifoldMouseKeywords are the two choices, in the order offered.
var manifoldMouseKeywords = []string{"double strike", "trample"}

func manifoldMouseChoose(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	options := make([]game.ChoiceOption, len(manifoldMouseKeywords))
	for i, kw := range manifoldMouseKeywords {
		options[i] = game.ChoiceOption{Label: "Gains " + kw + " until end of turn"}
	}
	return effects.PickOption{
		Question: "Manifold Mouse — the Mouse gains your choice of double strike or trample until end of turn",
		Options:  options,
		ThenKey:  manifoldMouseKeyword,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

var manifoldMouseKeyword = effects.OptionPickThen("option-pick/manifold-mouse-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(manifoldMouseKeywords) || len(r.Carry) == 0 || r.Carry[0] == uuid.Nil {
		return nil
	}
	if !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   r.Carry[0],
		Keywords: []string{manifoldMouseKeywords[r.Index]},
		Label:    "Manifold Mouse — " + manifoldMouseKeywords[r.Index],
	}.Apply(ctx)
})
