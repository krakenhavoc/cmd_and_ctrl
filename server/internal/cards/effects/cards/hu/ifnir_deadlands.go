package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ifnir Deadlands — Land — Desert (EDHREC rank 4473):
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add {B}.
//	 {2}{B}{B}, {T}, Sacrifice a Desert: Put two -1/-1 counters on
//	 target creature an opponent controls. Activate only as a
//	 sorcery."
//
// A painless colourless land, a painland for black, and a removal
// spell on a land. The third ability is why a mono-black deck plays
// it over a Swamp: four mana and the land itself answers a two-
// toughness creature outright and shrinks anything bigger, from a
// zone no counterspell reaches.
//
// Three abilities, and the split between them is the point:
//
//   - {T}: Add {C} is free.
//   - {T}, Pay 1 life: Add {B} is a real COST, not a rider — the
//     land is unactivatable at 1 life through this ability, which is
//     what "Pay 1 life" means and is why it is
//     ManaAbilityCost.Life rather than PainRider.
//   - The removal taps, costs {2}{B}{B}, and sacrifices A DESERT —
//     which the Deadlands itself is, so it can eat itself. That is
//     printed and it is usually what happens.
//
// "Sacrifice a Desert" is a fixed count of one over a land-subtype
// predicate, so the Deadlands is a legal choice and so is any other
// Desert on the battlefield under the same controller. Sacrificing a
// DIFFERENT Desert leaves the Deadlands in play — tapped, since the
// tap is part of the same cost.
//
// "Activate only as a sorcery" is SorcerySpeed: your main phase, your
// turn, empty stack.
//
// Two -1/-1 counters, not damage: they stay through the turn, they
// shrink toughness permanently, and they kill a 2/2 outright via the
// zero-toughness state-based action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af698bd5-5f56-4d2a-9f02-8c3e781210cd",
		Name:         "Ifnir Deadlands",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Life: 1},
				Produced: "{B}",
				Label:    "Pay 1 life: Add {B}",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}{B}, {T}, Sacrifice a Desert: Put two -1/-1 counters on target creature an opponent controls.",
			SorcerySpeed: true,
			Cost: effects.Plus(effects.ManaCost("{2}{B}{B}"), effects.TapCost(), game.AbilityCost{
				SacrificeOther: effects.SacrificeSpec("a Desert", effects.B02IsDesert),
			}),
			Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 2}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
