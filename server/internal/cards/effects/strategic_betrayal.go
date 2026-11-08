package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strategic Betrayal — Sorcery {1}{B} (Tarkir: Dragonstorm, #4463):
//
//	"Target opponent exiles a creature they control and their
//	 graveyard."
//
// Doomfall's exile edict (msc17ExileEdict): the targeted opponent chooses
// one of the creatures they control, and a creature that leaves before
// they answer is withdrawn from the pick. Their graveyard is exiled in the
// same move as the chosen creature, once they have answered, so the two
// leave together as the one instruction says. An opponent with no
// creature is not asked, and their graveyard is still exiled. With the
// target gone (an opponent who left the game) the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0127c13f-7b0f-44c3-8b35-97f306c5f239",
		Name:         "Strategic Betrayal",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opp := TargetedPlayer(ctx)
			if opp == uuid.Nil {
				return nil
			}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == opp && c.IsCreature() {
					return msc17ExileEdict(ctx, opp, "Strategic Betrayal — exile a creature you control (your graveyard is exiled too)",
						Creature(), strategicBetrayalExile)
				}
			}
			return ctx.Game.ExileCardsKeyedForEffect(strategicBetrayalGraveyard(ctx.Game, opp), game.KeyedTail{})
		},
	})
}

// strategicBetrayalExile exiles the creature the targeted opponent chose
// together with their graveyard. An on-disk key: never renamed, never
// reused.
var strategicBetrayalExile = game.RegisterChooseCardsThen("choose-cards/strategic-betrayal-exile",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered {
			return nil
		}
		ids := append(append([]uuid.UUID(nil), r.Picked...), strategicBetrayalGraveyard(g, r.Chooser)...)
		return g.ExileCardsKeyedForEffect(ids, game.KeyedTail{})
	})

// strategicBetrayalGraveyard is every card in `player`'s graveyard,
// snapshotted before anything moves.
func strategicBetrayalGraveyard(g *game.Game, player uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		ids = append(ids, c.InstanceID)
	}
	return ids
}
