package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b823_helpers.go — Homelands, "Pump and combat tricks" (S82 OLD
// batch 823, #5635). A file of its own so concurrent batches never collide.

// OldB823BlocksOrIsBlockedByBlack is the trigger condition of "Whenever this
// creature blocks or becomes blocked by one or more black creatures" (Serra
// Inquisitors), and with blocked=false of "Whenever this creature blocks one or
// more black creatures" (Rashka the Slayer). It is one trigger per block (CR
// 509.3a, 509.3c): the creature's first EventBlock when at least one attacker it
// blocks is black, or, when blocked is set, its one EventBecomesBlocked when at
// least one of the creatures blocking it is black.
func OldB823BlocksOrIsBlockedByBlack(blocked bool) When {
	black := func(c *game.Card) bool { return c.IsCreature() && c.HasColorWord("B") }
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		switch ev.Kind {
		case game.EventBlock:
			if ev.CardID != source.InstanceID || ev.Amount > 1 {
				return false
			}
			for _, a := range source.BlockedAttackers() {
				if c, ok := g.LookupCardForEffect(a); ok && black(&c) {
					return true
				}
			}
		case game.EventBecomesBlocked:
			if !blocked || ev.CardID != source.InstanceID {
				return false
			}
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.IsBlockingAttacker(source.InstanceID) && black(&c) {
					return true
				}
			}
		}
		return false
	}
}

// OldB823DuringCombat is "Activate only during combat": any step of the combat
// phase of any turn (CR 506.1).
func OldB823DuringCombat() ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool {
		return game.PhaseOf(g.Turn.Step) == game.PhaseCombat
	}
}

// OldB823TargetGetsThenSacrificeThisWhenItLeaves is "Target <creature> gets
// +P/+0 and gains <keyword> until end of turn. When that creature leaves the
// battlefield this turn, sacrifice this creature" (Heart Wolf): Runesword's
// delayed trigger, pinned to the target for the turn, whose body sacrifices the
// ability's source if it is still the same object (CR 603.7c).
func OldB823TargetGetsThenSacrificeThisWhenItLeaves(label string, power int, keyword string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: power, Label: label + " — +" + string(rune('0'+power)) + "/+0"}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword}, Label: label + " — " + keyword}).Apply(ctx); err != nil {
			return err
		}
		d := g.PinnedTo(g.UntilEndOfTurnDuration(), id)
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:   ctx.Controller(),
			SourceCardID: ctx.Source(),
			Label:        label + " — sacrifice it",
			On:           []game.EventKind{game.EventLTB},
			Condition:    theListedObjectLeftCondition,
			Cards:        []uuid.UUID{id},
			Duration:     &d,
			Body:         runeswordSacrificeBody,
		})
		return nil
	}
}
