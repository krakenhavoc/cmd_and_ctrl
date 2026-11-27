package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mercadia's Downfall — Instant {2}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Each attacking creature gets +1/+0 until end of turn for each nonbasic land
//	 defending player controls."
//
// The set of attackers is fixed as the spell resolves (CR 611.2c). "Defending
// player" is read per attacker, the player (or the controller of the
// planeswalker or battle) that creature is attacking, so in a multiplayer game
// each attacker is pumped by the lands of its own defender. A nonbasic land is a
// land without the basic supertype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c1201b2-5036-4e68-b36d-cdaf54b4993d",
		Name:         "Mercadia's Downfall",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx = ctx.AsGroupMember()
			g := ctx.Game
			type pump struct {
				id uuid.UUID
				n  int
			}
			var pumps []pump
			for _, c := range g.Battlefield.Cards {
				if !c.IsCreature() || c.AttackingTarget == uuid.Nil {
					continue
				}
				defender := g.DefendingPlayerForAttackerForEffect(c.InstanceID)
				n := 0
				for _, l := range g.Battlefield.Cards {
					if l.Controller == defender && effects.NonbasicLand()(g, defender, l) {
						n++
					}
				}
				if n > 0 {
					pumps = append(pumps, pump{c.InstanceID, n})
				}
			}
			for _, p := range pumps {
				if err := (effects.BoostUntilEOT{Target: p.id, Power: p.n, Label: "Mercadia's Downfall — +1/+0 for each nonbasic land defending player controls"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
