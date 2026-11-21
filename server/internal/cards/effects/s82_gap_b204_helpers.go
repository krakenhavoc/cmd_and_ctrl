package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// GapB204Dethrone is the dethrone triggered ability (CR 702.105a):
// "Whenever this creature attacks the player with the most life or tied for
// most life, put a +1/+1 counter on it." The defender is checked once, as
// the attack is declared; an attack at a planeswalker or battle never
// triggers it. label names the card for the stack overlay.
func GapB204Dethrone(label string) game.TriggeredAbility {
	return On(game.EventAttack, PreacherAttackedTheLifeLeader, label,
		Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))
}

// CopyTheFirstLegalTargetSpell is "Copy target spell. You may choose new
// targets for the copy." for the first target still legal as the ability
// resolves (CR 707.10c): Izzet Guildmage's and League Guildmage's body.
func CopyTheFirstLegalTargetSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return CopySpell{StackID: t.ID, Controller: item.Controller, ChooseNewTargets: true}.Apply(ctx)
	}
	return nil
}
