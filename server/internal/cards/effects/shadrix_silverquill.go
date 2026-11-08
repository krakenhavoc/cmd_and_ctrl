package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadrix Silverquill — {3}{W}{B} Legendary Creature — Elder Dragon 2/5
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying, double strike
//	 At the beginning of combat on your turn, you may choose two. Each
//	 mode must target a different player.
//	 • Target player creates a 2/1 white and black Inkling creature
//	   token with flying.
//	 • Target player draws a card and loses 1 life.
//	 • Target player puts a +1/+1 counter on each creature they control."
//
// "You may choose two" is exactly zero modes or exactly two (the
// ruling: "You can't choose only one mode"), and the choice is made as
// the ability is put on the stack (CR 603.3c), with the target players.
// So the "you may" is asked as the ability triggers
// (TriggerOptionalPrompt.BeforeStack): "no" is choosing no mode, and
// the ability is removed from the stack (CR 700.2b); "yes" asks for
// exactly two bullets (ChooseN 2, 2), then a target player for each.
// It is not CR 603.5's "you may" asked at resolution, which would let
// the controller decide after seeing every response.
//
// "Each mode must target a different player" is every clause Distinct,
// as on Casey & Raph, Hotheads: whichever bullet is targeted second
// can't name the player the first one did. Either may name the
// controller.
//
// The second bullet is two instructions (ModeSteps), so the life loss
// waits for a prompt the draw asks (CR 608.2c). The third puts the
// counters on each creature the target player controls as it resolves.
//
// No simplification.
func init() {
	combat := On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
		"Shadrix Silverquill — you may choose two; each mode must target a different player",
		func(*game.Game, *game.StackItem) error { return nil })
	combat.OptionalPrompt = &game.TriggerOptionalPrompt{
		Question:    "Shadrix Silverquill — choose two modes?",
		BeforeStack: true,
	}
	combat.Modes = ChooseN("Choose two", 2, 2,
		ModeDoing("Target player creates a 2/1 white and black Inkling creature token with flying.",
			Distinct(TargetPlayer("target player")), shadrixInkling),
		ModeSteps("Target player draws a card and loses 1 life.",
			Distinct(TargetPlayer("target player")),
			Instruction(shadrixDraw), Instruction(shadrixLoseLife)),
		ModeDoing("Target player puts a +1/+1 counter on each creature they control.",
			Distinct(TargetPlayer("target player")), shadrixCounters),
	)
	Register(Spec{
		OracleID:        "2143f413-7baa-4fd6-a7ca-32f52bfa553f",
		Name:            "Shadrix Silverquill",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike"},
		Triggered:       []game.TriggeredAbility{combat},
	})
}

// shadrixModePlayer is the target player of the bullet at `occ`.
func shadrixModePlayer(ctx *Context, occ int) (game.TargetRef, bool) {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return game.TargetRef{}, false
	}
	return t, true
}

// shadrixInkling is the first bullet: the Inkling under the target
// player's control.
func shadrixInkling(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := shadrixModePlayer(ctx, occ)
	if !ok {
		return nil
	}
	return CreateToken{Controller: t.ID, Template: TokenCard("2/1 white and black Inkling with flying"), N: 1}.Apply(ctx)
}

// shadrixDraw and shadrixLoseLife are the second bullet's two
// instructions, both to its target player.
func shadrixDraw(ctx *Context) error {
	t, ok := shadrixModePlayer(ctx, ctx.Occurrence())
	if !ok {
		return nil
	}
	return DrawCards{Player: t.ID, N: 1}.Apply(ctx)
}

func shadrixLoseLife(ctx *Context) error {
	t, ok := shadrixModePlayer(ctx, ctx.Occurrence())
	if !ok {
		return nil
	}
	return GainLife{Player: t.ID, Amount: -1}.Apply(ctx)
}

// shadrixCounters is the third bullet: a +1/+1 counter on each creature
// the target player controls as it resolves.
func shadrixCounters(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := shadrixModePlayer(ctx, occ)
	if !ok {
		return nil
	}
	var ids []game.Card
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == t.ID && c.IsCreature() {
			ids = append(ids, c)
		}
	}
	for _, c := range ids {
		if z := ctx.Game.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
