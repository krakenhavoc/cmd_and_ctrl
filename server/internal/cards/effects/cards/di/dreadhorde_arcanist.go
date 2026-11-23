package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreadhorde Arcanist — Creature — Zombie Wizard {1}{R}, 1/3:
//
//	"Trample
//	 Whenever this creature attacks, you may cast target instant or
//	 sorcery card with mana value less than or equal to this creature's
//	 power from your graveyard without paying its mana cost. If that
//	 spell would be put into your graveyard, exile it instead."
//
// Arcane Proxy's power-relative target and Daring Waverider's free cast
// from the graveyard with the exile rider.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c05b382-58ab-4a2d-a81c-408ea273b6b6",
		Name:            "Dreadhorde Arcanist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Dreadhorde Arcanist — you may cast target instant or sorcery card with mana value less than or equal to this creature's power from your graveyard without paying its mana cost",
					effects.FreeCastTargetFromYourGraveyardExilingIt("Dreadhorde Arcanist")),
				effects.RelativeToSource(
					effects.TargetCardInGraveyard("target instant or sorcery card with mana value less than or equal to Dreadhorde Arcanist's power from your graveyard",
						effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
					effects.ManaValueNoGreaterThanPower())),
		},
	})
}
