package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molten Sentry — Creature — Elemental {3}{R}, */* (Ravnica: City of Guilds, S82
// OLD batch 576, #5600):
//
//	"As this creature enters, flip a coin. If the coin comes up heads, this
//	 creature enters as a 5/2 creature with haste. If it comes up tails, this
//	 creature enters as a 2/5 creature with defender."
//
// The flip is the uncalled kind (CR 705.2: only the face matters; no one wins
// or loses it), made as the creature enters (CR 614.12) through Game.
// FlipCoinsForEffect, so a coin-flip rule (Krark's Thumb) applies. The result
// is a layer 7b base power and toughness and a layer 6 keyword on the
// permanent, pinned to it for good (CR 611.2a): there is never a moment on the
// battlefield with the printed */*, counters and anthems still apply to the
// base, and a copy of the Sentry flips for itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66742e88-88cd-4c89-b04b-806577fe5f67",
		Name:         "Molten Sentry",
		Completeness: effects.CompletenessFull,
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			faces, err := ctx.Game.FlipCoinsForEffect(game.RandomDraw{Player: card.Controller, Source: card.InstanceID}, 1)
			if err != nil || len(faces) != 1 {
				return err
			}
			power, toughness, keyword, label := 2, 5, "defender", "Molten Sentry — enters as a 2/5 creature with defender (tails)"
			if faces[0] {
				power, toughness, keyword, label = 5, 2, "haste", "Molten Sentry — enters as a 5/2 creature with haste (heads)"
			}
			mods := append(game.SetBasePTMods(power, toughness), game.AddKeywordsMod(keyword))
			ctx.Game.RegisterScopedEffectForEffect(card.InstanceID,
				[]game.AffectedObject{game.PinObject(card.InstanceID, card.EnteredBattlefieldAt)}, mods,
				ctx.Game.PinnedTo(game.IndefiniteDuration(), card.InstanceID), label)
			return nil
		},
	})
}
