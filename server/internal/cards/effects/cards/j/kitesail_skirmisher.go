package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitesail Skirmisher — {3}{U} Creature — Human Pirate 3/1, Commander Legends (#5507):
//
//	"Flying
//	 Whenever this creature attacks, another target creature attacking the same
//	 player or planeswalker gains flying until end of turn.
//	 Encore {4}{U}"
//
// The target clause reads the player the Skirmisher was declared against off the
// attack event and offers every other creature attacking that same defender
// (S82OldB082TargetAttackerOfTheSameTarget). Encore is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "462269c7-ca16-4aa7-bf52-f7fcda25a05b",
		Name:            "Kitesail Skirmisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Encore("{4}{U}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			},
			TargetsFrom: effects.S82OldB082TargetAttackerOfTheSameTarget,
			Key:         "Kitesail Skirmisher — another target creature attacking the same player gains flying until end of turn",
			Effect:      effects.FirstTargetGainsUntilEndOfTurn(0, "Kitesail Skirmisher — flying until end of turn", "flying"),
		}},
	})
}
