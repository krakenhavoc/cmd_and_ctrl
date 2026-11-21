package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Technomancer — Artifact Creature — Necron Wizard {5}{B}{B}, 5/1 (Warhammer
// 40,000 Commander, #5948):
//
//	"When this creature enters, mill three cards, then return any number of
//	 artifact creature cards with total mana value 6 or less from your graveyard
//	 to the battlefield."
//
// Teval's mill-then-pick: the cards are chosen after the mill, so a milled
// artifact creature is a legal pick, and the total-mana-value ceiling is a set
// rule on the pick. The cards enter together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f97541e-34f3-4f7f-8ac2-7167b2ec9b33",
		Name:         "Technomancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Technomancer — mill three cards, then return any number of artifact creature cards with total mana value 6 or less from your graveyard to the battlefield",
				effects.GapB189TechnomancerMillThenReturn),
		},
	})
}
