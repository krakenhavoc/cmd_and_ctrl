package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hostile Takeover — Sorcery {2}{U}{B}{R} (Streets of New Capenna, #5491):
//
//	"Up to one target creature has base power and toughness 1/1 until end
//	 of turn. Up to one other target creature has base power and toughness
//	 4/4 until end of turn. Then Hostile Takeover deals 3 damage to each
//	 creature."
//
// Two "up to one" clauses; the second must differ from the first ("other",
// Distinct). Each base size is a layer 7b effect for the turn on the object
// that was targeted, so counters and pumps still apply on top. A target that
// is gone or illegal on resolution gets nothing (CR 608.2b), and with no
// target the spell still resolves and deals its damage. The damage comes
// last, to every creature as one instance, so a creature set to 4/4 takes 3
// and lives while a 1/1 dies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c7f91af-8bac-4da1-b9e5-f15989c9d53c",
		Name:         "Hostile Takeover",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		Targets: Clauses(
			TargetCreature("up to one target creature").WithCount(0, 1),
			Distinct(TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for slot, size := range []int{1, 4} {
				t, ok := ctx.ClauseTarget(slot)
				if !ok || t.Kind != game.TargetCard {
					continue
				}
				if err := (ScopedEffectFor{
					Target: t.ID,
					Mods: []game.Mod{
						game.SetBasePowerMod(size),
						game.SetBaseToughnessMod(size),
					},
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Hostile Takeover — base power and toughness " + string(rune('0'+size)) + "/" + string(rune('0'+size)),
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return damageEachMatching(ctx, Creature(), 3)
		},
	})
}
