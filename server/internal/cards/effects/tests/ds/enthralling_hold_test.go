package testsds

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enthralling_hold_test.go — Enthralling Hold (#6168): "You can't choose
// an untapped creature as this spell's target as you cast it. You control
// enchanted creature."

const enthrallingHoldOracle = "2f94cb83-1f48-4293-b219-67830c074aeb"

func TestEnthrallingHoldIsFullWithNoCaveats(t *testing.T) {
	spec, ok := effects.Lookup(enthrallingHoldOracle)
	if !ok || spec.Name != "Enthralling Hold" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec %+v (registered %v)", spec, ok)
	}
}

func enthrallingHoldBear(g *game.Game, owner uuid.UUID, name string, tapped bool) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, Tapped: tapped,
	})
}

func enthrallingHoldInHand(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Enthralling Hold", TypeLine: effectstest.AuraTypeLine,
		OracleID: enthrallingHoldOracle, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func setTapped(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
			}
		}
	})
}

// As it is cast, only a tapped creature is a legal target: the legal set
// the picker reads lists the tapped one and not the untapped one, and an
// announce naming the untapped one is refused.
func TestEnthrallingHoldCanTargetOnlyATappedCreatureAsCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := enthrallingHoldBear(g, opp.ID, "Tapped Bear", true)
	untapped := enthrallingHoldBear(g, opp.ID, "Untapped Bear", false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hold := enthrallingHoldInHand(g, me)
	card := me.Hand.Cards[len(me.Hand.Cards)-1]
	if card.InstanceID != hold {
		t.Fatal("setup: Enthralling Hold is not on top of the hand")
	}
	legal := g.LegalTargetsFor(game.SourceCast(me.ID, &card), enthrallingHoldOracle)
	if legal == nil {
		t.Fatal("no legal-target set")
	}
	has := func(id uuid.UUID) bool {
		for _, c := range legal.Cards {
			if c == id {
				return true
			}
		}
		return false
	}
	if !has(tapped) || has(untapped) {
		t.Fatalf("legal set %v: want the tapped bear only", legal.Cards)
	}
	err := g.CastSpell(me.ID, hold, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(untapped)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("cast at an untapped creature = %v, want ErrIllegalTarget", err)
	}
}

// Once cast, tapped-ness no longer matters: the creature untapping in
// response does not counter the spell (CR 608.2b reads only the enchant
// clause), and the Aura stays attached, with control, after it untaps
// again (CR 704.5m).
func TestEnthrallingHoldStaysAttachedAfterTheCreatureUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := enthrallingHoldBear(g, opp.ID, "Bear", true)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hold := enthrallingHoldInHand(g, me)
	if err := g.CastSpell(me.ID, hold, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("cast at a tapped creature: %v", err)
	}
	setTapped(g, theirs, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hold) {
		t.Fatal("Enthralling Hold did not resolve onto the battlefield after its target untapped")
	}
	if host := effectstest.AttachmentHostOf(t, g, hold); host.Kind != game.TargetCard || host.ID != theirs {
		t.Fatalf("AttachedTo = %+v, want the bear", host)
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("controller %s, want the Aura's controller", got)
	}
	setTapped(g, theirs, true)
	setTapped(g, theirs, false)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(hold) {
		t.Fatal("the Aura fell off an untapped creature")
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Errorf("controller %s after the creature untapped, want the Aura's controller", got)
	}
}
