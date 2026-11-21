package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Enervation — Enchantment {3}{B} (Murders at Karlov Manor, #4469):
//
//	"Flash
//	 When this enchantment enters, target creature gets -4/-4 until end
//	 of turn.
//	 Whenever one or more creature cards leave your graveyard, each
//	 opponent loses 1 life and you gain 1 life."
//
// The entry trigger shrinks the still-legal target (CR 608.2b). The
// second trigger is Chalk Outline's batch trigger: a mass return drains
// once, not once per card, and the creature cards may leave in any
// direction. The drain is the footprinted pair EachOpponentLosesLife
// and GainLife.
//
// No simplification.
func init() {
	const shrink = "Soul Enervation — target creature gets -4/-4 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "5b4508a3-1d21-46ec-8a50-cd3171a7348a",
		Name:            "Soul Enervation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters(shrink, func(g *game.Game, item *game.StackItem) error {
				return effects.VsPumpTheTarget(effects.NewContext(g, item), -4, -4, shrink)
			}), effects.TargetCreature("target creature")), effects.ForTargets(effects.HarmUntilEOT(0, -4, -4))),
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.CreatureCardLeftYourGraveyard(ev, source, g)
				},
				"Soul Enervation — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}))),
		},
	})
}
