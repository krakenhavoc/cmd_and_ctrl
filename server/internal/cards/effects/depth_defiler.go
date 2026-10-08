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
// discards. Devoid is declared in PrintedKeywords, and the engine reads
// it as CR 702.114a's colour-defining ability (#2152), so the card is
// colourless in every zone. The trigger resolves above the spell, so a
// counterspell aimed
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
		// CR 608.2c: the engine runs the bullets in printed order,
		// whatever order the modes were picked in — the bounce lands
		// before the draw and the discard, so a creature bounced to
		// the target player's hand is a card they may discard.
		// The trigger's own body is empty; each bullet carries its
		// own (ModeDoing), which the engine runs after it. A row with a
		// Build declares its Effect (ADR 0041 P9).
		Effect: func(*game.Game, *game.StackItem) error { return nil },
		Modes: ChooseOne(
			ModeDoing("Return target creature to its owner's hand.",
				TargetCreature("target creature"),
				BounceTheModesTarget),
			ModeWithPurpose(ModeDoing("Target player draws two cards, then discards a card.",
				TargetPlayer("target player"),
				depthDefilerDrawThenDiscard),
				ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, Discards: 1})),
		).InsteadIf(2, WasKicked),
	}
	Register(Spec{
		OracleID:        "8ca4ca66-30b1-4074-a2e3-545b7682381b",
		Name:            "Depth Defiler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
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
	return drawThenDiscard(ctx.Game, t.ID, item.SourceCardID, 2, 1, "Depth Defiler — discard a card")
}
