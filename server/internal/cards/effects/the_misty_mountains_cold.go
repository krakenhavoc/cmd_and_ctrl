package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Misty Mountains Cold — Enchantment — Saga {2}{R} (The Hobbit):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III, IV — Create a Treasure token. Then if you control four
//	 or more Treasures, sacrifice this Saga. If you do, create a 6/6 red
//	 Dragon creature token with flying."
//
// Four chapters with one body. The Treasure is made first and the count
// is read after it exists, so the new Treasure counts (a Treasure token
// is a permanent with the subtype Treasure, however it came to be). The
// sacrifice is mandatory once the count is met, and the Dragon follows
// only if the Saga actually went (SacrificeThisThen: "if you do"): a Saga
// that already left the battlefield in response makes no Dragon. The
// Dragon is Utvara Hellkite's token at size six.
//
// No simplification.
func init() {
	const label = "The Misty Mountains Cold — create a Treasure; with four or more Treasures, sacrifice this Saga and create a 6/6 red Dragon with flying"
	chapters := make([]game.TriggeredAbility, 0, 4)
	for n := 1; n <= 4; n++ {
		chapters = append(chapters, ChapterTrigger(n, label, mistyMountainsColdChapter))
	}
	Register(Spec{
		OracleID:     "c50282bd-4b57-4c7b-9598-ea67ba49248b",
		Name:         "The Misty Mountains Cold",
		Completeness: CompletenessFull,
		Triggered:    chapters,
	})
}

// mistyMountainsColdSacrifice is "sacrifice this Saga. If you do, create a
// 6/6 red Dragon creature token with flying". An on-disk key.
var mistyMountainsColdSacrifice = SacrificeThisThen("then/the-misty-mountains-cold-dragon", func(ctx *Context) error {
	return CreateToken{Template: b14RedDragonToken(6), N: 1}.Apply(ctx)
})

// mistyMountainsColdChapter is every chapter: a Treasure, then the count.
func mistyMountainsColdChapter(g *game.Game, item *game.StackItem) error {
	if err := (CreateToken{Template: TreasureToken(), N: 1}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	if len(permanentsControlledByMatching(g, item.Controller, Subtype("Treasure"))) < 4 {
		return nil
	}
	return mistyMountainsColdSacrifice(g, item)
}
