package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evolution Witness — Creature — Elf Shaman Mutant {2}{G} (2/2):
//
//	"{1}{G}: Adapt 2. (If this creature has no +1/+1 counters on it,
//	 put two +1/+1 counters on it.)
//	 Whenever one or more +1/+1 counters are put on this creature,
//	 return target permanent card from your graveyard to your hand."
//
// The return is one trigger per placement event, not one per counter,
// so a Hardened Scales'd adapt still pays out once; the trigger is
// removed when the graveyard holds no permanent card (CR 603.3d).
// Adapt checks for counters on resolution (CR 701.46a).
//
// No simplification.
func init() {
	const label = "Evolution Witness — return target permanent card from your graveyard to your hand"
	effects.Register(effects.Spec{
		OracleID:     "0e07f1af-5ff6-4da5-8683-aea4cc390975",
		Name:         "Evolution Witness",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Adapt 2",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect:  effects.AdaptTwo,
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Target == source.InstanceID && effects.B33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
			}, label, effects.ReturnTargetGraveyardCardToHand), effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.Permanent(), effects.YouOwn())),
		},
	})
}
