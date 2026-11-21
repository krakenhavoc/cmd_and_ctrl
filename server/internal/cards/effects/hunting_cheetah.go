package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunting Cheetah — Creature — Cat {2}{G}, 2/3 (Mystery Booster 2):
//
//	"Whenever this creature deals damage to an opponent, you may search your
//	 library for a Forest card, reveal that card, put it into your hand, then
//	 shuffle."
//
// Any damage, combat or not, to an opponent (Looter il-Kor's trigger); the
// "may" is Avarax's Optional. The search finds any card with the Forest
// subtype and may find nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "057e1cfc-d528-49e1-a386-9fb1edc1d135",
		Name:         "Hunting Cheetah",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverThisDealsDamageToAnOpponent("Hunting Cheetah — search your library for a Forest card",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsLandWithSubtype("forest"),
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Hunting Cheetah — a Forest card",
					}.Apply(NewContext(g, item))
				}),
				"Search your library for a Forest card?"),
		},
	})
}
