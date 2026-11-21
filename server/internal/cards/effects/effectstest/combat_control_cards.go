package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_control_cards_test.go — the proof cards for #1317 (CR 724.2,
// end the combat phase) and #1329 (CR 508.7, reselect an attack).

const (
	MandateOfPeaceOracle     = "c309ec42-34a0-4083-a36c-7814643d7960"
	MisleadingSignpostOracle = "eaffdf95-8e20-408d-99fa-3adc9e19523d"
	PortalMageOracle         = "4d19ea6d-cabe-4a25-a911-dd744517dd73"
	WindshaperOracle         = "c47050c2-9b46-427e-85d6-058fd6a61e67"
)

// --- Reselecting an attack (#1329) ------------------------------------

// castFlashPermanentInPlace casts a flash permanent without moving the
// turn on, so the declare attackers step is still where it resolves.
func CastFlashPermanentInPlace(t *testing.T, g *game.Game, caster uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	var p *game.Player
	for _, s := range g.Seats {
		if s != nil && s.ID == caster {
			p = s
		}
	}
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		OracleID: oracle, Owner: caster, Controller: caster,
	})
	if err := g.CastSpell(caster, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// attackInDeclareAttackers declares the attacks and leaves the cursor
// in the declare attackers step.
func AttackInDeclareAttackers(t *testing.T, g *game.Game, attacks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for attacker, defender := range attacks {
		if err := g.DeclareAttacker(attacker, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
}

// settleUntilOptionPick passes priority, answering a pick_target with
// `target`, until an option pick is owed by `chooser` (returned) or
// the stack is empty (nil).
func SettleUntilOptionPick(t *testing.T, g *game.Game, chooser, target uuid.UUID) *game.PendingChoice {
	t.Helper()
	for range 8 {
		if c := LatestOptionPickFor(g, chooser); c != nil {
			return c
		}
		if pick := LatestPickTarget(g, chooser); pick != nil {
			if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if StackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			return nil
		}
		PassPriorityAroundTable(t, g)
	}
	return LatestOptionPickFor(g, chooser)
}

// optionFor is the index of the option naming `subject` as a seat.
func OptionFor(t *testing.T, c *game.PendingChoice, seat uuid.UUID) int {
	t.Helper()
	for i, opt := range c.PickOptions {
		if opt.Player == seat {
			return i
		}
	}
	t.Fatalf("no option for seat %s in %+v", seat, c.PickOptions)
	return -1
}
