package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exert_test.go — ADR 0130 test 10: the four PR 1 exert cards, driven
// through the declaration verb, the lock-in and the stack.

const (
	oketrasAvengerOracle    = "8f064160-3afe-408a-85b4-b335eae8571c"
	glorybringerOracle      = "b75c3902-633e-4d24-acde-d7a9cc8f466e"
	combatCelebrantOracle   = "5e15ff93-99a0-4000-918e-4bd2c257188d"
	resoluteSurvivorsOracle = "3d699db7-cc52-4ee3-947b-0ba291bf037a"
)

// declareExerting declares `id` attacking `defender` and chooses to
// exert it as it attacks, in the declare-attackers step.
func declareExerting(t *testing.T, g *game.Game, id, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: id, Target: defender, Exert: true}, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare %s exerting: %v", id, err)
	}
}

func exertedEvents(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventExert && ev.CardID == id {
			n++
		}
	}
	return n
}

func hasExerterSkip(g *game.Game, id, player uuid.UUID) bool {
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		return false
	}
	for _, s := range c.NextUntapSkips {
		if s.While == nil && s.Player == player {
			return true
		}
	}
	return false
}

// Oketra's Avenger, exerted, takes no combat damage this turn; other
// damage still lands.
func TestOketrasAvengerExertedPreventsCombatDamageToIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	avenger := pushDiesCreatureForTest(g, me.ID, "Oketra's Avenger", oketrasAvengerOracle, "Creature — Human Warrior", 3, 1)
	blocker := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	findBattlefieldCardForTest(g, blocker).Power = 3

	declareExerting(t, g, avenger, opp.ID)
	lockInAttacks(t, g)
	if triggerOnStack(g, avenger) == nil {
		t.Fatal("the linked trigger goes on the stack with the declaration")
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, avenger); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, avenger) == nil || damageMarkedOn(g, avenger) != 0 {
		t.Fatalf("the exerted Avenger took combat damage (marked %d)", damageMarkedOn(g, avenger))
	}
	if !hasExerterSkip(g, avenger, me.ID) {
		t.Error("it won't untap during its controller's next untap step")
	}
	pr6Damage(t, g, blocker, avenger, 1)
	if findBattlefieldCardForTest(g, avenger) != nil {
		t.Error("the shield is combat damage only; other damage still kills it")
	}
}

// Oketra's Avenger attacking without exerting is not shielded and
// skips nothing.
func TestOketrasAvengerNotExertedIsNotShielded(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	avenger := pushDiesCreatureForTest(g, me.ID, "Oketra's Avenger", oketrasAvengerOracle, "Creature — Human Warrior", 3, 1)
	blocker := pushVanillaCreature(g, opp.ID, "Bear", 2, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(avenger, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, avenger); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, avenger) != nil {
		t.Error("an Avenger that wasn't exerted dies to the block")
	}
	if exertedEvents(g, avenger) != 0 {
		t.Error("nothing exerted it")
	}
}

// Glorybringer with no legal target: it is exerted anyway, and the
// trigger is removed (the Amonkhet ruling, CR 603.3d).
func TestGlorybringerExertsWithNoLegalTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gb := pushDiesCreatureForTest(g, me.ID, "Glorybringer", glorybringerOracle, "Creature — Dragon", 4, 4)
	dragon := pushDiesCreatureForTest(g, opp.ID, "Their Dragon", "", "Creature — Dragon", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	declareExerting(t, g, gb, opp.ID)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if exertedEvents(g, gb) != 1 || !hasExerterSkip(g, gb, me.ID) {
		t.Fatal("Glorybringer is exerted with no legal target")
	}
	if damageMarkedOn(g, dragon) != 0 || damageMarkedOn(g, mine) != 0 {
		t.Error("a Dragon and your own creature are not legal targets")
	}
}

