package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sorin_deck_c_4077_test.go — the four Vampires of the Sorin's Birthday
// Party deck request (#4077, slice c): Bloodline Keeper // Lord of
// Lineage, Champion of Dusk, Elenda the Dusk Rose and Preacher of the
// Schism.

const (
	s4077BloodlineKeeperOracle = "434e82b5-abd4-41ac-a5b1-2dd01c667374"
	s4077ChampionOfDuskOracle  = "5f7f0a38-fb7e-4ee8-a712-9aa911f17dae"
	s4077ElendaOracle          = "ce35f262-b93f-4409-8b33-26a7ed7abac8"
	s4077PreacherOracle        = "74947e47-c4ff-4f47-9845-befa7403cbde"
)

func TestS4077CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s4077BloodlineKeeperOracle:        "Bloodline Keeper",
		s4077BloodlineKeeperOracle + "#1": "Lord of Lineage",
		s4077ChampionOfDuskOracle:         "Champion of Dusk",
		s4077ElendaOracle:                 "Elenda, the Dusk Rose",
		s4077PreacherOracle:               "Preacher of the Schism",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Bloodline Keeper // Lord of Lineage ---------------------------

func s4077KeeperRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Bloodline Keeper // Lord of Lineage",
		Layout:   "transform",
		TypeLine: "Creature — Vampire // Creature — Vampire",
		OracleID: uuid.MustParse(s4077BloodlineKeeperOracle),
		Keywords: []string{"Flying", "Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Bloodline Keeper", TypeLine: "Creature — Vampire", ManaCost: "{2}{B}{B}", Colors: []string{"B"}, Power: "3", Toughness: "3",
				OracleText: "Flying\n{T}: Create a 2/2 black Vampire creature token with flying.\n{B}: Transform this creature. Activate only if you control five or more Vampires."},
			{Name: "Lord of Lineage", TypeLine: "Creature — Vampire", Colors: []string{"B"}, Power: "5", Toughness: "5",
				OracleText: "Flying\nOther Vampire creatures you control get +2/+2.\n{T}: Create a 2/2 black Vampire creature token with flying."},
		},
	}
}

func s4077PutKeeper(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(s4077KeeperRow(), p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatal("the Keeper did not reach the hand")
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

func s4077VampireTokens(g *game.Game, controller uuid.UUID) int {
	n, _ := countTokensNamed(g, controller, "Vampire")
	return n
}

// The front face taps for a 2/2 black flying Vampire.
func TestS4077BloodlineKeeperMakesAFlyingVampire(t *testing.T) {
	g, me := iwTable(t)
	keeper := s4077PutKeeper(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the token ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := s4077VampireTokens(g, me.ID); got != 1 {
		t.Fatalf("%d Vampire tokens, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && IsToken(c) && c.Name == "Vampire" {
			if p, th := effectivePTOf(t, g, c.InstanceID); p != 2 || th != 2 {
				t.Errorf("token is %d/%d, want 2/2", p, th)
			}
			if !effectiveAbilitiesContain(t, g, c.InstanceID, "flying") {
				t.Error("the token has no flying")
			}
			if len(c.Colors) != 1 || c.Colors[0] != "B" {
				t.Errorf("token colours %v, want black", c.Colors)
			}
		}
	}
	if c, _ := battlefieldCard(g, keeper); !c.Tapped {
		t.Error("the {T} cost did not tap the Keeper")
	}
}

// The transform is gated on FIVE Vampires at activation: four refuse
// with nothing paid, five turn it over.
func TestS4077BloodlineKeeperTransformsOnlyWithFiveVampires(t *testing.T) {
	g, me := iwTable(t)
	keeper := s4077PutKeeper(t, g, me)
	for i := 0; i < 3; i++ {
		b12Creature(g, me.ID, "Fang", "Creature — Vampire", 1, 1)
	}
	b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Vampire", "Creature — Vampire", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{B}")
	// Keeper + three Fangs = four of mine; the opponent's Vampire and my
	// Bear do not count.
	err := g.ActivateCatalogAbility(me.ID, keeper, 1, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("four Vampires: got %v, want ErrConditionNotMet", err)
	}
	if iwFace(g, keeper) != 0 {
		t.Fatal("it turned over with four Vampires")
	}

	b12Creature(g, me.ID, "Fifth", "Creature — Vampire", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, keeper, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("five Vampires: %v", err)
	}
	passPriorityAroundTable(t, g)
	if iwFace(g, keeper) != 1 {
		t.Fatal("the Keeper did not turn into Lord of Lineage")
	}
	if c, _ := battlefieldCard(g, keeper); c.Name != "Lord of Lineage" {
		t.Errorf("name %q after transforming", c.Name)
	}
}

// Activation instructions are not re-checked at resolution (CR 602.1b):
// a Vampire dying in response does not stop the transform.
func TestS4077BloodlineKeeperTransformSurvivesAVampireDyingInResponse(t *testing.T) {
	g, me := iwTable(t)
	keeper := s4077PutKeeper(t, g, me)
	var fangs []uuid.UUID
	for i := 0; i < 4; i++ {
		fangs = append(fangs, b12Creature(g, me.ID, "Fang", "Creature — Vampire", 1, 1))
	}
	floatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fangs[0]) })
	passPriorityAroundTable(t, g)
	if iwFace(g, keeper) != 1 {
		t.Error("the transform was lost when a Vampire died in response")
	}
}

