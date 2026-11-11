package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathmist Raptor — Creature — Dinosaur Beast {1}{G}{G}, 3/3 (Murders at
// Karlov Manor Commander):
//
//	"Deathtouch
//	 Whenever a permanent you control is turned face up, you may return
//	 this card from your graveyard to the battlefield face up or face
//	 down.
//	 Megamorph {4}{G}"
//
// Bloodghast's graveyard-watching trigger (InGraveyard + Optional),
// on the turned-face-up event for ANY permanent its owner controls, the
// Raptor's own turning up included (it is on the battlefield then, so
// that never fires from the graveyard). "You may" is asked as the
// trigger fires; "face up or face down" is asked as it resolves. Face up
// is an ordinary return. Face down is the face-down reanimation Yedora
// and Ashcloud Phoenix use, so it enters with no ETB (CR 708.2a) and,
// because the Raptor prints megamorph, can be turned up for {4}{G}
// with the +1/+1 counter (CR 702.37e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a7d4f5aa-26d4-4363-aa5a-a26d6f791b82",
		Name:             "Deathmist Raptor",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		AlternativeCosts: []game.AlternativeCost{Megamorph("{4}{G}")},
		Triggered: []game.TriggeredAbility{
			Optional(InGraveyard(On(game.EventTurnedFaceUp, cmdr2S36YourPermanentTurnedFaceUpFromGraveyard,
				"Deathmist Raptor — return it from your graveyard to the battlefield face up or face down",
				cmdr2S36ReturnThisFaceUpOrFaceDown)),
				"Deathmist Raptor — return it to the battlefield?"),
		},
	})
}
