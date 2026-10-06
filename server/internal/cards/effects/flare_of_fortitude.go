package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flare of Fortitude — Instant {2}{W}{W}:
//
//	"You may sacrifice a nontoken white creature rather than pay this
//	 spell's mana cost.
//	 Until end of turn, your life total can't change, and permanents
//	 you control gain hexproof and indestructible."
//
// Three seams that each landed on their own. The alternative cost is
// Fireblast's sacrifice-instead component: the creature dies with the
// spell already on the stack, so a countered Flare still costs it. The
// life lock is Platinum Emperion's rule (CR 119.7, 119.8) on a
// until-end-of-turn duration. The grant is Heroic Intervention's mass
// hexproof-and-indestructible, snapshotted as the spell resolves
// (CR 611.2c) so a permanent that arrives later is unprotected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ff07b7c-e97a-4b57-bcf1-2f22c37a8bd6",
		Name:         "Flare of Fortitude",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(1, "a nontoken white creature", 0, Creature(), OfColor("W"), Not(IsTokenPredicate())),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (LockLifeTotal{
				Player:   ctx.Controller(),
				Label:    "Flare of Fortitude — your life total can't change",
				Duration: DurationUntilEndOfTurn(ctx),
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    YouControl(),
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "Flare of Fortitude — hexproof and indestructible",
			}.Apply(ctx)
		},
	})
}
