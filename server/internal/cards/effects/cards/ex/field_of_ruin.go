package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Field of Ruin — Land (EDHREC rank 896):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Sacrifice this land: Destroy target nonbasic land an
//	 opponent controls. Each player searches their library for a
//	 basic land card, puts it onto the battlefield, then shuffles."
//
// Demolition Field's older sibling: the same three-component cost
// and three-predicate target clause, but the searches are not "may"s
// and they are for EVERY player at the table, not just the two
// parties. One mandatory SearchLibrary per seated player, in seat
// order; each prompt is addressed to its own player and they can
// all sit open at once, because each is over a different library.
// A player whose library holds at most one basic is served without
// a prompt; a player with none finds nothing and still shuffles, as
// printed ("then shuffles" is unconditional).
//
// The basics enter UNTAPPED — the printed text says so.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f825c98f-a327-440b-8c0d-ebe02e23bfb7",
		Name:         "Field of Ruin",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this land: Destroy target nonbasic land an opponent controls. Each player searches for a basic land.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target nonbasic land an opponent controls", effects.Land(), effects.B03Nonbasic(), effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				if err := (effects.DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
				for _, p := range g.Seats {
					if p == nil || p.Eliminated {
						continue
					}
					if err := effects.B07SearchBasicOntoBattlefield(g, item, p.ID, false, false,
						"Field of Ruin — search for a basic land"); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
