package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b71a_helpers.go — shared vocabulary for Standard batch 71, group a
// (Lorwyn Eclipsed, #4458): a creature-type choice from a printed list,
// and a target clause narrowed to the source's chosen type.

// ChooseCreatureTypeFromAsEnters is "As this permanent enters, choose
// <types[0]>, <types[1]>, … or <types[n]>" (Dawn-Blessed Pennant's eight
// Lorwyn types): the CR 614.12 creature-type choice limited to the
// printed list. The prompt is an option pick of exactly those words, in
// printed order, and the answer lands where every other chosen creature
// type does (Card.NamedTribe), so "of the chosen type" reads it the same
// way. Words outside the CR 205.3m list are refused at boot.
func ChooseCreatureTypeFromAsEnters(label string, types ...string) func(*game.Card, *Context) error {
	words := make([]string, len(types))
	for i, t := range types {
		canonical, ok := game.CanonicalCreatureType(t)
		if !ok {
			panic("effects.ChooseCreatureTypeFromAsEnters: " + t + " is not a creature type")
		}
		words[i] = canonical
	}
	question := label + " — choose " + joinOr(words)
	return func(card *game.Card, ctx *Context) error {
		options := make([]game.ChoiceOption, len(words))
		for i, w := range words {
			options[i] = game.ChoiceOption{Label: w}
		}
		ctx.Game.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  card.Controller,
			Source:   card.InstanceID,
			Question: question,
			Options:  options,
			ThenKey:  b71aCreatureTypeFromListChosen,
			Carry:    []uuid.UUID{card.InstanceID},
			Words:    words,
		})
		return nil
	}
}

// b71aCreatureTypeFromListChosen stamps the answered type onto the
// permanent (Carry[0]). A dropped prompt (no option) names nothing, and a
// permanent that has left has nowhere for the answer to land (CR 608.2).
// An on-disk key: never renamed, never reused.
var b71aCreatureTypeFromListChosen = game.RegisterOptionPickThen("option-pick/choose-creature-type-from-list-as-enters",
	func(g *game.Game, r game.OptionPicked) error {
		if len(r.Carry) == 0 || r.Option == nil {
			return nil
		}
		g.SetNamedTribeForEffect(r.Carry[0], r.Option.Label)
		return nil
	})

// OfTheChosenType narrows a clause to the cards that have the creature
// type chosen for its own source (game.TargetSpec.OfSourcesChosenType):
// "target card of the chosen type". The source is read live, or as it
// last existed once a sacrifice cost has taken it (CR 608.2h), and a
// source with no type chosen admits nothing. Wrap the whole clause:
//
//	Targets: OfTheChosenType(TargetCardInGraveyard("target card of the chosen type from your graveyard", YouOwn()))
func OfTheChosenType(spec *game.TargetSpec) *game.TargetSpec {
	spec.OfSourcesChosenType = true
	return spec
}
