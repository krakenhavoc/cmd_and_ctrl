package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
		Apply:     pumpBy(1, 2),
	}
	vigilance := KeywordGrant(whileYouHave(drcDeliriumMet, selfCreature), "vigilance")
	vigilance.Label = "Delirium — Has vigilance" + delirium
	Register(Spec{
		OracleID:     "dcfa0476-e88f-43b4-8011-d78a3de97f2a",
		Name:         "Spineseeker Centipede",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spineseeker Centipede — search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Spineseeker Centipede — a basic land card",
				}.Apply(NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{pump, vigilance},
	})
}
