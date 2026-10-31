package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valentina Allegra de Fontaine — Legendary Creature — Human Spy Villain
// {2}{B}, 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Other Villains you control are Heroes in addition to their other types.
//	 {T}: Target Hero you control connives. Activate only as a sorcery."
//
// The static is a layer 4 type effect on every other Villain the controller
// controls, so the layered view of such a creature carries Hero as well as
// Villain (it can be targeted by her own ability, and is a Hero to anything
// else that asks, Heroic Return among them). Valentina is a Villain but not
// "other", so she is not made a Hero by it. The ability targets a Hero the
// controller controls as it goes on the stack and makes it connive on
// resolution (CR 701.50a): a target gone by then fizzles (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58ce3c79-8903-41be-8494-56d7953074ad",
		Name:         "Valentina Allegra de Fontaine",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Other Villains you control are Heroes in addition to their other types.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.HasSubtype(source.TextSubtype("Villain"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				hero := source.TextSubtype("Hero")
				for _, st := range c.Subtypes {
					if st == hero {
						return
					}
				}
				c.Subtypes = append(c.Subtypes, hero)
			},
		}},
		Activated: []ActivatedAbility{{
			Label:        "{T}: Target Hero you control connives. Activate only as a sorcery.",
			Cost:         TapCost(),
			Targets:      TargetCreature("target Hero you control", YouControl(), HasSubtype("Hero")),
			SorcerySpeed: true,
			Effect:       theTargetCreatureConnives,
		}},
	})
}
