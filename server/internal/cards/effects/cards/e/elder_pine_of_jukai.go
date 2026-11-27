package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elder Pine of Jukai — Creature — Spirit {2}{G}, 2/1 (Saviors of Kamigawa,
// #5601):
//
//	"Whenever you cast a Spirit or Arcane spell, reveal the top three cards
//	 of your library. Put all land cards revealed this way into your hand and
//	 the rest on the bottom of your library in any order.
//	 Soulshift 2 (When this creature dies, you may return target Spirit card
//	 with mana value 2 or less from your graveyard to your hand.)"
//
// Mulch's reveal-and-take: every land goes to the hand (no choice), and the
// rest go on the bottom in the order the player picks (CR 701.20a, 401.4).
// Soulshift is the engine's (CR 702.46a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1658528b-df40-44fd-b3f6-eb042706c522",
		Name:         "Elder Pine of Jukai",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Elder Pine of Jukai — reveal the top three cards of your library; lands to your hand, the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 3,
						"Elder Pine of Jukai — reveal the top three cards of your library")
					var lands, rest []uuid.UUID
					for _, id := range revealed {
						if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
							lands = append(lands, id)
						} else {
							rest = append(rest, id)
						}
					}
					g.TakeFromLibraryToHandForEffect(item.Controller, lands)
					rest = effects.CardsStillInALibrary(g, rest)
					if len(rest) == 0 {
						return nil
					}
					return effects.PutInLibraryInAnyOrder{
						Chooser:   item.Controller,
						Cards:     rest,
						From:      game.ZoneLibrary,
						Placement: game.LibraryPlaceBottom,
						Label:     "Elder Pine of Jukai — put the rest on the bottom of your library in any order",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.Soulshift(2, "Elder Pine of Jukai"),
		},
	})
}
