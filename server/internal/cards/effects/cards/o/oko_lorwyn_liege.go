package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oko, Lorwyn Liege // Oko, Shadowmoor Scion — {2}{U} Legendary
// Planeswalker — Oko, loyalty 3, both faces (Lorwyn Eclipsed, #4458):
//
//	Front: "At the beginning of your first main phase, you may pay {G}. If
//	        you do, transform Oko.
//	        +2: Up to one target creature gains all creature types. (This
//	            effect doesn't end.)
//	        +1: Target creature gets -2/-0 until your next turn."
//	Back:  "At the beginning of your first main phase, you may pay {U}. If
//	        you do, transform Oko.
//	        −1: Mill three cards. You may put a permanent card from among
//	            them into your hand.
//	        −3: Create two 3/3 green Elk creature tokens.
//	        −6: Choose a creature type. You get an emblem with "Creatures
//	            you control of the chosen type get +3/+3 and have vigilance
//	            and hexproof.""
//
// The transform is Ashling, Rekindled's MayPay whose paid branch is
// TransformThis (nothing happens if Oko has turned over since the trigger
// went on the stack, CR 701.27f); loyalty counters stay on a planeswalker
// that transforms. +2 is a continuous effect with no end (it does not end
// at the cleanup step, when Oko leaves or when Oko transforms): AllCreature-
// TypesMod on the target, which it keeps through other type changes only
// as the layers order them. +1 lasts until the controller's next turn
// (CR 611.2b), a layer 7c modify. −1 is Ainok Wayfarer's mill then take
// (only the cards that mill put into the graveyard, any permanent card).
// −3 makes two Elk (3/3 green). −6 asks for the type as the ability
// resolves and makes an emblem that remembers it (CreateEmblemWithTribe-
// ForEffect): the emblem's three statics are the tribal anthem and the two
// keyword grants over the emblem owner's creatures of that type, a
// changeling counting for every type.
//
// No simplification.
func init() {
	const oracle = "a0c956dc-d7de-4f11-b76f-7537fbee1731"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Oko, Lorwyn Liege",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Oko, Lorwyn Liege — you may pay {G} to transform Oko", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{G}",
					Question: "Oko, Lorwyn Liege — pay {G} to transform Oko?",
					Then:     effects.EirduTransformPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Up to one target creature gains all creature types. (This effect doesn't end.)",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AllCreatureTypesMod()},
						Duration: game.IndefiniteDuration(),
						Label:    "Oko, Lorwyn Liege — gains all creature types",
					}.Apply(ctx)
				})),
			},
			{
				Label:   "+1: Target creature gets -2/-0 until your next turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("target creature"),
				Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.ModifyPTMod(-2, 0)},
						Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
						Label:    "Oko, Lorwyn Liege — -2/-0 until your next turn",
					}.Apply(ctx)
				})),
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Oko, Shadowmoor Scion",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Oko, Shadowmoor Scion emblem",
			Text:  "Creatures you control of the chosen type get +3/+3 and have vigilance and hexproof.",
			Static: []game.StaticAbility{
				effects.TribalAnthem(okoEmblemTribe, 3, 3),
				effects.TribalKeywordGrant(okoEmblemTribe, "vigilance"),
				effects.TribalKeywordGrant(okoEmblemTribe, "hexproof"),
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Oko, Shadowmoor Scion — you may pay {U} to transform Oko", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{U}",
					Question: "Oko, Shadowmoor Scion — pay {U} to transform Oko?",
					Then:     effects.EirduTransformPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "−1: Mill three cards. You may put a permanent card from among them into your hand.",
				Cost:  effects.LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(okoMilled)}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Create two 3/3 green Elk creature tokens.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Tokens: 2},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Elk"), N: 2}),
			},
			{
				Label: "−6: Choose a creature type. You get an emblem with \"Creatures you control of the chosen type get +3/+3 and have vigilance and hexproof.\"",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					effects.ChooseCreatureTypeTail(g, item.Controller, item.SourceCardID,
						"Oko, Shadowmoor Scion — choose a creature type for the emblem", ctx.ItemTail(okoEmblemChosen))
					return nil
				},
			},
		},
	})
}

// okoEmblemTribe is the emblem's scope: creatures its owner controls of the
// type stored on the emblem.
var okoEmblemTribe = effects.TribeFilter{Chosen: true, YoursOnly: true}

// okoMilled offers a permanent card from among the cards the mill put into
// the graveyard. An on-disk key.
var okoMilled = effects.CardTail("oko-shadowmoor-scion/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.Permanent(),
		"Oko, Shadowmoor Scion — you may put a permanent card from among them into your hand")
})

// okoEmblemChosen creates the emblem once the type is chosen; the tail is not
// run for no answer. An on-disk key.
var okoEmblemChosen = effects.CreatureTypeChosen("oko-shadowmoor-scion/emblem-type",
	func(ctx *effects.Context, creatureType string, _ game.TailOutcome) error {
		return ctx.Game.CreateEmblemWithTribeForEffect(ctx.Controller(), ctx.Source(), creatureType)
	})
