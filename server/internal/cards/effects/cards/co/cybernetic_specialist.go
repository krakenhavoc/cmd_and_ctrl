package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cybernetic Specialist — Artifact Creature — Human Scientist {2}, 2/1
// (Star Trek, #4973):
//
//	"When this creature enters, surveil 1.
//	 {1}, {T}: Put target card from your graveyard on the bottom of your
//	 library."
//
// Coin of Fate's surveil 1 on entry, and Barkform Harvester's graveyard
// tuck behind a {1} and {T} cost (trekB27TuckTargetCardToBottom). "Your
// graveyard" is YouOwn: a card in a graveyard has no controller. The {T}
// is the tap symbol, so a Specialist that entered this turn cannot pay it
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fc49a69-f80c-478f-b237-66f155c0c09a",
		Name:         "Cybernetic Specialist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cybernetic Specialist — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Put target card from your graveyard on the bottom of your library",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target card in your graveyard", effects.YouOwn()),
			Effect:  effects.TrekB27TuckTargetCardToBottom,
		}},
	})
}
