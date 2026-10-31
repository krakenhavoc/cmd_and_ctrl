package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// molecule_man_test.go — Molecule Man's "Nonland cards in your hand have
// miracle {0}" (#4455), end to end on the real catalog card. The engine
// half is game/miracle.go (GrantedMiracleTrigger, AltCostKeyGrantedMiracle);
// what is pinned here is the whole road a player walks, and every way it is
// closed.

const moleculeManOracle = "ba944437-0b55-47cb-92dc-2477ce6726c3"

// moleculeManOnTable puts the real Molecule Man on `owner`'s battlefield.
func moleculeManOnTable(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Molecule Man", "Legendary Creature — Human Villain", moleculeManOracle, false)
}

// hasGrantedMiracleGrant reports whether p holds a live permission under
// `key` naming cardID as it sits in p's hand now.
func hasMiracleGrantUnder(p *game.Player, cardID uuid.UUID, key string) bool {
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
		if perm.Zone != game.ZoneHand || perm.AltCostKey != key {
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

const (
	mmSorceryType = "Sorcery"
	mmSorceryCost = "{4}{R}{R}"
)

func TestMoleculeManIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(moleculeManOracle)
	if !ok {
		t.Fatal("Molecule Man is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Molecule Man: completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
}

// The whole road: the first nonland card drawn this turn is offered a
// reveal, naming the card; the trigger resolves, and the card is cast with
// an EMPTY pool and nothing else paid, at sorcery speed in the draw step
// (CR 608.2g).
func TestMoleculeManFirstDrawIsCastForNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	moleculeManOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, "Pyre Sweep", mmSorceryType, mmSorceryCost, "")

	prompt := miraclePrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no reveal prompt for the first nonland card drawn this turn")
	}
	if !strings.Contains(prompt.Reason, "miracle cost {0}") {
		t.Errorf("prompt %q does not name the miracle cost {0}", prompt.Reason)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	resolveTopWithoutPassingHolder(t, g)

	if !hasMiracleGrantUnder(me, spell, game.AltCostKeyGrantedMiracle) {
		t.Fatal("the resolved trigger granted no miracle cast")
	}
	// Printed cost still closed in the draw step (a sorcery), and the
	// miracle claim is not.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the printed {4}{R}{R} sorcery was castable in the draw step")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: game.AltCostKeyGrantedMiracle,
	}); err != nil {
		t.Fatalf("the miracle {0} cast with an empty pool: %v", err)
	}
	if me.Hand.Contains(spell) {
		t.Error("the cast card is still in hand")
	}
	if hasMiracleGrantUnder(me, spell, game.AltCostKeyGrantedMiracle) {
		t.Error("the grant outlived the cast (CR 400.7)")
	}
}

// "You may reveal": a "no" leaves an ordinary card, and the {0} claim is
// refused.
func TestMoleculeManDeclinedRevealIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	moleculeManOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, "Pyre Sweep", mmSorceryType, mmSorceryCost, "")
	answerLatestTriggerPrompt(t, g, me.ID, false)
	if !stackFullyEmpty(g) {
		t.Error("a declined reveal put a trigger on the stack")
	}
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle})
	if err == nil {
		t.Fatal("the {0} miracle cast was accepted without a reveal")
	}
	if !errors.Is(err, game.ErrAltCostNotGranted) {
		t.Errorf("miracle cast without a reveal: %v, want ErrAltCostNotGranted", err)
	}
	if !me.Hand.Contains(spell) {
		t.Error("a refused cast took the card out of hand")
	}
}

// Without Molecule Man there is no miracle at all: no prompt, and the claim
// names an offer the card does not make.
func TestMoleculeManNothingWithoutIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	spell := drawMiracleInOwnDrawStep(t, g, 1, "Pyre Sweep", mmSorceryType, mmSorceryCost, "")
	if miraclePrompt(g, me.ID) != nil {
		t.Error("a card was offered a miracle reveal with no Molecule Man")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Error("the {0} miracle cast was accepted with no Molecule Man")
	}
}

// "In YOUR hand": an opponent's Molecule Man gives me nothing.
func TestMoleculeManOnlyGrantsToItsControllersHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	moleculeManOnTable(g, g.Seats[2].ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, "Pyre Sweep", mmSorceryType, mmSorceryCost, "")
	if miraclePrompt(g, me.ID) != nil {
		t.Error("an opponent's Molecule Man offered me a miracle reveal")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Error("the {0} miracle cast was accepted under an opponent's Molecule Man")
	}
}

// "Nonland": a land drawn first is not offered a reveal, and is not
// cast-able for {0} (a land is played, not cast).
func TestMoleculeManAFirstDrawnLandIsNotOffered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	moleculeManOnTable(g, me.ID)
	land := drawMiracleInOwnDrawStep(t, g, 1, "Mountain", "Basic Land — Mountain", "", "")
	if miraclePrompt(g, me.ID) != nil {
		t.Error("a land was offered a miracle reveal")
	}
	if hasMiracleGrantUnder(me, land, game.AltCostKeyGrantedMiracle) {
		t.Error("a land holds a miracle grant")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Error("a land was cast for miracle {0}")
	}
}

