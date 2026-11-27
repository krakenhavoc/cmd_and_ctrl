package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bogardan Phoenix — Creature — Phoenix {2}{R}{R}{R}, 3/3 (Visions, #5631):
//
//	"Flying
//	 When this creature dies, exile it if it had a death counter on it.
//	 Otherwise, return it to the battlefield under your control and put a
//	 death counter on it."
//
// The counter is read from the Phoenix as it last existed (CR 603.10a,
// CR 608.2h). With one, the card is exiled if it is still in the graveyard.
// Without, it returns under the trigger's controller with a death counter
// (it enters with it, so a counter replacement sees it and the next death
// finds it); a card that left the graveyard in the meantime is a new object
// and does nothing (CR 400.7).
//
// No simplification.
func init() {
	const deathCounter = "death"
	effects.Register(effects.Spec{
		OracleID:        "72586a7a-3c73-4012-8664-1db0f268eb60",
		Name:            "Bogardan Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Bogardan Phoenix — exile it if it had a death counter on it, otherwise return it with a death counter",
				func(g *game.Game, item *game.StackItem) error {
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if info, ok := ctx.TriggeringPermanent(); ok && info.Counters[deathCounter] > 0 {
						return effects.ExileTarget{Target: item.SourceCardID}.Apply(ctx)
					}
					_, err := g.ReturnFromGraveyardWithCountersForEffect(item.SourceCardID, item.Controller, false, map[string]int{deathCounter: 1})
					return effects.NothingIfGone(err)
				}),
		},
	})
}
