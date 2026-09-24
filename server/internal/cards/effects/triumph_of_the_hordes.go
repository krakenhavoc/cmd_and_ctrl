package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triumph of the Hordes — Sorcery {2}{G}{G}:
//
//	"Until end of turn, creatures you control get +1/+1 and gain
//	 trample and infect."
//
// Overrun's shape with infect on it: every creature you control deals
// its combat damage this turn as poison counters to players and -1/-1
// counters to creatures (ADR 0056, #748), which is a kill from ten
// power of attackers.
//
// CR 611.2c: the affected set is locked when it resolves, in the
// BoostUntilEOT / GrantKeywordUntilEOT primitives, so a creature that
// arrives afterwards gets none of it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3ded0c0c-40ce-4d14-a9a6-b023bc19ee0e",
		Name:         "Triumph of the Hordes",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			yours := And(Creature(), YouControl())
			if err := (BoostUntilEOT{
				Match:     yours,
				Power:     1,
				Toughness: 1,
				Label:     "Triumph of the Hordes — +1/+1",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"trample", "infect"},
				Label:    "Triumph of the Hordes — trample and infect",
			}.Apply(ctx)
		},
	})
}
