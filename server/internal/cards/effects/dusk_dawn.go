package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dusk // Dawn — Split card with aftermath (EDHREC top 1,000, #3654):
//
//	Dusk {2}{W}{W} Sorcery: "Destroy all creatures with power 3 or greater."
//	Dawn {3}{W}{W} Sorcery — Aftermath: "Return all creature cards with
//	 power 2 or less from your graveyard to your hand."
//
// Two separate specs, the back half under "<oracle_id>#1" as Insult //
// Injury is. The engine already restricts an aftermath half to a
// graveyard and exiles it as it resolves from there (CR 702.127a, ADR
// 0103), so the card file is only the two instructions.
//
// Dusk reads each creature's effective power as it resolves, so a
// creature pumped in response is swept. Dawn reads the printed power of
// the cards in the graveyard, since a card there has no layer effects,
// and takes a snapshot first so the set is fixed (CR 608.2c).
//
// No simplification.
func init() {
	const oracle = "7bc3f92f-68a2-4934-afc4-89f6d0e8cf98"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Dusk",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), PowerGE(3))}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Dawn",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			p := ctx.PlayerByID(ctx.Controller())
			if p == nil || p.Graveyard == nil {
				return nil
			}
			var ids []uuid.UUID
			for _, c := range p.Graveyard.Cards {
				if c.IsCreature() && c.Power <= 2 {
					ids = append(ids, c.InstanceID)
				}
			}
			for _, id := range ids {
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