// Glorybringer exerted deals 4 to the target non-Dragon creature an
// opponent controls.
func TestGlorybringerExertedDealsFourToTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gb := pushDiesCreatureForTest(g, me.ID, "Glorybringer", glorybringerOracle, "Creature — Dragon", 4, 4)
	giant := pushVanillaCreature(g, opp.ID, "Giant", 5, 5)
	pushVanillaCreature(g, opp.ID, "Other Giant", 5, 5)
	dragon := pushDiesCreatureForTest(g, opp.ID, "Their Dragon", "", "Creature — Dragon", 5, 5)

	declareExerting(t, g, gb, opp.ID)
	lockInAttacks(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the trigger asks for its target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dragon}); err == nil {
		t.Fatal("a Dragon is not a legal target")
	}
	pickTriggerTarget(t, g, me.ID, giant)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, giant); got != 4 {
		t.Fatalf("the target has %d damage, want 4", got)
	}
}

// Resolute Survivors triggers on another creature's exert and on its
// own: 1 damage to each opponent and 1 life each time.
func TestResoluteSurvivorsTriggersOnEveryExert(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	survivors := pushDiesCreatureForTest(g, me.ID, "Resolute Survivors", resoluteSurvivorsOracle, "Creature — Human Warrior", 3, 3)
	avenger := pushDiesCreatureForTest(g, me.ID, "Oketra's Avenger", oketrasAvengerOracle, "Creature — Human Warrior", 3, 1)
	lifeMe := me.Life
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: survivors, Target: opp.ID},
		{Attacker: avenger, Target: opp.ID, Exert: true},
	}, game.DeclareAttackersParams{}); err != nil {
		t.Fatal(err)
	}
	// The Avenger's linked trigger and Survivors' payoff are both mine.
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Life != lifeMe+1 {
		t.Fatalf("life %d, want %d: one exert, one trigger", me.Life, lifeMe+1)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Life != lives[p.ID]-1 {
			t.Errorf("opponent %s at %d, want %d", p.Name, p.Life, lives[p.ID]-1)
		}
	}
	if exertedEvents(g, survivors) != 0 {
		t.Error("Survivors attacked without exerting")
	}
}

func TestResoluteSurvivorsSeesItsOwnExert(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	survivors := pushDiesCreatureForTest(g, me.ID, "Resolute Survivors", resoluteSurvivorsOracle, "Creature — Human Warrior", 3, 3)
	lifeMe := me.Life
	declareExerting(t, g, survivors, opp.ID)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Life != lifeMe+1 {
		t.Fatalf("life %d, want %d", me.Life, lifeMe+1)
	}
}

// Combat Celebrant: exerted, it untaps your other creatures and adds a
// combat after this one; it stays tapped itself, and in the added
// combat it can't be exerted again this turn.
func TestCombatCelebrantUntapsOthersAndAddsACombatOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	celebrant := pushDiesCreatureForTest(g, me.ID, "Combat Celebrant", combatCelebrantOracle, "Creature — Human Warrior", 4, 1)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	findBattlefieldCardForTest(g, theirs).Tapped = true

	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: celebrant, Target: opp.ID, Exert: true},
		{Attacker: bear, Target: opp.ID},
	}, game.DeclareAttackersParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Error("Celebrant untaps the other creatures you control")
	}
	if !isTapped(g, celebrant) {
		t.Error("\"other\": Celebrant itself stays tapped")
	}
	if !isTapped(g, theirs) {
		t.Error("only creatures you control are untapped")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want the added combat", next)
	}
	findBattlefieldCardForTest(g, celebrant).Tapped = false
	advanceTo(t, g, game.StepDeclareAttackers)
	err := g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: celebrant, Target: opp.ID, Exert: true}, game.DeclareAttackersParams{})
	if !errors.Is(err, game.ErrCantExert) {
		t.Fatalf("a second exert this turn: err = %v, want ErrCantExert", err)
	}
	if err := g.DeclareAttacker(celebrant, opp.ID); err != nil {
		t.Fatalf("it may still attack without exerting: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the added combat: %+v, want the main phase (no second extra combat)", next)
	}
}

