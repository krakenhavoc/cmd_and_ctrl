package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Dark-Dwellers — Creature — Goblin {3}{R}{R}, 4/4 (Foundations
// Jumpstart):
//
//	"Menace
//	 When this creature enters, you may cast target instant or sorcery
//	 card with mana value 3 or less from your graveyard without paying
//	 its mana cost. If that spell would be put into your graveyard,
//	 exile it instead."
//
// Daring Waverider's clause (freeCastTargetFromYourGraveyardExilingIt)
// at mana value 3. No simplification.
func init() {
	enters := effects.WhenThisEnters("Goblin Dark-Dwellers — you may cast target instant or sorcery card with mana value 3 or less from your graveyard without paying its mana cost",
		effects.FreeCastTargetFromYourGraveyardExilingIt("Goblin Dark-Dwellers"))
	enters.Targets = effects.TargetCardInGraveyard("target instant or sorcery card with mana value 3 or less from your graveyard",
		effects.Or(effects.Instant(), effects.Sorcery()), effects.ManaValueLE(3), effects.YouOwn())
	effects.Register(effects.Spec{
		OracleID:        "bab0ab8d-74d2-49a3-b258-b93d19925d99",
		Name:            "Goblin Dark-Dwellers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
