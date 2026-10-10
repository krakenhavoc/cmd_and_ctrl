package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// costed_option_cards_2854_test.go — #2854: Winter's Chill, Lim-Dûl's
// Hex and Thrull Wizard, the cards whose "may pay {N} or {M}" is one
// option_pick whose options carry their mana cost.

const (
	s2854WintersChill = "ae78d498-11c7-427f-9d13-3f1b7096f803"
	s2854LimDulsHex   = "7f9116e1-9aab-470b-90eb-f51a3fbb3c0e"
	s2854ThrullWizard = "3cae0e4b-1827-4ed2-832f-98f8e79941a5"
)

func s2854Lands(g *game.Game, owner uuid.UUID, name, typeLine string, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, apaPush(g, owner, owner, game.Card{Name: name, TypeLine: typeLine}))
	}
	return out
}

// s2854Pick is the one open option_pick, which must be `chooser`'s.
func s2854Pick(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no option_pick open: %+v", g.PendingChoices)
	}
	if c.Chooser != chooser {
		t.Fatalf("the option_pick is %s's, want %s's", c.Chooser, chooser)
	}
	return c
}

func s2854Costs(c *game.PendingChoice) []string {
	out := make([]string, len(c.PickOptions))
	for i, o := range c.PickOptions {
		out[i] = o.ManaCost
	}
	return out
}

func s2854Answer(t *testing.T, g *game.Game, c *game.PendingChoice, index int) {
	t.Helper()
	if err := g.ResolveOptionPick(c.ID, c.Chooser, index); err != nil {
		t.Fatalf("ResolveOptionPick(%d): %v", index, err)
	}
}

func s2854Equal(a []string, b ...string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func s2854Untapped(g *game.Game, ids []uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if c := findBattlefieldCardForTest(g, id); c != nil && !c.Tapped {
			n++
		}
	}
	return n
}

// s2854Attack declares the attackers and locks them in, leaving the
// game in the declare attackers step, where Winter's Chill can be cast.
func s2854Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// The active player's pass locks the attack in and hands priority
	// on, still in the declare attackers step.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("at %s after declaring attackers", g.Turn.Step)
	}
}

// s2854CastChill has `caster` cast Winter's Chill with X = len(targets)
// at the attacking creatures, with the snow lands and mana to do it.
func s2854CastChill(t *testing.T, g *game.Game, caster *game.Player, targets ...uuid.UUID) {
	t.Helper()
	s2854Lands(g, caster.ID, "Snow-Covered Island", "Basic Snow Land — Island", len(targets))
	apaMana(caster, "U")
	for range targets {
		apaMana(caster, "C")
	}
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Winter's Chill", TypeLine: "Instant", ManaCost: "{X}{U}",
		OracleID: s2854WintersChill, Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{XValue: len(targets), Targets: pr7bTargets(targets...)}); err != nil {
		t.Fatalf("CastSpell Winter's Chill: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// Winter's Chill, each creature in turn: its controller is offered
// what they can pay, the {2} creature fights as normal, the {1} one
// neither deals nor is dealt combat damage, and the one nobody paid
// for deals its damage and is destroyed at end of combat.
func TestWintersChillAsksAboutEachCreatureAndSettlesEachAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	paid2 := pr7Creature(g, me.ID, "Paid Two", 3, "R")
	paid1 := pr7Creature(g, me.ID, "Paid One", 4, "R")
	unpaid := pr7Creature(g, me.ID, "Unpaid", 2, "R")
	lands := s2854Lands(g, me.ID, "Mountain", "Basic Land — Mountain", 3)
	s2854Attack(t, g, opp.ID, paid2, paid1, unpaid)
	s2854CastChill(t, g, opp, paid2, paid1, unpaid)

	c := s2854Pick(t, g, me.ID)
	if !s2854Equal(s2854Costs(c), "", "{1}", "{2}") {
		t.Fatalf("offered %v to a player with three lands, want nothing, {1}, {2}", s2854Costs(c))
	}
	s2854Answer(t, g, c, 2)
	if n := s2854Untapped(g, lands); n != 1 {
		t.Fatalf("%d lands untapped after paying {2}, want 1", n)
	}

	c = s2854Pick(t, g, me.ID)
	if !s2854Equal(s2854Costs(c), "", "{1}") {
		t.Fatalf("offered %v with one land left, want nothing and {1}", s2854Costs(c))
	}
	s2854Answer(t, g, c, 1)

	c = s2854Pick(t, g, me.ID)
	if !s2854Equal(s2854Costs(c), "") {
		t.Fatalf("offered %v with no mana left, want only paying nothing", s2854Costs(c))
	}
	s2854Answer(t, g, c, 0)
	if pendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		t.Fatal("a fourth question was asked about three creatures")
	}

	if n := pr7bSourceShields(g); n != 1 {
		t.Fatalf("%d shields, want one, for the creature paid {1} for", n)
	}
	life := opp.Life
	pr7bDamageStep(t, g)
	if want := life - 3 - 2; opp.Life != want {
		t.Fatalf("defender at %d, want %d: the {2} and unpaid creatures deal damage, the {1} one doesn't", opp.Life, want)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, unpaid) != nil {
		t.Error("the creature nobody paid for survived end of combat")
	}
	for _, id := range []uuid.UUID{paid1, paid2} {
		if findBattlefieldCardForTest(g, id) == nil {
			t.Errorf("a creature paid for was destroyed")
		}
	}
}

