package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:      "b254ca74-2a96-4590-91bc-650f407daf5d",
		Name:          "Manifold Mouse",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Manifold Mouse"),
			Targeting(
				AtBeginningOfYourCombat("Manifold Mouse — target Mouse you control gains your choice of double strike or trample until end of turn",
					manifoldMouseChoose),
				TargetCreature("target Mouse you control", YouControl(), OfCreatureType("Mouse"))),
		},
	})
}

// manifoldMouseKeywords are the two choices, in the order offered.
var manifoldMouseKeywords = []string{"double strike", "trample"}

func manifoldMouseChoose(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	options := make([]game.ChoiceOption, len(manifoldMouseKeywords))
	for i, kw := range manifoldMouseKeywords {
		options[i] = game.ChoiceOption{Label: "Gains " + kw + " until end of turn"}
	}
	return PickOption{
		Question: "Manifold Mouse — the Mouse gains your choice of double strike or trample until end of turn",
		Options:  options,
		ThenKey:  manifoldMouseKeyword,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

var manifoldMouseKeyword = OptionPickThen("option-pick/manifold-mouse-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(manifoldMouseKeywords) || len(r.Carry) == 0 || r.Carry[0] == uuid.Nil {
		return nil
	}
	if !onBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   r.Carry[0],
		Keywords: []string{manifoldMouseKeywords[r.Index]},
		Label:    "Manifold Mouse — " + manifoldMouseKeywords[r.Index],
	}.Apply(ctx)
})
