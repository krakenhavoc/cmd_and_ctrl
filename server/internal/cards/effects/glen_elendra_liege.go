package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glen Elendra Liege — Creature — Faerie Knight {1}{U/B}{U/B}{U/B}, 2/3:
//
//	"Flying
//	 Other blue creatures you control get +1/+1.
//	 Other black creatures you control get +1/+1."
//
// Two printed abilities, so two statics: a creature that is both blue
// and black gets +2/+2, which is the card's ruling (each ability
// applies on its own). Colours are read through the layered view, so a
// colour-changing effect moves a creature in or out of the bonus. "Other"
// excludes the Liege itself, by instance, so a second Liege buffs the
// first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "946bba74-0951-408c-b06f-167739b10934",
		Name:            "Glen Elendra Liege",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			b16Anthem(g2OtherCreaturesYouControlOfColor("U"), 1, 1),
			b16Anthem(g2OtherCreaturesYouControlOfColor("B"), 1, 1),
		},
	})
}

// g2OtherCreaturesYouControlOfColor scopes a static to the other
// creatures its source's controller controls that are `color` as they
// are now.
func g2OtherCreaturesYouControlOfColor(color string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller &&
			target.InstanceID != source.InstanceID &&
			slices.Contains(target.EffectiveColors(), color)
	}
}
