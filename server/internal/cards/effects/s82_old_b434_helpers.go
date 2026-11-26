package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ControlsExactlyOneCreatureAndIsIt is the "As long as you control exactly
// one creature, that creature ..." subject (Homicidal Seclusion): the
// source's controller controls exactly one creature, and `target` is it.
func ControlsExactlyOneCreatureAndIsIt(target *game.Card, g *game.Game, source *game.Card) bool {
	if !target.IsCreature() || target.Controller != source.Controller {
		return false
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == source.Controller && c.IsCreature() {
			n++
		}
	}
	return n == 1
}

// TargetPlayerLosesLifeAndYouGainThatMuch is Essence Harvest's body, with X
// the greatest power among creatures its controller controls, read once as
// it resolves: the target player loses X life and its controller gains X.
func TargetPlayerLosesLifeAndYouGainThatMuch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := B42GreatestPowerControlledBy(g, item.Controller)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if x <= 0 {
			return nil
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -x); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: x}.Apply(ctx)
	}
	return nil
}
