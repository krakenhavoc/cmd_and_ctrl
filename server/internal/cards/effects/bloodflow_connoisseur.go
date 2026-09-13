package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodflow Connoisseur — Creature — Vampire {2}{B}, 1/1:
//
//	"Sacrifice a creature: Put a +1/+1 counter on this creature."
//
// Carrion Feeder's twin at one more mana, and the bridge between
// this sprint's two halves: a free sacrifice outlet whose payoff is
// a +1/+1 counter, which is then something to PROLIFERATE. Karn's
// Bastion plus this is a mana sink that eats a token and grows two
// counters.
//
// Like the Feeder, the ability can eat the Connoisseur itself: legal
// and pointless, because the counter lands on a card that is already
// in the graveyard and AddCounter no-ops there.
func init() {
	Register(Spec{
		OracleID:     "ddb2fb87-235a-4365-aa25-40c197425e43",
		Name:         "Bloodflow Connoisseur",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "Sacrifice a creature: put a +1/+1 counter on this creature",
			Cost:  SacrificeACreature(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return AddCounter{
					Target: ctx.Source(),
					Kind:   game.CounterPlusOne,
					N:      1,
				}.Apply(ctx)
			},
		}},
	})
}
