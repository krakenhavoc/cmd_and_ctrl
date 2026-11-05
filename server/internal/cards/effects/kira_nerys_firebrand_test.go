package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// kira_nerys_firebrand_test.go — "If a creature you control would enter
// while you control an attacking creature, instead it enters tapped and
// attacking" (CR 614.1d), with CR 508.4's choice of what it attacks made
// before it enters (CR 614.12a), and "Whenever another creature you
// control enters, it gets +1/+1 until end of turn."

const kiraNerysOracle = "ff9f5bbb-43c4-481e-9bd7-db9f246355ae"

func entryAttackPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryAttackTarget {
			return c
		}
	}
	return nil
}

// kiraCombat seats Kira and a Bear for seat 0, declares the Bear
// attacking seat 1, and moves to the declare blockers step. A Soldier
// card waits in seat 0's hand.
func kiraCombat(t *testing.T) (g *game.Game, me *game.Player, kira, bear, soldier uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	kira = pushCatalogPermanent(g, me.ID, "Kira Nerys, Firebrand", "Legendary Creature — Bajoran Soldier", kiraNerysOracle, false)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	soldier = pushNinjaToHand(me, "Soldier", "Creature — Human Soldier", "", 2, 2)
	declareAttack(t, g, g.Seats[1].ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	return g, me, kira, bear, soldier
}

// putFromHand puts a card from its owner's hand onto the battlefield.
func putFromHand(t *testing.T, g *game.Game, card, controller uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		_, err = g.PutFromHandOntoBattlefieldForEffect(card, game.HandEntryOptions{Controller: controller})
	})
	if err != nil {
		t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
	}
}

// answerEntryAttack picks the option naming `defender`.
func answerEntryAttack(t *testing.T, g *game.Game, defender uuid.UUID) {
	t.Helper()
	c := entryAttackPrompt(g)
	if c == nil {
		t.Fatal("no entry_attack_target prompt is open")
	}
	for i, opt := range c.PickOptions {
		if opt.Player == defender || (len(opt.Cards) == 1 && opt.Cards[0] == defender) {
			if err := g.ResolveEntryAttackTarget(c.ID, c.Chooser, i); err != nil {
				t.Fatalf("ResolveEntryAttackTarget: %v", err)
			}
			return
		}
	}
	t.Fatalf("no option names %s: %+v", defender, c.PickOptions)
}

// While Kira's controller has an attacker, a creature entering under
// them is asked about before it enters, offered every defender with the
// one already attacked first, and enters tapped and attacking the one
// chosen. It was never declared, so it announces no attack (CR 508.4).
// Then Kira's trigger gives it +1/+1.
func TestKiraNerysACreatureEntersTappedAndAttackingTheChosenDefender(t *testing.T) {
	g, me, _, _, soldier := kiraCombat(t)
	target := g.Seats[2].ID
	putFromHand(t, g, soldier, me.ID)

	c := entryAttackPrompt(g)
	if c == nil {
		t.Fatal("the creature entered without being asked what it attacks")
	}
	if c.Chooser != me.ID {
		t.Fatalf("chooser = %s, want Kira's controller", c.Chooser)
	}
	if len(c.PickOptions) != 3 || c.PickOptions[0].Player != g.Seats[1].ID {
		t.Fatalf("options = %+v, want the three opponents with seat 1 (already attacked) first", c.PickOptions)
	}
	if _, onBF := battlefieldCardByID(g, soldier); onBF {
		t.Fatal("the creature is on the battlefield before its defender was chosen (CR 614.12a)")
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("the table moved on with the defender question open")
	}

	answerEntryAttack(t, g, target)
	card, ok := battlefieldCardByID(g, soldier)
	if !ok {
		t.Fatal("the answer did not land the creature")
	}
	if !card.Tapped {
		t.Error("the creature entered untapped")
	}
	if card.AttackingTarget != target {
		t.Errorf("the creature attacks %s, want seat 2 (%s)", card.AttackingTarget, target)
	}
	if n := attackEventsFor(g, soldier); n != 0 {
		t.Errorf("the creature announced %d attacks; it was put onto the battlefield attacking, never declared", n)
	}

	passPriorityAroundTable(t, g)
	card, _ = battlefieldCardByID(g, soldier)
	if p, tough := card.Effective().Power, card.Effective().Toughness; p != 3 || tough != 3 {
		t.Errorf("the creature is %d/%d after Kira's trigger, want 3/3", p, tough)
	}
}

