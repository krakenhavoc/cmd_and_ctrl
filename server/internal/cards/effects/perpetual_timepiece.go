package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perpetual Timepiece — Artifact {2} (EDHREC rank 2025):
//
//	"{T}: Mill two cards. (Put the top two cards of your library
//	 into your graveyard.)
//	 {2}, Exile this artifact: Shuffle any number of target cards
//	 from your graveyard into your library."
//
// The self-mill rock that doubles as graveyard insurance. The mill
// is live: a tap ability on the stack, two cards, every mill watcher
// sees them.
//
// DECLARED SIMPLIFICATION, the Dragon's Hoard posture: the shuffle
// ability is not implemented. "Exile this artifact" is a cost
// component the engine still cannot express, even after #1381 gave
// AbilityCost.ExileSelf a real cost shape: that field is scavenge's
// and embalm's "exile this card from your GRAVEYARD" (CR 702.96a /
// CR 702.128a), and both effects.Register (registry.go) and the
// activation-time validator (game.validateExileSelfCostLocked,
// exile_cost.go) hardcode the graveyard as the zone the source has to
// be in — confirmed by trying it here: Register panics with "does not
// function from the graveyard" regardless of what Zones the ability
// declares. Perpetual Timepiece's exile is a plain CR 406 move of a
// permanent OFF THE BATTLEFIELD as a cost, which is a different rule
// with no shape yet, and exile-self is not sacrifice either way (a
// sacrifice would fire every "whenever you sacrifice" payoff and put
// the card in the graveyard, both of which the printed card avoids on
// purpose). Shipping it with the exile deferred to resolution would
// leave the Timepiece on the battlefield with the ability on the
// stack, which is the #259 direction. The Timepiece mills and stops
// there until a battlefield exile-self cost component lands; weaker
// than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "17b4778b-82b1-4845-ad08-00f3ff66877b",
		Name:         "Perpetual Timepiece",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Exiling it to shuffle cards from your graveyard into your library isn't implemented — only the tap-to-mill works."},
		Activated: []ActivatedAbility{{
			Label: "{T}: Mill two cards.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return MillCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
