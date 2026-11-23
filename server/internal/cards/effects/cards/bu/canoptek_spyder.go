package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canoptek Spyder — Artifact Creature — Spider {5}, 4/4 (Warhammer 40,000
// Commander, S82 GAP batch 187, #5948):
//
//	"Flying
//	 Fabricator Claw Array — Whenever another nontoken artifact creature or
//	 Vehicle you control enters, draw a card."
//
// A Vehicle is judged as it enters, a creature or not (CR 301.7): the
// trigger reads the entering permanent's types, not its animation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bcad5b83-5513-4455-a9c9-716281505eaa",
		Name:            "Canoptek Spyder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || c.Controller != source.Controller || c.IsToken() {
					return false
				}
				return (c.IsArtifact() && c.IsCreature()) || c.HasSubtype("Vehicle")
			}, "Canoptek Spyder — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