// With no attacking creature, Kira changes nothing about the entry: no
// question, untapped, not attacking. Her trigger still pumps it.
func TestKiraNerysWithNoAttackerTheCreatureEntersNormally(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Kira Nerys, Firebrand", "Legendary Creature — Bajoran Soldier", kiraNerysOracle, false)
	soldier := pushNinjaToHand(me, "Soldier", "Creature — Human Soldier", "", 2, 2)
	putFromHand(t, g, soldier, me.ID)

	if c := entryAttackPrompt(g); c != nil {
		t.Fatalf("asked what it attacks with nothing attacking: %+v", c)
	}
	card, ok := battlefieldCardByID(g, soldier)
	if !ok {
		t.Fatal("the creature did not enter")
	}
	if card.Tapped || card.AttackingTarget != uuid.Nil {
		t.Errorf("tapped = %v, attacking = %s; want an ordinary entry", card.Tapped, card.AttackingTarget)
	}
	passPriorityAroundTable(t, g)
	card, _ = battlefieldCardByID(g, soldier)
	if p := card.Effective().Power; p != 3 {
		t.Errorf("power = %d after Kira's trigger, want 3", p)
	}
}

// Kira's replacement reads only her controller's creatures: an
// opponent's creature entering during her controller's attack, and a
// noncreature permanent of her controller's, enter as they would.
func TestKiraNerysTouchesOnlyHerControllersCreatures(t *testing.T) {
	g, me, _, _, _ := kiraCombat(t)
	opp := g.Seats[1]
	theirs := pushNinjaToHand(opp, "Their Soldier", "Creature — Human Soldier", "", 2, 2)
	putFromHand(t, g, theirs, opp.ID)
	relic := pushNinjaToHand(me, "Relic", "Artifact", "", 0, 0)
	putFromHand(t, g, relic, me.ID)

	if c := entryAttackPrompt(g); c != nil {
		t.Fatalf("asked what something attacks: %+v", c)
	}
	for _, id := range []uuid.UUID{theirs, relic} {
		card, ok := battlefieldCardByID(g, id)
		if !ok {
			t.Fatalf("%s did not enter", id)
		}
		if card.Tapped || card.AttackingTarget != uuid.Nil {
			t.Errorf("%s: tapped = %v, attacking = %s; want an ordinary entry", card.Name, card.Tapped, card.AttackingTarget)
		}
	}
}

// With one defender left to attack the choice is forced: no prompt, and
// the creature enters tapped and attacking it.
func TestKiraNerysWithOneDefenderTheChoiceIsForced(t *testing.T) {
	g, me, _, _, soldier := kiraCombat(t)
	g.Seats[2].Eliminated = true
	g.Seats[3].Eliminated = true
	putFromHand(t, g, soldier, me.ID)
	if c := entryAttackPrompt(g); c != nil {
		t.Fatalf("asked with one defender: %+v", c)
	}
	card, ok := battlefieldCardByID(g, soldier)
	if !ok || !card.Tapped || card.AttackingTarget != g.Seats[1].ID {
		t.Fatalf("the creature is %+v (on the battlefield %v), want tapped and attacking seat 1", card, ok)
	}
}

