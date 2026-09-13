package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Millikin — Artifact Creature — Construct {2}, 0/1 (EDHREC rank
// 1213):
//
//	"{T}, Mill a card: Add {C}. (Activate only as an instant. To mill
//	 a card, put the top card of your library into your graveyard.)"
//
// A mana Myr that feeds the graveyard — the self-mill deck's rock.
// ManaAbilityCost has no "mill a card" component, so the cost is
// built from the two slots it does have: Condition gates the
// activation on a non-empty library (CR 601.2h / 701.13a — a cost
// that cannot be paid cannot be activated, and with no card to mill
// there is nothing to pay), and Rider mills the card the moment the
// mana lands. The order within one atomic mana-ability resolution is
// not observable, and a "whenever you mill" payoff sees the same
// EventMill either way.
//
// A Rider drops the ability out of auto-tap planning, which is the
// right call twice over: auto-tap should not mill a library to pay
// for a spell, and the tap is on a CREATURE, so summoning sickness
// applies (CR 302.1) — the engine enforces that from the type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa4dffda-6f04-4d0b-829d-28a1a5794dee",
		Name:         "Millikin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}, Mill a card: Add {C}",
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				p := g.PlayerByIDForEffect(controller)
				return p != nil && p.Library != nil && p.Library.Size() > 0
			},
			Rider: func(g *game.Game, controller, _ uuid.UUID) error {
				return g.MillNForEffect(controller, 1)
			},
		}},
	})
}
