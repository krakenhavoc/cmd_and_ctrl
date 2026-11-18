package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shambling Ghast — Creature — Zombie {B}, 1/1 (AFR):
//
//	"When this creature dies, choose one —
//	 • Target creature an opponent controls gets -1/-1 until end of turn.
//	 • Create a Treasure token."
func init() {
	ghast := WhenThisDies("Shambling Ghast — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	ghast.Modes = ChooseOne(
		ModeDoing("Target creature an opponent controls gets -1/-1 until end of turn.",
			TargetCreature("target creature an opponent controls", OpponentControls()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return BoostUntilEOT{Target: t.ID, Power: -1, Toughness: -1,
					Label: "Shambling Ghast — -1/-1 until end of turn"}.Apply(ctx)
			}),
		ModeWithPurpose(ModeDoing("Create a Treasure token.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
			}), game.Purpose{Tokens: 1}),
	)
	Register(Spec{
		OracleID:     "e5e6229a-3422-4afd-9a87-86daa2d31d3a",
		Name:         "Shambling Ghast",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{ghast},
	})
}
