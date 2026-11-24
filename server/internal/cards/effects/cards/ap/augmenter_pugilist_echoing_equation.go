package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Augmenter Pugilist // Echoing Equation — modal double-faced card (Strixhaven:
// School of Mages, #5502):
//
//	Augmenter Pugilist — Creature — Troll Druid {1}{G}{G}, 3/3
//	 "Trample
//	  As long as you control eight or more lands, this creature gets +5/+5."
//	Echoing Equation — Sorcery {3}{U}{U}
//	 "Choose target creature you control. Each other creature you control
//	  becomes a copy of it until end of turn, except those creatures aren't
//	  legendary."
//
// # Two keys, one card
//
// The creature keeps the bare oracle ID and the sorcery takes
// "<oracle_id>#1" (Wandering Archaic's keyspace).
//
// # The front face
//
// Gigantoad's self-pump at +5/+5 with the land count read live (CR 604.2).
//
// # The back face
//
// Mirrorweave's mass copy: the model is the target, and the creatures that become
// copies are the other creatures its controller controls as the spell resolves
// (CR 611.2c fixes the set). Each copy lasts until end of turn and is not
// legendary (CR 707.9b's "except"). A model that is no longer a legal target
// makes the spell do nothing (CR 608.2b).
//
// No simplification on either face.
func init() {
	const id = "9110339d-72ba-4132-801f-cd2fd738b71d"
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Augmenter Pugilist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "As long as you control eight or more lands, this creature gets +5/+5.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.B10LandsControlled(g, source.Controller) >= 8
			},
			Apply: effects.PumpBy(5, 5),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Echoing Equation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			model, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			var others []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.Controller == ctx.Controller() && c.InstanceID != model {
					others = append(others, c.InstanceID)
				}
			}
			return effects.BecomeCopy{
				Targets: others,
				Of:      model,
				Except:  func(v *game.PrintedValues) { v.RemoveSupertype("Legendary") },
				Label:   "Echoing Equation — becomes a copy until end of turn",
			}.Apply(ctx)
		},
	})
}
