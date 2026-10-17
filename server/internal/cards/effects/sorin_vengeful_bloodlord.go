package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorin, Vengeful Bloodlord — Legendary Planeswalker — Sorin
// {2}{W}{B}, loyalty 4 (#4077):
//
//	"During your turn, creatures and planeswalkers you control have
//	 lifelink.
//	 +2: Sorin deals 1 damage to target player or planeswalker.
//	 −X: Return target creature card with mana value X from your
//	     graveyard to the battlefield. That creature is a Vampire in
//	     addition to its other types."
//
// THE STATIC is a layer-6 keyword grant (CR 613.1f) whose AppliesTo asks,
// on every recompute, whether the source's controller is the active
// player, the Anara, Wolvid Familiar shape: it is a window on a
// continuous effect (CR 611.2), not a trigger, so nothing is unwound when
// the turn passes. It reaches planeswalkers as well as creatures, and
// Sorin himself: the engine reads lifelink off whatever object dealt the
// damage, so his own +2 gains his controller 1 life on their turn
// (CR 702.15b, CR 120.3), and a planeswalker that is a creature through a
// type-changing effect has it once.
//
// THE −X is a loyalty cost of X (LoyaltyMinusX, CR 107.3a, CR 606.6): the
// activator announces X, and the target clause holds the graveyard card to
// mana value exactly X (WithManaValueEqualsX), at announce and again as the
// ability resolves. The creature returns under its owner's control, which
// is yours because it comes from your graveyard. "That creature is a
// Vampire in addition to its other types" has no stated duration (CR
// 611.2a), so it is a layer-4 subtype add pinned to the returned object
// (CR 611.2c, CR 400.7) and read through Sorin's own text (CR 612.2). It is
// applied from the return's continuation, after any entry question.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cffb615c-2de0-4c65-9d0b-08a649ba0695",
		Name:            "Sorin, Vengeful Bloodlord",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Static: []game.StaticAbility{{
			Label: "During your turn, creatures and planeswalkers you control have lifelink.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller &&
					(target.IsCreature() || target.IsPlaneswalker()) &&
					isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Abilities, "lifelink") {
					c.Abilities = append(c.Abilities, "lifelink")
				}
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Sorin deals 1 damage to target player or planeswalker.",
				Cost:    LoyaltyCost(2),
				Targets: targetPlayerOrPlaneswalker(),
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Effect:  sourceDealsDamageToEachLegalTarget(1),
			},
			{
				Label: "−X: Return target creature card with mana value X from your graveyard to the battlefield. That creature is a Vampire in addition to its other types.",
				Cost:  LoyaltyMinusX(),
				Targets: TargetCardInGraveyard("target creature card with mana value X from your graveyard",
					Creature(), YouOwn()).WithManaValueEqualsX(),
				Effect: sorinVengefulReturn,
			},
		},
	})
}

// sorinVengefulReturn is the −X: the creature comes back, then it is a
// Vampire as well.
func sorinVengefulReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets: ids,
		Then: func(ctx *Context, entered []uuid.UUID) error {
			for _, id := range entered {
				if err := (ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Vampire")...)},
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Sorin, Vengeful Bloodlord — it is a Vampire in addition to its other types",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(ctx)
}
