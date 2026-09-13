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
// component the engine cannot express — AbilityCost carries tap,
// sacrifice, mana, life, loyalty and crew, and exile-self is not
// sacrifice (a sacrifice would fire every "whenever you sacrifice"
// payoff and put the card in the graveyard, both of which the
// printed card avoids on purpose). Shipping it with the exile
// deferred to resolution would leave the Timepiece on the
// battlefield with the ability on the stack, which is the #259
// direction. The Timepiece mills and stops there until an exile-self
// cost lands; weaker than printed, never stronger.
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
