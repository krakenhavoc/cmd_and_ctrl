package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f334087a-deda-4e9e-ab58-68f7e226d7ce",
		Name:         "Ant-Man, Elusive Avenger",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedByComparing(OnSelf(), GreaterPowerThanAttacker(), "creatures with greater power"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Ant-Man — create that many Treasure tokens",
				Do(CreateTokensForEach{
					Template: TreasureToken(),
					Name:     "the combat damage dealt",
					Count:    func(ctx *Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
	})
}
