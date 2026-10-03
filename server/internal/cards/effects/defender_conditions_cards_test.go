package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// defender_conditions_cards_test.go — the rest of #1879 (ADR 0107 §2):
// the cards whose "can't attack unless" is a condition on the defending
// player other than "controls a permanent", their "can't block unless"
// halves, and Veiled Serpent, which gains the Island restriction from a
// resolved trigger. The engine half is pinned in
// game/attack_unless_defender_is_test.go.

const (
	chainedThroatseekerOracle = "f971a597-caa7-4289-ae8f-1efdc9abb593"
	crownHunterHirelingOracle = "fece36a3-deb6-4fef-a456-b6514056bfd2"
	goblinGoonOracle          = "e91bcf01-113b-4532-9f78-2a155b17be4f"
	moggToadyOracle           = "77c7ba06-d247-421d-9e0c-54c3000c1b78"
	monstrousHoundOracle      = "97589fbb-3bd9-49c9-b8d8-bdc895a7f1e6"
	vantressGargoyleOracle    = "fc7c0197-bfa1-43b8-87c8-f134192b14c0"
	veiledSerpentOracle       = "ff59a95c-28de-44cf-abbe-772417851ffa"
)

func wantAttackRefused(t *testing.T, g *game.Game, attacker, target uuid.UUID, why string) {
	t.Helper()
	err := declareResult(g, attacker, target)
	if !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Fatalf("attack %s: err = %v, want ErrIllegalAttackTarget", target, err)
	}
	if why != "" && !strings.Contains(err.Error(), why) {
		t.Errorf("refusal %q does not say %q", err.Error(), why)
	}
}

// TestChainedThroatseekerAttacksOnlyAPoisonedOpponent — one poison counter
// is enough (CR 122.1f); an opponent with none can't be attacked.
func TestChainedThroatseekerAttacksOnlyAPoisonedOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	horror := pushTestPermanent(g, me.ID, game.Card{Name: "Chained Throatseeker", OracleID: chainedThroatseekerOracle,
		TypeLine: "Creature — Phyrexian Horror", Power: 5, Toughness: 5})
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(b.ID, game.CounterPoison, 1); err != nil {
			t.Fatal(err)
		}
	})
	advanceToDeclareAttackersOf(t, g, 0)

	if err := declareResult(g, horror, b.ID); err != nil {
		t.Errorf("attack the poisoned opponent: %v", err)
	}
	wantAttackRefused(t, g, horror, c.ID, "isn't poisoned")
	if !hasString(effectiveAbilities(t, g, horror), "infect") {
		t.Error("Chained Throatseeker has no infect")
	}
}

// TestCrownHunterHirelingWaitsForAnOpponentToTakeTheCrown — it makes you
// the monarch as it enters, so it can't attack at all until an opponent
// is the monarch, and then only that opponent.
func TestCrownHunterHirelingWaitsForAnOpponentToTakeTheCrown(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	ogre := castCatalogSpell(t, g, "Crown-Hunter Hireling", "Creature — Ogre Mercenary", crownHunterHirelingOracle, nil)
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %s, want the Hireling's controller", g.Monarch)
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, ogre).SummonedThisTurn = false })
	advanceToDeclareAttackersOf(t, g, 0)

	for _, opp := range []uuid.UUID{b.ID, c.ID} {
		wantAttackRefused(t, g, ogre, opp, "isn't the monarch")
	}
	monCrown(t, g, b.ID)
	if err := declareResult(g, ogre, b.ID); err != nil {
		t.Errorf("attack the monarch: %v", err)
	}
	wantAttackRefused(t, g, ogre, c.ID, "isn't the monarch")
}

// TestMoreThanDefendingPlayerCards — Goblin Goon, Mogg Toady (creatures)
// and Monstrous Hound (lands). I control two of the counted kind, the
// card included when it counts itself; B controls one, C two. It may
// attack B and not C.
func TestMoreThanDefendingPlayerCards(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		kind                   string
	}{
		{"Goblin Goon", goblinGoonOracle, "Creature — Goblin Mutant", "creature"},
		{"Mogg Toady", moggToadyOracle, "Creature — Goblin", "creature"},
		{"Monstrous Hound", monstrousHoundOracle, "Creature — Dog", "land"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
			id := pushTestPermanent(g, me.ID, game.Card{Name: tc.name, OracleID: tc.oracle, TypeLine: tc.typeLine, Power: 4, Toughness: 4})
			counted := func(p uuid.UUID, n int) {
				for i := 0; i < n; i++ {
					if tc.kind == "creature" {
						pushTestPermanent(g, p, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
					} else {
						pushTestPermanent(g, p, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
					}
				}
			}
			mine := 2
			if tc.kind == "creature" {
				mine = 1 // the card itself is the second
			}
			counted(me.ID, mine)
			counted(b.ID, 1)
			counted(c.ID, 2)
			advanceToDeclareAttackersOf(t, g, 0)

			if err := declareResult(g, id, b.ID); err != nil {
				t.Errorf("attack the opponent with fewer %ss: %v", tc.kind, err)
			}
			wantAttackRefused(t, g, id, c.ID, "doesn't control more "+tc.kind+"s than")
		})
	}
}

