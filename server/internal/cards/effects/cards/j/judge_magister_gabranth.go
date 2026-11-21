package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Judge Magister Gabranth — Legendary Creature — Human Advisor Knight
// {W}{B}, 2/2 (Final Fantasy, #4462):
//
//	"Menace
//	 Whenever another creature or artifact you control dies, put a +1/+1
//	 counter on Judge Magister Gabranth."
//
// Agent of the Iron Throne's death condition (b21ArtifactOrCreatureYouControlDied)
// without the source itself: the permanent's types and controller are read as
// it last existed, so a Treasure, a token creature and a nontoken one all count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1a846ad6-15c9-4bf9-9127-612018a557b9",
		Name:            "Judge Magister Gabranth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.B21ArtifactOrCreatureYouControlDied(ev, source, g)
				return ok && c.InstanceID != source.InstanceID
			}, "Judge Magister Gabranth — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
