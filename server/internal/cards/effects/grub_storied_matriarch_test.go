package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const grubOracle = "180286e4-67fa-44d5-9e90-4a4e99c016a7"

func grubRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Grub, Storied Matriarch // Grub, Notorious Auntie",
		Layout:   "transform",
		TypeLine: "Legendary Creature — Goblin Warlock // Legendary Creature — Goblin Warrior",
		OracleID: uuid.MustParse(grubOracle),
		Keywords: []string{"Blight", "Transform", "Menace"},
		CardFaces: []cards.CardFace{
			{Name: "Grub, Storied Matriarch", TypeLine: "Legendary Creature — Goblin Warlock", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: "2", Toughness: "1",
				OracleText: "Menace\nWhenever this creature enters or transforms into Grub, Storied Matriarch, return up to one target Goblin card from your graveyard to your hand.\nAt the beginning of your first main phase, you may pay {R}. If you do, transform Grub."},
			{Name: "Grub, Notorious Auntie", TypeLine: "Legendary Creature — Goblin Warrior", Colors: []string{"R"}, Power: "2", Toughness: "1",
				OracleText: "Menace\nWhenever Grub attacks, you may blight 1. If you do, create a tapped and attacking token that's a copy of the blighted creature, except it has \"At the beginning of the end step, sacrifice this token.\"\nAt the beginning of your first main phase, you may pay {B}. If you do, transform Grub."},
		},
	}
}

// grubPut puts Grub onto p's battlefield, front face up, without entering
// triggers and not summoning sick.
func grubPut(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(grubRow(), p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatal("Grub did not reach the hand")
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

// grubTransformed puts Grub down and pays {R} at the first main phase.
func grubTransformed(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := grubPut(t, g, me)
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	if f := iwFace(g, id); f != 1 {
		t.Fatalf("Grub is on face %d after paying {R}, want the back face", f)
	}
	return id
}

func TestGrubIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{grubOracle, grubOracle + "#1"} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Fatalf("%s: spec = %+v, %v", oracle, spec, ok)
		}
		if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "menace" {
			t.Errorf("%s: keywords %v, want menace", oracle, spec.PrintedKeywords)
		}
	}
}

// Entering returns a Goblin card from your graveyard to your hand; a
// non-Goblin card is not a legal target.
func TestGrubStoriedMatriarchReturnsAGoblinCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := game.Card{InstanceID: uuid.New(), Name: "Goblin Card", TypeLine: "Creature — Goblin", Owner: me.ID, Controller: me.ID}
	elf := game.Card{InstanceID: uuid.New(), Name: "Elf Card", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(goblin)
	me.Graveyard.PushTop(elf)
	importAndCast(t, g, grubRow(), me)
	monSettle(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("entering asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: elf.InstanceID}}); err == nil {
		t.Fatal("a non-Goblin card was a legal target")
	}
	answerTriggerTargets(t, g, me.ID, goblin.InstanceID)
	monSettle(t, g)
	if _, ok := cardInZone(me.Hand, goblin.InstanceID); !ok {
		t.Error("the Goblin card did not return to your hand")
	}
	if _, ok := cardInZone(me.Graveyard, elf.InstanceID); !ok {
		t.Error("the Elf card left the graveyard")
	}
}

// Attacking with Notorious Auntie blights a creature you control and makes
// a tapped and attacking copy of it, which is sacrificed at the end step.
func TestGrubNotoriousAuntieCopiesTheBlightedCreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	ox := pushVanillaCreature(g, me.ID, "Ox", 3, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Their Ox", 3, 3)
	grub := grubTransformed(t, g, me)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(grub, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	monSettle(t, g)
	p := blightPrompt(g, me.ID)
	if p == nil {
		t.Fatal("the attack asked no blight")
	}
	if p.ChooseMin != 0 || hasID(p.ChooseCards, theirs) || !hasID(p.ChooseCards, ox) || !hasID(p.ChooseCards, grub) {
		t.Fatalf("blight offered %v (min %d)", p.ChooseCards, p.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, ox)
	passAndSettleAttackSplit(t, g, me.ID)

	if got := minusCounters(g, ox); got != 1 {
		t.Errorf("-1/-1 counters on the Ox: %d, want 1", got)
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == me.ID && c.IsToken() && c.Name == "Ox" {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no token copy of the Ox")
	}
	if !token.Tapped || token.AttackingTarget != opp.ID {
		t.Errorf("the copy should be tapped and attacking: tapped=%v attacking=%s", token.Tapped, token.AttackingTarget)
	}
	if token.Counters[game.CounterMinusOne] != 0 {
		t.Error("the copy copied the -1/-1 counter")
	}
	id := token.InstanceID
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, id) != nil {
		t.Error("the copy was not sacrificed at the beginning of the end step")
	}
	if findBattlefieldCardForTest(g, ox) == nil {
		t.Error("the Ox itself was sacrificed")
	}
}

// Declining the blight makes no copy and puts no counter.
func TestGrubNotoriousAuntieDeclinedMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	ox := pushVanillaCreature(g, me.ID, "Ox", 3, 3)
	grub := grubTransformed(t, g, me)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(grub, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	monSettle(t, g)
	answerChooseCards(t, g, me.ID)
	passAndSettleAttackSplit(t, g, me.ID)
	if n, _ := countTokensNamed(g, me.ID, "Ox"); n != 0 {
		t.Errorf("%d copies without a blight", n)
	}
	if got := minusCounters(g, ox); got != 0 {
		t.Errorf("-1/-1 counters on the Ox: %d, want 0", got)
	}
}

// Paying {B} turns Notorious Auntie back, which returns a Goblin card.
func TestGrubTransformsBackAndReturnsAGoblin(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := grubTransformed(t, g, me)
	goblin := game.Card{InstanceID: uuid.New(), Name: "Goblin Card", TypeLine: "Creature — Goblin", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(goblin)
	advanceToStepOf(t, g, seat, game.StepUpkeep)
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	if f := iwFace(g, id); f != 0 {
		t.Fatalf("Grub is on face %d after paying {B}, want the front face", f)
	}
	answerTriggerTargets(t, g, me.ID, goblin.InstanceID)
	monSettle(t, g)
	if _, ok := cardInZone(me.Hand, goblin.InstanceID); !ok {
		t.Error("transforming into Storied Matriarch returned no Goblin card")
	}
}
