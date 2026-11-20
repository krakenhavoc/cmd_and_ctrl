package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost Quarter — Land (EDHREC rank 823):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Destroy target land. Its controller
//	 may search their library for a basic land card, put it onto the
//	 battlefield, then shuffle."
//
// Strip Mine with a consolation prize — the utility-land answer that
// costs the victim nothing but their Cabal Coffers. Strip Mine's
// tap-and-sacrifice cost and land target; the "may search" is
// Demolition Field's victim-side Optional SearchLibrary, addressed
// to whoever controlled the land, which can be the activator (a
// Ghost Quarter on your own land is the classic "fetch a basic"
// line and works here). The controller is read BEFORE the destroy,
// because afterwards the land is in a graveyard and its controller
// field no longer says who had it. The search is not conditional on
// the land actually dying: an indestructible land's controller still
// gets the offer, as printed.
//
// The basic enters UNTAPPED — the printed text says so.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ec4288e-34c6-4831-a2c0-ba1ca1d9d1dc",
		Name:         "Ghost Quarter",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Destroy target land. Its controller may search for a basic land.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				target := item.Targets[0].ID
				victim, ok := effects.ControllerOfTarget(ctx, target)
				if !ok {
					return nil
				}
				if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
					return err
				}
				return effects.B07SearchBasicOntoBattlefield(g, item, victim, false, true,
					"Ghost Quarter — you may search for a basic land")
			},
		}},
	})
}
