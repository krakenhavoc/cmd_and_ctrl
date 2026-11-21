package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skitterbeam Battalion — Artifact Creature — Construct {9}, 4/4 (The
// Brothers' War, #5656): "Prototype {3}{R}{R} — 2/2", "Trample, haste",
// "When this creature enters, if you cast it, create two tokens that are
// copies of it."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out. Trample
// and haste are printed keywords. "If you cast it" is an intervening if
// (CR 603.4), rechecked as the ability resolves. The two tokens are made
// together (CR 608.2f) and copy the Battalion as it was: a prototyped one
// makes two 2/2 tokens (CR 718.3d), a normally cast one two 4/4s. The
// tokens were not cast, so they make no more tokens, and they copy the
// Battalion's last-known values if it has left (CR 707.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff1dd9c1-d1e5-49dd-b8b2-df3e4a41d3e5",
		Name:            "Skitterbeam Battalion",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2},
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.S82BroB21EntersIfCast("Skitterbeam Battalion — create two tokens that are copies of it",
				func(ctx *effects.Context) error {
					return effects.CreateTokenCopiesTogether{Copies: []uuid.UUID{ctx.Source()}, N: 2}.Apply(ctx)
				}),
		},
	})
}
