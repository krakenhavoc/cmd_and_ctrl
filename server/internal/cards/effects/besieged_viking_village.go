package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Besieged Viking Village — Plane — Earth (Doctor Who, #4870):
//
//	"All creatures have "Boast — {1}: Put a +1/+1 counter on this
//	 creature." (Activate only if that creature attacked this turn and
//	 only once each turn.)
//	 Whenever chaos ensues, put an indestructible counter on target
//	 creature you control that attacked this turn."
//
// A plane (ADR 0216 §6). The static is a layer-6 grant of a catalog bundle
// (ADR 0093) to every creature on the battlefield, read from the command
// zone while the plane is face up (CR 901.7); each creature's controller
// activates its own boast (CR 702.142), which game.BoastBlockLocked gates
// on the creature's attack this turn and its once-each-turn tally. The
// chaos ability (CR 311.7) targets a creature the planar controller
// controls that attacked this turn; an indestructible counter is a keyword
// counter (CR 122.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75206757-dae7-4a1e-a391-862699a100d7",
		Name:         "Besieged Viking Village",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: besiegedVikingVillageBoast,
			Activated: []ActivatedAbility{
				BoastAnswering(game.AnswerPump, "{1}: Put a +1/+1 counter on this creature.", ManaCost("{1}"), plusOneCountersOnThis(1)),
			},
			Text: "Boast — {1}: Put a +1/+1 counter on this creature. (Activate only if this creature attacked this turn and only once each turn.)",
		}},
		Static: []game.StaticAbility{GrantAbilities(allCreatures, besiegedVikingVillageBoast)},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Besieged Viking Village — put an indestructible counter on target creature you control that attacked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: t, Kind: game.CounterIndestructible, N: 1}.Apply(ctx)
				}),
				TargetCreature("target creature you control that attacked this turn", YouControl(), AttackedThisTurn())),
		},
	})
}

// besiegedVikingVillageBoast is the granted bundle's key. An on-disk key.
const besiegedVikingVillageBoast = "besieged-viking-village/boast"

// allCreatures is "All creatures": every creature on the battlefield.
func allCreatures(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() }