// TestMoreThanAttackingPlayerBlockHalf — "can't block unless you control
// more creatures than attacking player" (CR 509.1b), read as blockers are
// declared. The verb and the enumerator agree.
func TestMoreThanAttackingPlayerBlockHalf(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, attacker.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b12Creature(g, attacker.ID, "Hill Giant", "Creature — Giant", 3, 3)
	goon := b12Push(g, defender.ID, "Goblin Goon", "Creature — Goblin Mutant", goblinGoonOracle, 6, 6)
	b12Creature(g, defender.ID, "Elf", "Creature — Elf", 1, 1)
	brAttack(t, g, bear)

	// Two each: not more.
	if brOffers(brOffered(t, g, defender.ID), goon, bear) {
		t.Error("the enumerator offers a block the verb refuses")
	}
	refused := brRefusal(t, g.Clone().DeclareBlocker(goon, bear), game.BlockReasonCantBlockAttacker)
	if !strings.Contains(refused.Sentence(uuid.Nil), "more creatures than the attacking player") {
		t.Errorf("sentence %q does not say why", refused.Sentence(uuid.Nil))
	}
	// A third creature: three against two.
	b12Creature(g, defender.ID, "Elf", "Creature — Elf", 1, 1)
	if !brOffers(brOffered(t, g, defender.ID), goon, bear) {
		t.Error("the enumerator does not offer the now-legal block")
	}
	if err := g.DeclareBlocker(goon, bear); err != nil {
		t.Errorf("block with three creatures against two: %v", err)
	}
}

// TestMonstrousHoundBlocksOnlyWithMoreLands — the land count's block half.
func TestMonstrousHoundBlocksOnlyWithMoreLands(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, attacker.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b12Permanent(g, attacker.ID, "Mountain", "Basic Land — Mountain")
	hound := b12Push(g, defender.ID, "Monstrous Hound", "Creature — Dog", monstrousHoundOracle, 4, 4)
	b12Permanent(g, defender.ID, "Mountain", "Basic Land — Mountain")
	// #1501: a creature that can block keeps the defender's declaration
	// open while the land count changes under it; with no legal block
	// at all it would complete as the step began, and the block below
	// would be refused as late.
	b12Creature(g, defender.ID, "Elf", "Creature — Elf", 1, 1)
	brAttack(t, g, bear)

	brRefusal(t, g.Clone().DeclareBlocker(hound, bear), game.BlockReasonCantBlockAttacker)
	b12Permanent(g, defender.ID, "Mountain", "Basic Land — Mountain")
	if err := g.DeclareBlocker(hound, bear); err != nil {
		t.Errorf("block with two lands against one: %v", err)
	}
}

