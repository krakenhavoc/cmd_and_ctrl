package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cunning Maneuver — Instant {1}{R} (Avatar: The Last Airbender, #4459):
//
//	"Target creature gets +3/+1 until end of turn.
//	 Create a Clue token."
//
// Creating the Clue is not investigating (the ruling). With its only
// target gone the spell doesn't resolve, so no Clue (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96335212-ec57-4beb-bab1-11c8eaa62119",
		Name:         "Cunning Maneuver",
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 1)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{
					Target: t.ID,
					Power:  3, Toughness: 1,
					Label: "Cunning Maneuver — +3/+1 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return CreateToken{Template: ClueToken(), N: 1}.Apply(ctx)
		},
	})
}
