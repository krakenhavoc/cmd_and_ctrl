package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TARDIS — Artifact — Vehicle {2}, 2/4 (Doctor Who):
//
//		"Flying
//		 Whenever this Vehicle attacks, if you control a Time Lord, the next spell
//		 you cast this turn has cascade and you may planeswalk.
//		 Crew 2"
//
//	  - The attack trigger has an intervening "if" (CR 603.4): a Time Lord (the
//	    creature type is the two words Time and Lord) must be under your control
//	    as the Vehicle attacks and again as the trigger resolves.
//	  - It schedules two delayed triggers on your next cast this turn (CR 603.7),
//	    one for each half of the sentence. The cascade is Dark Apostle's: it
//	    fires as the spell is cast, above it on the stack, and cascades from the
//	    spell's mana value, X counted (CR 702.85a). The planeswalk fires on the
//	    same cast and is a "you may": outside a Planechase game there is no
//	    planar deck, so the answer does nothing (CR 701.31). A turn in which you
//	    cast no spell uses neither up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9167a585-b742-427d-b840-e024f6878a5c",
		Name:            "TARDIS",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Time Lord", func(g *game.Game, you uuid.UUID) bool {
				return effects.GapB95YouControlATimeLord(g, you, uuid.Nil)
			}, effects.WheneverThisAttacks("TARDIS — the next spell you cast this turn has cascade and you may planeswalk",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if err := effects.WhenYouNextCast("TARDIS — cascade", game.CastFilter{}, effects.Cmdr2S41CascadeTheSpellBody).Apply(ctx); err != nil {
						return err
					}
					d := effects.WhenYouNextCast("TARDIS — you may planeswalk", game.CastFilter{}, TardisPlaneswalkBody)
					d.OptionalQuestion = "TARDIS — planeswalk?"
					return d.Apply(ctx)
				})))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("TARDIS"),
		}},
	})
}

// TardisPlaneswalkBody is the planeswalk the next cast offers (the "you may" is
// the delayed trigger's own question). An on-disk key.
var TardisPlaneswalkBody = game.SimpleDelayedBody("planeswalk/tardis-next-cast", func(g *game.Game, item *game.StackItem) error {
	if g.Planechase == nil {
		return nil
	}
	return g.PlaneswalkForEffect(item.Controller)
})
