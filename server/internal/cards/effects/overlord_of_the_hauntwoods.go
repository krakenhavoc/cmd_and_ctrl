package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overlord of the Hauntwoods — Enchantment Creature — Avatar Horror
// {3}{G}{G}:
//
//	"Impending 4—{1}{G}{G} (If you cast this spell for its impending
//	 cost, it enters with four time counters and isn't a creature
//	 until the last is removed. At the beginning of your end step,
//	 remove a time counter from it.)
//	 Whenever this permanent enters or attacks, create a tapped
//	 colorless land token named Everywhere that is every basic land
//	 type."
//
// Impending is the shared constructor (cost, the not-a-creature static
// and the countdown); what the card adds is the trigger and the token.
// "Everywhere" is a land token with all five basic land types and no
// colour. Its mana is not declared anywhere: the engine derives a
// land's intrinsic mana abilities from its effective subtypes
// (CR 305.6), so the token taps for any of the five colours. It enters
// tapped through the creation's own entry clause, so an enters-tapped
// replacement and a token doubler see it like any other token.
//
// While impending the Overlord is not a creature and so cannot attack;
// the enter half still fires, which is the point of paying {1}{G}{G}.
//
// No simplification.
func init() {
	Register(Impending(4, "{1}{G}{G}", Spec{
		OracleID:     "4669e8c7-fc37-4b97-9cb7-8d29b43d9176",
		Name:         "Overlord of the Hauntwoods",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Overlord of the Hauntwoods — create a tapped colorless land token named Everywhere",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenAdvanced{
						Controller: item.Controller,
						Spec:       Token(everywhereToken()).EntersTapped(),
						N:          1,
					}.Apply(NewContext(g, item))
				}),
		},
	}))
}

// everywhereToken is the colorless land token named Everywhere that is
// every basic land type.
func everywhereToken() game.Card {
	return game.Card{
		Name:     "Everywhere",
		TypeLine: "Token Land — Plains Island Swamp Mountain Forest",
	}
}
