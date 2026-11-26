package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fountain of Ichor — Artifact {3} (Modern Horizons):
//
//	"{T}: Add one mana of any color.
//	 {3}: This artifact becomes a 3/3 Dinosaur artifact creature until end of
//	 turn."
//
// The mana ability is Abstergo Entertainment's any-color tap (CR 605.1a). The
// animation is Living Brain's BecomeCreatureUntilEOT with a creature type: the
// Creature type and Dinosaur subtype are added (layer 4) and base power and
// toughness set to 3/3 (layer 7b) until end of turn, pinned to this object so
// a flicker ends it (CR 400.7). It stays an artifact, so it keeps its mana
// ability, and it can't attack the turn it arrived (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9082d364-9fe4-47bd-9097-8888ae75b2e1",
		Name:         "Fountain of Ichor",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This artifact becomes a 3/3 Dinosaur artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.BecomeCreatureUntilEOT{
					Target:       item.SourceCardID,
					Subtypes:     ctx.TextSubtypes("Dinosaur"),
					SetPower:     3,
					SetToughness: 3,
					Label:        "Fountain of Ichor — becomes a 3/3 Dinosaur artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
