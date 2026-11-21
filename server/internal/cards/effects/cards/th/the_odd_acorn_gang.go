package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Odd Acorn Gang — Legendary Creature — Squirrel Warrior {3}{B}{G},
// 5/5:
//
//	"Reach, menace, trample
//	 Squirrels you control have "{T}: Target Squirrel gets +2/+2 and
//	 gains trample until end of turn. Activate only as a sorcery."
//	 Whenever one or more Squirrels you control deal combat damage to a
//	 player, draw a card."
//
// The grant is a layer 6 activated ability on every Squirrel you control
// (ADR 0093), the Gang itself included; the target is any Squirrel, an
// opponent's too. The draw is the batch trigger (CR 603.2c): one draw
// per player the Squirrels connect with per damage step.
//
// No simplification.
const cmdr2B17OddAcornGrant = "the-odd-acorn-gang/squirrel-pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d9d8028-d946-4245-ae0c-7508023f619a",
		Name:            "The Odd Acorn Gang",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "menace", "trample"},
		Grants: []effects.AbilityGrant{{
			Key: cmdr2B17OddAcornGrant,
			Activated: []effects.ActivatedAbility{{
				Label:        "{T}: Target Squirrel gets +2/+2 and gains trample until end of turn. Activate only as a sorcery.",
				Cost:         effects.TapCost(),
				Targets:      effects.TargetCreature("target Squirrel", effects.HasSubtype("Squirrel")),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "The Odd Acorn Gang — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "The Odd Acorn Gang — trample"}.Apply(ctx)
					}
					return nil
				},
			}},
			Text: "{T}: Target Squirrel gets +2/+2 and gains trample until end of turn. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Squirrel"}, YoursOnly: true}, cmdr2B17OddAcornGrant),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(effects.HasSubtype("Squirrel"),
				"The Odd Acorn Gang — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
