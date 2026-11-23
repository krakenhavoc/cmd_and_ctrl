package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assassin Initiate — Creature — Human Assassin {B}, 1/1 (Assassin's Creed,
// #5948):
//
//	"{1}: This creature gains your choice of flying, deathtouch, or lifelink
//	 until end of turn."
//
// Argivian Avenger's keyword pick without the shrink: the keyword is chosen as
// the ability resolves (CR 608.2c), not as it is activated, and the Initiate
// gains it until end of turn. An Initiate that left in response gets nothing
// (CR 400.7). Each activation is its own choice, so the three keywords can all
// be had by paying three times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23be5c68-6530-474d-b677-551da4745b3c",
		Name:         "Assassin Initiate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature gains your choice of flying, deathtouch, or lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.PickOption{
					Question: "Assassin Initiate gains which until end of turn?",
					Options:  []game.ChoiceOption{{Label: "Flying"}, {Label: "Deathtouch"}, {Label: "Lifelink"}},
					ThenKey:  assassinInitiateAnswer,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// assassinInitiateKeywords are the three choices, in the order the question
// lists them.
var assassinInitiateKeywords = []string{"flying", "deathtouch", "lifelink"}

// assassinInitiateAnswer grants the keyword chosen to the Initiate (the
// prompt's source) until end of turn. An on-disk key: never renamed.
var assassinInitiateAnswer = effects.OptionPickThen("option-pick/s82-gap-b36-assassin-initiate-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(assassinInitiateKeywords) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: r.Source, Keywords: []string{assassinInitiateKeywords[r.Index]}}.Apply(ctx)
})
