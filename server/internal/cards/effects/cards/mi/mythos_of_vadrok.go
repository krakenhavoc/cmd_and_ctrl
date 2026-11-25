package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mythos of Vadrok — Sorcery {2}{R}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Mythos of Vadrok deals 5 damage divided as you choose among any number of
//	 target creatures and/or planeswalkers. If {W}{U} was spent to cast this
//	 spell, until your next turn, those permanents can't attack or block and
//	 their activated abilities can't be activated."
//
// The damage is divided as the spell is cast (CR 601.2d) and each part is
// dealt as it resolves to the targets still legal (CR 608.2b). The rider
// reads the mana actually spent: at least one white and one blue. "Those
// permanents" are the targets still legal; the restriction ends as your next
// turn begins and covers mana abilities too, an activated ability being any
// ability with a cost and an effect (CR 602.1, CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0b72fdeb-80b2-4691-93b9-8df5307b4641",
		Name:           "Mythos of Vadrok",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"W", "U"},
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("any number of target creatures and/or planeswalkers", effects.Or(effects.Creature(), effects.Planeswalker())).
			WithCount(0, 0).Dividing(effects.Divide(5)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DealDividedDamage(ctx); err != nil {
				return err
			}
			if ctx.ManaSpentOfColor("W") == 0 || ctx.ManaSpentOfColor("U") == 0 {
				return nil
			}
			until := effects.DurationUntilYourNextTurn(ctx, ctx.Controller())
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard || !effects.OnBattlefield(ctx.Game, t.ID) {
					continue
				}
				if err := (effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana)},
					Duration: until,
					Label:    "Mythos of Vadrok — can't attack or block, and its activated abilities can't be activated, until your next turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
