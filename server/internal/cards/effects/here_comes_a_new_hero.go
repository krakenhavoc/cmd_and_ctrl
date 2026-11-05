package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Here Comes a New Hero! — Sorcery {X}{2}{U} (TMNT Commander):
//
//	"Target player draws X cards. Create a token that's a copy of up to
//	 one target creature with mana value X or less."
//
// Two target clauses (CR 601.2c), Homesickness's shape: slot 0 is the
// player, slot 1 the creature. The creature clause is bound to the
// announced X with WithManaValueAtMostX, Kozilek's Command's binding,
// so a creature of a greater mana value is refused at announce and
// re-checked on resolution (CR 608.2b). A creature with {X} in its mana
// cost counts that X as 0 (CR 202.3e), as the rulings say.
//
// The instructions run in printed order (CR 608.2c): the player draws,
// then the caster creates the copy. CreateTokenCopy takes the creature's
// copiable values (CR 707.2): what is printed, or what it is copying, or
// a token's original characteristics, and not its counters, tapped
// state, Auras or non-copy effects. The token runs the ordinary entry,
// so the copied "when this enters" and "enters with" clauses work. With
// no creature chosen, or the creature gone, the draw still happens; with
// the player gone, the copy is still made.
//
// XMatters stays unset (#810): X=0 is a real move, a token copy of a
// creature with mana value 0: most token creatures, or one whose cost
// is only {X} (CR 202.3e).
//
// Purpose: the draw is X, and ADR 0126 declares no X amount (Stroke of
// Genius, Blue Sun's Zenith), so no Purpose is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4fb12cb-05e1-4df2-b62e-3391621a601b",
		Name:         "Here Comes a New Hero!",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target player"),
			TargetCreature("up to one target creature with mana value X or less").WithCount(0, 1).WithManaValueAtMostX(),
		),
		OnResolve: hereComesANewHeroResolve,
	})
}

// hereComesANewHeroResolve draws X for the target player, then copies
// the creature target for the caster.
func hereComesANewHeroResolve(item *game.StackItem, ctx *Context) error {
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
		if err := (DrawCards{Player: t.ID, N: ctx.X()}).Apply(ctx); err != nil {
			return err
		}
	}
	t, ok := ctx.ClauseTarget(1)
	if !ok || t.Kind != game.TargetCard || !onBattlefield(ctx.Game, t.ID) {
		return nil
	}
	return CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1}.Apply(ctx)
}
