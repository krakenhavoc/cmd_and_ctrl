package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brisela, Voice of Nightmares — Legendary Creature — Eldrazi Angel, 9/10
// (Innistrad Remastered): the melded back face of Bruna, the Fading Light
// and Gisela, the Broken Blade (CR 701.42, 712.4; ADR 0145).
//
//	"Flying, first strike, vigilance, lifelink
//	 Your opponents can't cast spells with mana value 3 or less."
//
// The restriction reads the spell's mana value off the card on the stack.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b23587e0-9fb7-492e-8bb4-f218e8e4ce6c",
		Name:            "Brisela, Voice of Nightmares",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "vigilance", "lifelink"},
		CastRestrictions: []game.CastRestriction{
			effects.OpponentsCantCast("Your opponents can't cast spells with mana value 3 or less.", effects.ManaValueLE(3)),
		},
	})
}
