package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// restrictions_test.go is the #544 half of the S24 restriction
// vocabulary, and it is the half that can hang a table.
//
// The package's promise is that every Move EnumerateFor returns is
// one actions.Dispatch accepts. A restriction breaks that promise in
// the direction that matters: the engine refuses a declaration the
// enumerator offered, the seat's policy re-picks the same argmax,
// and because `legal` returns only a choice's answers while a seat
// owes one, there is nothing else for a bot to do. #544 was a search
// prompt; the declare-attackers step is the same shape with more
// creatures on it.
//
// The mechanism that keeps the two in agreement is a single
// predicate each, not a mirrored rule: game.AttackerEligible for
// attacks (the engine's bulk DeclareAttackers runs the same
// function) and game.CanBlock for blocks (DeclareBlocker runs the
// same function). These tests would fail if either were re-derived
// here.

const (
	oraclePacifism      = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
	oracleWhispersilk   = "9ad4f730-a18e-4a7c-a468-a926c718c741"
	oracleCarrionFeeder = "a1cc5e37-b09a-4b7f-afd5-77c1c35aa425"
	oracleFaithsFetters = "2b2d76f5-4c9b-49dc-b202-68095e2d9b29"
	oracleRoguesPassage = "f29dc596-2121-4421-8463-15f6c2e8b9b3"
)

// attachTo hangs an Aura or Equipment on a host and settles the
// layer engine, the way a resolution would.
func attachTo(t *testing.T, g *game.Game, attachment, host uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: host})
	})
	if err != nil {
		t.Fatalf("AttachForEffect: %v", err)
	}
}

// freshCreature drops an unsick creature on the battlefield for a
// seat. Cards pushed straight onto the zone never entered, so they
// carry no summoning sickness and the tests fail on the restriction
// they are about.
func freshCreature(g *game.Game, p *game.Player, name string) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
}

// attacksBy counts the declare-attacker moves offered for one
// creature.
func attacksBy(moves []legal.Move, creature uuid.UUID) int {
	n := 0
	for _, m := range moves {
		if m.Kind == legal.KindAttack && m.Source == creature {
			n++
		}
	}
	return n
}

func blocksBy(moves []legal.Move, blocker uuid.UUID) int {
	n := 0
	for _, m := range moves {
		if m.Kind == legal.KindBlock && m.Source == blocker {
			n++
		}
	}
	return n
}

// TestPacifiedCreatureIsNeverOfferedAsAnAttacker is the invariant,
// stated where it would break: the enumerator does not offer an
// attack the engine refuses, and the creature beside it is still
// offered so the test cannot pass by enumerating nothing.
func TestPacifiedCreatureIsNeverOfferedAsAnAttacker(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)

	pacified := freshCreature(g, me, "Pacified Bear")
	free := freshCreature(g, me, "Free Bear")
	aura := battlefieldCard(g, me, game.Card{
		Name: "Pacifism", TypeLine: "Enchantment — Aura", OracleID: oraclePacifism,
	})
	attachTo(t, g, aura, pacified)
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)

	// The invariant: nothing offered may be refused.
	dispatchAll(t, g, me.ID, moves)

	if n := attacksBy(moves, pacified); n != 0 {
		t.Errorf("the enumerator offered %d attacks with a pacified creature: %v", n, labels(moves))
	}
	if n := attacksBy(moves, free); n == 0 {
		t.Fatalf("no attack was offered for the unrestricted creature, so this test proves nothing: %v", labels(moves))
	}

	// And the other side of the same coin: the omission is real,
	// because the engine really does refuse the move. Without this a
	// future enumerator could satisfy the test by going silent on
	// both creatures.
	if err := g.DeclareAttacker(pacified, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID); err != game.ErrCantAttack {
		t.Errorf("engine accepted the attack the enumerator withheld: %v", err)
	}
}

