package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// K-9, Mark I — Legendary Artifact Creature — Robot Dog {U}, 1/1 (Doctor
// Who, #5948):
//
//	"Negative — As long as K-9 is untapped, other legendary creatures you
//	 control have ward {1}.
//	 Affirmative — {1}{U}, {T}: Target legendary creature can't be blocked
//	 this turn.
//	 Doctor's companion"
//
// The ward grant is WardGranted over the other legendary creatures the
// controller has, and it is gated on K-9 being untapped when the ward
// would trigger. The activated ability targets a legendary creature (any
// controller's) and is RestrictUntilEOT. Doctor's companion is a
// deck-construction rule.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "413bc18b-1579-41b5-a778-3a29def52bb0",
		Name:         "K-9, Mark I",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "K-9, Mark I — other legendary creatures you control have ward {1}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return !source.Tapped && target.InstanceID != source.InstanceID &&
						target.Controller == source.Controller && target.IsCreature() && target.IsLegendary()
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Target legendary creature can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    game.AbilityCost{Mana: "{1}{U}", Tap: true},
			Targets: TargetCreature("target legendary creature", Legendary()),
			Effect:  targetCantBeBlockedThisTurn("K-9, Mark I"),
		}},
	})
}
