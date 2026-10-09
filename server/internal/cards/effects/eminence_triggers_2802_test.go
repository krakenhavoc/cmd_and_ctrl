package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eminence_triggers_2802_test.go — #2802's last three eminence
// commanders, each on EminenceTrigger: Sidar Jabari of Zhalfir, Arahbo,
// Roar of the World and Inalla, Archmage Ritualist.

const (
	sidarJabariOracle = "b229f314-a5e5-41a3-a8c6-217a3c5a61c3"
	arahboOracle      = "66944a11-40a1-4f3a-9f83-52324e0edfef"
	inallaOracle      = "21bdba6e-3f9d-4ead-8212-0cbb0ce7f8cc"
)

// eminenceCommander puts a commander card into p's command zone.
func eminenceCommander(p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, IsCommander: true,
		Owner: p.ID, Controller: p.ID, Power: power, Toughness: toughness,
	}
	p.Command.PushTop(c)
	return c.InstanceID
}

func TestEminenceCardsAreRegisteredWithACommandZoneTrigger(t *testing.T) {
	for _, oracle := range []string{sidarJabariOracle, arahboOracle, inallaOracle} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull {
			t.Fatalf("%s is not registered complete", oracle)
		}
		if !game.TriggerWatchesFromZone(spec.Triggered[0], game.ZoneCommand) ||
			!game.TriggerWatchesFromZone(spec.Triggered[0], game.ZoneBattlefield) {
			t.Errorf("%s: the eminence trigger watches from %v", spec.Name, game.TriggerZones(spec.Triggered[0]))
		}
		for i, tr := range spec.Triggered[1:] {
			if game.TriggerWatchesFromZone(tr, game.ZoneCommand) {
				t.Errorf("%s: trigger %d works from the command zone", spec.Name, i+1)
			}
		}
	}
}

// --- Sidar Jabari of Zhalfir ---------------------------------------------

// From the command zone: an attack with two Knights is one loot.
func TestSidarJabariLootsOncePerAttackFromTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sidar := eminenceCommander(me, "Sidar Jabari of Zhalfir", "Legendary Creature — Human Knight", sidarJabariOracle, 4, 3)
	a := b12Push(g, me.ID, "Knight A", "Creature — Human Knight", "", 2, 2)
	b := b12Push(g, me.ID, "Knight B", "Creature — Human Knight", "", 2, 2)

	declareAttack(t, g, opp.ID, a, b)
	if n := triggersOnStackFrom(g, sidar); n != 1 {
		t.Fatalf("Sidar Jabari triggered %d times, want 1", n)
	}
	hand := me.Hand.Size()
	passPriorityUntilChoice(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d after the draw", got, hand+1)
	}
	if discardPromptFor(g, me) == nil {
		t.Error("no discard after the draw")
	}
}

// No Knight attacking, or Sidar Jabari in an opponent's command zone:
// nothing.
func TestSidarJabariNeedsYourKnight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := eminenceCommander(me, "Sidar Jabari of Zhalfir", "Legendary Creature — Human Knight", sidarJabariOracle, 4, 3)
	theirs := eminenceCommander(opp, "Sidar Jabari of Zhalfir", "Legendary Creature — Human Knight", sidarJabariOracle, 4, 3)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)

	declareAttack(t, g, opp.ID, bear)
	if n := triggersOnStackFrom(g, mine) + triggersOnStackFrom(g, theirs); n != 0 {
		t.Errorf("%d triggers for an attack with no Knight, want none", n)
	}
}

// On the battlefield: its combat damage returns a Knight creature card
// from its controller's graveyard.
func TestSidarJabariReturnsAKnightOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sidar := b12Push(g, me.ID, "Sidar Jabari of Zhalfir", "Legendary Creature — Human Knight", sidarJabariOracle, 4, 3)
	knight := game.Card{InstanceID: uuid.New(), Name: "Dead Knight", TypeLine: "Creature — Human Knight",
		Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2}
	me.Graveyard.PushTop(knight)

	declareAttack(t, g, opp.ID, sidar)
	passPriorityUntilChoice(t, g)
	if discardPromptFor(g, me) == nil {
		t.Fatal("Sidar Jabari's own attack did not loot")
	}
	answerChooseCards(t, g, me.ID, me.Hand.Cards[0].InstanceID)

	life := opp.Life
	for i := 0; i < 40 && latestPickTarget(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if opp.Life != life-4 {
		t.Fatalf("opponent life %d, want %d", opp.Life, life-4)
	}
	pickTriggerTarget(t, g, me.ID, knight.InstanceID)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, knight.InstanceID); !ok || c.Controller != me.ID {
		t.Error("the Knight is not back on the battlefield under my control")
	}
}

// --- Arahbo, Roar of the World -------------------------------------------

// From the command zone, at the beginning of combat on its owner's
// turn: another target Cat they control gets +3/+3. A non-Cat is not a
// legal target.
func TestArahboPumpsACatFromTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eminenceCommander(me, "Arahbo, Roar of the World", "Legendary Creature — Cat Avatar", arahboOracle, 5, 5)
	cat := b12Push(g, me.ID, "Cat", "Creature — Cat", "", 2, 2)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)

	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := sgAnswerTargets(t, g, bear); err == nil {
		t.Fatal("a Bear is not a Cat")
	}
	if err := sgAnswerTargets(t, g, cat); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, cat); p != 5 || th != 5 {
		t.Errorf("Cat P/T = %d/%d, want 5/5", p, th)
	}
}