// Lord of Lineage: +2/+2 to OTHER Vampire creatures its controller
// controls, and the same token ability.
func TestS4077LordOfLineageAnthemAndToken(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keeper := s4077PutKeeper(t, g, me)
	mine := b12Creature(g, me.ID, "Fang", "Creature — Vampire", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, _ := effectivePTOf(t, g, mine); p != 1 {
		t.Fatalf("the front face pumped a Vampire: power %d", p)
	}
	iwTransform(t, g, keeper)
	if p, th := effectivePTOf(t, g, mine); p != 3 || th != 3 {
		t.Errorf("my Vampire is %d/%d under the Lord, want 3/3", p, th)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Vampire got the bonus: power %d", p)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a non-Vampire got the bonus: power %d", p)
	}
	if p, th := effectivePTOf(t, g, keeper); p != 5 || th != 5 {
		t.Errorf("the Lord is %d/%d, want its printed 5/5 (it pumps only others)", p, th)
	}
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Lord's token ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := s4077VampireTokens(g, me.ID); got != 1 {
		t.Fatalf("%d tokens, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && IsToken(c) {
			if p, th := effectivePTOf(t, g, c.InstanceID); p != 4 || th != 4 {
				t.Errorf("the new token is %d/%d under the Lord, want 4/4", p, th)
			}
		}
	}
}

// --- Champion of Dusk ----------------------------------------------

func TestS4077ChampionOfDuskDrawsAndLosesPerVampire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Fang", "Creature — Vampire", 1, 1)
	b12Creature(g, me.ID, "Fang Two", "Creature — Vampire Knight", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[1].ID, "Their Vampire", "Creature — Vampire", 1, 1)
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Champion of Dusk", "Creature — Vampire Knight", s4077ChampionOfDuskOracle, nil)
	passPriorityAroundTable(t, g)
	// Two of my Vampires + the Champion itself; the opponent's Vampire
	// and my Bear do not count.
	x := 3
	// The Champion left the hand when cast: hand - 0 + X.
	if got := me.Hand.Size(); got != hand+x {
		t.Errorf("hand %d, want %d (drew X=%d)", got, hand+x, x)
	}
	if me.Life != life-x {
		t.Errorf("life %d, want %d (lost X=%d)", me.Life, life-x, x)
	}
}

// A lone Champion counts itself: X is one.
func TestS4077ChampionOfDuskCountsItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Champion of Dusk", "Creature — Vampire Knight", s4077ChampionOfDuskOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d (want %d), life %d (want %d)", me.Hand.Size(), hand+1, me.Life, life-1)
	}
}

// --- Elenda, the Dusk Rose -----------------------------------------

func TestS4077ElendaGrowsOnAnyDeathAndLeavesLifelinkVampires(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elenda := b12Push(g, me.ID, "Elenda, the Dusk Rose", "Legendary Creature — Vampire Knight", s4077ElendaOracle, 1, 1)
	if !effectiveAbilitiesContain(t, g, elenda, "lifelink") {
		t.Error("Elenda has no lifelink")
	}
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	advanceToMain(t, g)
	b15Destroy(t, g, theirs)
	b15Destroy(t, g, mine)
	if got := counterCount(g, elenda, "+1/+1"); got != 2 {
		t.Fatalf("Elenda has %d counters after two deaths, want 2", got)
	}
	// An anthem raises her power as she dies: X is last-known power.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment", OracleID: gloriousAnthemOracle,
		Owner: me.ID, Controller: me.ID,
	})
	if p := effectivePowerWithCounters(t, g, elenda); p != 4 {
		t.Fatalf("Elenda is %d power under the anthem, want 4", p)
	}
	b15Destroy(t, g, elenda)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && IsToken(c) && c.Name == "Vampire" {
			n++
			if !effectiveAbilitiesContain(t, g, c.InstanceID, "lifelink") {
				t.Error("a token has no lifelink")
			}
			if len(c.Colors) != 1 || c.Colors[0] != "W" {
				t.Errorf("token colours %v, want white", c.Colors)
			}
		}
	}
	if n != 4 {
		t.Errorf("%d Vampire tokens, want 4 (her last-known power, counters and anthem included)", n)
	}
}

// A 1/1 Elenda leaves one token; her own death does not grow her.
func TestS4077ElendaDyingAloneMakesPowerManyTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elenda := b12Push(g, me.ID, "Elenda, the Dusk Rose", "Legendary Creature — Vampire Knight", s4077ElendaOracle, 1, 1)
	advanceToMain(t, g)
	b15Destroy(t, g, elenda)
	if got := s4077VampireTokens(g, me.ID); got != 1 {
		t.Errorf("%d tokens, want 1", got)
	}
}

