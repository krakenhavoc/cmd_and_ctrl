package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thelonite Druid — Creature — Human Cleric Druid {2}{G}, 1/1 (Fallen Empires, S82
// OLD batch 852, #5637):
//
//	"{1}{G}, {T}, Sacrifice a creature: Forests you control become 2/3 creatures
//	 until end of turn. They're still lands."
//
// Life's land animation (a Creature type and a base 2/3, the land types and
// abilities kept) over the Forests you control: the set is the Forests as the
// ability resolves, locked then (CR 611.2c), so one played afterwards is not
// animated. A Forest you haven't controlled continuously since your turn began
// can't attack or use a {T} ability (CR 302.6). The sacrifice is a cost; it may
// be the Druid itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adbc8970-1bbc-4f30-9a2e-eebaf074eae6",
		Name:         "Thelonite Druid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}, Sacrifice a creature: Forests you control become 2/3 creatures until end of turn. They're still lands.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost(), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Match: effects.And(effects.Land(), effects.HasSubtype("Forest"), effects.YouControl()),
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.SetBasePowerMod(2),
						game.SetBaseToughnessMod(3),
					},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Thelonite Druid — Forests you control are 2/3 creatures",
				}.Apply(ctx)
			},
		}},
	})
}
