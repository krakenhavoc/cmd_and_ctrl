package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beifong's Bounty Hunters — Creature — Human Mercenary {2}{B}{G}, 4/4:
//
//	"Whenever a nonland creature you control dies, earthbend X, where X
//	 is that creature's power. (Target land you control becomes a 0/0
//	 creature with haste that's still a land. Put X +1/+1 counters on
//	 it. When it dies or is exiled, return it to the battlefield
//	 tapped.)"
//
// A Zulaport-shaped dies trigger (ACreatureYouControlDied, the
// Hunters themselves included) narrowed to a permanent that was not a
// land as it last existed (CR 603.10a), so an earthbent land dying
// does not feed the next one. The target is the earthbend keyword's
// own clause, picked as the trigger is put on the stack (CR 603.3d).
//
// X is the dead creature's power as it last existed on the battlefield
// (CR 608.2h, TriggeringPermanent): counters and pumps it had count. A
// negative power earthbends 0 (CR 107.1b) — the land still animates as
// a 0/0 and comes back tapped, as the keyword says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8561b088-56b3-4732-82e9-8da1e3affd55",
		Name:         "Beifong's Bounty Hunters",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !ACreatureYouControlDied(ev, source, lki, g) {
					return false
				}
				dead, _ := g.LookupCardForEffect(ev.CardID)
				return !leftAsType(ev, dead, "land")
			}, "Beifong's Bounty Hunters — earthbend X, where X is that creature's power", beifongsBountyHuntersEarthbend),
				EarthbendTargets()),
		},
	})
}

// beifongsBountyHuntersEarthbend earthbends the chosen land by the
// dead creature's last-known power.
func beifongsBountyHuntersEarthbend(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	x := 0
	if info, ok := ctx.TriggeringPermanent(); ok && info.Power > 0 {
		x = info.Power
	}
	return Earthbend{Target: target, N: x}.Apply(ctx)
}
