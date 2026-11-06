package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raph & Leo, Sibling Rivals — Legendary Creature — Mutant Ninja Turtle
// {1}{R/W}{R/W}, 2/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Whenever Raph & Leo attack, if it's the first combat phase of the
//	 turn, untap one or two target attacking creatures. After this
//	 phase, there is an additional combat phase."
//
// Karlach, Fury of Avernus's trigger on a single attacker with targets.
// "If it's the first combat phase of the turn" is an intervening if
// (CR 603.4), read as Raph & Leo are declared and again as the trigger
// resolves, so the added combat's own attack does not add a third. The
// targets are any one or two attacking creatures, whoever controls
// them; both chosen as the ability goes on the stack (CR 603.3d). When
// every target has left combat or the battlefield by resolution the
// ability does nothing at all (CR 608.2b), the additional combat
// included. No main phase comes with the combat (CR 500.8).
//
// No simplification.
func init() {
	const label = "Raph & Leo, Sibling Rivals — untap one or two target attacking creatures; an additional combat phase"
	Register(Spec{
		OracleID:     "57f39ce2-ecbb-4262-8140-03aba5c4303f",
		Name:         "Raph & Leo, Sibling Rivals",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return ThisAttacked(ev, source, lki, g) && IsFirstCombatPhase(g)
				}, label, raphAndLeoUntapAndCombat),
				TargetCreature("one or two target attacking creatures", AttackingCreature()).WithCount(1, 2),
			),
		},
	})
}

// raphAndLeoUntapAndCombat is the trigger at resolution: the
// intervening if again, then each still-legal target untaps and another
// combat follows this one.
func raphAndLeoUntapAndCombat(g *game.Game, item *game.StackItem) error {
	if !IsFirstCombatPhase(g) {
		return nil
	}
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	for _, t := range legal {
		if err := (UntapTarget{Target: t.ID}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return ExtraCombatAfterThisPhase().Apply(ctx)
}
