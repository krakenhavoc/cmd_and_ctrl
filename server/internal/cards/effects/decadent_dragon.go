package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decadent Dragon // Expensive Taste — an adventure card (CR 715)
// (Smaug Unchained deck, #4724):
//
//	Decadent Dragon — Creature — Dragon {2}{R}{R}, 4/4
//	  "Flying, trample
//	   Whenever this creature attacks, create a Treasure token."
//	Expensive Taste — Instant — Adventure {2}{B}
//	  "Exile the top two cards of target opponent's library face down.
//	   You may look at and play those cards for as long as they remain
//	   exiled."
//
// Two keys, one card, as Bonecrusher Giant's: the creature under the
// bare oracle ID and the Adventure under "<oracle_id>#1". The adventure
// lifecycle (exile on resolution, then cast the creature from exile) is
// the engine's (game/adventure.go) and this file says nothing about it.
//
// The creature is Captain Lannery Storm's attack trigger beside two
// printed keywords.
//
// Expensive Taste is Outrageous Robbery's exile-and-play grant
// (ExileTopWithPermission) with a fixed count of two and without
// Robbery's "mana of any type" rider, which this card does not print.
// The cards go face down (game.FaceDownPermitted): only the caster may
// look at them. "Play", so a land among them is a land drop, and the
// grant lasts for as long as each card stays in exile. The caster
// exiles them (Robbery has the opponent do it), which changes nothing a
// player sees. An opponent who has left the game by resolution is an
// illegal target and the spell does nothing.
//
// No simplification.

// decadentDragonOracleID is shared by both faces of the card.
const decadentDragonOracleID = "5df44b8d-b337-49d8-8427-34253f52cb47"

func init() {
	// Face 0 — the creature.
	Register(Spec{
		OracleID:        decadentDragonOracleID,
		Name:            "Decadent Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Decadent Dragon — create a Treasure",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})

	// Face 1 — the Adventure, an instant while it is on the stack
	// (CR 715.3a).
	Register(Spec{
		OracleID:     decadentDragonOracleID + "#1",
		Name:         "Expensive Taste",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return ExileTopWithPermission{
					From:        t.ID,
					GrantTo:     item.Controller,
					N:           2,
					FaceDown:    true,
					WhileExiled: true,
				}.Apply(ctx)
			}
			return nil
		},
	})
}
