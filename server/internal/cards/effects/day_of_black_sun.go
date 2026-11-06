package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Day of Black Sun — Sorcery {X}{B}{B}:
//
//	"Each creature with mana value X or less loses all abilities until
//	 end of turn. Destroy those creatures."
//
// The set is fixed once, as the spell resolves (CR 611.2c): every
// creature whose mana value on the battlefield is X or less (an {X} in a
// permanent's cost counts as 0, CR 202.3e; a token's is 0). Those
// creatures lose all abilities until end of turn (a layer 6 effect,
// CR 613.1f), and then exactly those creatures are destroyed, at once.
// Losing indestructible first means an indestructible creature in the
// set dies; a creature that would have said "when this dies" no longer
// has that ability as it leaves, so it does not trigger (CR 603.10a, the
// ruling). Regeneration still applies, since the card doesn't forbid it.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID: "dc857129-533c-41d1-8e8b-1c0443030d69",
		Name:     "Day of Black Sun",
		// ADR 0126 §6: only creatures with mana value X or less.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var those []uuid.UUID
			for _, c := range MatchingBattlefield(ctx, And(Creature(), ManaValueLE(ctx.X()))) {
				those = append(those, c.InstanceID)
			}
			if len(those) == 0 {
				return nil
			}
			if err := (ScopedEffectFor{
				Match:    b10OneOf(those),
				Mods:     []game.Mod{game.LoseAllAbilitiesMod()},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Day of Black Sun — loses all abilities until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return DestroyAllMatching{Match: b10OneOf(those)}.Apply(ctx)
		},
	})
}
