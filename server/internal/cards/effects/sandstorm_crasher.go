package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandstorm Crasher — Creature — Minotaur Berserker Wizard {3}{R}, 3/4:
//
//	"Trample
//	 You may exert this creature as it attacks. When you do, create a
//	 tapped and attacking token that's a copy of target creature you
//	 control. Sacrifice the token at the beginning of the next end
//	 step. (An exerted creature won't untap during your next untap
//	 step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a targeted linked
// trigger (CR 607.2h). The token is a copy of the target as it exists
// as the trigger resolves, and it was never declared as an attacker, so
// it triggers nothing that watches attacks and can't be exerted (CR
// 508.4). It is sacrificed by a delayed trigger (CR 603.7) at the
// beginning of the next end step. If the target is gone as the trigger
// resolves it is removed from the stack and the Crasher stays exerted
// (CR 603.3d, 608.2b).
//
// Caveat: CR 508.4 lets the token's controller choose which defending
// player, planeswalker or battle it attacks. The engine has it join
// whatever the Crasher is attacking (the same reading Sauron, the
// Necromancer and Dalkovan Encampment use), so at a table of several
// opponents the choice is narrower than printed.
func init() {
	const label = "Sandstorm Crasher — a tapped and attacking token copy of target creature you control; sacrifice it at the next end step"
	Register(Spec{
		OracleID:        "d1597137-5609-4012-93bb-19c49c9cf2c3",
		Name:            "Sandstorm Crasher",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The token copy always attacks whatever Sandstorm Crasher is attacking; you can't send it at another opponent."},
		PrintedKeywords: []string{"trample"},
		ExertOnAttack:   ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			// No Purpose: the token is gone at the end step, so no
			// printed amount of Tokens would be true.
			Targeting(
				WhenExerted(label, sandstormCrasherCopy),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// sandstormCrasherCopy creates the tapped, attacking token copy of the
// trigger's target and schedules its sacrifice.
func sandstormCrasherCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(g, ids[0])
	if !ok {
		return nil
	}
	tmpl.Tapped = true
	cursor := b25LastEventSeq(g)
	if err := g.CreateTokensAttackingForEffect(item.Controller, tmpl, 1, item.Trigger.Event.Target); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Sandstorm Crasher — sacrifice the token",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