// A table with the defender question open is a restore point: the
// restored answer lands the creature tapped and attacking the defender
// picked.
func TestKiraNerysRestoreWithTheDefenderChoicePending(t *testing.T) {
	g, me, _, _, soldier := kiraCombat(t)
	putFromHand(t, g, soldier, me.ID)
	if entryAttackPrompt(g) == nil {
		t.Fatal("Kira did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	c := entryAttackPrompt(g)
	if c == nil || len(c.PickOptions) != 3 {
		t.Fatalf("the restored question is %+v, want three defenders", c)
	}
	if _, onBF := battlefieldCardByID(g, soldier); onBF {
		t.Fatal("the restored table landed the creature before the answer")
	}
	target := g.Seats[3].ID
	answerEntryAttack(t, g, target)
	card, ok := battlefieldCardByID(g, soldier)
	if !ok {
		t.Fatal("the restored answer did not land the creature")
	}
	if !card.Tapped || card.AttackingTarget != target {
		t.Fatalf("tapped = %v, attacking = %s; want tapped and attacking seat 3", card.Tapped, card.AttackingTarget)
	}
}

// The bot's half (#544's rule): while the question blocks the table the
// chooser is offered one answer per defender and nothing else, and every
// answer is one the dispatcher accepts.
func TestKiraNerysEnumeratorOffersEachDefenderAndDispatches(t *testing.T) {
	g, me, _, _, soldier := kiraCombat(t)
	putFromHand(t, g, soldier, me.ID)
	var answers []legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("the chooser was offered %q while the defender prompt blocks the table", m.Type)
		}
		answers = append(answers, m)
	}
	if len(answers) != 3 {
		t.Fatalf("enumerated %d answers, want one per defender (3)", len(answers))
	}
	last := answers[2]
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(last.Type), Player: last.Player, Caller: last.Player, Params: last.Params,
	}); err != nil {
		t.Fatalf("dispatch the enumerated answer: %v", err)
	}
	card, ok := battlefieldCardByID(g, soldier)
	if !ok || !card.Tapped || card.AttackingTarget == uuid.Nil {
		t.Fatalf("the dispatched answer left the creature %+v (on the battlefield %v)", card, ok)
	}
}

// A creature token created while Kira's controller attacks is asked
// about too (the replacement reads every creature entering under them),
// and a token created already attacking keeps the defender its effect
// named (CR 508.4's "unless the effect ... specifies") but still enters
// tapped.
func TestKiraNerysTokensEnterTappedAndAttacking(t *testing.T) {
	g, me, _, _, _ := kiraCombat(t)
	tmpl := game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}
	var created []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		created, err = g.CreateTokensForEffect(me.ID, tmpl, 1, game.TokenEntryOptions{})
	})
	if err != nil {
		t.Fatalf("CreateTokensForEffect: %v", err)
	}
	if entryAttackPrompt(g) == nil {
		t.Fatalf("the token entered (%v) without being asked what it attacks", created)
	}
	target := g.Seats[2].ID
	answerEntryAttack(t, g, target)
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" {
			token = c
		}
	}
	if !token.Tapped || token.AttackingTarget != target {
		t.Fatalf("the token is tapped %v and attacking %s, want tapped and attacking seat 2", token.Tapped, token.AttackingTarget)
	}

	named := g.Seats[3].ID
	g.WithWriteLock(func() {
		err = g.CreateTokensAttackingForEffect(me.ID, game.Card{Name: "Ensign", TypeLine: "Token Creature — Human", Power: 1, Toughness: 1}, 1, named)
	})
	if err != nil {
		t.Fatalf("CreateTokensAttackingForEffect: %v", err)
	}
	if c := entryAttackPrompt(g); c != nil {
		t.Fatalf("asked what a token created attacking attacks: %+v", c)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ensign" && (!c.Tapped || c.AttackingTarget != named) {
			t.Fatalf("the created-attacking token is tapped %v and attacking %s, want tapped and attacking seat 3", c.Tapped, c.AttackingTarget)
		}
	}
}
