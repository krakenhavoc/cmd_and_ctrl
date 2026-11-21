package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Fire Crystal — Legendary Artifact {2}{R}{R}:
//
//	"Red spells you cast cost {1} less to cast.
//	 Creatures you control have haste.
//	 {4}{R}{R}, {T}: Create a token that's a copy of target creature
//	 you control. Sacrifice it at the beginning of the next end step."
//
// The red Crystal — see the_water_crystal.go / the_wind_crystal.go /
// the_earth_crystal.go for the cycle's other members. The static
// haste grant reuses Hammer of Purphoros's Layer 6 shape
// (b16GrantKeywords over b16CreaturesYouControl). The activated
// ability is Reflection of Kiki-Jiki's shape (CreateTokenCopy + a
// delayed sacrifice trigger), without the "another"/"nonlegendary"
// restriction and without the haste exception, since this Crystal's
// copy is plain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57276db8-9d8d-4587-ae88-dfdc44343f17",
		Name:         "The Fire Crystal",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Red spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.ColoredSpell("R")),
		},
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B16CreaturesYouControl, "haste"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}{R}, {T}: Create a token that's a copy of target creature you control. Sacrifice it at the beginning of the next end step.",
			Cost:    effects.Plus(effects.ManaCost("{4}{R}{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  theFireCrystalCopyAndSacrifice,
		}},
	})
}

// theFireCrystalCopyAndSacrifice mints the token copy and schedules
// its sacrifice at the next end step. The target is re-read out of
// ctx.LegalTargets() so a creature that left in response is skipped
// (CR 608.2b).
func theFireCrystalCopyAndSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var copyOf uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			copyOf = t.ID
			break
		}
	}
	if copyOf == uuid.Nil {
		return nil
	}
	cursor := effects.B25LastEventSeq(g)
	if err := (effects.CreateTokenCopy{
		Controller: item.Controller,
		Copy:       copyOf,
		N:          1,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := effects.B27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label: "The Fire Crystal — sacrifice the token",
		Cards: tokens,
		Body:  effects.SacrificeListedCardsBody,
	}.Apply(ctx)
}
