package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodvial Purveyor — Creature — Vampire {2}{B}{B}, 5/6 (Innistrad:
// Crimson Vow, #5495):
//
//	"Flying, trample
//	 Whenever an opponent casts a spell, that player creates a Blood token.
//	 Whenever this creature attacks, it gets +1/+0 until end of turn for
//	 each Blood token defending player controls."
//
// The Blood token belongs to the caster (the cast event's actor), not to
// the Purveyor's controller. The attack bonus counts the Blood tokens
// controlled by the player the Purveyor attacks (a planeswalker or battle
// it attacks names that planeswalker's or battle's defending player) as the
// trigger resolves; none, or no defending player any more, is +0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7292272d-6511-466b-aadd-c4fe7ed8df09",
		Name:            "Bloodvial Purveyor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.B15OpponentCastSpell(ev, source)
				},
				Key: "Bloodvial Purveyor — that player creates a Blood token",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Trigger.Event.Actor, Template: effects.BloodToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			effects.WheneverThisAttacks("Bloodvial Purveyor — it gets +1/+0 until end of turn for each Blood token defending player controls",
				func(g *game.Game, item *game.StackItem) error {
					defender := effects.B17DefendingPlayer(g, item.Trigger.Event)
					if defender == uuid.Nil {
						return nil
					}
					return effects.BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  effects.VowB04BloodTokensControlledBy(g, defender),
						Label:  "Bloodvial Purveyor — +1/+0 per Blood token",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
