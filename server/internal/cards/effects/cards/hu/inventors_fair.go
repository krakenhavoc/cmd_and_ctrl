package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inventors' Fair — Legendary Land (EDHREC rank 281):
//
//	"At the beginning of your upkeep, if you control three or more
//	 artifacts, you gain 1 life.
//	 {T}: Add {C}.
//	 {4}, {T}, Sacrifice Inventors' Fair: Search your library for an
//	 artifact card, reveal it, put it into your hand, then shuffle.
//	 Activate only if you control three or more artifacts."
//
// Metalcraft on a land, twice. The upkeep life is an intervening-if
// (CR 603.4): checked as the upkeep begins and again on resolution.
// The tutor's "Activate only if you control three or more artifacts"
// is its activation condition (CR 602.1b, #743) — Mox Opal's
// ControlsAtLeast(3, artifacts), on a non-mana ability.
//
// No simplification.
func init() {
	threeArtifacts := effects.ControlsAtLeast(3, effects.MatchArtifact)
	effects.Register(effects.Spec{
		OracleID:     "91d4a5fe-fd6d-4b14-a63f-61b4d0ecd9c4",
		Name:         "Inventors' Fair",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return threeArtifacts(g, source.Controller, source.InstanceID)
			}), "Inventors' Fair — gain 1 life", effects.Do(effects.If{
				Cond: effects.CondBoard{Name: "three or more artifacts", Test: func(ctx *effects.Context) bool {
					return threeArtifacts(ctx.Game, ctx.Controller(), ctx.Source())
				}},
				Then: []effects.Applier{effects.GainLife{Amount: 1}},
			})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{4}, {T}, Sacrifice Inventors' Fair: Search your library for an artifact card, reveal it, put it into your hand, then shuffle. Activate only if you control three or more artifacts.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Condition: threeArtifacts,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B06TutorToHand("Inventors' Fair — an artifact card", func(c game.Card) bool { return c.IsArtifact() })(item, effects.NewContext(g, item))
			},
		}},
	})
}
