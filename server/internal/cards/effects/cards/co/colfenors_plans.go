package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Colfenor's Plans — Enchantment {2}{B}{B} (Lorwyn, S82 OLD b546, #5593):
//
//	"When this enchantment enters, exile the top seven cards of your library
//	 face down.
//	 You may look at the cards exiled with this enchantment, and you may
//	 play lands and cast spells from among those cards.
//	 Skip your draw step.
//	 You can't cast more than one spell each turn."
//
// The enters trigger looks at the top seven and exiles each face down, linked
// to this object (CR 607.2a, CR 406.3; ExileHiddenLinkedForEffect: only its
// controller may look). A standing cast permission over exile, narrowed to the
// cards exiled with this enchantment, lets its controller play lands and cast
// spells from them whenever they could from hand, paying costs as usual. The
// draw step is skipped by Solitary Confinement's replacement, and the
// one-spell limit is Moderation's cast restriction, which counts every spell
// the player has cast this turn, from any zone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad786230-4045-46af-bc89-078d45891e5b",
		Name:         "Colfenor's Plans",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Colfenor's Plans — exile the top seven cards of your library face down",
				func(g *game.Game, item *game.StackItem) error {
					obj, ok := g.SourceObjectForEffect(item)
					if !ok {
						return nil
					}
					ref := game.PermissionCardRef(obj)
					for _, id := range g.LookAtTopOfLibraryForEffect(item.Controller, 7) {
						if _, err := g.ExileHiddenLinkedForEffect(ref, item.Controller, id); err != nil {
							return err
						}
					}
					return nil
				}),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: game.CastPermission{
				Zone:   game.ZoneExile,
				Filter: game.PermissionFilter{ExiledWithSource: true},
				Label:  "You may play lands and cast spells from among the cards exiled with this enchantment (Colfenor's Plans)",
			},
		}},
		Replacements: []game.ReplacementEffect{effects.SkipYourDrawStep()},
		CastRestrictions: []game.CastRestriction{{
			Label: "Colfenor's Plans — you can't cast more than one spell each turn.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.Controller == q.Controller && q.Game.CastTallyFor(q.Controller).Total >= 1
			},
		}},
	})
}
