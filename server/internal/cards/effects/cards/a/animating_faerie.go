package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Animating Faerie // Bring to Life — Creature — Faerie {2}{U}, 2/2 //
// Sorcery — Adventure {2}{U} (Throne of Eldraine, #5516):
//
//	Animating Faerie — "Flying"
//	Bring to Life — "Target noncreature artifact you control becomes a
//	 0/0 artifact creature. Put four +1/+1 counters on it."
//
// Both faces register, the Adventure under "<oracle_id>#1" (Besotted
// Knight's shape). The type change has no stated duration, so it is
// indefinite and pinned to that object (CR 611.2a; Case of the Filched
// Falcon's shape). It becomes a creature first and the counters follow, in
// printed order (CR 608.2c); an artifact that became a creature in response
// is no longer a legal target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	const oracle = "db6fbdc9-1c45-416f-a08e-2082f65adb65"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Animating Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Bring to Life",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target noncreature artifact you control", effects.Artifact(), effects.Noncreature(), effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			mods := append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(0, 0)...)
			duration := ctx.Game.PinnedTo(game.IndefiniteDuration(), id)
			if err := (effects.ScopedEffectFor{Target: id, Mods: mods, Duration: duration,
				Label: "Bring to Life — becomes a 0/0 artifact creature"}).Apply(ctx); err != nil {
				return err
			}
			return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 4}.Apply(ctx)
		},
	})
}
