package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armory Paladin — Creature — Human Knight {1}{R}{W}, 3/3 (Fallout,
// #5948):
//
//	"Trample
//	 Whenever you cast an Aura or Equipment spell, exile the top card of your
//	 library. You may play that card until the end of your next turn."
//
// The cast trigger is Mister Gutsy's reading of an Aura or Equipment spell
// (WheneverYouCast over the spell's subtypes); the exile is Alania's
// Pathmaker's impulse draw for one card over the seat-turn window, the card
// playable (a land too) until the end of its controller's next turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "360f7bfd-3b4c-4e87-a251-fd55c8a62320",
		Name:            "Armory Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.HasSubtype("Aura"), effects.HasSubtype("Equipment")),
				"Armory Paladin — exile the top card of your library; you may play it until the end of your next turn",
				effects.Do(effects.Instruction(effects.BnsB50ImpulseTopUntilYourNextTurnEnds))),
		},
	})
}
