package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sokka, Lateral Strategist — Legendary Creature — Human Warrior Ally
// {1}{W/U}{W/U}, 2/4 (Avatar: The Last Airbender, #4459):
//
//	"Vigilance
//	 Whenever Sokka and at least one other creature attack, draw a card."
//
// The engine stamps every attacker of a declaration before it emits the
// declaration's EventAttacks, so Sokka's own EventAttack sees the whole
// declaration: the ability triggers on it when another creature of his
// controller's is attacking too. Only Sokka's own event matches, so one
// declaration is one trigger. A creature put onto the battlefield
// attacking was never declared (CR 508.4) and is not yet there when
// Sokka is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "427f8ef9-12b3-4fa0-b303-28153566b6dc",
		Name:            "Sokka, Lateral Strategist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && len(b13AttackingCreaturesYouControl(g, source.Controller)) >= 2
			}, "Sokka, Lateral Strategist — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
