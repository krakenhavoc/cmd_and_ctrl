package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Promise of Aclazotz // Foul Rebirth — an Adventure card, Lost Caverns of
// Ixalan Commander reprint (#5368):
//
//	Promise of Aclazotz {1}{B}, Enchantment:
//	"At the beginning of your end step, you may sacrifice a non-Demon
//	 creature. If you do, populate."
//	Foul Rebirth {2}{B}, Sorcery — Adventure:
//	"Sacrifice a non-Demon creature. If you do, create a 4/3 white and
//	 black Vampire Demon creature token with flying."
//
// Both sacrifices are asked as the ability or spell resolves (CR 608.2c),
// over the non-Demon creatures you control then, read through the layers
// (Dreadfeast Demon's shape); "if you do" runs only for a creature that
// really left (CR 701.21a). The Promise's is optional (a floor of zero),
// Foul Rebirth's is not, and with no non-Demon creature neither makes
// anything. Populate copies a creature token you control, the sacrificed
// one included if it was a token and was the only choice — it is gone by
// then (CR 701.36), so it can't be copied. The engine owns the adventure:
// the sorcery resolves into exile and the Promise may be cast from there.
//
// No simplification.
func init() {
	const oracleID = "4ef4b4f9-47cd-4862-8cbb-b6e624d6e6fc"
	Register(Spec{
		OracleID:     oracleID,
		Name:         "Promise of Aclazotz",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Promise of Aclazotz — you may sacrifice a non-Demon creature; if you do, populate",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Promise of Aclazotz — sacrifice a non-Demon creature to populate?",
						And(Creature(), Not(HasSubtype("Demon"))), false, cmdr2B09PromisePopulate)
				}),
		},
	})
	Register(Spec{
		OracleID:     oracleID + "#1",
		Name:         "Foul Rebirth",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return mustSacrificeOne(ctx.Game, item, "Foul Rebirth — sacrifice a non-Demon creature",
				And(Creature(), Not(HasSubtype("Demon"))), cmdr2B09FoulRebirthToken)
		},
	})
}

var cmdr2B09PromisePopulate = RegisterSacrificedOne("prompt-run/cmdr2-b09-promise-of-aclazotz-populate", func(ctx *Context, _ uuid.UUID) error {
	return Populate{}.Apply(ctx)
})

var cmdr2B09FoulRebirthToken = RegisterSacrificedOne("prompt-run/cmdr2-b09-foul-rebirth-token", func(ctx *Context, _ uuid.UUID) error {
	return CreateToken{Template: TokenCard("4/3 white and black Vampire Demon with flying"), N: 1}.Apply(ctx)
})
