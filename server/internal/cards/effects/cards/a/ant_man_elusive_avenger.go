package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ant-Man, Elusive Avenger — Legendary Creature — Human Rogue Hero
// {1}{U}{R}, 1/2 (Marvel Super Heroes Commander, #4455):
//
//	"Ant-Man can't be blocked by creatures with greater power.
//	 Whenever Ant-Man deals combat damage to a player, create that many
//	 Treasure tokens."
//
// The evasion is Locke's block rule (skulk's text without the keyword);
// the Treasures are one per point of the damage that triggered it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f334087a-deda-4e9e-ab58-68f7e226d7ce",
		Name:         "Ant-Man, Elusive Avenger",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedByComparing(effects.OnSelf(), effects.GreaterPowerThanAttacker(), "creatures with greater power"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Ant-Man — create that many Treasure tokens",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TreasureToken(),
					Name:     "the combat damage dealt",
					Count:    func(ctx *effects.Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
	})
}
