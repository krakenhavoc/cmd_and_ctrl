package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ravenous Slime — Creature — Ooze {2}{G}, 1/1:
//
//	"This creature can't be blocked by creatures with power 2 or less.
//	 If a creature an opponent controls would die, instead exile it and put a
//	 number of +1/+1 counters equal to that creature's power on this
//	 creature."
//
// The block restriction is Questing Beast's shared CantBeBlockedBy. The
// replacement is Head of the Hunt's (RfOpponentCreatureWouldDie, tokens
// included): the creature is exiled instead of dying (CR 700.4, so its
// dies-triggers do not fire) and the Slime gets counters equal to that
// creature's power as it last existed on the battlefield (CR 608.2h), read as
// the replacement applies, while the creature is still there. A creature of
// power 0 or less gives none, and the counters are put through the counter
// pipeline so a doubler applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4696decb-7bff-4c6b-8a7b-9ec324faefde",
		Name:         "Ravenous Slime",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less")},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.RfOpponentCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				dying, ok := g.LookupCardForEffect(ev.CardID)
				if !ok {
					return nil
				}
				power := g.PowerAnywhereForEffect(dying)
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				if power <= 0 {
					return nil
				}
				_, err := g.AddCounterMustSettleNowForEffect(src.InstanceID, game.CounterPlusOne, power)
				return err
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Ravenous Slime: exile instead of dying and put +1/+1 counters equal to its power on this creature",
		}},
	})
}
