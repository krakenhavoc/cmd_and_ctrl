package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pulsemage Advocate — Creature — Human Cleric {2}{W}, 1/3 (Judgment,
// #5611):
//
//	"{T}: Return three target cards from an opponent's graveyard to their
//	 hand. Return target creature card from your graveyard to the
//	 battlefield."
//
// Two target clauses: exactly three cards from ONE opponent's graveyard
// (a set rule, as "from a single graveyard"), and a creature card of yours.
// Each returns on its own if the other is gone (CR 608.2b); the first three
// go to their owners' hands (CR 400.3), the creature enters under your
// control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06e4355b-02d4-4238-a6e6-b03a5773346c",
		Name:         "Pulsemage Advocate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Return three target cards from an opponent's graveyard to their hand. Return target creature card from your graveyard to the battlefield.",
			Cost:  effects.TapCost(),
			Targets: effects.Clauses(
				effects.TargetCardInGraveyard("three target cards from an opponent's graveyard", effects.OldB621OpponentOwns()).
					WithCount(3, 3).AllShare(effects.FromASingleGraveyard()),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			),
			Effect: effects.OldB621PulsemageAdvocate,
		}},
	})
}
