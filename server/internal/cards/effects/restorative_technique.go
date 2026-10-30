package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restorative Technique — {2}{G} Sorcery (Marvel Super Heroes, #4455):
//
//	"Target player gains 2 life, then searches their library for a basic
//	 land card, puts it onto the battlefield tapped, then shuffles. Put a
//	 +1/+1 counter on up to one target creature."
//
// Two target clauses (CR 601.2c): slot 0 is the player, who gains the life
// and does the search (Fertilid's search for a basic land, run by the
// TARGET and not by the caster), slot 1 an "up to one" creature that gets
// the counter. Each slot is checked at resolution (CR 608.2b), so a player
// who has left the game gets nothing but the creature still gets its
// counter, and the other way round.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7b2606c-42b8-46d6-bcaa-8d4506b348eb",
		Name:         "Restorative Technique",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target player"),
			TargetCreature("up to one target creature").WithCount(0, 1),
		),
		Purpose: ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 2}),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if p, ok := ctx.ClauseTarget(0); ok && p.Kind == game.TargetPlayer {
				if err := (GainLife{Player: p.ID, Amount: 2}).Apply(ctx); err != nil {
					return err
				}
				if err := (SearchLibrary{
					Player: p.ID, Predicate: IsBasicLand, Dest: game.ZoneBattlefield, Limit: 1,
					Shuffle: true, TappedOnEntry: true,
					Reason: "Restorative Technique — a basic land, tapped",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if c, ok := ctx.ClauseTarget(1); ok && c.Kind == game.TargetCard {
				return AddCounter{Target: c.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
