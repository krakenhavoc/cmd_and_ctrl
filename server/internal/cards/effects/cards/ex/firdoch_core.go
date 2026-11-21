package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firdoch Core — Kindred Artifact — Shapeshifter {3} (Lorwyn Eclipsed,
// #4458):
//
//	"Changeling (This card is every creature type.)
//	 {T}: Add one mana of any color.
//	 {4}: This artifact becomes a 4/4 artifact creature until end of
//	 turn."
//
// Changeling is a characteristic-defining ability honoured in every zone
// (CR 702.73a), so the Core is every creature type as a kindred artifact
// and stays every type while animated. The mana ability is Birds of
// Paradise's pick. The animation is Iron Suitcase's BecomeCreatureUntilEOT
// (types and base 4/4 for the turn, layer 7b), which keeps the Core's
// other types and abilities; a 4/4-setting effect that starts later
// overwrites it, as its ruling says, by timestamp.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "126c75c0-414b-4e55-8747-2b5566ce716f",
		Name:            "Firdoch Core",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}: This artifact becomes a 4/4 artifact creature until end of turn.",
			Cost:    effects.ManaCost("{4}"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.BecomeCreatureUntilEOT{
					SetPower: 4, SetToughness: 4,
					Label: "Firdoch Core — 4/4 artifact creature",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
