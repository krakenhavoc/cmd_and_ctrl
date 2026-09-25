package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Golgari Charm — Instant {B}{G}:
//
//	"Choose one —
//	 • All creatures get -1/-1 until end of turn.
//	 • Destroy target enchantment.
//	 • Regenerate each creature you control."
//
// Three unrelated modes, none of them needing anything the catalog
// didn't already have: the board-wide -1/-1 is Languish's
// BoostUntilEOT{Match: Creature()} narrowed to one turn's Languish;
// the enchantment destruction is Naya Charm's single-target shape;
// the regeneration shields every creature the caster controls with
// Regenerate, one shield per creature, the same primitive a
// single-target regenerate spell uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0ab7166-9fcf-46e5-af2c-b7f628db4789",
		Name:         "Golgari Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("All creatures get -1/-1 until end of turn."),
			Mode("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment())),
			Mode("Regenerate each creature you control."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return BoostUntilEOT{
					Match:     Creature(),
					Power:     -1,
					Toughness: -1,
					Label:     "Golgari Charm — -1/-1",
				}.Apply(ctx)
			case ctx.HasMode(1):
				return destroyChosenPermanent(ctx.Game, item)
			case ctx.HasMode(2):
				return regenerateEachCreatureControlledBy(ctx, ctx.Controller())
			}
			return nil
		},
	})
}

// regenerateEachCreatureControlledBy is "regenerate each creature you
// control": snapshot the set, then shield each one, the same posture
// putPlusOneCounterOnEachCreatureControlledBy uses.
func regenerateEachCreatureControlledBy(ctx *Context, player uuid.UUID) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (Regenerate{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
