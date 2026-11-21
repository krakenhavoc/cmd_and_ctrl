package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vampires' Vengeance — Instant {2}{R} (Innistrad: Crimson Vow, #5495):
//
//	"Vampires' Vengeance deals 2 damage to each non-Vampire creature.
//	 Create a Blood token."
//
// A one-shot sweep: the non-Vampire creatures are read once, as it
// resolves, through the spell's own text (Subtype reads the word the card
// now says), and are dealt their damage together. The Blood token is made
// whether or not any creature was hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a82e724-28f6-430f-a474-ab4276197963",
		Name:         "Vampires' Vengeance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			vampire := ctx.TextSubtype("Vampire")
			nonVampire := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !c.HasSubtype(vampire) }
			var ids []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), nonVampire)) {
				ids = append(ids, c.InstanceID)
			}
			if err := effects.DealDamageToEachThen(ctx, ids, 2, effects.RecipientThen{}); err != nil {
				return err
			}
			return effects.CreateToken{Controller: item.Controller, Template: effects.BloodToken(), N: 1}.Apply(ctx)
		},
	})
}
