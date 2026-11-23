package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const hezrouOracleID = "65d257c2-b3be-4244-a74c-bc4b5d7cedc3"

// Hezrou // Demonic Stench — Adventure.
//
//	Hezrou {5}{B}{B} Creature — Frog Demon 6/6
//	  "Whenever one or more creatures you control become blocked, each
//	   blocking creature gets -1/-1 until end of turn."
//	Demonic Stench {B} Instant — Adventure
//	  "Each creature that blocked this turn gets -1/-1 until end of turn."
//
// The trigger is one per batch of attackers becoming blocked (CR 509.3d).
// "Each blocking creature" is read as it resolves: every creature then
// blocking, whoever controls it. Demonic Stench reads this turn's block
// declarations, so a creature that blocked and was removed from combat since
// is still hit.
func init() {
	effects.Register(effects.Spec{
		OracleID:     hezrouOracleID,
		Name:         "Hezrou",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventBecomesBlocked,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := g.LookupCardForEffect(ev.Target)
					return ok && ev.Kind == game.EventBecomesBlocked && c.Controller == source.Controller
				},
				"Hezrou — each blocking creature gets -1/-1 until end of turn",
				effects.Do(effects.BoostUntilEOT{Match: effects.BlockingCreature(), Power: -1, Toughness: -1, Label: "Hezrou — -1/-1"}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     hezrouOracleID + "#1",
		Name:         "Demonic Stench",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			blocked := effects.GapB213CreaturesThatBlockedThisTurn(ctx.Game)
			return effects.BoostUntilEOT{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return blocked[c.InstanceID] },
				Power: -1, Toughness: -1, Label: "Demonic Stench — -1/-1",
			}.Apply(ctx)
		},
	})
}
