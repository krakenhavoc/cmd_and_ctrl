package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psychic Puppetry — Instant — Arcane {1}{U}:
//
//	"You may tap or untap target permanent.
//	Splice onto Arcane {U}"
func init() {
	Register(Spec{
		OracleID:     "8d3c4e50-6709-42f9-97e7-3165ee501684",
		Name:         "Psychic Puppetry",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{U}"),
		Targets:      TargetPermanent("target permanent"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			target := t.ID
			return PickOption{
				Question: "Psychic Puppetry — you may tap or untap " + cardNameFor(ctx.Game, target),
				Options: []game.ChoiceOption{
					{Label: "Tap it", Cards: []uuid.UUID{target}},
					{Label: "Untap it", Cards: []uuid.UUID{target}},
					{Label: "Leave it as it is", Cards: []uuid.UUID{target}},
				},
				ThenKey: psychicPuppetryAnswer,
				Carry:   []uuid.UUID{target},
			}.Apply(ctx)
		},
	})
}

// psychicPuppetryAnswer is the tap-or-untap choice answered, keyed (#4253):
// the permanent rides the prompt as its one carried ID.
var psychicPuppetryAnswer = OptionPickThen("option-pick/psychic-puppetry-tap-or-untap", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	target := r.Carry[0]
	switch r.Index {
	case 0:
		return TapTarget{Target: target}.Apply(ctx)
	case 1:
		return UntapTarget{Target: target}.Apply(ctx)
	}
	return nil
})
