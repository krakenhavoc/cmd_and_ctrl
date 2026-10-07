package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glyph of Destruction — Instant {R}:
//
//	"Target blocking Wall you control gets +10/+0 until end of combat.
//	 Prevent all damage that would be dealt to it this turn. Destroy it
//	 at the beginning of the next end step."
//
// ADR 0108 amendment 2026-10-07 (#2027): the +10/+0 is a layer-7c
// modification that lasts until the combat phase it was made in ends
// (game.UntilEndOfCombat, CR 511.3), not "this turn" — in a turn with an
// additional combat phase the Wall is a 1/x again for the second one. The
// prevention shield is the ordinary this-turn one and the destruction is
// a CR 603.7 delayed trigger that names the Wall as the object it is now
// (CR 400.7), so a Wall that was flickered or has already died is left
// alone.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "cca68900-2093-4cae-b757-efc7dd807e18",
		Name:         "Glyph of Destruction",
		Completeness: CompletenessFull,
		Targets: TargetCreature("target blocking Wall you control",
			And(BlockingCreature(), OfCreatureType("Wall"), YouControl())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				d, ok := DurationUntilEndOfCombat(ctx)
				if !ok {
					return nil
				}
				if err := (ScopedEffectFor{
					Target:   t.ID,
					Mods:     []game.Mod{game.ModifyPTMod(10, 0)},
					Duration: d,
					Label:    "Glyph of Destruction — +10/+0 until end of combat",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (PreventDamageFromSource{Protect: ShieldObject(t.ID), Label: "Glyph of Destruction — prevent all damage to it this turn"}).Apply(ctx); err != nil {
					return err
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				return ScheduleDelayedTrigger{
					Label:  "Glyph of Destruction — destroy the Wall",
					Cards:  []uuid.UUID{t.ID},
					Body:   destroyTheObjectBody,
					Params: game.EffectParams{Object: game.ObjectRef{ID: t.ID, Epoch: c.ObjectEpoch}},
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// destroyTheObject is destroyTheObjectBody: destroy the permanent
// p.Object names if it is still that object on the battlefield.
func destroyTheObject(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !onBattlefield(g, id) {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(NewContext(g, item))
}
