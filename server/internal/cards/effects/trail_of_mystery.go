package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trail of Mystery — Enchantment {1}{G} (Duskmourn Commander):
//
//	"Whenever a face-down creature you control enters, you may search your
//	 library for a basic land card, reveal it, put it into your hand, then
//	 shuffle.
//	 Whenever a permanent you control is turned face up, if it's a creature,
//	 it gets +2/+2 until end of turn."
//
// The first trigger reads the entering permanent as it entered: face down
// it is a 2/2 creature, whatever it hides (CR 708.2a), so a face-down land
// or a manifested noncreature card counts. The second is rechecked as it
// resolves (CR 603.4): the permanent is face up and the same object
// (CR 708.8), so its real front face decides "if it's a creature".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e09aaba-5c32-4a45-8a07-b69e873510ef",
		Name:         "Trail of Mystery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, cmdr2S37FaceDownCreatureEntered,
				"Trail of Mystery — search for a basic land card, reveal it, put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Trail of Mystery — a basic land card, to your hand",
					}.Apply(NewContext(g, item))
				}),
				"Trail of Mystery — search your library for a basic land card?"),
			On(game.EventTurnedFaceUp, ACreatureYouControlWasTurnedFaceUp,
				"Trail of Mystery — it gets +2/+2 until end of turn", cmdr2S37PlusTwoIfCreature),
		},
	})
}
