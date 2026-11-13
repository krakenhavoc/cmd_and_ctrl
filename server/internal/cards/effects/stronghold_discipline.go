package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stronghold Discipline — Sorcery {2}{B}{B}
// (Nemesis, #5618):
//
//	"Each player loses 1 life for each creature they control."
//
// Every player, the caster included, counts the creatures they control
// as the spell resolves, and loses that much life; a player with none
// loses nothing. The counts are taken before any life is lost. Life
// loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90e46d04-9d53-4a24-aff0-f8c61e1db97d",
		Name:         "Stronghold Discipline",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			players := tablePlayers(ctx)
			counts := make([]int, len(players))
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				for i, id := range players {
					if c.Controller == id && c.IsCreature() {
						counts[i]++
					}
				}
			}
			for i, id := range players {
				if counts[i] == 0 {
					continue
				}
				if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -counts[i]); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
