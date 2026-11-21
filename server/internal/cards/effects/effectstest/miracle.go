package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// miracle_test.go — miracle (CR 702.94, #1665), end to end on real
// catalog cards. The keyword's engine half is game/miracle.go; what is
// pinned here is the whole road a player walks: draw it first, answer
// "Reveal Terminus for its miracle cost {W}?", let the trigger resolve,
// cast it for {W} — and every way that road is closed.

const (
	TerminusOracle         = "3dd196b6-a85a-4e3e-bb57-ec34241f8117"
	ThunderousWrathOracle  = "78260893-c443-44c8-ab45-ce86ef347d98"
	DevastationTideOracle  = "4245ee98-2d4c-49d1-8d07-80760cae2bf9"
	ReforgeTheSoulOracle   = "ece854f8-8c60-4f30-894f-2286d3dd61b9"
	EntreatTheAngelsOracle = "b349f018-c20b-48b0-9e65-d5fd56b24b88"
	BanishingStrokeOracle  = "a6898364-c29e-4b97-a500-344efa3ec24a"
)

// pushMiracleOnTop puts a catalog card on top of p's library, so the
// next draw is it.
func PushMiracleOnTop(g *game.Game, p *game.Player, name, typeLine, manaCost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{
			InstanceID: id,
			Name:       name,
			TypeLine:   typeLine,
			ManaCost:   manaCost,
			OracleID:   oracle,
			Owner:      p.ID,
			Controller: p.ID,
		})
	})
	return id
}

// miraclePrompt is the pending "Reveal X for its miracle cost?" prompt
// addressed to chooser, or nil.
func MiraclePrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// hasMiracleGrant reports whether p holds a miracle permission naming
// cardID AS IT SITS IN p's HAND NOW — the instance and its CR 400.7
// epoch. A permission over an object that has since left is not one.
func HasMiracleGrant(p *game.Player, cardID uuid.UUID) bool {
	epoch := -1
	for _, c := range p.Hand.Cards {
		if c.InstanceID == cardID {
			epoch = c.ObjectEpoch
		}
	}
	if epoch < 0 {
		return false
	}
	for _, perm := range p.CastPermissions {
		if perm.Zone != game.ZoneHand || perm.AltCostKey != game.AltCostKeyMiracle {
			continue
		}
		for _, ref := range perm.Cards {
			if ref.ID == cardID && ref.Epoch == epoch {
				return true
			}
		}
	}
	return false
}

// drawMiracleInOwnDrawStep walks to seat's upkeep, stacks the card on
// top, and enters the draw step so the turn-based draw is the first
// card seat draws this turn.
func DrawMiracleInOwnDrawStep(t *testing.T, g *game.Game, seat int, name, typeLine, manaCost, oracle string) uuid.UUID {
	t.Helper()
	AdvanceToUpkeepOf(t, g, seat)
	PassPriorityAroundTable(t, g)
	p := g.Seats[seat]
	id := PushMiracleOnTop(g, p, name, typeLine, manaCost, oracle)
	for g.Turn.Step != game.StepDraw {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to the draw step: %v", err)
		}
	}
	if !p.Hand.Contains(id) {
		t.Fatalf("%s was not drawn in the draw step", name)
	}
	return id
}

// revealAndResolve answers the reveal prompt "yes" and resolves the
// miracle trigger.
func RevealAndResolve(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	if MiraclePrompt(g, p.ID) == nil {
		t.Fatal("no \"Reveal … for its miracle cost?\" prompt for the first card drawn this turn")
	}
	AnswerLatestTriggerPrompt(t, g, p.ID, true)
	ResolveTopWithoutPassingHolder(t, g)
}

// resolveTopWithoutPassingHolder resolves the miracle trigger. Passing
// priority is the decline once the grant exists (closeMiracleWindow),
// so the passes are made while the trigger is still on the stack and
// the resolving pass is the last one — which is how the table plays it.
func ResolveTopWithoutPassingHolder(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8 && !StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !StackFullyEmpty(g) {
		t.Fatal("the miracle trigger never resolved")
	}
}
