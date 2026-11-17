package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kethek, Crucible Goliath — Legendary Creature — Phyrexian Beast
// {2}{B}{R}, 4/4 (ONE):
//
//	"At the beginning of your end step, you may sacrifice another
//	 creature. If you do, reveal cards from the top of your library until
//	 you reveal a nonlegendary creature card with lesser mana value, put
//	 it onto the battlefield, then put the rest on the bottom of your
//	 library in a random order."
//
// The "may" is the sacrifice prompt, asked on resolution (CR 603.5); the
// "if you do" is its keyed continuation, run only for a creature that
// really left. "Lesser" is read off the sacrificed creature as it last
// existed on the battlefield (CR 608.2h, departedManaValue). The reveal
// run is the shared reveal-until-put-onto-the-battlefield sentence: the
// hit enters, the rest go to the bottom in a random order, and a library
// with no such card puts every revealed card on the bottom.
//
// No simplification.
func init() {
	const label = "Kethek, Crucible Goliath — you may sacrifice another creature; if you do, reveal until a nonlegendary creature card with lesser mana value and put it onto the battlefield"
	Register(Spec{
		OracleID:     "26747bdc-6bd4-4218-ae83-bd46c7668880",
		Name:         "Kethek, Crucible Goliath",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep(label, func(g *game.Game, item *game.StackItem) error {
				return maySacrificeOne(g, item, "Kethek, Crucible Goliath — sacrifice another creature?",
					Creature(), true, kethekSacrificed)
			}),
		},
	})
}

// kethekSacrificed is "If you do, reveal cards from the top of your
// library until you reveal a nonlegendary creature card with lesser mana
// value": keyed, so a table waiting on the sacrifice is a restore point.
var kethekSacrificed = RegisterSacrificedOne("prompt-run/kethek-crucible-goliath-sacrifice", func(ctx *Context, sacrificed uuid.UUID) error {
	mv := departedManaValue(ctx.Game, sacrificed)
	return RevealUntilThenPutOntoBattlefield{
		Match: func(c game.Card) bool {
			return c.IsCreature() && !c.IsLegendary() && c.ManaValue() < mv
		},
		Reason: "Kethek, Crucible Goliath — a nonlegendary creature card with lesser mana value",
	}.Apply(ctx)
})
