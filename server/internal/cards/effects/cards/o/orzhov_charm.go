package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orzhov Charm — Instant {W}{B} (Gatecrash, #5565):
//
//	"Choose one —
//	 • Return target creature you control and all Auras you control attached
//	   to it to their owner's hand.
//	 • Destroy target creature and you lose life equal to its toughness.
//	 • Return target creature card with mana value 1 or less from your
//	   graveyard to the battlefield."
//
// Mode 0 is Hubris's simultaneous bounce, narrowed to the Auras the caster
// controls: the creature and each such Aura leave together (CR 400.7), so an
// Aura is never left to fall off first. Mode 1 reads the creature's
// toughness before it is destroyed (CR 608.2h) and the caster loses that much
// life whether or not the creature was destroyed (an indestructible creature
// still costs the life, as the rulings say); a negative toughness loses
// nothing. Mode 2 is Bishop of Rebirth's reanimation, the card checked again
// as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26f8626a-ce64-450c-bda5-833df2fbb4fb",
		Name:         "Orzhov Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Return target creature you control and all Auras you control attached to it to their owner's hand.",
				effects.TargetCreature("target creature you control", effects.YouControl())),
			effects.ModeWithPurpose(effects.Mode("Destroy target creature and you lose life equal to its toughness.",
				effects.TargetCreature("target creature")), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.Mode("Return target creature card with mana value 1 or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target creature card with mana value 1 or less from your graveyard",
					effects.YouOwn(), effects.Creature(), effects.ManaValueLE(1))), effects.ForTargets(effects.Reanimates(0))),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if !orzhovCharmTargetLegal(ctx, target) {
				return nil
			}
			switch {
			case ctx.HasMode(0):
				ids := []uuid.UUID{target}
				for _, a := range ctx.Game.BattlefieldCardsForEffect() {
					if a.IsAttachedTo(target) && a.IsAura() && a.Controller == ctx.Controller() {
						ids = append(ids, a.InstanceID)
					}
				}
				return ctx.Game.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
			case ctx.HasMode(1):
				toughness := effects.DepartedCreatureToughness(ctx.Game, target)
				if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
					return err
				}
				if toughness <= 0 {
					return nil
				}
				return ctx.Game.LoseLifeEachKeyedForEffect(ctx.Source(), []uuid.UUID{ctx.Controller()}, toughness, game.KeyedTail{})
			case ctx.HasMode(2):
				return effects.ReturnFromGraveyard{Target: target, Dest: game.ZoneBattlefield}.Apply(ctx)
			}
			return nil
		},
	})
}

// orzhovCharmTargetLegal reports whether the chosen mode's target is still
// legal as the spell resolves (CR 608.2b).
func orzhovCharmTargetLegal(ctx *effects.Context, id uuid.UUID) bool {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && t.ID == id {
			return true
		}
	}
	return false
}
