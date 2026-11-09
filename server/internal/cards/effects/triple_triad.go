package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triple Triad — Enchantment {3}{R}{R}{R} (Final Fantasy, #4462):
//
//	"At the beginning of your upkeep, each player exiles the top card of
//	 their library. Until end of turn, you may play the card you own exiled
//	 this way and each other card exiled this way with lesser mana value
//	 than it without paying their mana costs."
//
// The exile is a chain over every library in APNAP order with no permission
// stamped (Game.ExileTopKeyedForEffect), so a replacement's question can
// pause it. Once the last library is done, the grant is worked out from what
// landed: the card the controller owns is playable free this turn, and so is
// each other exiled card whose mana value is less than it (X is 0 in exile,
// CR 202.3e). Play, not cast, so a land may be played from exile and spends
// the land drop. With no card of the controller's own (an empty library) "it"
// is nothing and no other card may be played.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9567d963-4815-466b-a319-6d138b6a0cd2",
		Name:         "Triple Triad",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Triple Triad — each player exiles the top card of their library; until end of turn you may play your card and each lesser card free",
				tripleTriadUpkeep),
		},
	})
}
