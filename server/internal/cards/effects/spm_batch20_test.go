package effects

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_batch20_test.go — Marvel's Spider-Man batch 20 (#4460): Amazing
// Alliance and Spinneret and Spiderling, each against the readings the
// owner settled on 2026-11-02.

const (
	spm20AmazingAllianceOracle = "5a824aff-52fe-4402-b074-884afb5286d4"
	spm20SpinneretOracle       = "daca1d5e-349d-4e8a-abc7-427dab43fb0c"
)

func spm20Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: typeLine, Power: power, Toughness: toughness})
}

// --- Amazing Alliance ----------------------------------------------

func TestSpm20AmazingAllianceGivesYourCreaturesPlusOnePlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := spm20Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := spm20Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	pushCatalogPermanent(g, me.ID, "Amazing Alliance", "Enchantment", spm20AmazingAllianceOracle, false)

	if c := apaLive(g, mine); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("my creature is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := apaLive(g, theirs); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSpm20AmazingAllianceGainsOneLifePerLegendaryAttacker(t *testing.T) {
	cases := []struct {
		name      string
		legends   int
		bears     int
		wantGain  int
		wantStack bool
	}{
		{"one legend", 1, 0, 1, true},
		{"two legends", 2, 0, 2, true},
		{"three legends and two bears", 3, 2, 3, true},
		{"a legend among bears", 1, 3, 1, true},
		{"only bears", 0, 3, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			alliance := pushCatalogPermanent(g, me.ID, "Amazing Alliance", "Enchantment", spm20AmazingAllianceOracle, false)
			var attackers []uuid.UUID
			for i := 0; i < tc.legends; i++ {
				attackers = append(attackers, spm20Creature(g, me.ID, fmt.Sprintf("Legend %d", i), "Legendary Creature — Human", 1, 1))
			}
			for i := 0; i < tc.bears; i++ {
				attackers = append(attackers, spm20Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1))
			}
			life := me.Life
			declareAttack(t, g, opp.ID, attackers...)
			if got := triggersOnStackFrom(g, alliance) > 0; got != tc.wantStack {
				t.Fatalf("Amazing Alliance on the stack = %v, want %v", got, tc.wantStack)
			}
			if n := triggersOnStackFrom(g, alliance); n > 1 {
				t.Fatalf("Amazing Alliance triggered %d times, want one trigger per declaration", n)
			}
			passPriorityAroundTable(t, g)
			if got := me.Life - life; got != tc.wantGain {
				t.Errorf("life gained = %d, want %d", got, tc.wantGain)
			}
		})
	}
}

// Another player's legendary attack is not yours.
func TestSpm20AmazingAllianceIgnoresAnOpponentsLegendaryAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alliance := pushCatalogPermanent(g, opp.ID, "Amazing Alliance", "Enchantment", spm20AmazingAllianceOracle, false)
	legend := spm20Creature(g, me.ID, "Legend", "Legendary Creature — Human", 1, 1)
	life := opp.Life
	declareAttack(t, g, opp.ID, legend)
	passPriorityAroundTable(t, g)
	if n := triggersOnStackFrom(g, alliance); n != 0 || opp.Life != life {
		t.Fatalf("an opponent's attack triggered it (%d on the stack, life %d -> %d)", n, life, opp.Life)
	}
}

// "Legendary" is read through the layered view: a creature that is only
// legendary because an effect says so counts.
func TestSpm20AmazingAllianceReadsLegendaryThroughTheLayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Amazing Alliance", "Enchantment", spm20AmazingAllianceOracle, false)
	plain := spm20Creature(g, me.ID, "Plain Bear", "Creature — Bear", 1, 1)
	serum := pushCatalogPermanent(g, me.ID, "Super-Soldier Serum", "Enchantment — Aura", "6f2ba142-5e7c-4709-ae03-e028cb2a9872", false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(serum, game.TargetRef{Kind: game.TargetCard, ID: plain}); err != nil {
			t.Fatal(err)
		}
	})
	life := me.Life
	declareAttack(t, g, opp.ID, plain)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1 for a creature that is legendary through an effect", life, me.Life)
	}
}

