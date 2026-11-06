package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightmares and Daydreams — Enchantment — Saga {2}{U} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III — Until your next turn, whenever you cast an instant or
//	 sorcery spell, target player mills cards equal to that spell's mana
//	 value.
//	 IV — Draw a card. If a graveyard has twenty or more cards in it,
//	 draw three cards instead."
//
// Chapters I to III each schedule a delayed trigger (CR 603.7) held by
// the chapter's controller, firing on every instant or sorcery spell
// they cast until their next turn begins (avB29NightmaresChapter). Each
// fire targets a player as it goes on the stack and mills them the
// spell's mana value as it resolves (X counted while the spell is still
// on the stack). Chapter IV counts every graveyard as it resolves.
//
// No simplification.
func init() {
	const name = "Nightmares and Daydreams"
	const mills = "until your next turn, whenever you cast an instant or sorcery spell, target player mills cards equal to that spell's mana value"
	Register(Spec{
		OracleID:     "0403bfb0-2174-4360-994d-68d8ca96fc55",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, mills), avB29NightmaresChapter),
			ChapterTrigger(2, SagaChapterLabel(name, 2, mills), avB29NightmaresChapter),
			ChapterTrigger(3, SagaChapterLabel(name, 3, mills), avB29NightmaresChapter),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "draw a card, or three cards if a graveyard has twenty or more cards in it"),
				func(g *game.Game, item *game.StackItem) error {
					n := 1
					if avB29AGraveyardHasTwenty(g) {
						n = 3
					}
					return DrawCards{N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