// TestRestrictedBlocksAreNeverOffered covers both block-side bits at
// once, because they are enforced in the same predicate and a single
// declaration exercises both sides of it: a Cloaked attacker that
// cannot be blocked, and a Carrion Feeder that cannot block.
func TestRestrictedBlocksAreNeverOffered(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(them)

	cloaked := freshCreature(g, me, "Cloaked Bear")
	plain := freshCreature(g, me, "Plain Bear")
	cloak := battlefieldCard(g, me, game.Card{
		Name: "Whispersilk Cloak", TypeLine: "Artifact — Equipment", OracleID: oracleWhispersilk,
	})
	attachTo(t, g, cloak, cloaked)

	feeder := battlefieldCard(g, them, game.Card{
		Name: "Carrion Feeder", TypeLine: "Creature — Zombie", Power: 1, Toughness: 1,
		OracleID: oracleCarrionFeeder,
	})
	wall := freshCreature(g, them, "Wall")

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{cloaked, plain} {
		if err := g.DeclareAttacker(a, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	moves := legal.EnumerateFor(g, them.ID)
	dispatchAll(t, g, them.ID, moves)

	if n := blocksBy(moves, feeder); n != 0 {
		t.Errorf("the enumerator offered %d blocks with a can't-block creature: %v", n, labels(moves))
	}
	for _, m := range moves {
		if m.Kind == legal.KindBlock && m.Source == wall {
			// The only attacker the Wall may block is the plain one.
			if !hasLabel([]legal.Move{m}, "Block Plain Bear") {
				t.Errorf("offered a block against an unblockable attacker: %q", m.Label)
			}
		}
	}
	if n := blocksBy(moves, wall); n != 1 {
		t.Errorf("the Wall was offered %d blocks, want exactly 1 (the unrestricted attacker): %v", n, labels(moves))
	}

	// Both refusals are real.
	if err := g.DeclareBlocker(feeder, plain); err != game.ErrIllegalBlock {
		t.Errorf("engine accepted a block by a can't-block creature: %v", err)
	}
	if err := g.DeclareBlocker(wall, cloaked); err != game.ErrIllegalBlock {
		t.Errorf("engine accepted a block against an unblockable attacker: %v", err)
	}
}

// TestFetteredPermanentsAbilitiesAreNeverOffered is the activation
// half. Faith's Fetters is the interesting case rather than Arrest,
// because it stops one kind of activation and not the other, so an
// enumerator that simply dropped every ability of a restricted
// permanent would fail here too.
func TestFetteredPermanentsAbilitiesAreNeverOffered(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)

	// Rogue's Passage has both shapes on one card: a mana ability
	// Fetters spares, and an activated ability it stops.
	passage := battlefieldCard(g, me, game.Card{
		Name: "Rogue's Passage", TypeLine: "Land", OracleID: oracleRoguesPassage,
	})
	freshCreature(g, me, "Bear")
	// Five lands so the {4} is affordable and the ability would
	// otherwise be enumerated.
	for i := 0; i < 5; i++ {
		battlefieldCard(g, me, basic("Forest", "Forest"))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	before := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, before)
	activations, manaMoves := 0, 0
	for _, m := range before {
		if m.Source != passage {
			continue
		}
		switch m.Kind {
		case legal.KindActivate:
			activations++
		case legal.KindMana:
			manaMoves++
		}
	}
	if activations == 0 || manaMoves == 0 {
		t.Fatalf("baseline is wrong: %d activations, %d mana moves for Rogue's Passage: %v",
			activations, manaMoves, labels(before))
	}

	fetters := battlefieldCard(g, me, game.Card{
		Name: "Faith's Fetters", TypeLine: "Enchantment — Aura", OracleID: oracleFaithsFetters,
	})
	attachTo(t, g, fetters, passage)

	after := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, after)
	activations, manaMoves = 0, 0
	for _, m := range after {
		if m.Source != passage {
			continue
		}
		switch m.Kind {
		case legal.KindActivate:
			activations++
		case legal.KindMana:
			manaMoves++
		}
	}
	if activations != 0 {
		t.Errorf("a Fettered permanent's activated ability is still offered: %v", labels(after))
	}
	if manaMoves == 0 {
		t.Error(`Fetters says "unless they're mana abilities" — the mana ability must stay on offer`)
	}
}
