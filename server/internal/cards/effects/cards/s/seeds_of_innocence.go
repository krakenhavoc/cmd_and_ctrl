package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seeds of Innocence — Sorcery {1}{G}{G} (Mirage):
//
//	"Destroy all artifacts. They can't be regenerated. The controller of each of
//	 those artifacts gains life equal to its mana value."
//
// Every artifact is destroyed together ignoring regeneration (CR 701.19c). The
// life is gained by the controller each artifact had as it left, equal to its
// mana value (the mass tail's last-known copies, CR 603.10), so an artifact
// that was not destroyed (indestructible) gives its controller nothing: "those
// artifacts" are the ones destroyed. A token or an artifact land of mana value
// 0 gives 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5ebad5c-9675-4a37-b76c-760416083939",
		Name:         "Seeds of Innocence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{
				Match:             effects.Artifact(),
				CantBeRegenerated: true,
				ThenTail:          ctx.ItemTail(seedsOfInnocenceLife),
			}.Apply(ctx)
		},
	})
}

// seedsOfInnocenceLife is "the controller of each of those artifacts gains
// life equal to its mana value". An on-disk key.
var seedsOfInnocenceLife = effects.MassTail("mass/seeds-of-innocence-life", func(ctx *effects.Context, swept []game.Card, _ int) error {
	for _, c := range swept {
		if err := (effects.GainLife{Player: c.Controller, Amount: c.ManaValue()}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
