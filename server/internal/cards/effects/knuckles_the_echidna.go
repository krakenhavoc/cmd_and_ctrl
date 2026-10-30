package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Knuckles the Echidna — Legendary Creature — Echidna Warrior {2}{R}{R},
// 2/4 (Smaug Unchained deck, #4724):
//
//	"Double strike, trample, haste
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, create a Treasure token.
//	 Treasure Hunter — At the beginning of your upkeep, if you control
//	 thirty or more artifacts, you win the game."
//
// Three keywords on PrintedKeywords, and two triggers the catalog
// already has the shapes for:
//
//   - The Treasure is Grim Hireling's batch trigger
//     (WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer):
//     one trigger per player your creatures connect with in a combat
//     damage step, however many connected (CR 603.2c). Double strike
//     makes two combat damage steps (CR 510.4), so a double striker that
//     connects in both makes two Treasures.
//   - Treasure Hunter is Hellkite Tyrant's upkeep win at thirty rather
//     than twenty: an intervening if (CR 603.4) checked as the upkeep
//     begins, so with twenty-nine artifacts nothing goes on the stack,
//     and checked again as the ability resolves, so losing artifacts in
//     response stops the win. The win itself is WinTheGame, which a
//     "can't win" effect still stops (ADR 0057). Every artifact counts,
//     not only Treasures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "71ae0fab-eb3e-4226-9fcc-29186f98a569",
		Name:            "Knuckles the Echidna",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Knuckles the Echidna — create a Treasure",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && knucklesControlsThirtyArtifacts(g, source.Controller)
			}, "Knuckles the Echidna — Treasure Hunter: if you control thirty or more artifacts, you win the game",
				func(g *game.Game, item *game.StackItem) error {
					if !knucklesControlsThirtyArtifacts(g, item.Controller) {
						return nil
					}
					return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
				}),
		},
	})
}

// knucklesControlsThirtyArtifacts is Treasure Hunter's intervening if.
func knucklesControlsThirtyArtifacts(g *game.Game, player uuid.UUID) bool {
	return len(permanentsControlledByMatching(g, player, Artifact())) >= 30
}
