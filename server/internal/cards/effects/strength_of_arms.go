package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strength of Arms — Instant {W} (Innistrad Remastered):
//
//	"Target creature gets +2/+2 until end of turn. If you control an
//	 Equipment, create a 1/1 white Human Soldier creature token."
//
// The Equipment is looked for as the spell resolves, after the pump. If the
// target is gone the whole spell does nothing (CR 608.2b).
func init() {
	purpose := ForTargets(BenefitUntilEOT(0, 2, 2))
	purpose.Tokens, purpose.CreatureTokens = 1, 1
	Register(Spec{
		OracleID:     "40ff84c7-cd72-49a6-86ba-a8d23c819ee7",
		Name:         "Strength of Arms",
		Purpose:      purpose,
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Strength of Arms — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			if !b05ControlsSubtype(ctx.Game, ctx.Controller(), "Equipment") {
				return nil
			}
			return CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1}.Apply(ctx)
		},
	})
}
