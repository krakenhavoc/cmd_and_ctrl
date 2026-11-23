package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spineseeker Centipede — Creature — Insect {2}{G}, 2/1 (Duskmourn: House
// of Horror, #4466):
//
//	"When this creature enters, search your library for a basic land card,
//	 reveal it, put it into your hand, then shuffle.
//	 Delirium — This creature gets +1/+2 and has vigilance as long as there
//	 are four or more card types among cards in your graveyard."
//
// The enters trigger is Horizon Seeker's search (SearchLibrary to the
// hand, revealed, then shuffled); it is not a "may", so the library is
// shuffled even when no basic land is found (CR 701.23b).
//
// Delirium is an ability word (CR 207.2c). The bonus is Dragon's Rage
// Channeler's: a layer 7c modification and a layer 6 grant on the
// Centipede alone (selfCreature), both gated on the controller's
// graveyard (drcDeliriumMet) and re-read on every recompute, so they turn
// on and off together as the graveyard changes.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	pump := game.StaticAbility{
		Label:     "Delirium — Gets +1/+2" + delirium,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: effects.WhileYouHave(effects.DrcDeliriumMet, effects.SelfCreature),
		Apply:     effects.PumpBy(1, 2),
	}
	vigilance := effects.KeywordGrant(effects.WhileYouHave(effects.DrcDeliriumMet, effects.SelfCreature), "vigilance")
	vigilance.Label = "Delirium — Has vigilance" + delirium
	effects.Register(effects.Spec{
		OracleID:     "dcfa0476-e88f-43b4-8011-d78a3de97f2a",
		Name:         "Spineseeker Centipede",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spineseeker Centipede — search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Spineseeker Centipede — a basic land card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{pump, vigilance},
	})
}
