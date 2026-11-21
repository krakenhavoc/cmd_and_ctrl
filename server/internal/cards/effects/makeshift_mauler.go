package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Makeshift Mauler — Creature — Zombie Horror {3}{U}, 4/5 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"As an additional cost to cast this spell, exile a creature card from
//	 your graveyard."
//
// Cobbled Lancer's additional cost: the caster names the creature card as
// the spell is cast, and it is exiled with the spell on the stack
// (CR 601.2b, 601.2h). A caster with no creature card in their graveyard
// can't cast it (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03025ec0-75fe-439d-b0c1-9bdb0f0ff336",
		Name:         "Makeshift Mauler",
		Completeness: CompletenessFull,
		AdditionalCost: ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool {
			return c.IsCreature()
		}),
	})
}
