package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "390320f7-11b8-4b61-8316-36c396470a09",
		Name:         "Fowl Strike",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with flying", HasKeyword("flying")),
		OnResolve:    destroyTheTargetPermanent,
		Activated: []ActivatedAbility{{
			Label:   "Reinforce 2—{2}{G} ({2}{G}, Discard this card: Put two +1/+1 counters on target creature.)",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{2}{G}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target creature"),
			Effect:  renewCountersOnTarget(game.CounterPlusOne, 2),
		}},
	})
}
