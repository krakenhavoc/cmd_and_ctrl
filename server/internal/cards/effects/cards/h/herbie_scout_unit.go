package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// H.E.R.B.I.E. Scout Unit — Artifact Creature — Robot Scout {4}, 2/1
// (Marvel Super Heroes, #4455):
//
//	"Flying
//	 When this creature enters, draw a card, then you may put a land card
//	 from your hand onto the battlefield tapped."
//
// One enter trigger run as a Do: the draw first, so the card just drawn
// is a legal pick, then the optional land (PutFromHandOntoBattlefield,
// entering tapped as part of its entry, so nothing sees a tap). The pick
// opens only once the draw has fully resolved. Putting a land onto the
// battlefield is not a land play (CR 305.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "393ff8c5-4da0-4290-b498-222cc0244026",
		Name:         "H.E.R.B.I.E. Scout Unit",
		Completeness: effects.CompletenessFull,

		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("H.E.R.B.I.E. Scout Unit — draw a card, then you may put a land card from your hand onto the battlefield tapped",
				effects.Do(effects.DrawCards{N: 1}, effects.PutFromHandOntoBattlefield{
					Match:    effects.Land(),
					Tapped:   true,
					Optional: true,
					Label:    "H.E.R.B.I.E. Scout Unit — you may put a land card from your hand onto the battlefield tapped",
				})),
		},
	})
}
