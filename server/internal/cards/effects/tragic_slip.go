package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tragic Slip — Instant {B}:
//
//	"Target creature gets -1/-1 until end of turn.
//	 Morbid — That creature gets -13/-13 until end of turn instead if a
//	 creature died this turn."
//
// Morbid is read as the spell resolves, off the turn tally of creatures
// that went to a graveyard from the battlefield (tokens included), the
// same cell Gruesome Discovery reads. A creature that dies in response
// to the Slip, including one its own controller sacrifices, turns morbid
// on. The size is chosen at resolution and then locked in, so a creature
// dying after the Slip resolves does not grow the penalty.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa6e2a47-a8cb-4224-ae3e-f61c90477285",
		Name:         "Tragic Slip",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			n, label := 1, "Tragic Slip — -1/-1"
			if b11CreaturesDiedThisTurn(ctx.Game) > 0 {
				n, label = 13, "Tragic Slip — morbid, -13/-13"
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -n,
				Toughness: -n,
				Label:     label,
			}.Apply(ctx)
		},
	})
}
