package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodfire Enforcers — {3}{R} Creature — Human Monk 5/2 (Fate Reforged):
//
//	"This creature has first strike and trample as long as an instant card and
//	 a sorcery card are in your graveyard."
//
// Two layer-6 grants (CR 613.1f) on this creature alone, gated on the
// controller's own graveyard and re-read on every recompute (CR 611.3a): a card
// entering or leaving the graveyard drops the layer cache. One instant card and
// one sorcery card are needed; a single card is never both.
func init() {
	label := "This creature has first strike and trample as long as an instant card and a sorcery card are in your graveyard."
	first := effects.KeywordGrant(effects.WhileYouHave(effects.SocB25GraveyardHasInstantAndSorcery, effects.SelfCreature), "first strike")
	first.Label = label
	trample := effects.KeywordGrant(effects.WhileYouHave(effects.SocB25GraveyardHasInstantAndSorcery, effects.SelfCreature), "trample")
	trample.Label = label
	effects.Register(effects.Spec{
		OracleID:     "5e4facd9-d416-4ade-ba71-6da32b12a891",
		Name:         "Bloodfire Enforcers",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{first, trample},
	})
}
