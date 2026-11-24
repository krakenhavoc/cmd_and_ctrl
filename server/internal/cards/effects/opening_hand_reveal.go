package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// opening_hand_reveal.go — the reveal shape of Spec.OpeningHand (CR 103.6,
// #6101, ADR 0133 amendment 2026-11-24). The engine asks and arms the
// trigger (game/opening_hand_reveal.go); a card file only declares it.

// RevealFromOpeningHandForYourFirstUpkeep is "You may reveal this card from
// your opening hand. If you do, at the beginning of your first upkeep, …"
// (Devourer of Destiny):
//
//	OpeningHand: RevealFromOpeningHandForYourFirstUpkeep(label, body),
//
// `label` is the delayed trigger's stack copy and `body` a registered body
// (game.SimpleDelayedBody) that does the rest of the sentence; the trigger
// is controlled by the player who revealed the card.
func RevealFromOpeningHandForYourFirstUpkeep(label string, body game.BodyRef) *game.OpeningHandAction {
	return &game.OpeningHandAction{Reveal: &game.OpeningHandReveal{Label: label, Body: body}}
}
