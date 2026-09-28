package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Depth Defiler — {3}{U}{U} 3/5 Creature — Eldrazi:
//
//	"Devoid
//	 Kicker {C}
//	 When you cast this spell, choose one. If it was kicked, choose
//	 both instead.
//	 • Return target creature to its owner's hand.
//	 • Target player draws two cards, then discards a card."
//
// The trigger half of #1655: a "when you cast this spell" ability
// (FromStack, Desolation Twin's shape) whose mode count reads the
// SPELL's kicker. The mode_pick prompt is queued as the trigger goes
// on the stack (CR 603.3c), and the engine answers "was it kicked"
// from the spell's own PaidCost on the stack — so a kicked Defiler
// must take both bullets and an unkicked one exactly one. No "may":
// InsteadIf raises the minimum with the maximum.
//
// A bullet whose target clause has no legal target cannot be chosen
// (CR 603.3c), so a kicked cast onto an empty board still draws and
// discards. Devoid is colour data: the dump's colour list is already
// empty. The trigger resolves above the spell, so a counterspell aimed
// at the Defiler does not stop it. No simplification.
func init() {
	onCast := game.TriggeredAbility{
		FromStack: true,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID
		},
		// The CASTER controls the trigger (Desolation Twin's reason).
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, "Depth Defiler — choose one; both if kicked")
			item.Controller, item.Owner = ev.Actor, ev.Actor
			return item
		},
		// CR 608.2c: printed order, whatever order the modes were
		// picked in — the bounce lands before the draw and the
		// discard, so a creature bounced to the target player's hand
		// is a card they may discard.
		Effect: func(g *game.Game, item *game.StackItem) error {
			return BulletsInPrintedOrder(item, NewContext(g, item),
				BounceTheModesTarget,
				depthDefilerDrawThenDiscard)
		},
		Modes: ChooseOne(
			Mode("Return target creature to its owner's hand.",
				TargetCreature("target creature")),
			Mode("Target player draws two cards, then discards a card.",
				TargetPlayer("target player")),
		).InsteadIf(2, WasKicked),
	}
	Register(Spec{
		OracleID:     "8ca4ca66-30b1-4074-a2e3-545b7682381b",
		Name:         "Depth Defiler",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			Kicker("{C}"),
		},
		Triggered: []game.TriggeredAbility{onCast},
	})
}

// depthDefilerDrawThenDiscard is the second bullet.
func depthDefilerDrawThenDiscard(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	if err := (DrawCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   t.ID,
		Source:   item.SourceCardID,
		N:        1,
		Question: "Depth Defiler — discard a card",
	})
	return nil
}
