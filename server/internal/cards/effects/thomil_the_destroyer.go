package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thomil, the Destroyer — Legendary Planeswalker — Thomil {3}{B}{B}, loyalty 4
// (Mystery Booster Commander, #5948):
//
//	"+2: Create a 2/2 black Zombie creature token.
//	 0: You may sacrifice a creature. If you do, add {B}{B}{B}.
//	 −5: Create a Lord of the Pit token. (It's a {4}{B}{B}{B} 7/7 Demon
//	 creature with flying, trample, and "At the beginning of your upkeep,
//	 sacrifice another creature. If you can't, this token deals 7 damage to
//	 you.")
//	 Thomil, the Destroyer can be your commander."
//
// The 0 is Ashnod's maySacrificeOne asked as the ability resolves, with the
// mana added only for a creature that really went. The mana is added during
// the main phase and empties with it, as any mana from a resolving ability
// does. The Lord of the Pit is a catalog token (gapB05LordOfThePitToken).
// "Can be your commander" is deck-building and is read from the oracle text by
// the deck validator, not here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd769cfb-668c-4557-a907-c6c17539c96a",
		Name:         "Thomil, the Destroyer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Create a 2/2 black Zombie creature token.",
				Cost:    LoyaltyCost(2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}),
			},
			{
				Label: "0: You may sacrifice a creature. If you do, add {B}{B}{B}.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Thomil, the Destroyer — sacrifice a creature to add {B}{B}{B}?",
						Creature(), false, thomilAddBBB)
				},
			},
			{
				Label:   "−5: Create a Lord of the Pit token.",
				Cost:    LoyaltyCost(-5),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  Do(CreateToken{Template: gapB05LordOfThePitToken(), N: 1}),
			},
		},
	})
}

// thomilAddBBB is "if you do, add {B}{B}{B}". An on-disk key.
var thomilAddBBB = RegisterSacrificedOne("prompt-run/thomil-the-destroyer-add-bbb", func(ctx *Context, _ uuid.UUID) error {
	return AddMana{Player: ctx.Controller(), Produced: "{B}{B}{B}"}.Apply(ctx)
})
