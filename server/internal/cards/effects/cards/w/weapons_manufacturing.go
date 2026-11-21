package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weapons Manufacturing — Enchantment {1}{R} (EDHREC rank 3342):
//
//	"Whenever a nontoken artifact you control enters, create a
//	 colorless artifact token named Munitions with "When this token
//	 leaves the battlefield, it deals 2 damage to any target.""
//
// Every real artifact brings a Shock that fires when the Munitions
// is sacrificed, bounced or blown up. The entry trigger is
// b31NontokenArtifactYouControlEntered — post-layer types, tokens
// excluded, so the Munitions themselves never chain. The token is a
// colorless artifact named Munitions and nothing else.
//
// The printed trigger lives on the TOKEN (MunitionsToken in
// tokens.go, ADR 0083), not on this enchantment: the Munitions
// shoots whether or not the Manufacturing is still on the
// battlefield, and for whoever controls the token when it leaves —
// an opponent who stole it gets the 2 damage, as printed. The
// damage is dealt by the token, a colorless source, so a red-damage
// payoff does not see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e48a8160-cffe-4e00-a3e9-a59d7fc7b3a2",
		Name:         "Weapons Manufacturing",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B31NontokenArtifactYouControlEntered(ev, source, g)
			}, "Weapons Manufacturing — create a Munitions token", effects.Do(effects.CreateToken{Template: effects.MunitionsToken(), N: 1})),
		},
	})
}
