package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cold-Eyed Selkie — Creature — Merfolk Rogue, {1}{G/U}{G/U}, 1/1:
//
//	"Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)
//	 Whenever this creature deals combat damage to a player, you may
//	 draw that many cards."
//
// Islandwalk is a printed keyword, enforced by the block check since
// #705 (game/landwalk.go). It is declared in PrintedKeywords as well as
// arriving on Scryfall's keyword array, so a fixture or a token copy
// that never went through deck import still has it.
//
// The draw is a CR 603.4 "you may": one yes/no prompt, and a yes draws
// the whole amount — "that many" is the damage dealt, not an upper
// bound to choose under. The amount is read off the triggering event
// (item.Trigger.Event.Amount, ADR 0041 P9), fixed as of the moment the
// trigger fired, so a pump or shrink after damage doesn't change it:
// "that many" is the damage that was dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f33cd975-ab70-4209-a6d8-e0d727772bf2",
		Name:            "Cold-Eyed Selkie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: ThisDealtCombatDamageToAPlayer,
				Key:       "Cold-Eyed Selkie — draw that many cards",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
				},
			}, "Cold-Eyed Selkie — draw that many cards?"),
		},
	})
}
