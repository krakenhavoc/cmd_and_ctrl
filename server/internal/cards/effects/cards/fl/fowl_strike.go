package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fowl Strike — Instant {1}{G} (Modern Horizons 3, #5467):
//
//	"Destroy target creature with flying.
//	 Reinforce 2—{2}{G} ({2}{G}, Discard this card: Put two +1/+1
//	 counters on target creature.)"
//
// The spell is Plummet. Reinforce is an activated ability that works
// only from the hand (ADR 0062), written the way Cycling is: discarding
// this card is part of the cost, so it is in the graveyard before the
// ability resolves, and the counters go on a target still legal at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "390320f7-11b8-4b61-8316-36c396470a09",
		Name:         "Fowl Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Activated: []effects.ActivatedAbility{{
			Label:   "Reinforce 2—{2}{G} ({2}{G}, Discard this card: Put two +1/+1 counters on target creature.)",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RenewCountersOnTarget(game.CounterPlusOne, 2),
		}},
	})
}
