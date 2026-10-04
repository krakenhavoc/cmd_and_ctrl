package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// There and Back Again — Enchantment — Saga {3}{R}{R}:
//
//	"(As this Saga enters and after your draw step, add a lore
//	 counter. Sacrifice after III.)
//	 I — Up to one target creature can't block for as long as you
//	     control this Saga. The Ring tempts you.
//	 II — Search your library for a Mountain card, put it onto the
//	     battlefield, then shuffle.
//	 III — Create Smaug, a legendary 6/6 red Dragon creature token with
//	     flying, haste, and "When Smaug dies, create fourteen Treasure
//	     tokens.""
//
// Chapter I's restriction lasts while you control the Saga (CR
// 611.2b): it ends when the Saga is sacrificed after chapter III. If
// you no longer control the Saga as chapter I resolves, the
// restriction never begins; the Ring still tempts. Smaug is a token
// with its own printed trigger (SmaugToken, tokens.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de4e5120-b958-4dcb-ad78-d0f726b7a881",
		Name:         "There and Back Again",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "There and Back Again — a creature can't block, then the Ring tempts you",
				TargetCreature("up to one target creature").WithCount(0, 1), thereAndBackAgainCantBlock),
			ChapterTrigger(2, "There and Back Again — search for a Mountain card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype("Mountain") },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "There and Back Again — a Mountain card",
				}.Apply(NewContext(g, item))
			}),
			ChapterTrigger(3, "There and Back Again — create Smaug", Do(CreateToken{Template: SmaugToken(), N: 1})),
		},
	})
}

// thereAndBackAgainCantBlock is chapter I.
func thereAndBackAgainCantBlock(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetIDs(ctx) {
		d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
		if !ok {
			break
		}
		if err := (ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBlock)},
			Duration: d,
			Label:    "There and Back Again — can't block",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return TheRingTemptsYou{}.Apply(ctx)
}
