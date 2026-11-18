package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigarda's Summons — Enchantment {4}{W}{W} (Innistrad: Crimson Vow, S82):
//
//	"Creatures you control with +1/+1 counters on them have base power
//	 and toughness 4/4, have flying, and are Angels in addition to their
//	 other types."
//
// Three statics over one set, each in its own layer: the Angel subtype in
// layer 4 (CR 613.1d), flying in layer 6 (CR 613.1f) and the 4/4 base in
// layer 7b (CR 613.4b), so a +1/+1 counter still adds on top of the 4/4.
// The set is read off the live card's counters, as Exava and Drix do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fd06710-0721-4e3e-90ed-d0ccecb1919a",
		Name:         "Sigarda's Summons",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Creatures you control with +1/+1 counters on them are Angels in addition to their other types.",
			Layer:     game.Layer4Type,
			AppliesTo: sigardasSummonsCreature,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				angel := source.TextSubtype("Angel")
				if !eotHasType(c.Subtypes, angel) {
					c.Subtypes = append(c.Subtypes, angel)
				}
			},
		}, {
			Label:     "Creatures you control with +1/+1 counters on them have flying.",
			Layer:     game.Layer6Ability,
			AppliesTo: sigardasSummonsCreature,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
			},
		}, {
			Label:     "Creatures you control with +1/+1 counters on them have base power and toughness 4/4.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: sigardasSummonsCreature,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = 4
				c.Toughness = 4
			},
		}},
	})
}

// sigardasSummonsCreature is "creature you control with a +1/+1 counter on it".
func sigardasSummonsCreature(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller && hasCounterOfKind(*target, game.CounterPlusOne)
}
