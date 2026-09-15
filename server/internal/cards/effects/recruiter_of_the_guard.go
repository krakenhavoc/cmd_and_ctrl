package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Recruiter of the Guard — Creature — Human Soldier {2}{W}, 1/1
// (EDHREC rank 1152):
//
//	"When this creature enters, you may search your library for a
//	 creature card with toughness 2 or less, reveal it, put it into
//	 your hand, then shuffle."
//
// Imperial Recruiter's white twin, with toughness in place of power.
// The tutor is optional, so the prompt always opens and "fail to
// find" is always an answer. The toughness check reads the printed
// value: a card in the library has no battlefield characteristics
// for the layer engine to modify, so there is nothing else it could
// mean.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d521a329-a53a-4962-810a-2abed80df260",
		Name:         "Recruiter of the Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Recruiter of the Guard — search for a creature with toughness 2 or less", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && c.Toughness <= 2
					},
					Dest:     game.ZoneHand,
					Limit:    1,
					Reveal:   true,
					Shuffle:  true,
					Optional: true,
					Reason:   "Recruiter of the Guard — a creature card with toughness 2 or less, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
