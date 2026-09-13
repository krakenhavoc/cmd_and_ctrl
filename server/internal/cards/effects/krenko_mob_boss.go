package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krenko, Mob Boss — 3/3 Legendary Creature — Goblin Warrior for
// {2}{R}{R}:
//
//	"{T}: Create X 1/1 red Goblin creature tokens, where X is the
//	number of Goblins you control."
//
// S21 sub-PR 2's tap-cost card: it exercises summoning sickness on
// the cost (a freshly-cast Krenko can't tap) and a count taken at
// RESOLUTION, not announce — a Goblin that dies in response makes
// the ability produce one fewer token, and Krenko counts himself.
func init() {
	Register(Spec{
		OracleID:     "68418069-f615-40ef-ae0d-764192acae00",
		Name:         "Krenko, Mob Boss",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Create X 1/1 Goblins, where X is the number of Goblins you control",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && isGoblin(c) {
						n++
					}
				}
				if n == 0 {
					return nil
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   RedGoblinToken(),
					N:          n,
				}.Apply(ctx)
			},
		}},
	})
}

// isGoblin reports whether a card is a Goblin creature. Reads the
// post-layer subtypes so a creature that was turned into a Goblin
// counts (CR 205.3 — type-changing effects apply).
func isGoblin(c game.Card) bool {
	return c.IsCreature() && c.HasSubtype("Goblin")
}
