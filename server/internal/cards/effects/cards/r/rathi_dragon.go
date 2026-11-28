package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rathi Dragon — Creature — Dragon {2}{R}{R}, 5/5 (Tempest, S82 OLD batch 768,
// #5628):
//
//	"Flying
//	 When this creature enters, sacrifice it unless you sacrifice two
//	 Mountains."
//
// Primeval Force's tax with two Mountains: with fewer than two the Dragon is
// sacrificed outright; otherwise its controller chooses between the two Mountains
// and the Dragon. Both sacrifices are real choices, so the step cannot pass
// unanswered. A Mountain is a land with the subtype, so a Mountain an effect has
// made is one (CR 305.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6a180f07-7afa-4605-9c58-4521a5f3fcc1",
		Name:            "Rathi Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rathi Dragon — sacrifice it unless you sacrifice two Mountains", rathiDragonTax),
		},
	})
}

func rathiDragonTax(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Subtype("Mountain"))) < 2 {
		return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	return effects.PickOption{
		Question: "Rathi Dragon — sacrifice two Mountains, or sacrifice Rathi Dragon",
		Options: []game.ChoiceOption{
			{Label: "Sacrifice two Mountains"},
			{Label: "Sacrifice Rathi Dragon"},
		},
		ThenKey: rathiDragonAnswer,
	}.Apply(ctx)
}

var rathiDragonAnswer = effects.OptionPickThen("option-pick/rathi-dragon-two-mountains-or-this", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index != 0 {
		return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
	}
	return effects.ChoosePermanents{
		Question: "Rathi Dragon — sacrifice two Mountains",
		Keyed:    rathiDragonMountains,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return effects.PermanentsControlledByMatching(g, of, effects.Subtype("Mountain")), 2, 2
		},
	}.Apply(ctx)
})

var rathiDragonMountains = effects.RegisterPermanentsPick("prompt-run/rathi-dragon-two-mountains", effects.PermanentsDef{Sacrifice: true})