// --- Spinneret and Spiderling --------------------------------------

func spm20Spinneret(g *game.Game, owner uuid.UUID, power int, keywords ...string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{
		Name: "Spinneret and Spiderling", OracleID: spm20SpinneretOracle,
		TypeLine: "Legendary Creature — Spider Human Hero", Power: power, Toughness: 5,
		Keywords: keywords,
	})
}

func spm20Counters(g *game.Game, id uuid.UUID) int {
	return findBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]
}

func TestSpm20SpinneretCountsSpidersThatAttackTogether(t *testing.T) {
	cases := []struct {
		name         string
		spinneretAtk bool
		spiders      int
		bears        int
		want         int
	}{
		{"two Spiders, Spinneret home", false, 2, 0, 1},
		{"Spinneret and one other Spider", true, 1, 0, 1},
		{"three Spiders and a bear", true, 2, 1, 1},
		{"one Spider and a bear", false, 1, 1, 0},
		{"Spinneret alone", true, 0, 0, 0},
		{"two bears", false, 0, 2, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			spin := spm20Spinneret(g, me.ID, 1)
			var attackers []uuid.UUID
			if tc.spinneretAtk {
				attackers = append(attackers, spin)
			}
			for i := 0; i < tc.spiders; i++ {
				attackers = append(attackers, spm20Creature(g, me.ID, "Spider", "Creature — Spider", 1, 1))
			}
			for i := 0; i < tc.bears; i++ {
				attackers = append(attackers, spm20Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1))
			}
			declareAttack(t, g, opp.ID, attackers...)
			passPriorityAroundTable(t, g)
			if got := spm20Counters(g, spin); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// spm20Combat attacks with `attacker` into opp, blocked by `blockers`,
// answers any damage assignment prompt with `toBlockers` to each blocker
// and `toPlayer` to the player, and runs the combat damage steps
// (`firstStrike` adds the first-strike step).
func spm20Combat(t *testing.T, g *game.Game, attacker uuid.UUID, opp *game.Player, blockers []uuid.UUID, toBlockers, toPlayer int, firstStrike bool) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatal(err)
		}
	}
	steps := []game.Step{game.StepCombatDamage}
	if firstStrike {
		steps = []game.Step{game.StepFirstStrikeDamage, game.StepCombatDamage}
	}
	for _, step := range steps {
		advanceTo(t, g, step)
		for i := 0; i < 4; i++ {
			var c *game.PendingChoice
			for _, p := range g.PendingChoices {
				if p != nil && p.Kind == game.PendingChoiceDamageAssignment {
					c = p
				}
			}
			if c == nil {
				break
			}
			var entries []game.DamageAssignmentEntry
			for _, b := range blockers {
				entries = append(entries, game.DamageAssignmentEntry{BlockerID: b, Amount: toBlockers})
			}
			if err := g.ResolveDamageAssignment(c.ID, c.Chooser, entries, toPlayer); err != nil {
				t.Fatalf("assign: %v", err)
			}
		}
	}
}

func spm20Exiled(g *game.Game) int { return len(g.Exile.Cards) }

