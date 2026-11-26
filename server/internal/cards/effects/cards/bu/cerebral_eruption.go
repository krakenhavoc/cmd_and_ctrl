package cardsbu

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cerebral Eruption — Sorcery {2}{R}{R} (Scars of Mirrodin, S82 OLD batch 471,
// #5577):
//
//	"Target opponent reveals the top card of their library. Cerebral Eruption
//	 deals damage equal to the revealed card's mana value to that player and
//	 each creature that player controls. If a land card is revealed this way,
//	 return Cerebral Eruption to its owner's hand."
//
// The reveal is one instruction, its mana value read as it is revealed (a
// land is 0), and the damage goes to the player and each creature they control
// at once (Mob Verdict's strike). A land makes the spell go to its owner's hand
// as its resolution ends (Molten Birth's exit). An empty library reveals
// nothing: no damage, and no return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8143642e-ce93-47f1-a601-3a4f5731e1d7",
		Name:         "Cerebral Eruption",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			ids := ctx.Game.RevealTopOfLibraryForEffect(victim, ctx.Source(), 1, "Cerebral Eruption — reveal the top card")
			if len(ids) == 0 {
				return nil
			}
			top, ok := ctx.Game.LookupCardForEffect(ids[0])
			if !ok {
				return nil
			}
			if top.IsLand() {
				ctx.Game.ReturnResolvingSpellToHandForEffect(ctx.Source())
			}
			hit := []uuid.UUID{victim}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == victim && c.IsCreature() {
					hit = append(hit, c.InstanceID)
				}
			}
			return effects.DealDamageToEachThen(ctx, hit, top.ManaValue(), effects.RecipientThen{})
		},
	})
}
