package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "ea96fe35-88ed-432b-a495-08e787d23d0b",
		Name:          "Piercing Exhale",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{OptionalBehold("a", "Dragon")},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := oneSidedBite(item, ctx); err != nil {
				return err
			}
			if !ctx.WasBeheld() {
				return nil
			}
			return Surveil{N: 2}.Apply(ctx)
		},
	})
}
