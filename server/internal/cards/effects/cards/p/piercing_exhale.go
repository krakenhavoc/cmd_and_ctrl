package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Piercing Exhale — Instant {1}{G} (Tarkir: Dragonstorm, #4463):
//
//	"As an additional cost to cast this spell, you may behold a Dragon.
//	 (You may choose a Dragon you control or reveal a Dragon card from
//	 your hand.)
//	 Target creature you control deals damage equal to its power to
//	 target creature or planeswalker. If a Dragon was beheld, surveil 2."
//
// Bite Down's one-sided bite (oneSidedBite) with Hulk's Thunderclap's
// optional behold (OptionalBehold, CR 701.4a). The second clause is any
// creature or planeswalker, yours included, and it is a separate word
// "target", so the same creature may be chosen for both (CR 115.3). The
// power is read as the spell resolves; if the biter has become an illegal
// target it deals no damage, and the surveil still happens when the spell
// resolves with the other target legal (CR 608.2b). "If a Dragon was
// beheld" is a fact about the announcement (CR 701.4b), read with
// ctx.WasBeheld(). The surveil (CR 701.25a) comes after the damage, in
// printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "ea96fe35-88ed-432b-a495-08e787d23d0b",
		Name:          "Piercing Exhale",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{effects.OptionalBehold("a", "Dragon")},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.OneSidedBite(item, ctx); err != nil {
				return err
			}
			if !ctx.WasBeheld() {
				return nil
			}
			return effects.Surveil{N: 2}.Apply(ctx)
		},
	})
}
