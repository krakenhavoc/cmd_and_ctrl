package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wickersmith's Tools — Artifact {3}, (Lorwyn Eclipsed Commander, #5368):
//
//	"Whenever one or more -1/-1 counters are put on a creature, put a
//	 charge counter on this artifact.
//	 {T}: Add one mana of any color.
//	 {5}, {T}, Sacrifice this artifact: Create X tapped 2/2 colorless
//	 Scarecrow artifact creature tokens, where X is the number of charge
//	 counters on this artifact."
//
// The first is one trigger per placement event, whoever put the counters and
// whichever creature took them (cmdr2B25MinusCountersPlacedOnACreature), and
// the engine's CounterOnThis primitive. X is the charge counters the artifact
// last had, read from the departure record since sacrificing it is a cost
// (CR 608.2h, LastKnownPermanentForEffect).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c89cd328-b4cc-4760-8b6c-59e82eed8dac",
		Name:         "Wickersmith's Tools",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.Cmdr2B25MinusCountersPlacedOnACreature(ev, g)
			}, "Wickersmith's Tools — put a charge counter on this artifact",
				effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this artifact: Create X tapped 2/2 colorless Scarecrow artifact creature tokens, where X is the number of charge counters on this artifact.",
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				info, ok := g.LastKnownPermanentForEffect(item.SourceCardID)
				if !ok || info.Counters[game.CounterCharge] <= 0 {
					return nil
				}
				return effects.CreateTokenAdvanced{
					Spec: effects.Token(effects.TokenCard("2/2 colorless Scarecrow artifact")).EntersTapped(),
					N:    info.Counters[game.CounterCharge],
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