// Arahbo in an opponent's command zone does nothing on my turn.
func TestArahboWorksOnlyOnItsOwnersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	eminenceCommander(opp, "Arahbo, Roar of the World", "Legendary Creature — Cat Avatar", arahboOracle, 5, 5)
	b12Push(g, opp.ID, "Their Cat", "Creature — Cat", "", 2, 2)
	b12Push(g, me.ID, "My Cat", "Creature — Cat", "", 2, 2)

	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if p := latestPickTarget(g, opp.ID); p != nil {
		t.Error("an opponent's Arahbo triggered on my turn")
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("an opponent's Arahbo asked me for a target")
	}
}

// On the battlefield: another Cat attacking offers {1}{G}{W} for
// trample and +X/+X, X its power as the trigger resolves.
func TestArahboPaidAttackDoublesTheCatAndGivesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	arahbo := b12Push(g, me.ID, "Arahbo, Roar of the World", "Legendary Creature — Cat Avatar", arahboOracle, 5, 5)
	cat := b12Push(g, me.ID, "Cat", "Creature — Cat", "", 2, 2)

	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := sgAnswerTargets(t, g, arahbo); err == nil {
		t.Fatal("Arahbo targeted itself")
	}
	if err := sgAnswerTargets(t, g, cat); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)

	declareAttack(t, g, opp.ID, arahbo, cat)
	if n := triggersOnStackFrom(g, arahbo); n != 1 {
		t.Fatalf("Arahbo's attack trigger fired %d times, want 1 (for the other Cat only)", n)
	}
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{G}{W}") })
	passPriorityUntilChoice(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, cat); p != 10 || th != 10 {
		t.Errorf("Cat P/T = %d/%d, want 10/10 (5/5, then +5/+5)", p, th)
	}
	if c, _ := battlefieldCard(g, cat); !game.HasKeyword(&c, "trample") {
		t.Error("the Cat did not gain trample")
	}
}

// --- Inalla, Archmage Ritualist ------------------------------------------

// From the command zone: a nontoken Wizard entering offers {1} for a
// hasty token copy, exiled at the next end step.
func TestInallaCopiesAWizardFromTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eminenceCommander(me, "Inalla, Archmage Ritualist", "Legendary Creature — Human Wizard", inallaOracle, 4, 5)

	castCatalogSpell(t, g, "Test Wizard", "Creature — Human Wizard", "", nil)
	passPriorityUntilChoice(t, g)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}") })
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	wizards := battlefieldIDsNamed(g, "Test Wizard")
	if len(wizards) != 2 {
		t.Fatalf("%d Test Wizards, want the card and its copy", len(wizards))
	}
	var token game.Card
	for _, id := range wizards {
		if c, _ := battlefieldCard(g, id); c.IsToken() {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil || !game.HasKeyword(&token, "haste") || token.Controller != me.ID {
		t.Fatalf("the copy is missing, has no haste, or is not mine: %+v", token)
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, token.InstanceID); ok {
		t.Error("the copy was not exiled at the end step")
	}
	if n := len(battlefieldIDsNamed(g, "Test Wizard")); n != 1 {
		t.Errorf("%d Test Wizards after the end step, want the card only", n)
	}
}

// Declining makes nothing; a non-Wizard does not trigger.
func TestInallaDeclinedOrNotAWizard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eminenceCommander(me, "Inalla, Archmage Ritualist", "Legendary Creature — Human Wizard", inallaOracle, 4, 5)

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a Bear entering offered a copy")
	}

	castCatalogSpell(t, g, "Test Wizard", "Creature — Human Wizard", "", nil)
	passPriorityUntilChoice(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Test Wizard")); n != 1 {
		t.Errorf("%d Test Wizards after declining, want 1", n)
	}
}

// Tap five untapped Wizards, Inalla among them: target player loses 7.
func TestInallaTapsFiveWizardsToDrain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inalla := b12Push(g, me.ID, "Inalla, Archmage Ritualist", "Legendary Creature — Human Wizard", inallaOracle, 4, 5)
	tap := []uuid.UUID{inalla}
	for i := 0; i < 4; i++ {
		tap = append(tap, b12Push(g, me.ID, "Wizard", "Creature — Human Wizard", "", 1, 1))
	}
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	if err := g.ActivateCatalogAbility(me.ID, inalla, 0, game.ActivateAbilityParams{
		Targets: target, TapIDs: append(append([]uuid.UUID(nil), tap[:4]...), bear),
	}); err == nil {
		t.Fatal("a Bear paid for a Wizard")
	}
	life := opp.Life
	pr7Activate(t, g, me.ID, inalla, 0, game.ActivateAbilityParams{Targets: target, TapIDs: tap})
	if opp.Life != life-7 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-7)
	}
	for _, id := range tap {
		if c, _ := battlefieldCard(g, id); !c.Tapped {
			t.Errorf("%s is untapped", c.Name)
		}
	}
}
