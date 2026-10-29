package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Down in the Valley — Enchantment — Saga {2}{G} (The Hobbit):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Search your library for a basic land card, reveal it, put it
//	 into your hand, then shuffle.
//	 II — This Saga gains "Landfall — Whenever a land you control enters,
//	 create a 1/1 green Elf creature token."
//	 III, IV — Elves you control get +1/+0 and gain vigilance until end
//	 of turn."
//
// Chapter II is a duration grant with no stated duration (ADR 0093,
// CR 611.2a), pinned to this Saga as Urza's Saga's chapters are: the
// landfall trigger is the Saga's own until it leaves the battlefield, and
// a Saga that comes back is a new object without it (CR 400.7). The Elf
// is Thranduil, the Strategist's token. Chapters III and IV fix the set
// of Elves as they resolve (CR 611.2c): an Elf that enters later in the
// turn gets nothing.
//
// No simplification.
const downInTheValleyGrant = "down-in-the-valley/landfall-elf"

func init() {
	elves := And(OfCreatureType("Elf"), YouControl())
	Register(Spec{
		OracleID:     "e3491542-569e-48a8-b625-fa7c4aa2792a",
		Name:         "Down in the Valley",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: downInTheValleyGrant,
			Triggered: []game.TriggeredAbility{
				LandfallPaying("Down in the Valley — create a 1/1 green Elf (landfall)", game.LandfallPayoff{Tokens: 1},
					b34CreateTokens(thranduilElfToken, 1)),
			},
			Text: "Landfall — Whenever a land you control enters, create a 1/1 green Elf creature token.",
		}},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "Down in the Valley — search for a basic land card, reveal it, put it into your hand", downInTheValleySearch),
			ChapterTrigger(2, "Down in the Valley — this Saga gains the landfall Elf trigger", downInTheValleyGains),
			ChapterTrigger(3, "Down in the Valley — Elves you control get +1/+0 and vigilance until end of turn", downInTheValleyRally(elves)),
			ChapterTrigger(4, "Down in the Valley — Elves you control get +1/+0 and vigilance until end of turn", downInTheValleyRally(elves)),
		},
	})
}

// downInTheValleySearch is chapter I.
func downInTheValleySearch(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: IsBasicLand,
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Down in the Valley — a basic land card",
		Source:    item.SourceCardID,
	}.Apply(NewContext(g, item))
}

// downInTheValleyGains is chapter II: the Saga gives itself the trigger
// for the rest of the game, pinned to this object.
func downInTheValleyGains(g *game.Game, item *game.StackItem) error {
	if sourceIsNewObject(g, item) {
		return nil // "this Saga" left and came back: a new object
	}
	saga := item.SourceCardID
	return GrantAbilitiesFor{
		Target:   saga,
		Keys:     []string{downInTheValleyGrant},
		Duration: g.PinnedTo(game.IndefiniteDuration(), saga),
		Label:    "Down in the Valley — gains landfall: create a 1/1 green Elf",
	}.Apply(NewContext(g, item))
}

// downInTheValleyRally is chapters III and IV: +1/+0 and vigilance for the
// Elves its controller has now.
func downInTheValleyRally(elves CardPredicate) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := (BoostUntilEOT{Match: elves, Power: 1, Label: "Down in the Valley — Elves you control get +1/+0"}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Match: elves, Keywords: []string{"vigilance"}, Label: "Down in the Valley — Elves you control gain vigilance"}.Apply(ctx)
	}
}
