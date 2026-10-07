package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fomori Vault — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}, Discard a card: Look at the top X cards of your library,
//	 where X is the number of artifacts you control. Put one of those
//	 cards into your hand and the rest on the bottom of your library in
//	 a random order."
//
// The discard is a cost, paid at announce (the card picker opens before
// the ability goes on the stack), and X is counted as the ability
// resolves — "where X is the number of artifacts you control" is not an
// announced X, so the artifacts that are there at resolution are the
// ones that count. Taking a card is mandatory when there is one to take;
// with no artifacts the ability looks at nothing and does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "622a53bd-d894-422c-a606-7126041afa02",
		Name:          "Fomori Vault",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}, Discard a card: Look at the top X cards of your library, where X is the number of artifacts you control. Put one of those cards into your hand and the rest on the bottom of your library in a random order",
			Cost:  Plus(ManaCost("{3}"), TapCost(), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Controller
				x := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == player && c.IsArtifact() {
						x++
					}
				}
				if x == 0 {
					return nil
				}
				return TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, x),
					Max:    1,
					Label:  "Fomori Vault — put one of those cards into your hand",
					Then:   TakeRestOnBottomInRandomOrder,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