// TestVantressGargoyle — it attacks only a defending player with seven or
// more cards in their graveyard, blocks only while its controller has
// four or more cards in hand, and taps to make each player mill a card.
func TestVantressGargoyle(t *testing.T) {
	t.Run("attack", func(t *testing.T) {
		g := newCatalogGame(t)
		me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
		gargoyle := pushTestPermanent(g, me.ID, game.Card{Name: "Vantress Gargoyle", OracleID: vantressGargoyleOracle,
			TypeLine: "Artifact Creature — Gargoyle", Power: 5, Toughness: 4})
		for i := 0; i < 7; i++ {
			b.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", Owner: b.ID})
		}
		for i := 0; i < 6; i++ {
			c.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", Owner: c.ID})
		}
		advanceToDeclareAttackersOf(t, g, 0)
		if err := declareResult(g, gargoyle, b.ID); err != nil {
			t.Errorf("attack the seven-card graveyard: %v", err)
		}
		wantAttackRefused(t, g, gargoyle, c.ID, "fewer than seven cards in their graveyard")
		if !hasString(effectiveAbilities(t, g, gargoyle), "flying") {
			t.Error("Vantress Gargoyle has no flying")
		}
	})
	t.Run("block", func(t *testing.T) {
		g := newCatalogGame(t)
		attacker, defender := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, attacker.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
		gargoyle := b12Push(g, defender.ID, "Vantress Gargoyle", "Artifact Creature — Gargoyle", vantressGargoyleOracle, 5, 4)
		for len(defender.Hand.Cards) > 3 {
			defender.Hand.Cards = defender.Hand.Cards[1:]
		}
		for len(defender.Hand.Cards) < 3 {
			defender.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", Owner: defender.ID, Controller: defender.ID})
		}
		// #1501: a creature that can block keeps the declaration open
		// while the hand grows under it (see Monstrous Hound above).
		b12Creature(g, defender.ID, "Elf", "Creature — Elf", 1, 1)
		brAttack(t, g, bear)
		brRefusal(t, g.Clone().DeclareBlocker(gargoyle, bear), game.BlockReasonCantBlockAttacker)
		defender.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", Owner: defender.ID, Controller: defender.ID})
		if err := g.DeclareBlocker(gargoyle, bear); err != nil {
			t.Errorf("block with four cards in hand: %v", err)
		}
	})
	t.Run("mill", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		gargoyle := pushTestPermanent(g, me.ID, game.Card{Name: "Vantress Gargoyle", OracleID: vantressGargoyleOracle,
			TypeLine: "Artifact Creature — Gargoyle", Power: 5, Toughness: 4})
		before := map[uuid.UUID]int{}
		for _, p := range g.Seats {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", Owner: p.ID})
			before[p.ID] = p.Graveyard.Size()
		}
		if err := g.ActivateCatalogAbility(me.ID, gargoyle, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		for _, p := range g.Seats {
			if got := p.Graveyard.Size(); got != before[p.ID]+1 {
				t.Errorf("%s graveyard = %d, want %d", p.Name, got, before[p.ID]+1)
			}
		}
		if c := findBattlefieldCardForTest(g, gargoyle); c == nil || !c.Tapped {
			t.Error("the Gargoyle did not tap")
		}
	})
}

// TestVeiledSerpentBecomesARestrictedSerpentWhenAnOpponentCasts — an
// opponent's spell turns it into a blue 4/4 Serpent creature that is no
// longer an enchantment (CR 205.1a) and can't attack unless defending
// player controls an Island. It does not trigger again (CR 603.4), and
// my own spell never triggers it.
func TestVeiledSerpentBecomesARestrictedSerpentWhenAnOpponentCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	serpent := pushTestPermanent(g, me.ID, game.Card{Name: "Veiled Serpent", OracleID: veiledSerpentOracle,
		TypeLine: "Enchantment", ManaCost: "{2}{U}"})
	pushTestPermanent(g, b.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})

	batch01OpponentCasts(t, g, b, "Opt", "", "{U}", nil)
	if triggerOnStack(g, serpent) == nil {
		t.Fatal("an opponent's spell did not trigger Veiled Serpent")
	}
	passPriorityAroundTable(t, g)

	s := findBattlefieldCardForTest(g, serpent)
	if s == nil {
		t.Fatal("the Serpent left the battlefield")
	}
	eff := s.Effective()
	if !s.IsCreature() || s.HasCardType("enchantment") || !s.HasSubtype("Serpent") || eff.Power != 4 || eff.Toughness != 4 {
		t.Fatalf("it is %v %v %d/%d, want a 4/4 Serpent creature and not an enchantment", eff.Types, eff.Subtypes, eff.Power, eff.Toughness)
	}
	if !s.HasColor("U") {
		t.Error("it is no longer blue")
	}

	batch01OpponentCasts(t, g, c, "Opt", "", "{U}", nil)
	if triggerOnStack(g, serpent) != nil {
		t.Error("the creature triggered again: it is not an enchantment any more")
	}
	passPriorityAroundTable(t, g)

	advanceToDeclareAttackersOf(t, g, 0)
	if err := declareResult(g, serpent, b.ID); err != nil {
		t.Errorf("attack the opponent with an Island: %v", err)
	}
	wantAttackRefused(t, g, serpent, c.ID, "controls no Island")
}

// TestVeiledSerpentIgnoresItsControllersSpells — "an opponent casts".
func TestVeiledSerpentIgnoresItsControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	serpent := pushTestPermanent(g, me.ID, game.Card{Name: "Veiled Serpent", OracleID: veiledSerpentOracle,
		TypeLine: "Enchantment", ManaCost: "{2}{U}"})
	batch01OpponentCasts(t, g, me, "Opt", "", "{U}", nil)
	if triggerOnStack(g, serpent) != nil {
		t.Error("its controller's spell triggered it")
	}
}
