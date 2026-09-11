package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rise of the Dark Realms — Sorcery {7}{B}{B} (EDHREC rank 501):
//
//	"Put all creature cards from all graveyards onto the battlefield
//	 under your control."
//
// Nine mana to take every dead creature at the table. Every
// graveyard is walked, the caster's included, and each creature
// card returns UNDER THE CASTER'S CONTROL — the Controller field on
// ReturnFromGraveyard, which is what makes an opponent's dead bomb
// yours rather than theirs (the Reanimate lesson). The instance IDs
// are snapshotted before any card moves, because each return
// removes a card from the pile being ranged over, and every returned
// creature fires its own ETB.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5223a09-f732-4747-8914-e6546ab0ef4c",
		Name:         "Rise of the Dark Realms",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var dead []uuid.UUID
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.IsCreature() {
						dead = append(dead, c.InstanceID)
					}
				}
			}
			for _, id := range dead {
				if err := (ReturnFromGraveyard{
					Target:     id,
					Dest:       game.ZoneBattlefield,
					Controller: ctx.Controller(),
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
