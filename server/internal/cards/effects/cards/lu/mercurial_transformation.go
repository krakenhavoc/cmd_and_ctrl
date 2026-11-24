package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mercurial Transformation — Sorcery — Lesson {1}{U} (Strixhaven: School of
// Mages, S82 batch 047, #5502):
//
//	"Until end of turn, target nonland permanent loses all abilities and
//	 becomes your choice of a blue Frog creature with base power and
//	 toughness 1/1 or a blue Octopus creature with base power and toughness
//	 4/4."
//
// Turn to Frog's one until-end-of-turn effect with the shape picked as the
// spell resolves (an option pick, the caster's, asked only when the target
// is still legal): layer 4 the permanent becomes a creature of the chosen
// subtype and nothing else (CR 205.1a, so an artifact or planeswalker stops
// being one), layer 5 blue, layer 6 no abilities and layer 7b the chosen
// base power and toughness. Counters and +N/+N effects still apply on top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7efe6c6-3da1-47a6-a8ad-30c36502a7ab",
		Name:         "Mercurial Transformation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok || !effects.OnBattlefield(ctx.Game, id) {
				return nil
			}
			return effects.PickOption{
				Question: "Mercurial Transformation — becomes a blue Frog 1/1 or a blue Octopus 4/4?",
				Options: []game.ChoiceOption{
					{Label: "Frog (1/1)"},
					{Label: "Octopus (4/4)"},
				},
				ThenKey:  mercurialTransformationShape,
				Carry:    []uuid.UUID{id},
				WithItem: true,
			}.Apply(ctx)
		},
	})
}

// mercurialTransformationShape applies the shape chosen: option 0 is the
// Frog, option 1 the Octopus. An on-disk key.
var mercurialTransformationShape = effects.OptionPickThen("option-pick/mercurial-transformation-shape", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	subtype, p := "Frog", 1
	switch r.Index {
	case 0:
	case 1:
		subtype, p = "Octopus", 4
	default:
		return nil
	}
	return effects.UntilEndOfTurn(ctx, r.Carry[0], nil,
		"Mercurial Transformation — loses all abilities and becomes a blue "+subtype+" creature",
		game.SetTypesMod([]string{"Creature"}, ctx.TextSubtypes(subtype)),
		game.SetColorsMod(ctx.TextColor("U")),
		game.LoseAllAbilitiesMod(),
		game.SetBasePowerMod(p),
		game.SetBaseToughnessMod(p),
	)
})
