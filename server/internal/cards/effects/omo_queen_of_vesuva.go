package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Omo, Queen of Vesuva — Legendary Creature — Shapeshifter Noble {2}{G/U},
// 1/5:
//
//	"Whenever Omo enters or attacks, put an everything counter on each of up
//	 to one target land and up to one target creature.
//	 Each land with an everything counter on it is every land type in
//	 addition to its other types.
//	 Each nonland creature with an everything counter on it is every
//	 creature type."
//
// One triggered ability with two conditions (WhenThisEntersOrAttacks), and
// two "up to one" clauses announced separately, each re-checked as it
// resolves (CR 608.2b). The counter is the "everything" kind. The two
// statics are layer-4 type changes read off the counter on the object
// itself, whoever controls it (CR 205.3i, CR 205.3m): a land with one gets
// every land type in addition to its own (and so the mana abilities those
// types carry, CR 305.6, for the basic ones), and a creature with one that
// is not also a land is every creature type, as a changeling is.
func init() {
	const name = "Omo, Queen of Vesuva"
	const kind = "everything"
	trigger := Targeting(WhenThisEntersOrAttacks(name+" — put an everything counter on each of up to one target land and up to one target creature",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for slot := 0; slot < 2; slot++ {
				for _, t := range ctx.ClauseTargets(slot) {
					if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
						continue
					}
					if err := (AddCounter{Target: t.ID, Kind: kind, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		}), Clauses(
		TargetPermanent("up to one target land", Land()).WithCount(0, 1),
		TargetCreature("up to one target creature").WithCount(0, 1),
	))
	Register(Spec{
		OracleID:     "021b91ea-18d8-4b50-97da-dd003380cd68",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
		Static: []game.StaticAbility{
			{
				Label: "Each land with an everything counter on it is every land type in addition to its other types.",
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
					return target.IsLand() && target.Counters[kind] > 0
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, t := range game.LandTypes {
						if !slices.Contains(c.Subtypes, t) {
							c.Subtypes = append(c.Subtypes, t)
						}
					}
				},
			},
			{
				Label: "Each nonland creature with an everything counter on it is every creature type.",
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
					return target.IsCreature() && !target.IsLand() && target.Counters[kind] > 0
				},
				Apply: applyAllCreatureTypes,
			},
		},
	})
}
