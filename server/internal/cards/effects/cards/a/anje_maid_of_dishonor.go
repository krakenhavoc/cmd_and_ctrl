package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anje, Maid of Dishonor — Legendary Creature — Vampire {2}{B}{R}, 4/5
// (Innistrad: Crimson Vow, #5495):
//
//	"Whenever Anje and/or one or more other Vampires you control enter,
//	 create a Blood token. This ability triggers only once each turn.
//	 {2}, Sacrifice another creature or a Blood token: Each opponent loses
//	 2 life and you gain 2 life."
//
// The trigger is Baron Bertram Graywater's shape: one trigger per batch of
// entering permanents (Anje is a Vampire herself, so she counts), held to
// once each turn. The activated ability's cost is ONE sacrifice of a
// permanent that is another creature or a Blood token (a Blood token that
// is also a creature counts once); Anje cannot sacrifice herself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1829f1dc-1fa9-4361-b318-d4dee280e6fd",
		Name:         "Anje, Maid of Dishonor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.HasSubtypeWord("Vampire")
				},
				"Anje, Maid of Dishonor — create a Blood token (a Vampire entered)",
				effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1})))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice another creature or a Blood token: Each opponent loses 2 life and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 2},
			Cost: effects.Plus(effects.ManaCost("{2}"), game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or a Blood token",
					effects.Or(effects.Creature(), effects.HasSubtype("Blood")))),
			}),
			Effect: effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2}),
		}},
	})
}
