package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Friendly Fire — Instant {3}{R} (Fate Reforged):
//
//	"Target creature's controller reveals a card at random from their hand.
//	 Friendly Fire deals damage to that creature and that player equal to the
//	 revealed card's mana value."
//
// The controller is read as the spell resolves; the card is picked from the
// game's keyed random stream (so an undo rewinds it) and revealed to the table,
// and an empty hand reveals nothing and deals no damage. Both are hit at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4ab3e1d-e1c4-43c3-8068-1809938ab094",
		Name:         "Friendly Fire",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			creature, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			owner := ctx.Game.PlayerByIDForEffect(creature.Controller)
			if owner == nil || owner.Hand.Size() == 0 {
				return nil
			}
			ids := make([]uuid.UUID, 0, owner.Hand.Size())
			for _, c := range owner.Hand.Cards {
				ids = append(ids, c.InstanceID)
			}
			pick := ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
			if len(pick) == 0 {
				return nil
			}
			revealed, ok := ctx.Game.LookupCardForEffect(pick[0])
			if !ok {
				return nil
			}
			ctx.Game.RevealForEffect(game.RevealSpec{Player: owner.ID, Source: ctx.Source(), Cards: pick, Reason: "Friendly Fire"})
			n := revealed.ManaValue()
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: n}).Apply(ctx); err != nil {
					return err
				}
				return effects.DealDamage{Source: ctx.Source(), Target: owner.ID, Amount: n}.Apply(ctx)
			})
		})),
	})
}
