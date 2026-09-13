package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f825c98f-a327-440b-8c0d-ebe02e23bfb7",
		Name:         "Field of Ruin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this land: Destroy target nonbasic land an opponent controls. Each player searches for a basic land.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target nonbasic land an opponent controls", Land(), b03Nonbasic(), OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
				for _, p := range g.Seats {
					if p == nil || p.Eliminated {
						continue
					}
					if err := b07SearchBasicOntoBattlefield(g, item, p.ID, false, false,
						"Field of Ruin — search for a basic land"); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
