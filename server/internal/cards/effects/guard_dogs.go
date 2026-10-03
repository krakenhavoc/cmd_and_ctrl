package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guard Dogs — Creature — Dog {3}{W}, 2/2:
//
//	"{2}{W}, {T}: Choose a permanent you control. Prevent all combat damage target creature would deal this turn if it shares a color with that permanent."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the target is chosen as the
// ability is activated, the permanent as it resolves (it does not
// target), and the colours are compared then and only then (its
// rulings): a match makes a combat-damage shield with the target as its
// source, pinned now (CR 400.7), that a later colour change does not
// undo; no match makes nothing.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "490af644-2e85-49c7-af63-d5f0a9babff8",
		Name:         "Guard Dogs",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}, {T}: Choose a permanent you control. Prevent all combat damage target creature would deal this turn if it shares a color with that permanent.",
			Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ChoosePermanents{
					Question: "Guard Dogs — choose a permanent you control",
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						var ids []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == of {
								ids = append(ids, c.InstanceID)
							}
						}
						return ids, 1, 1
					},
					Then: guardDogsCompare,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// guardDogsCompare shields the target when it shares a colour with the
// chosen permanent, both read now.
func guardDogsCompare(ctx *Context, picked game.PromptedPicks) error {
	chosen := picked.Cards()
	if len(chosen) == 0 {
		return nil
	}
	ctx.Game.RecomputeLayersIfStaleLocked()
	perm, ok := ctx.Game.LookupCardForEffect(chosen[0])
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok || !sharesAColorNow(c, perm) {
			return nil
		}
		return PreventDamageFromSource{From: t.ID, CombatOnly: true, Protect: ShieldAnything}.Apply(ctx)
	}
	return nil
}

// sharesAColorNow reports whether two objects have a colour in common,
// read off their current characteristics.
func sharesAColorNow(a, b game.Card) bool {
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		if a.HasColor(color) && b.HasColor(color) {
			return true
		}
	}
	return false
}
