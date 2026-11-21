package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const tenthDistrictHeroMilevaGrant = "tenth-district-hero/mileva-indestructible"

// Tenth District Hero — Creature — Human {1}{W}, 2/3 (Murders at Karlov
// Manor, #4469):
//
//	"{1}{W}, Collect evidence 2: This creature becomes a Human Detective
//	 with base power and toughness 4/4 and gains vigilance.
//	 {2}{W}, Collect evidence 4: If this creature is a Detective, it
//	 becomes a legendary creature named Mileva, the Stalwart, it has base
//	 power and toughness 5/5, and it gains "Other creatures you control
//	 have indestructible.""
//
// Both are ScopedEffects with no duration (the rulings: they last until
// the Hero leaves or another effect overrides them): the creature types
// in layer 4, the name in layer 3, base P/T in layer 7b, and the gained
// abilities in layer 6, so later effects that set P/T win. The second
// checks "is a Detective" as it resolves, through the layered view (it
// does nothing otherwise). "Becomes a legendary creature" adds no creature
// subtype and takes none away, so the Hero stays a Human Detective
// (ruling). The quoted ability is an ADR 0093 grant of a layer-6 static
// that gives other creatures its controller has indestructible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0953772c-9374-4616-9c04-5838aaf626c9",
		Name:         "Tenth District Hero",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: tenthDistrictHeroMilevaGrant,
			Static: []game.StaticAbility{{
				Label:     "Other creatures you control have indestructible.",
				Layer:     game.Layer6Ability,
				AppliesTo: effects.OtherCreaturesYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, "indestructible")
				},
			}},
			Text: "Other creatures you control have indestructible.",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W}, Collect evidence 2: This creature becomes a Human Detective with base power and toughness 4/4 and gains vigilance.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.CollectEvidenceN(2)),
				Effect:  tenthDistrictHeroDetective,
			},
			{
				Label:   "{2}{W}, Collect evidence 4: If this creature is a Detective, it becomes a legendary creature named Mileva, the Stalwart, it has base power and toughness 5/5, and it gains \"Other creatures you control have indestructible.\"",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSurvive},
				Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.CollectEvidenceN(4)),
				Effect:  tenthDistrictHeroMileva,
			},
		},
	})
}

func tenthDistrictHeroDetective(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := item.SourceCardID
	if !effects.OnBattlefield(g, src) || ctx.IsNewSourceObject(src) {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: src,
		Mods: []game.Mod{
			game.SetTypesMod([]string{"Creature"}, []string{"Human", "Detective"}),
			game.SetBasePowerMod(4),
			game.SetBaseToughnessMod(4),
			game.AddKeywordsMod("vigilance"),
		},
		Duration: game.IndefiniteDuration(),
		Label:    "Tenth District Hero — becomes a Human Detective 4/4 with vigilance",
	}.Apply(ctx)
}

func tenthDistrictHeroMileva(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := item.SourceCardID
	if !effects.OnBattlefield(g, src) || ctx.IsNewSourceObject(src) {
		return nil
	}
	if c, ok := g.LookupCardForEffect(src); !ok || !c.HasSubtypeWord("Detective") {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: src,
		Mods: []game.Mod{
			game.SetNameMod("Mileva, the Stalwart"),
			game.SetTypesMod([]string{"Creature"}, nil, "Legendary"),
			game.SetBasePowerMod(5),
			game.SetBaseToughnessMod(5),
			game.GrantAbilitiesMod(tenthDistrictHeroMilevaGrant),
		},
		Duration: game.IndefiniteDuration(),
		Label:    "Tenth District Hero — becomes Mileva, the Stalwart",
	}.Apply(ctx)
}
