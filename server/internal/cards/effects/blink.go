package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blink — Enchantment — Saga {2}{U}{B} (Doctor Who, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, III — Choose target creature. Its owner shuffles it into their
//	 library, then investigates.
//	 II, IV — Create a 2/2 black Alien Angel artifact creature token with first
//	 strike, vigilance, and 'Whenever an opponent casts a creature spell, this
//	 token isn't a creature until end of turn.'"
//
// Chapters I and III put the target into its OWNER's library (a token ceases
// to exist) and then that owner shuffles and investigates, even when the
// creature was a token or has gone (the owner is read as the ability
// resolves from the target's last-known record). Chapters II and IV create
// the Alien Angel that Aplan Mortarium also makes.
//
// No simplification.
func init() {
	const name = "Blink"
	shuffleAway := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		owner := c.Owner
		if owner == uuid.Nil {
			owner = c.Controller
		}
		return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{},
			game.KeyedTail{Ref: gapB92ShuffleThenInvestigate, Player: owner, Source: item.SourceCardID})
	}
	angel := Do(CreateToken{Template: AlienAngelToken(), N: 1})
	target := func() *game.TargetSpec { return TargetCreature("target creature") }
	Register(Spec{
		OracleID:     "f5fba11b-7669-40a0-97c7-6cb7c13acedd",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, gapB92ChapterLabel(name, 1, "choose target creature; its owner shuffles it into their library, then investigates"), target(), shuffleAway),
			ChapterTrigger(2, gapB92ChapterLabel(name, 2, "create a 2/2 black Alien Angel artifact creature token"), angel),
			ChapterTriggerTargeting(3, gapB92ChapterLabel(name, 3, "choose target creature; its owner shuffles it into their library, then investigates"), target(), shuffleAway),
			ChapterTrigger(4, gapB92ChapterLabel(name, 4, "create a 2/2 black Alien Angel artifact creature token"), angel),
		},
	})
}
