package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Birth of Meletis — Enchantment — Saga for {1}{W}:
//
//	"I — Search your library for a basic Plains card, reveal it, put
//	     it into your hand, then shuffle.
//	 II — Create a 0/4 colorless Wall artifact creature token with
//	      defender.
//	 III — You gain 2 life."
//
// Three chapters, three existing primitives — the simplest Saga in
// the S27 batch and the one the lifecycle tests lean on, because
// nothing in it can fail for a reason other than the Saga machinery.
func init() {
	Register(Spec{
		OracleID: "1ae54fe7-b1d3-4c13-a8ef-f502cf3eb1a0",
		Name:     "The Birth of Meletis",
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Birth of Meletis — I: search for a basic Plains", meletisFetchPlains),
			ChapterTrigger(2, "The Birth of Meletis — II: create a 0/4 Wall", meletisWall),
			ChapterTrigger(3, "The Birth of Meletis — III: gain 2 life", meletisGainLife),
		},
	})
}

func meletisFetchPlains(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: isBasicPlains,
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "The Birth of Meletis — search for a basic Plains",
	}.Apply(NewContext(g, item))
}

// isBasicPlains matches a BASIC Plains, both halves enforced: the
// printed text names the supertype as well as the land type, so a
// Sacred Foundry is not a legal fetch even though it is a Plains.
func isBasicPlains(c game.Card) bool {
	return IsBasicLand(c) && c.HasSubtype("Plains")
}

func meletisWall(g *game.Game, item *game.StackItem) error {
	return CreateToken{
		Controller: item.Controller,
		Template:   WallDefenderToken(),
		N:          1,
	}.Apply(NewContext(g, item))
}

func meletisGainLife(g *game.Game, item *game.StackItem) error {
	return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
}
