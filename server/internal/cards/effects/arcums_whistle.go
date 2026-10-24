package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcum's Whistle — Artifact {3} (#4062):
//
//	"{3}, {T}: Choose target non-Wall creature the active player has
//	 controlled continuously since the beginning of the turn. That player
//	 may pay {X}, where X is that creature's mana value. If they don't
//	 pay, the creature attacks this turn if able, and at the beginning of
//	 the next end step, destroy it if it didn't attack this turn.
//	 Activate only before attackers are declared."
//
// Nettling Imp's forced attack with a ransom: the ACTIVE player is asked
// whether to pay {X}, X the creature's mana value as the ability
// resolves (CR 118.12). Paid, nothing happens. Declined - or unable to
// pay - the creature attacks if able and is destroyed at the next end
// step unless it did. A mana value of zero costs {0}, which the player
// always pays. The window is BeforeAttackersAreDeclared alone: the
// printed clause names no turn, so it works on any player's turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de691eed-d90b-48f9-bae4-f982767f2bfd",
		Name:         "Arcum's Whistle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{3}, {T}: Choose target non-Wall creature the active player has controlled continuously since the beginning of the turn. That player may pay {X}, where X is that creature's mana value. If they don't pay, the creature attacks this turn if able, and at the beginning of the next end step, destroy it if it didn't attack this turn. Activate only before attackers are declared.",
			Cost:      Plus(ManaCost("{3}"), TapCost()),
			Targets:   forcedAttackTarget(),
			Condition: BeforeAttackersAreDeclared(),
			Effect:    arcumsWhistleAsks,
		}},
	})
}

// arcumsWhistleAsks asks the active player to pay the creature's mana
// value; their refusal is the forced attack.
func arcumsWhistleAsks(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return nil
	}
	payer := g.Seats[g.Turn.ActiveSeat].ID
	mv := c.ManaValue()
	if mv <= 0 {
		return nil // {0} is always paid
	}
	return PayUnless{
		Chooser:  payer,
		Cost:     fmt.Sprintf("{%d}", mv),
		Question: fmt.Sprintf("Arcum's Whistle - pay {%d} so %s doesn't have to attack?", mv, c.Name),
		Else:     arcumsWhistleDecline,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// arcumsWhistleDecline is the forced attack once the payer declines,
// keyed (#4256): a table waiting on the payer is a restore point. The
// creature rides on the prompt as its Carry.
var arcumsWhistleDecline = PayUnlessElse("pay-unless/arcums-whistle-must-attack", func(ctx *Context, r game.PayUnlessAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return mustAttackOrBeDestroyed(ctx, r.Carry[0], "Arcum's Whistle")
})
