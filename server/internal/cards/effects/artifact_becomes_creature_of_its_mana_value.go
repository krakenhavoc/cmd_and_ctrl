package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// artifactBecomesCreatureOfItsManaValue is "target noncreature artifact
// becomes an artifact creature with power and toughness each equal to its
// mana value" for the duration `until`: Karn, the Great Creator's +1
// (until your next turn) and Xenic Poltergeist (until your next upkeep,
// #5737). The creature type is added in layer 4 and the base power and
// toughness set in layer 7b, both on the first still-legal card target.
// The mana value is read once, as the ability resolves (CR 608.2h).
func artifactBecomesCreatureOfItsManaValue(ctx *Context, until game.Duration, label string) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		mv := c.ManaValue()
		return ScopedEffectFor{
			Target:   t.ID,
			Mods:     append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(mv, mv)...),
			Duration: until,
			Label:    label,
		}.Apply(ctx)
	}
	return nil
}
