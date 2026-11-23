package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venat, Heart of Hydaelyn // Hydaelyn, the Mothercrystal (Final
// Fantasy, #4462). Front, Legendary Creature — Elder Wizard {1}{W}{W},
// 3/3:
//
//	"Whenever you cast a legendary spell, draw a card. This ability
//	 triggers only once each turn.
//	 Hero's Sundering — {7}, {T}: Exile target nonland permanent.
//	 Transform Venat. Activate only as a sorcery."
//
// Back, Legendary Creature — God, 4/4:
//
//	"Indestructible
//	 Blessing of Light — At the beginning of combat on your turn, put a
//	 +1/+1 counter on another target creature you control. Until your
//	 next turn, it gains indestructible. If that creature is legendary,
//	 draw a card."
//
// Front: Reki's legendary-spell trigger under TriggersOnlyOnceEachTurn;
// Hero's Sundering exiles the target and then turns Venat over in place
// (TransformThis, a real transform: the same object, CR 701.28). A
// target that is gone by resolution fizzles the whole ability, so Venat
// does not transform, as the ability's text would have it.
//
// Back: the counter, the "until your next turn" grant
// (DurationUntilYourNextTurn) and the draw all depend on the one target
// (ruling): illegal at resolution and nothing happens, no draw. The
// legendary check reads the creature as it is after the counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     VenatOracleID,
		Name:         "Venat, Heart of Hydaelyn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.WheneverYouCast(effects.Legendary(), "Venat, Heart of Hydaelyn — draw a card",
				effects.Do(effects.DrawCards{N: 1}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "Hero's Sundering — {7}, {T}: Exile target nonland permanent. Transform Venat. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{7}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				return effects.TransformThis{}.Apply(ctx)
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        VenatOracleID + "#1",
		Name:            "Hydaelyn, the Mothercrystal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Hydaelyn, the Mothercrystal — Blessing of Light: +1/+1 counter and indestructible until your next turn on another target creature you control; draw if legendary",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, found := effects.B16FirstLegalTargetCard(ctx)
					if !found {
						return nil
					}
					if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
						Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
						Label:    "Hydaelyn, the Mothercrystal — gains indestructible until your next turn",
					}).Apply(ctx); err != nil {
						return err
					}
					if c, seen := g.LookupCardForEffect(id); seen && c.IsLegendary() {
						return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
					}
					return nil
				}), venatAnotherCreatureSpec()),
		},
	})
}

// venatOracleID is shared by both faces.
const VenatOracleID = "2e5adbcf-6b5c-4ded-a339-8757040ef878"

// venatAnotherCreatureSpec is "another target creature you control"
// for an ability: never the source itself.
func venatAnotherCreatureSpec() *game.TargetSpec {
	spec := effects.PermanentYouControl("another target creature you control", effects.Creature())
	spec.ExcludeSource = true
	return spec
}
