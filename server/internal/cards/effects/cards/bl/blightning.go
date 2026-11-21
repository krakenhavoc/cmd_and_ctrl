package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blightning — Sorcery {1}{B}{R} (Shards of Alara, #5588):
//
//	"Blightning deals 3 damage to target player or planeswalker. That
//	 player or that planeswalker's controller discards two cards."
//
// Ozai's Cruelty's damage-then-discard with Lavalanche's player-or-
// planeswalker target. The planeswalker's controller is read before the
// damage, so a planeswalker the 3 damage kills still names its controller
// (CR 608.2h). The discard is that player's own choice of two cards. An
// illegal target makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6496440-dc0c-4d9b-bf37-f537b6f0187b",
		Name:         "Blightning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			t := targets[0]
			victim := t.ID
			if t.Kind == game.TargetCard {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				victim = c.Controller
			}
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if victim == uuid.Nil {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   item.SourceCardID,
				N:        2,
				Question: "Blightning — discard two cards",
			})
			return nil
		},
	})
}