func TestSpm20SpinneretDamageThresholds(t *testing.T) {
	cases := []struct {
		name        string
		power       int
		keywords    []string
		blockers    int
		blockerTgh  int
		toBlocker   int // assigned to each blocker
		toPlayer    int // trampled over
		firstStrike bool
		wantExile   int
		wantLife    int
	}{
		{"4 unblocked", 4, nil, 0, 0, 0, 0, false, 1, -4},
		{"3 unblocked does not trigger", 3, nil, 0, 0, 0, 0, false, 0, -3},
		{"2 to a blocker and 2 to the player is one 4", 4, []string{"trample"}, 1, 2, 2, 2, false, 1, -2},
		{"2 to a blocker and 1 to the player is 3", 3, []string{"trample"}, 1, 2, 2, 1, false, 0, -1},
		{"4 into a blocker only", 4, nil, 1, 5, 4, 0, false, 1, 0},
		{"2 and 2 across two blockers", 4, nil, 2, 3, 2, 0, false, 1, 0},
		{"double strike 2 is two separate 2s", 2, []string{"double strike"}, 0, 0, 0, 0, true, 0, -4},
		{"double strike 4 is two separate 4s", 4, []string{"double strike"}, 0, 0, 0, 0, true, 2, -8},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			spin := spm20Spinneret(g, me.ID, tc.power, tc.keywords...)
			var blockers []uuid.UUID
			for i := 0; i < tc.blockers; i++ {
				blockers = append(blockers, apaPush(g, opp.ID, opp.ID, game.Card{
					Name: "Blocker", TypeLine: "Creature — Test", Power: 0, Toughness: tc.blockerTgh,
				}))
			}
			life := opp.Life
			before := spm20Exiled(g)
			spm20Combat(t, g, spin, opp, blockers, tc.toBlocker, tc.toPlayer, tc.firstStrike)
			passPriorityAroundTable(t, g)
			if got := spm20Exiled(g) - before; got != tc.wantExile {
				t.Errorf("exiled %d cards, want %d", got, tc.wantExile)
			}
			if got := opp.Life - life; got != tc.wantLife {
				t.Errorf("opponent life changed by %d, want %d", got, tc.wantLife)
			}
		})
	}
}

// Outside combat: one damage instruction is one event however many
// recipients it has, and two instructions are two events that are never
// summed, even in the same resolution.
func TestSpm20SpinneretSumsOneInstructionButNotTwo(t *testing.T) {
	deal := func(g *game.Game, src uuid.UUID, victim, player uuid.UUID, n int) {
		if err := g.DealDamageToCreatureForEffect(src, victim, n); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToPlayerForEffect(src, player, n); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("one instruction, 2 + 2", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spin := spm20Spinneret(g, me.ID, 1)
		victim := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 0, Toughness: 9})
		before := spm20Exiled(g)
		g.WithWriteLock(func() {
			if err := g.DamageInstanceForEffect(func() error { deal(g, spin, victim, opp.ID, 2); return nil }); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
		if got := spm20Exiled(g) - before; got != 1 {
			t.Errorf("exiled %d cards, want 1", got)
		}
	})
	t.Run("two instructions, 2 then 2", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spin := spm20Spinneret(g, me.ID, 1)
		victim := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 0, Toughness: 9})
		before := spm20Exiled(g)
		g.WithWriteLock(func() {
			deal(g, spin, victim, opp.ID, 1)
			deal(g, spin, victim, opp.ID, 1)
		})
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
		if got := spm20Exiled(g) - before; got != 0 {
			t.Errorf("exiled %d cards, want 0: 2+2 dealt by two instructions is not summed", got)
		}
	})
	t.Run("another source's damage is not summed in", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spin := spm20Spinneret(g, me.ID, 1)
		other := spm20Creature(g, me.ID, "Other", "Creature — Test", 1, 1)
		before := spm20Exiled(g)
		g.WithWriteLock(func() {
			if err := g.DamageInstanceForEffect(func() error {
				if err := g.DealDamageToPlayerForEffect(spin, opp.ID, 2); err != nil {
					return err
				}
				return g.DealDamageToPlayerForEffect(other, opp.ID, 2)
			}); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
		if got := spm20Exiled(g) - before; got != 0 {
			t.Errorf("exiled %d cards, want 0", got)
		}
	})
}

// The exiled card is playable by its controller, and only through the end
// of their next turn.
func TestSpm20SpinneretExilesTheTopCardToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spin := spm20Spinneret(g, me.ID, 4)
	top := plTop(me, "Top Card", "Sorcery", "{R}")
	spm20Combat(t, g, spin, opp, nil, 0, 0, false)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if !b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the exiled card is not playable by its owner")
	}
	if b12ExiledPlayableBy(t, g, top, opp.ID) {
		t.Error("an opponent can play the exiled card")
	}
}
