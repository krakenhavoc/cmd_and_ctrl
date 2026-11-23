package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "75206757-dae7-4a1e-a391-862699a100d7",
		Name:         "Besieged Viking Village",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: besiegedVikingVillageBoast,
			Activated: []effects.ActivatedAbility{
				effects.BoastAnswering(game.AnswerPump, "{1}: Put a +1/+1 counter on this creature.", effects.ManaCost("{1}"), effects.PlusOneCountersOnThis(1)),
			},
			Text: "Boast — {1}: Put a +1/+1 counter on this creature. (Activate only if this creature attacked this turn and only once each turn.)",
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(allCreatures, besiegedVikingVillageBoast)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverChaosEnsues("Besieged Viking Village — put an indestructible counter on target creature you control that attacked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: t, Kind: game.CounterIndestructible, N: 1}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you control that attacked this turn", effects.YouControl(), effects.AttackedThisTurn())),
		},
	})
}

// besiegedVikingVillageBoast is the granted bundle's key. An on-disk key.
const besiegedVikingVillageBoast = "besieged-viking-village/boast"

// allCreatures is "All creatures": every creature on the battlefield.
func allCreatures(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() }
