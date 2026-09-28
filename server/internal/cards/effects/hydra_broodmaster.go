package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hydra Broodmaster — 7/7 Hydra for {4}{G}{G}:
//
//	"{X}{X}{G}: Monstrosity X. (If this creature isn't monstrous, put X
//	 +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, create X X/X green Hydra
//	 creature tokens."
//
// X is the monstrosity's announced X (CR 701.37c), read off the
// triggering event — so a Doubling Season that doubles the counters
// does not change the tokens' size (it doubles the token creation
// instead, which is its own replacement).
//
// The token is built by hand, as Phyrexian Rebirth's is: a token
// table keyed on printed characteristics has no row for a size decided
// at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e4fb4b1-9b28-4a49-b4ac-d5549c08594d",
		Name:         "Hydra Broodmaster",
		Completeness: CompletenessFull,
		// X=0 makes no tokens and spends the monstrosity for good.
		XMatters:  true,
		Activated: []ActivatedAbility{MonstrosityX(ManaCost("{X}{X}{G}"))},
		Triggered: []game.TriggeredAbility{
			WhenBecomesMonstrous("Hydra Broodmaster — create X X/X Hydra tokens", hydraBroodmasterTokens),
		},
	})
}

func hydraBroodmasterTokens(g *game.Game, item *game.StackItem) error {
	x := MonstrosityXOf(item)
	if x <= 0 {
		return nil
	}
	return CreateToken{
		Controller: item.Controller,
		Template: game.Card{
			Name:      "Hydra",
			TypeLine:  "Token Creature — Hydra",
			Colors:    []string{"G"},
			Power:     x,
			Toughness: x,
		},
		N: x,
	}.Apply(NewContext(g, item))
}
