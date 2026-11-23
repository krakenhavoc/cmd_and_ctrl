package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demolisher Spawn — Enchantment Creature — Horror {5}{G}{G}, 7/7 (Duskmourn
// Commander):
//
//	"Trample, haste
//	 Delirium — Whenever this creature attacks, if there are four or more card
//	 types among cards in your graveyard, other attacking creatures get +4/+4
//	 until end of turn."
//
// An intervening if (CR 603.4): checked as the Spawn attacks and again as the
// trigger resolves. The set of other attackers is locked as the effect begins
// (CR 611.2c), so a creature that joins the attack later gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "17bdd827-2e89-4559-939a-fc5c153189ed",
		Name:            "Demolisher Spawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.B24CondDelirium{}, effects.On(game.EventAttack, effects.AllOf(effects.ThisAttacked, effects.YouHaveDeliriumWhen),
				"Demolisher Spawn — other attacking creatures get +4/+4 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.AttackingCreature(), effects.OtherThan(item.SourceCardID)),
						Power:     4,
						Toughness: 4,
						Label:     "Demolisher Spawn",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
