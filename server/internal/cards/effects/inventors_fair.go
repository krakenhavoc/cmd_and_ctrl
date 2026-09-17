package effects

import (
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
	threeArtifacts := ControlsAtLeast(3, MatchArtifact)
	Register(Spec{
		OracleID:     "91d4a5fe-fd6d-4b14-a63f-61b4d0ecd9c4",
		Name:         "Inventors' Fair",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return threeArtifacts(g, source.Controller, source.InstanceID)
			}), "Inventors' Fair — gain 1 life", func(g *game.Game, item *game.StackItem) error {
				if !threeArtifacts(g, item.Controller, item.SourceCardID) {
					return nil
				}
				return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
			}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{4}, {T}, Sacrifice Inventors' Fair: Search your library for an artifact card, reveal it, put it into your hand, then shuffle. Activate only if you control three or more artifacts.",
			Cost:      Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Condition: threeArtifacts,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b06TutorToHand("Inventors' Fair — an artifact card", func(c game.Card) bool { return c.IsArtifact() })(item, NewContext(g, item))
			},
		}},
	})
}
