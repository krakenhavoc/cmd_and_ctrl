package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splinter & Leo, Father & Son — Legendary Creature — Mutant Ninja Rat
// Turtle {2}{W}, 2/2 (TMNT Commander, #5011):
//
//	"When Splinter & Leo enter, choose one or both. Each mode must target
//	 a different player.
//	 • Target player creates a 2/2 red Mutant creature token.
//	 • Put a +1/+1 counter on each other creature target player controls."
//
// A modal enters trigger whose two "target player" clauses are Distinct
// (tmcB7EachModeADifferentPlayer). Each bullet reads its own player,
// re-checked as it resolves (CR 608.2b). "Each other creature" is every
// creature that player controls but Splinter & Leo itself, read as the
// bullet resolves; a Splinter & Leo that left and came back in response
// is a new object (CR 400.7) and so one of the others.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Splinter & Leo — choose one or both; each mode targets a different player",
		func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = tmcB7EachModeADifferentPlayer(
		ModeDoing("Target player creates a 2/2 red Mutant creature token.", tmcB7TargetPlayer(),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				p, ok := tmcB7ModePlayer(ctx, occ)
				if !ok {
					return nil
				}
				return CreateToken{Controller: p, Template: TokenCard("2/2 red Mutant"), N: 1}.Apply(ctx)
			}),
		ModeDoing("Put a +1/+1 counter on each other creature target player controls.", tmcB7TargetPlayer(),
			splinterAndLeoCounters),
	)
	Register(Spec{
		OracleID:     "c01095ba-b9c9-44e0-97d1-35cc47c4ed04",
		Name:         "Splinter & Leo, Father & Son",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

// splinterAndLeoCounters is the second bullet: one +1/+1 counter on each
// creature the target player controls other than Splinter & Leo.
func splinterAndLeoCounters(item *game.StackItem, ctx *Context, occ int) error {
	p, ok := tmcB7ModePlayer(ctx, occ)
	if !ok {
		return nil
	}
	var ids []game.Card
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != p || !c.IsCreature() {
			continue
		}
		if c.InstanceID == item.SourceCardID && !ctx.isNewSourceObject(c.InstanceID) {
			continue
		}
		ids = append(ids, c)
	}
	for _, c := range ids {
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
