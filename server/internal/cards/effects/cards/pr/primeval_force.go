package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primeval Force — Creature — Elemental {2}{G}{G}{G} 8/8 (Eighth Edition,
// S82 older sets batch 619, #5607):
//
//	"When this creature enters, sacrifice it unless you sacrifice three
//	 Forests."
//
// Scythe Tiger's tax with three of a kind: with fewer than three Forests the
// creature is sacrificed with no question; otherwise the controller picks
// between the creature and three Forests of their choice (the same kind of
// pick as Sacred Mesa's, then ChoosePermanents for the three).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4170db0-3adf-4744-b0ec-e889c713bb93",
		Name:         "Primeval Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Primeval Force — sacrifice it unless you sacrifice three Forests", primevalForceTax),
		},
	})
}

func primevalForceTax(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Subtype("Forest"))) < 3 {
		return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	return effects.PickOption{
		Question: "Primeval Force — sacrifice three Forests, or sacrifice Primeval Force",
		Options: []game.ChoiceOption{
			{Label: "Sacrifice three Forests"},
			{Label: "Sacrifice Primeval Force"},
		},
		ThenKey: primevalForceAnswer,
	}.Apply(ctx)
}

// primevalForceAnswer is the tax's choice answered, keyed: index 0 asks for
// three Forests, any other sacrifices the creature. An on-disk key: never
// renamed, never reused.
var primevalForceAnswer = effects.OptionPickThen("option-pick/primeval-force-three-forests-or-this", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index != 0 {
		return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
	}
	return effects.ChoosePermanents{
		Question: "Primeval Force — sacrifice three Forests",
		Keyed:    primevalForceForests,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return effects.PermanentsControlledByMatching(g, of, effects.Subtype("Forest")), 3, 3
		},
	}.Apply(ctx)
})

// primevalForceForests sacrifices the three chosen Forests and nothing else.
var primevalForceForests = effects.RegisterPermanentsPick("prompt-run/primeval-force-three-forests", effects.PermanentsDef{Sacrifice: true})
