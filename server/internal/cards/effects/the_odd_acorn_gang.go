package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "1d9d8028-d946-4245-ae0c-7508023f619a",
		Name:            "The Odd Acorn Gang",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "menace", "trample"},
		Grants: []AbilityGrant{{
			Key: cmdr2B17OddAcornGrant,
			Activated: []ActivatedAbility{{
				Label:        "{T}: Target Squirrel gets +2/+2 and gains trample until end of turn. Activate only as a sorcery.",
				Cost:         TapCost(),
				Targets:      TargetCreature("target Squirrel", HasSubtype("Squirrel")),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "The Odd Acorn Gang — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "The Odd Acorn Gang — trample"}.Apply(ctx)
					}
					return nil
				},
			}},
			Text: "{T}: Target Squirrel gets +2/+2 and gains trample until end of turn. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Squirrel"}, YoursOnly: true}, cmdr2B17OddAcornGrant),
		},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(HasSubtype("Squirrel"),
				"The Odd Acorn Gang — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
