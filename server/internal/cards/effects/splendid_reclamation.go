package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Splendid Reclamation — Sorcery {3}{G} (EDHREC rank 874):
//
//	"Return all land cards from your graveyard to the battlefield
//	 tapped."
//
// The lands deck's mass recursion — every fetchland cracked, every
// land discarded to a wheel, back at once. The IDs are snapshotted
// before the first move (ReturnFromGraveyard mutates the pile being
// walked), and each land returns under its owner's control, which
// is the caster's: "your graveyard".
//
// Sandbox simplification, declared: the lands enter untapped and are
// tapped a beat later (Victimize's posture — ReturnFromGraveyard has
// no tapped flag). Anything watching for a tap event sees one; both
// steps happen inside one resolution, so nothing gets a window to
// tap the land for mana in between.
func init() {
	Register(Spec{
		OracleID:     "13fe5e46-77a6-45d8-ac0b-c3d740eccf86",
		Name:         "Splendid Reclamation",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The lands enter untapped and are tapped immediately afterwards, so anything watching for a land being tapped sees one."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			p := ctx.PlayerByID(ctx.Controller())
			if p == nil || p.Graveyard == nil {
				return nil
			}
			var lands []uuid.UUID
			for _, c := range p.Graveyard.Cards {
				if c.IsLand() {
					lands = append(lands, c.InstanceID)
				}
			}
			for _, id := range lands {
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
					return err
				}
				if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
