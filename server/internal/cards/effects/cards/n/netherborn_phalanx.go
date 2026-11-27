package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Netherborn Phalanx — Creature — Horror {5}{B}, 2/4 (Ravnica: City of Guilds):
//
//	"When this creature enters, each opponent loses 1 life for each creature they
//	 control.
//	 Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your library for a
//	 card with the same mana value as this card, reveal it, put it into your hand,
//	 then shuffle. Transmute only as a sorcery.)"
//
// Each opponent, in turn order, loses life equal to the creatures they control as
// the ability resolves (the Phalanx itself counts for its controller only, who
// loses nothing). Transmute is the keyword's constructor (CR 702.53a), finding a
// mana value 6 card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a6ed882-ea6c-4dc2-8614-541e4a56b43c",
		Name:         "Netherborn Phalanx",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Netherborn Phalanx — each opponent loses 1 life for each creature they control",
				effects.OldB580EachOpponentLosesLifeForEachCreatureTheyControl),
		},
		Activated: []effects.ActivatedAbility{effects.Transmute("{1}{B}{B}")},
	})
}
