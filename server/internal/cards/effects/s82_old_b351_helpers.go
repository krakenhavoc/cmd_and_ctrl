package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB351TappedAttackingTokenLikeTheAttacker is "create a <template> token
// that's tapped and attacking" off an attack trigger: Adeline's shape, shared by
// Battle Cry Goblin's pack tactics and Mardu Ascendancy. The token enters tapped
// attacking the player or planeswalker the triggering creature attacked, and
// was never declared as an attacker (CR 508.4), so nothing that watches attacks
// sees it.
func OldB351TappedAttackingTokenLikeTheAttacker(template string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		tmpl := TokenCard(template)
		tmpl.Tapped = true
		tmpl.AttackingTarget = ctx.Trigger().Event.Target
		return CreateToken{Controller: item.Controller, Template: tmpl, N: 1}.Apply(ctx)
	}
}
