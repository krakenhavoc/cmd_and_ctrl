package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
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
			game.KeyedTail{Ref: effects.GapB92ShuffleThenInvestigate, Player: owner, Source: item.SourceCardID})
	}
	angel := effects.Do(effects.CreateToken{Template: effects.AlienAngelToken(), N: 1})
	target := func() *game.TargetSpec { return effects.TargetCreature("target creature") }
	effects.Register(effects.Spec{
		OracleID:     "f5fba11b-7669-40a0-97c7-6cb7c13acedd",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.GapB92ChapterLabel(name, 1, "choose target creature; its owner shuffles it into their library, then investigates"), target(), shuffleAway),
			effects.ChapterTrigger(2, effects.GapB92ChapterLabel(name, 2, "create a 2/2 black Alien Angel artifact creature token"), angel),
			effects.ChapterTriggerTargeting(3, effects.GapB92ChapterLabel(name, 3, "choose target creature; its owner shuffles it into their library, then investigates"), target(), shuffleAway),
			effects.ChapterTrigger(4, effects.GapB92ChapterLabel(name, 4, "create a 2/2 black Alien Angel artifact creature token"), angel),
		},
	})
}
