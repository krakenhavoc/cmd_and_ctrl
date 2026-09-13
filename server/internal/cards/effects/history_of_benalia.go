package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// History of Benalia — Enchantment — Saga for {1}{W}{W}:
//
//	"I, II — Create a 2/2 white Knight creature token with vigilance.
//	 III — Knights you control get +2/+1 until end of turn."
//
// The two identical early chapters are two declarations, not one
// with a count, because that is how the Saga machinery counts: a
// chapter fires when the lore counter LANDS on its number, so a Saga
// that gains two counters at once (Doubling Season on the way in)
// fires I and II separately and in order. One declaration covering
// "I, II" would fire once.
//
// Chapter III is a turn-scoped Layer 7c effect over the printed
// subtype, so it catches the two Knights this Saga made AND any
// other Knight on the board, and it expires at cleanup (CR 514.2).
// CR 611.2c: the affected set is snapshotted at resolution, so a
// Knight cast after the chapter resolves is not pumped.
func init() {
	Register(Spec{
		OracleID:     "c15bb7eb-aaaa-4468-9641-8f706d6137e8",
		Name:         "History of Benalia",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Knight token is created colorless instead of white, so anything that cares about a creature's color doesn't see it."},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "History of Benalia — I: create a 2/2 Knight", benaliaKnight),
			ChapterTrigger(2, "History of Benalia — II: create a 2/2 Knight", benaliaKnight),
			ChapterTrigger(3, "History of Benalia — III: Knights get +2/+1", benaliaPumpKnights),
		},
	})
}

func benaliaKnight(g *game.Game, item *game.StackItem) error {
	return CreateToken{
		Controller: item.Controller,
		Template:   KnightVigilanceToken(),
		N:          1,
	}.Apply(NewContext(g, item))
}

func benaliaPumpKnights(g *game.Game, item *game.StackItem) error {
	return BoostUntilEOT{
		Match:     And(Creature(), YouControl(), OfSubtype("Knight")),
		Power:     2,
		Toughness: 1,
		Label:     "History of Benalia — Knights get +2/+1",
	}.Apply(NewContext(g, item))
}
