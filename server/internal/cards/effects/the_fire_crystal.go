package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "57276db8-9d8d-4587-ae88-dfdc44343f17",
		Name:         "The Fire Crystal",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Red spells you cast cost {1} less to cast.",
				YourSpell(), ColoredSpell("R")),
		},
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "haste"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}{R}, {T}: Create a token that's a copy of target creature you control. Sacrifice it at the beginning of the next end step.",
			Cost:    Plus(ManaCost("{4}{R}{R}"), TapCost()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  theFireCrystalCopyAndSacrifice,
		}},
	})
}

// theFireCrystalCopyAndSacrifice mints the token copy and schedules
// its sacrifice at the next end step. The target is re-read out of
// ctx.LegalTargets() so a creature that left in response is skipped
// (CR 608.2b).
func theFireCrystalCopyAndSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
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
	cursor := b25LastEventSeq(g)
	if err := (CreateTokenCopy{
		Controller: item.Controller,
		Copy:       copyOf,
		N:          1,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "The Fire Crystal — sacrifice the token",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
