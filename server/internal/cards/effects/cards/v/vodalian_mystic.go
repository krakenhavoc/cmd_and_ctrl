package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vodalian Mystic — Creature — Merfolk Wizard {1}{U}, 1/1 (Apocalypse, S82 OLD,
// #5614):
//
//	"{T}: Target instant or sorcery spell becomes the color of your choice."
//
// Illusion's colour change on a spell: the colour is chosen as the ability
// resolves (CR 608.2d), then set in layer 5 on the spell on the stack
// (CR 613.1e, game.SetSpellColorsForEffect). The spell is an instant or sorcery,
// which leaves the stack as it resolves or is countered, so the change lasts as
// long as the spell can be affected by it; the card in a graveyard is a new
// object (CR 400.7). A target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7feecf0-5229-4c12-806a-16c9ab38e147",
		Name:         "Vodalian Mystic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target instant or sorcery spell becomes the color of your choice.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetSpell("target instant or sorcery spell", effects.Or(effects.Instant(), effects.Sorcery())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				effects.ChooseColorTail(game.ColorForBenefit, g, ctx.Controller(), ctx.Source(),
					"Vodalian Mystic — choose a color", ctx.ItemTail(vodalianMysticColorChosen))
				return nil
			},
		}},
	})
}

// vodalianMysticColorChosen sets the chosen colour on the target spell. An
// on-disk key.
var vodalianMysticColorChosen = effects.CardTail("vodalian-mystic/color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	return effects.Gap2B05TargetBecomesColor(ctx, []string{color}, "Vodalian Mystic — "+game.ColorName(color))
})
