package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Hunger Tide Rises — Enchantment — Saga {2}{G} (Modern Horizons 3):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III — Create a 1/1 black and green Insect creature token.
//	 IV — Sacrifice any number of creatures. Search your library and/or
//	 graveyard for a creature card with mana value less than or equal to the
//	 number of creatures sacrificed this way and put it onto the
//	 battlefield. If you search your library this way, shuffle."
//
// I-III are one body on three chapters (The Misty Mountains Cold's shape).
// IV is Plunge into Darkness's "any number" sacrifice (ChoosePermanents, a
// choice made as the chapter resolves) with Finale of Devastation's
// "library and/or graveyard" search: the player picks which zones to search
// before anything is shown, the library is shuffled only if it was
// searched, and X is the number of creatures that really left the
// battlefield. Sacrificing none still searches, for a creature card with
// mana value 0 or less. The Saga's own sacrifice after IV is the engine's
// (CR 714.4).
//
// No simplification.
func init() {
	const label = "The Hunger Tide Rises — create a 1/1 black and green Insect creature token"
	chapters := make([]game.TriggeredAbility, 0, 4)
	for n := 1; n <= 3; n++ {
		chapters = append(chapters, ChapterTrigger(n, label, Do(CreateToken{Template: TokenCard("1/1 black and green Insect"), N: 1})))
	}
	chapters = append(chapters, ChapterTrigger(4,
		"The Hunger Tide Rises — sacrifice any number of creatures; search your library and/or graveyard for a creature card with mana value up to that number and put it onto the battlefield",
		theHungerTideRisesSacrifice))
	Register(Spec{
		OracleID:     "356edd68-7dd5-4ab6-a738-6cb0eb2d4e56",
		Name:         "The Hunger Tide Rises",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		Triggered:    chapters,
	})
}

// theHungerTideRisesSacrifice is chapter IV's first sentence: the "any
// number" pick over your own creatures, no floor and no ceiling.
func theHungerTideRisesSacrifice(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question: "The Hunger Tide Rises — sacrifice any number of creatures",
		Keyed:    theHungerTideRisesSacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return permanentsControlledByMatching(g, of, Creature()), 0, 0
		},
	}.Apply(NewContext(g, item))
}

// theHungerTideRisesSacrificed is the sacrifice and then the search, keyed:
// X is the number that really left. An on-disk key.
var theHungerTideRisesSacrificed = RegisterPermanentsPick("prompt-run/gap-b55-the-hunger-tide-rises-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		x := picked.Count()
		return SearchLibrary{
			Player:        ctx.Controller(),
			Predicate:     func(c game.Card) bool { return c.IsCreature() && c.ManaValue() <= x },
			Dest:          game.ZoneBattlefield,
			Limit:         1,
			Shuffle:       true,
			AlsoGraveyard: true,
			Reason:        "The Hunger Tide Rises — a creature card with mana value up to the number of creatures sacrificed",
		}.Apply(ctx)
	},
})
