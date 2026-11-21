package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// T'Pring, Never Parted — Legendary Creature — Vulcan Citizen {3}{G}{U},
// 3/3 (Star Trek Commander, #4974):
//
//	"Kal-if-fee — When T'Pring enters, choose another target creature you
//	 control and target creature an opponent controls. You may have the
//	 chosen creatures fight each other. Whenever a creature that fought
//	 this way dies this turn, you draw a card."
//
// Kal-if-fee is an ability word (no rules meaning). Both targets are
// chosen as the trigger goes on the stack; the "you may" is asked as it
// resolves (CR 603.5). The fight happens only while both creatures are
// still legal targets (CR 701.14b); when it does, each of the two gets
// its own "when that creature dies this turn, you draw a card" (CR
// 603.7a), about the creature as the object it is now, so a creature
// that leaves and comes back is not watched (CR 603.7c). Two creatures
// dying at once draw two cards, as the one "whenever" would. The
// watches end at cleanup (CR 514.2).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("T'Pring, Never Parted — the chosen creatures fight; draw when one that fought dies this turn",
		tPringFight)
	enters.Targets = effects.Clauses(
		effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
		effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
	)
	effects.Register(effects.Spec{
		OracleID:     "29775eb6-f658-4839-a434-2fa063069f1c",
		Name:         "T'Pring, Never Parted",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(enters, "T'Pring, Never Parted — have the chosen creatures fight each other?"),
		},
	})
}

// tPringFight is the resolution: the two chosen creatures fight, and
// each one that fought is watched for dying this turn. The watches are
// created before the damage is dealt, so a creature the fight kills is
// still the object they are about.
func tPringFight(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	mine, ok := ctx.ClauseTarget(0)
	if !ok || mine.Kind != game.TargetCard {
		return nil
	}
	theirs, ok := ctx.ClauseTarget(1)
	if !ok || theirs.Kind != game.TargetCard {
		return nil
	}
	if ctx.IsNewSourceObjectAsThis(mine.ID) || ctx.IsNewSourceObjectAsThis(theirs.ID) {
		return nil
	}
	a, okA := effects.BattlefieldCardForEffect(g, mine.ID)
	b, okB := effects.BattlefieldCardForEffect(g, theirs.ID)
	if !okA || !okB || !a.IsCreature() || !b.IsCreature() {
		return nil
	}
	for _, c := range []game.Card{a, b} {
		if err := effects.WhenThatCreatureDiesThisTurn(ctx, c.InstanceID,
			"T'Pring, Never Parted — a creature that fought died: draw a card", effects.DrawOneBody); err != nil {
			return err
		}
	}
	return effects.B10Fight(ctx, mine.ID, theirs.ID)
}