// --- Preacher of the Schism ----------------------------------------

func s4077PreacherTable(t *testing.T, myLife, leaderLife, otherLife int) (g *game.Game, me *game.Player, preacher uuid.UUID, leader, other *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	leader, other = opps[0], opps[1]
	opps[2].Life = 1
	me.Life, leader.Life, other.Life = myLife, leaderLife, otherLife
	preacher = b12Push(g, me.ID, "Preacher of the Schism", "Creature — Vampire Cleric", s4077PreacherOracle, 2, 4)
	return
}

func s4077Hand(me *game.Player) int { return me.Hand.Size() }

// s4077Attack declares the Preacher attacking defender and answers the
// CR 603.3b ordering prompt when both of its triggers fired.
func s4077Attack(t *testing.T, g *game.Game, me *game.Player, defender, preacher uuid.UUID) {
	t.Helper()
	declareAttack(t, g, defender, preacher)
	if len(g.PendingTriggers) > 1 {
		if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
			t.Fatal("two triggers were pending but no ordering prompt came")
		}
	}
}

// Ahead of the table and attacking the other leader (tied): both fire.
func TestS4077PreacherAttackingTheLeaderWhileLeadingFiresBoth(t *testing.T) {
	g, me, preacher, leader, _ := s4077PreacherTable(t, 40, 40, 30)
	hand, life := s4077Hand(me), me.Life
	s4077Attack(t, g, me, leader.ID, preacher)
	passPriorityAroundTable(t, g)
	if got := s4077VampireTokens(g, me.ID); got != 1 {
		t.Errorf("%d tokens, want 1 (attacked a player tied for most life)", got)
	}
	if s4077Hand(me) != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d (want %d), life %d (want %d): the lead trigger draws and loses 1", s4077Hand(me), hand+1, me.Life, life-1)
	}
}

// Attacking a player who is NOT at the top, while leading: draw only.
func TestS4077PreacherAttackingAnyoneWhileLeadingDrawsOnly(t *testing.T) {
	g, me, preacher, _, other := s4077PreacherTable(t, 40, 45, 30)
	// me 40 < leader 45: not leading. Raise me to lead outright.
	me.Life = 50
	hand := s4077Hand(me)
	s4077Attack(t, g, me, other.ID, preacher)
	passPriorityAroundTable(t, g)
	if got := s4077VampireTokens(g, me.ID); got != 0 {
		t.Errorf("%d tokens, want 0 (the defender is not the life leader)", got)
	}
	if s4077Hand(me) != hand+1 || me.Life != 49 {
		t.Errorf("hand %d (want %d), life %d (want 49)", s4077Hand(me), hand+1, me.Life)
	}
}

// Behind the table, attacking the leader: the token only.
func TestS4077PreacherBehindAttackingTheLeaderMakesTheTokenOnly(t *testing.T) {
	g, me, preacher, leader, _ := s4077PreacherTable(t, 20, 45, 30)
	hand := s4077Hand(me)
	s4077Attack(t, g, me, leader.ID, preacher)
	passPriorityAroundTable(t, g)
	if got := s4077VampireTokens(g, me.ID); got != 1 {
		t.Errorf("%d tokens, want 1", got)
	}
	if s4077Hand(me) != hand || me.Life != 20 {
		t.Errorf("hand %d (want %d), life %d (want 20): not leading, so no draw", s4077Hand(me), hand, me.Life)
	}
}

// Behind the table, attacking someone who is not the leader: nothing.
func TestS4077PreacherBehindAttackingAnOrdinaryPlayerDoesNothing(t *testing.T) {
	g, me, preacher, _, other := s4077PreacherTable(t, 20, 45, 30)
	hand := s4077Hand(me)
	s4077Attack(t, g, me, other.ID, preacher)
	passPriorityAroundTable(t, g)
	if got := s4077VampireTokens(g, me.ID); got != 0 || s4077Hand(me) != hand || me.Life != 20 {
		t.Errorf("tokens %d, hand %d (want %d), life %d: neither trigger should fire", got, s4077Hand(me), hand, me.Life)
	}
}

// An eliminated player's life total is not "the most": a dead seat with
// a huge total must not stop the controller counting as the leader.
func TestS4077PreacherIgnoresEliminatedPlayers(t *testing.T) {
	g, me, preacher, leader, other := s4077PreacherTable(t, 20, 10, 90)
	other.Eliminated = true
	hand := s4077Hand(me)
	s4077Attack(t, g, me, leader.ID, preacher)
	passPriorityAroundTable(t, g)
	if s4077Hand(me) != hand+1 {
		t.Errorf("hand %d, want %d: with the 90-life seat out the controller leads", s4077Hand(me), hand+1)
	}
	if got := s4077VampireTokens(g, me.ID); got != 0 {
		t.Errorf("%d tokens, want 0: the defender (10) is not the leader (20)", got)
	}
}
