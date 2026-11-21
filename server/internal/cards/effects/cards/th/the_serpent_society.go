package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Serpent Society — Legendary Creature — Human Snake Villain {1}{B}{G},
// 3/4 (Marvel Super Heroes):
//
//	"Deathtouch
//	 Ward—Get five poison counters. (A player with ten or more poison
//	 counters loses the game.)
//	 Whenever another creature you control with deathtouch dies, each
//	 opponent sacrifices a nontoken creature of their choice."
//
// The ward is effects.WardPoison (#4613, ADR 0038 C6): the caster of the
// spell or ability that targeted the Society is offered five poison counters
// or has it countered (CR 702.21a, CR 118.12a). A player can always get
// counters, so the offer is never withheld, and a caster at five who takes it
// reaches ten and loses at the next state-based check (CR 704.5c).
//
// The dies trigger is judged on the creature as it last existed on the
// battlefield (CR 603.10a): leftWithKeyword reads the deathtouch the exit
// stamped on the event, so a creature that had it through an effect counts,
// and the Society's own death does not ("another"). The edict is Butcher of
// Malakir's: each opponent chooses their own, a creature with hexproof is a
// legal choice, and tokens are not (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3688ad0d-653b-4d5d-a57e-444db679d83e",
		Name:            "The Serpent Society",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardPoison(5), "The Serpent Society — ward, get five poison counters"),
			effects.On(game.EventLTB, anotherDeathtouchCreatureYouControlDied,
				"The Serpent Society — each opponent sacrifices a nontoken creature",
				effects.Do(effects.EachPlayerSacrifices{
					ExceptController: true,
					Match:            effects.And(effects.Creature(), effects.Not(effects.IsTokenPredicate())),
					Label:            "a nontoken creature",
				})),
		},
	})
}

// anotherDeathtouchCreatureYouControlDied is "Whenever another creature you
// control with deathtouch dies".
func anotherDeathtouchCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := effects.DiedCreature(ev, g)
	if !ok || effects.LeftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	return effects.LeftWithKeyword(ev, dead, "deathtouch")
}
