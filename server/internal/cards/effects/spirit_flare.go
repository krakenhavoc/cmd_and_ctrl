package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spirit Flare — Instant {3}{W} (Torment, #5612):
//
//	"Tap target untapped creature you control. If you do, it deals damage
//	 equal to its power to target attacking or blocking creature an
//	 opponent controls.
//	 Flashback—{1}{W}, Pay 3 life. (You may cast this card from your
//	 graveyard for its flashback cost. Then exile it.)"
//
// Two clauses, as Bite Down's: the creature you control, untapped, then
// the attacker or blocker an opponent controls. Each is checked again as
// the spell resolves (CR 608.2b). The first is tapped only while it is
// still a legal target, which means still untapped, so the tap always
// happens when it is legal; "if you do" then fails only when it is not,
// and then nothing happens at all. Both targets are judged once, before
// the tap, since the tap itself ends "untapped". The damage is dealt by
// the tapped creature at its power as it now is, and only to a second
// target that is still legal: a creature removed from combat in response
// is spared, though yours is still tapped.
//
// The flashback price is mana and life, batch 1's
// torB1FlashbackPayingLife: a caster with less than 3 life can't claim it
// (CR 118.3), and the card is exiled as it leaves the stack (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d19a2c28-8d09-40bf-aeb1-e699672b009f",
		Name:         "Spirit Flare",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target untapped creature you control", Untapped(), YouControl()),
			TargetCreature("target attacking or blocking creature an opponent controls",
				AttackingOrBlocking(), OpponentControls()),
		),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{torB1FlashbackPayingLife("{1}{W}", 3)},
		OnResolve:        spiritFlareTapThenDamage,
	})
}

// spiritFlareTapThenDamage taps Spirit Flare's first target and, having
// tapped it, has it deal damage equal to its power to the second. Both
// targets are judged once, before the tap: the tap is what makes the
// first one stop being "untapped", and it must not undo its own "if you
// do" (oneSidedBite would read the first target again and find it
// illegal).
func spiritFlareTapThenDamage(_ *game.StackItem, ctx *Context) error {
	mine, ok := ctx.ClauseTarget(0)
	if !ok || mine.Kind != game.TargetCard {
		return nil
	}
	victim, victimLegal := ctx.ClauseTarget(1)
	if err := ctx.Game.TapTargetForEffectBy(mine.ID, ctx.Controller()); err != nil {
		return err
	}
	if !victimLegal || victim.Kind != game.TargetCard {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(mine.ID)
	if !ok {
		return nil
	}
	return DealDamage{Source: mine.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
}
