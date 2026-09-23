package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Legend of Roku — Enchantment — Saga, {2}{R}{R}:
//
//	(As this Saga enters and after your draw step, add a lore counter.)
//	I — Exile the top three cards of your library. Until the end of
//	    your next turn, you may play those cards.
//	II — Add one mana of any color.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// The FRONT face of a transforming Saga (ADR 0034, ADR 0079). Chapter
// III is the exile-and-return verb, as on every other transforming
// Saga in this batch; see avatar_roku.go for the back face.
//
// Chapter I is Reckless Impulse's shape with the count changed from
// two to three: exile off the top, grant lets the controller PLAY
// (not just cast, so a land isn't stranded) until the end of their
// next turn — ADR 0063's seat-turn counter, so "your next turn" means
// the same thing regardless of table order.
//
// Chapter II is Dark Ritual's primitive with a colour pick instead of
// a fixed colour: AddMana with a bare pipe slot queues the ordinary
// mana_pick prompt, one colour, one mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     theLegendOfRokuOracleID,
		Name:         "The Legend of Roku",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Legend of Roku — exile the top three cards of your library",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopNUntilYourNextTurn(NewContext(g, item), 3)
				}),
			ChapterTrigger(2, "The Legend of Roku — add one mana of any color",
				Do(AddMana{Produced: "{W|U|B|R|G}"})),
			ChapterTrigger(3, "The Legend of Roku — exile it, then return it transformed",
				ChapterExileAndReturnTransformed),
		},
	})
}

// theLegendOfRokuOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const theLegendOfRokuOracleID = "2f26f270-26d9-46d3-957f-19f52d51eb03"
