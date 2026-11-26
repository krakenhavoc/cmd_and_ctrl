package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Explosive Revelation — Sorcery {3}{R}{R} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Choose any target. Reveal cards from the top of your library until you
//	 reveal a nonland card. Explosive Revelation deals damage equal to that
//	 card's mana value to that permanent or player. Put the nonland card into
//	 your hand and the rest on the bottom of your library in any order."
//
// Calibrated Blast's reveal (a token is never the hit), with the damage dealt
// by this spell to the target chosen as it was cast, read from the hit's mana
// value while it is still in the library. The hit then goes to the hand and
// every other revealed card to the bottom in the order the player chooses
// (Alrund's take with the rest on the bottom). A library with no nonland card
// reveals itself whole, deals nothing and goes to the bottom.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49f01296-7b30-4d84-bfa3-6ed58cd23a33",
		Name:         "Explosive Revelation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return !c.IsLand() }, "Explosive Revelation")
			if hit == uuid.Nil {
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  run,
					Match:  func(_ *game.Game, _ uuid.UUID, _ game.Card) bool { return false },
					All:    true,
					Label:  "Explosive Revelation — the revealed cards go to the bottom of your library",
					Key:    effects.TakeRestBottomAny,
				}.Apply(ctx)
			}
			if c, ok := ctx.Game.LookupCardForEffect(hit); ok {
				if targets := ctx.LegalTargets(); len(targets) > 0 {
					if err := (effects.DealDamage{Source: ctx.Source(), Target: targets[0].ID, Amount: c.ManaValue()}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  run,
				Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == hit },
				All:    true,
				Label:  "Explosive Revelation — the nonland card to your hand, the rest on the bottom",
				Key:    effects.TakeRestBottomAny,
			}.Apply(ctx)
		},
	})
}