// Two Celebrants exerted in one combat give two added combats, and the
// added combat happens even if the Celebrant has died (the Amonkhet
// rulings).
func TestTwoCombatCelebrantsGiveTwoCombats(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := pushDiesCreatureForTest(g, me.ID, "Combat Celebrant", combatCelebrantOracle, "Creature — Human Warrior", 4, 1)
	b := pushDiesCreatureForTest(g, me.ID, "Combat Celebrant", combatCelebrantOracle, "Creature — Human Warrior", 4, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: a, Target: opp.ID, Exert: true},
		{Attacker: b, Target: opp.ID, Exert: true},
	}, game.DeclareAttackersParams{}); err != nil {
		t.Fatal(err)
	}
	if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("two linked triggers, one controller: an ordering prompt")
	}
	// One Celebrant dies with both triggers waiting.
	pr6Damage(t, g, b, a, 1)
	if findBattlefieldCardForTest(g, a) != nil {
		t.Fatal("setup: the Celebrant should have died")
	}
	passPriorityAroundTable(t, g)
	combats := 1
	for i := 0; i < 40; i++ {
		before := g.Turn
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
		if g.Turn.Step == game.StepBeginCombat && g.Turn != before {
			combats++
		}
		if g.Turn.Step == game.StepEnd {
			break
		}
	}
	if combats != 3 {
		t.Errorf("%d combat phases this turn, want 3", combats)
	}
}

// TestExertPurposeGuard is ADR 0130's amendment of 2026-10-07: a pump
// and a self shield are an ability's, about its own source.
func TestExertPurposeGuard(t *testing.T) {
	cases := []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string // "" = accepted
	}{
		{"pump on a triggered row", purposeOnTriggered, game.Purpose{Pump: &game.Pump{Power: 2, Keywords: []string{"first strike"}}}, ""},
		{"pump on an activated row", purposeOnActivated, game.Purpose{Pump: &game.Pump{Toughness: 1}}, ""},
		{"pump on the card", purposeOnCard, game.Purpose{Pump: &game.Pump{Power: 1}}, "off a triggered or activated row"},
		{"pump on a mode", purposeOnMode, game.Purpose{Pump: &game.Pump{Power: 1}}, "off a triggered or activated row"},
		{"shield on the card", purposeOnCard, game.Purpose{PreventCombatDamageToSelf: true}, "off a triggered or activated row"},
		{"empty pump", purposeOnTriggered, game.Purpose{Pump: &game.Pump{}}, "gives nothing"},
		{"capital keyword", purposeOnTriggered, game.Purpose{Pump: &game.Pump{Keywords: []string{"Trample"}}}, "not lowercase"},
		{"negative extra combat", purposeOnCard, game.Purpose{ExtraCombat: -1}, "negative amount"},
		{"negative life", purposeOnTriggered, game.Purpose{LifeGain: -1}, "negative amount"},
		{"damage on a spell", purposeOnCard, game.Purpose{DamageToCreature: 3, DamageEachOpponent: 1, LifeGain: 1, ExtraCombat: 1}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				checkPurpose("Test Card", "triggered ability 0", c.slot, c.p)
			}()
			switch {
			case c.want == "" && msg != "":
				t.Errorf("refused: %s", msg)
			case c.want != "" && !strings.Contains(msg, c.want):
				t.Errorf("got %q, want a refusal containing %q", msg, c.want)
			}
		})
	}
}

// TestExertHelpersStampTheirRows: WhenExerted's row is "linked",
// WheneverYouExert's is "payoff" (ADR 0130's amendment of 2026-10-07).
func TestExertHelpersStampTheirRows(t *testing.T) {
	if got := WhenExerted("x", nil).Exert; got != game.ExertRowLinked {
		t.Errorf("WhenExerted stamps %q", got)
	}
	if got := WheneverYouExert("x", nil).Exert; got != game.ExertRowPayoff {
		t.Errorf("WheneverYouExert stamps %q", got)
	}
}
