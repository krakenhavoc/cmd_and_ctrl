package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadow in the Warp — Enchantment {1}{R}{G}:
//
//	"The first creature spell you cast each turn costs {2} less to
//	 cast.
//	 Whenever an opponent casts their first noncreature spell each
//	 turn, this enchantment deals 2 damage to that player."
//
// Both halves read the per-player cast tally (Game.SpellsCastThisTurn),
// which counts every spell cast this turn whether or not Shadow in the
// Warp was there to see it (the 2022-10-07 ruling).
//
// The discount is priced at CR 601.2f, before the spell becomes cast
// and is counted (CR 601.2i), so "the first" is "no creature spell
// cast yet this turn": Total minus Noncreature is zero. It reduces
// generic mana only (CostsLess).
//
// The trigger is Esper Sentinel's condition
// (AnOpponentCastTheirFirstNoncreatureSpellThisTurn): an opponent's
// spell that is not a creature spell, and the tally, bumped before
// EventCast fires, says it is that player's first noncreature spell. The
// enchantment deals the damage, so it is noncombat damage from a
// red-green source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d81db0d-2f56-42ac-92ce-249a34d262d4",
		Name:         "Shadow in the Warp",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "The first creature spell you cast each turn costs {2} less to cast.",
				YourSpell(), CreatureSpell(), func(q game.CostQuery) bool {
					if q.Game == nil {
						return false
					}
					t := q.Game.CastTallyFor(q.Controller)
					return t.Total-t.Noncreature == 0
				}),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnOpponentCastTheirFirstNoncreatureSpellThisTurn,
				"Shadow in the Warp — 2 damage to that player", func(g *game.Game, item *game.StackItem) error {
					return DealDamage{
						Source: item.SourceCardID,
						Target: item.Trigger.Event.Actor,
						Amount: 2,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
