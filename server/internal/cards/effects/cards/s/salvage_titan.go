package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Salvage Titan — {4}{B}{B} Artifact Creature — Golem 6/4 (Shards of Alara, S82
// OLD b516):
//
//	"You may sacrifice three artifacts rather than pay this spell's mana cost.
//	 Exile three artifact cards from your graveyard: Return this card from your
//	 graveyard to your hand."
//
// An alternative cost (CR 118.9) paid with a sacrifice of exactly three
// artifacts, and a graveyard ability whose cost exiles three OTHER artifact
// cards (the source is never a legal pick of its own cost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "823cd586-92db-4e44-ba1f-394b2404cafc",
		Name:         "Salvage Titan",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.SacrificeInstead(3, "three artifacts", 0, effects.Artifact()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile three artifact cards from your graveyard: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ExileFromGraveyard(3, "three artifact cards", func(c game.Card) bool { return c.IsArtifact() }),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
