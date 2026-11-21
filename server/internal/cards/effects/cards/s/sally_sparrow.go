package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sally Sparrow — Legendary Creature — Human Detective {2}{W}{U}, 2/3
// (Doctor Who, #5948):
//
//	"You may cast creature spells as though they had flash.
//	 Whenever one or more other creatures you control leave the battlefield,
//	 investigate. This ability triggers only once each turn."
//
// The first line is a cast-timing rule over creature spells (Yeva, Nature's
// Herald's shape). The trigger is once per batch of simultaneous departures
// (CR 603.2c) and once each turn (TriggersOnlyOnceEachTurn, used up when it
// triggers, as printed). A creature that left is read as it last existed
// (CR 603.10a), so a token, a creature put into the graveyard, exiled or
// returned to hand all count, and Sally leaving by herself does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c924f9f-84e2-4051-ae18-0c2cc57cf464",
		Name:         "Sally Sparrow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{CreatureOnly: true},
				"You may cast creature spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventLTB, sallySparrowAnotherCreatureLeft,
				"Sally Sparrow — investigate", effects.Do(effects.Investigate{N: 1})))),
		},
	})
}

// sallySparrowAnotherCreatureLeft is "one or more other creatures you
// control leave the battlefield": another creature, a creature as it last
// existed, controlled by the source's controller when it left.
func sallySparrowAnotherCreatureLeft(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && effects.LeftAsType(ev, c, "creature") && effects.LeftUnderControlOf(ev, c) == source.Controller
}
