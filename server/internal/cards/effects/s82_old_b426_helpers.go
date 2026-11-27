package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b426_helpers.go — helpers for Return to Ravnica, batch 426 (S82
// older sets, #5566).

// OldB426SlaughterGames is Slaughter Games' body: "Choose a nonland card name.
// Search target opponent's graveyard, hand, and library for any number of cards
// with that name and exile them. Then that player shuffles." Infinite
// Obliteration's search over a name its caster chooses as the spell resolves
// (OldB321InfiniteObliteration), with "nonland" where that says "creature". The
// name is free text, so "nonland card" is kept where it is read: a name only a
// land has finds nothing. A target gone by resolution makes the spell do
// nothing and no name is asked (CR 608.2b).
func OldB426SlaughterGames(label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		victim := TargetedPlayerOf(ctx)
		if victim == uuid.Nil {
			return nil
		}
		ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
			label+" — choose a nonland card name", WithCarry(ctx.ItemTail(oldB426Named), victim))
		return nil
	}
}

// oldB426Named is the search once the name (the tail's Word) is known. An
// on-disk key: never renamed, never reused.
var oldB426Named = CardTail("card-name/s82-old-b426-slaughter-games", func(ctx *Context, o game.TailOutcome) error {
	name := strings.TrimSpace(o.Tail.Word)
	victim := TailCarry0(o)
	if victim == uuid.Nil {
		return nil
	}
	return SearchLibrary{
		Player:       ctx.Controller(),
		LibraryOwner: victim,
		Predicate: func(c game.Card) bool {
			return name != "" && !c.FaceDown && !c.IsLand() && game.CardNameMatches(c, name)
		},
		Dest:                 game.ZoneExile,
		Unbounded:            true,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "Slaughter Games — search their graveyard, hand and library for nonland cards named " + name,
	}.Apply(ctx)
})