// Every question Winter's Chill asks is a restore point, and the
// restored game carries on asking.
func TestWintersChillQuestionSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, me.ID, "A", 3, "R")
	b := pr7Creature(g, me.ID, "B", 3, "R")
	s2854Lands(g, me.ID, "Mountain", "Basic Land — Mountain", 1)
	s2854Attack(t, g, opp.ID, a, b)
	s2854CastChill(t, g, opp, a, b)
	first := s2854Pick(t, g, me.ID)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("an open Winter's Chill question blocks the restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	c := s2854Pick(t, restored, me.ID)
	if c.ID != first.ID || !s2854Equal(s2854Costs(c), "", "{1}") {
		t.Fatalf("restored question %v, want the same one offering nothing and {1}", s2854Costs(c))
	}
	s2854Answer(t, restored, c, 1)
	next := s2854Pick(t, restored, me.ID)
	if !s2854Equal(s2854Costs(next), "") {
		t.Fatalf("second question offers %v with the land spent, want only paying nothing", s2854Costs(next))
	}
	s2854Answer(t, restored, next, 0)
	if n := pr7bSourceShields(restored); n != 1 {
		t.Fatalf("%d shields on the restored game, want 1", n)
	}
}

// Winter's Chill can be cast only during combat before blockers.
func TestWintersChillIsCastOnlyBeforeBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, me.ID, "A", 3, "R")
	s2854Attack(t, g, opp.ID, a)
	advanceTo(t, g, game.StepDeclareBlockers)
	s2854Lands(g, opp.ID, "Snow-Covered Island", "Basic Snow Land — Island", 1)
	apaMana(opp, "U", "C")
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Winter's Chill", TypeLine: "Instant", ManaCost: "{X}{U}",
		OracleID: s2854WintersChill, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{XValue: 1, Targets: pr7bTargets(a)}); err == nil {
		t.Fatal("Winter's Chill was cast in the declare blockers step")
	}
}

// Lim-Dûl's Hex: every player is asked in APNAP order, each offered what
// they can pay, and the players who paid nothing are dealt 1 damage
// once everyone has answered.
func TestLimDulsHexAsksEveryPlayerThenDamagesThoseWhoDidNotPay(t *testing.T) {
	g := newCatalogGame(t)
	// The next upkeep is seat 1's, so the Hex is theirs.
	owner := g.Seats[1]
	pushCatalogPermanent(g, owner.ID, "Lim-Dûl's Hex", "Enchantment", s2854LimDulsHex, false)
	s2854Lands(g, owner.ID, "Swamp", "Basic Land — Swamp", 1)
	s2854Lands(g, g.Seats[2].ID, "Mountain", "Basic Land — Mountain", 3)
	advanceTo(t, g, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}

	c := s2854Pick(t, g, g.Seats[1].ID)
	if !s2854Equal(s2854Costs(c), "", "{B}") {
		t.Fatalf("the Swamp's controller is offered %v, want nothing and {B}", s2854Costs(c))
	}
	s2854Answer(t, g, c, 1)
	c = s2854Pick(t, g, g.Seats[2].ID)
	if !s2854Equal(s2854Costs(c), "", "{3}") {
		t.Fatalf("three Mountains are offered %v, want nothing and {3}", s2854Costs(c))
	}
	s2854Answer(t, g, c, 1)
	c = s2854Pick(t, g, g.Seats[3].ID)
	s2854Answer(t, g, c, 0)
	if g.Seats[3].Life != lives[g.Seats[3].ID] {
		t.Fatal("damage was dealt before every player had answered")
	}
	c = s2854Pick(t, g, g.Seats[0].ID)
	s2854Answer(t, g, c, 0)

	for i, want := range []int{-1, 0, 0, -1} {
		p := g.Seats[i]
		if got := p.Life - lives[p.ID]; got != want {
			t.Errorf("seat %d life changed by %d, want %d", i, got, want)
		}
	}
}

// Thrull Wizard: the spell's controller pays {B} or {3}, or the spell is
// countered.
func TestThrullWizardCountersUnlessItsControllerPays(t *testing.T) {
	for _, tc := range []struct {
		name    string
		swamps  int
		answer  int
		offered []string
		counter bool
	}{
		{"pays {B}", 1, 1, []string{"", "{B}"}, false},
		{"pays nothing", 1, 0, []string{"", "{B}"}, true},
		{"cannot pay", 0, 0, []string{""}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			wizard := pushCatalogPermanent(g, opp.ID, "Thrull Wizard", "Creature — Thrull Wizard", s2854ThrullWizard, false)
			spell := batch01OpponentCasts(t, g, me, "Black Spell", lightningBoltOracle, "{B}",
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
			s2854Lands(g, me.ID, "Swamp", "Basic Land — Swamp", tc.swamps)
			apaMana(opp, "B", "C")
			if err := g.ActivateCatalogAbility(opp.ID, wizard, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			c := s2854Pick(t, g, me.ID)
			if !s2854Equal(s2854Costs(c), tc.offered...) {
				t.Fatalf("offered %v, want %v", s2854Costs(c), tc.offered)
			}
			s2854Answer(t, g, c, tc.answer)
			if got := g.StackItemForEffect(spell) == nil; got != tc.counter {
				t.Fatalf("spell countered = %v, want %v", got, tc.counter)
			}
		})
	}
}
