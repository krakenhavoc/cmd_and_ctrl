package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// White Widow, Free Agent — Legendary Creature — Human Hero Villain
// {3}{W}, 2/3 (Marvel Super Heroes, #4455):
//
//	"When White Widow enters, choose one —
//	 • Put a +1/+1 counter on each of up to two target creatures.
//	 • Return target artifact or enchantment card from your graveyard to
//	   your hand."
//
// A modal enter trigger (CR 700.2b): the mode is chosen as it goes on
// the stack, and only a mode with a legal target can be chosen. The first
// bullet's targets are "up to two", so it is always choosable; each target
// is re-checked as the bullet resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f0c055f-f614-44b0-9374-c53541f89185",
		Name:         "White Widow, Free Agent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "White Widow, Free Agent — choose one",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeDoing("Put a +1/+1 counter on each of up to two target creatures.",
					TargetCreature("up to two target creatures").WithCount(0, 2), mshB11CountersOnEachModeTarget),
				ModeDoing("Return target artifact or enchantment card from your graveyard to your hand.",
					TargetCardInGraveyard("target artifact or enchantment card in your graveyard", Or(Artifact(), Enchantment()), YouOwn()),
					mshB11ReturnModeTargetToHand),
			),
		}},
	})
}

// mshB11CountersOnEachModeTarget puts a +1/+1 counter on every still-legal
// target of the bullet.
func mshB11CountersOnEachModeTarget(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if t.ID == uuid.Nil || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// mshB11ReturnModeTargetToHand returns the bullet's graveyard target to
// its owner's hand, if it is still in a graveyard.
func mshB11ReturnModeTargetToHand(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
}