// "The first card you drew this turn": a land drawn first uses the miracle
// up, so the nonland card drawn after it is the second and not offered.
func TestMoleculeManTheSecondCardDrawnIsNotOffered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	moleculeManOnTable(g, me.ID)
	drawMiracleInOwnDrawStep(t, g, 1, "Mountain", "Basic Land — Mountain", "", "")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	spell := pushMiracleOnTop(g, me, "Pyre Sweep", mmSorceryType, mmSorceryCost, "")
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if !me.Hand.Contains(spell) {
		t.Fatal("the second draw did not draw the spell")
	}
	if miraclePrompt(g, me.ID) != nil {
		t.Error("the second card drawn this turn was offered a miracle reveal")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Error("the second card drawn was cast for miracle {0}")
	}
}

// Molecule Man leaving ends the ability: a granted-but-uncast miracle can
// no longer be claimed once it is gone.
func TestMoleculeManLeavingEndsTheGrant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	mm := moleculeManOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, "Pyre Sweep", mmSorceryType, mmSorceryCost, "")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	resolveTopWithoutPassingHolder(t, g)
	if !hasMiracleGrantUnder(me, spell, game.AltCostKeyGrantedMiracle) {
		t.Fatal("the resolved trigger granted no miracle cast")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(mm); err != nil {
			t.Errorf("destroying Molecule Man: %v", err)
		}
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Error("the {0} miracle cast was accepted after Molecule Man left")
	}
}

// Passing priority is the decline once the grant exists, as for the
// printed keyword.
func TestMoleculeManPassingPriorityClosesTheWindow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	moleculeManOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, "Pyre Sweep", mmSorceryType, mmSorceryCost, "")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	resolveTopWithoutPassingHolder(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if hasMiracleGrantUnder(me, spell, game.AltCostKeyGrantedMiracle) {
		t.Error("passing priority left the miracle grant open")
	}
}

// A card that prints a miracle of its own has two (CR 113.2c): Terminus
// drawn under Molecule Man raises two reveals, and each is its own claim at
// its own price, {W} printed and {0} granted. Passing priority to resolve
// the second trigger is the decline of the first one's cast (miracle.go),
// so the trigger the holder orders LAST is the cast they keep, and ordering
// each way keeps each.
func TestMoleculeManOnACardThatPrintsMiracleHasBoth(t *testing.T) {
	kept := map[string]bool{}
	for _, reverse := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[1]
		moleculeManOnTable(g, me.ID)
		term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)

		// Two prompts: the printed reveal and the granted one.
		answerLatestTriggerPrompt(t, g, me.ID, true)
		answerLatestTriggerPrompt(t, g, me.ID, true)
		// Two triggers of one controller go on the stack in an order they pick.
		for _, c := range g.PendingChoices {
			if c == nil || c.Kind != game.PendingChoiceTriggerOrder || c.Chooser != me.ID {
				continue
			}
			order := append([]uuid.UUID(nil), c.TriggerOrderIDs...)
			if reverse {
				order[0], order[1] = order[1], order[0]
			}
			if err := g.ResolveTriggerOrder(c.ID, me.ID, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			break
		}
		resolveTopWithoutPassingHolder(t, g)

		printed := hasMiracleGrantUnder(me, term, game.AltCostKeyMiracle)
		granted := hasMiracleGrantUnder(me, term, game.AltCostKeyGrantedMiracle)
		if printed == granted {
			t.Fatalf("reverse=%v: grants printed %v, granted %v, want the last-resolved one alone", reverse, printed, granted)
		}
		if granted {
			kept["granted"] = true
			if err := g.CastSpell(me.ID, term, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err != nil {
				t.Fatalf("the granted miracle {0} cast on Terminus: %v", err)
			}
			continue
		}
		kept["printed"] = true
		// The printed claim still costs {W}: an empty pool cannot pay it.
		if err := g.CastSpell(me.ID, term, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle}); err == nil {
			t.Fatal("the printed miracle {W} cast was accepted with an empty pool")
		}
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
			t.Fatal(err)
		}
		if err := g.CastSpell(me.ID, term, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle}); err != nil {
			t.Fatalf("the printed miracle {W} cast on Terminus: %v", err)
		}
	}
	if !kept["printed"] || !kept["granted"] {
		t.Errorf("ordering each way kept %v, want both the printed and the granted cast", kept)
	}
}

// The granted reveal is asked before the stack, so a table waiting on it
// is a restore point (the trigger frame names Molecule Man's row), and the
// restored "yes" still reveals and grants the {0} cast. A trigger the
// reveal put on the stack is a restore point too (its body is keyed).
func TestMoleculeManRevealPromptAndTriggerAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	moleculeManOnTable(g, g.Seats[1].ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, "Pyre Sweep", mmSorceryType, mmSorceryCost, "")
	if miraclePrompt(g, g.Seats[1].ID) == nil {
		t.Fatal("no granted miracle prompt to restore across")
	}

	g = restoreRoundTrip(t, g, true)
	me := g.Seats[1]
	answerLatestTriggerPrompt(t, g, me.ID, true)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[1]
	resolveTopWithoutPassingHolder(t, g)
	if !hasMiracleGrantUnder(me, spell, game.AltCostKeyGrantedMiracle) {
		t.Fatal("the restored reveal granted no miracle cast")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err != nil {
		t.Fatalf("the granted miracle cast after the restore: %v", err)
	}
}
