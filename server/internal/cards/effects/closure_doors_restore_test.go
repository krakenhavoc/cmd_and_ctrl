package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_doors_restore_test.go — #4253 slice 6: the put and return doors
// and token creation take their "then" as a registered tail, not a closure,
// so a table waiting on the entry question of a card a spell just put onto
// the battlefield, or on the ordering prompt of a token creation, is a
// restore point. Each test takes a STRICT restore (the boot path) while the
// prompt is open, answers it on the restored table, and checks the card
// still finishes: the tail ran off the disk, not off a closure.
//
// The entry question is a shockland's "pay 2 life?" or a Clone's copy
// pick; the ordering prompt is a Doubling Season beside a Parallel Lives.

// libTop pushes a card on top of a player's library and returns its ID.
func libTop(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

// tokenOrderBoard puts a Doubling Season and a Parallel Lives under `me`:
// every token creation then asks for the CR 616 order.
func tokenOrderBoard(g *game.Game, me uuid.UUID) (ds, pl uuid.UUID) {
	ds = pushTokenReplacementCard(g, doublingSeasonOracle, "Doubling Season", "Enchantment", me)
	pl = pushTokenReplacementCard(g, parallelLivesOracle, "Parallel Lives", "Enchantment", me)
	return ds, pl
}

// --- library puts: the shared rest-of-the-pile tails -------------------------

// Open the Way: the revealed land is a shockland, whose "pay 2 life?" is a
// restore point, and the restored answer puts the land onto the battlefield
// and the rest on the bottom of the library (put-from-library/rest-on-bottom-
// in-random-order, shared with fifteen cards).
func TestOpenTheWayIsARestorePointAtThePutLandsQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	libTop(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	crypt := libTop(me, basicShockland())
	bear := libTop(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	if _, err := castOpenTheWay(t, g, 1); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the revealed shockland did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[g.Turn.ActiveSeat]
	if plBottomIDs(me, 1)[bear] {
		t.Fatal("the restored table put the rest on the bottom before the answer")
	}
	answerEntryPayLife(t, g, me.ID, false)
	card, ok := battlefieldCard(g, crypt)
	if !ok || !card.Tapped {
		t.Fatalf("the land is on the battlefield %v, tapped %v, want tapped", ok, card.Tapped)
	}
	if !plBottomIDs(me, 1)[bear] {
		t.Error("the restored tail did not put the revealed Bear on the bottom of the library")
	}
}

// Genesis Wave: the chosen permanents enter together, a shockland among them
// asks, and the restored answer puts the rest into the graveyard
// (put-from-library/rest-into-graveyard, shared).
func TestGenesisWaveIsARestorePointAtAChosenPermanentsQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := plTop(me, "Big", "Creature — Giant", "{5}")
	bear := plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	fountain := libTop(me, game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island", OracleID: hallowedFountainOracle})

	b12PlayFromHand(t, g, "Genesis Wave", "Sorcery", plGenesisWaveOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, fountain, bear)
	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the Fountain did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[0]
	if me.Graveyard.Contains(big) {
		t.Fatal("the restored table binned the rest before the answer")
	}
	answerEntryPayLife(t, g, me.ID, true)
	if !g.Battlefield.Contains(fountain) || !g.Battlefield.Contains(bear) {
		t.Fatal("the chosen permanents did not enter after the restored answer")
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the restored tail did not put the revealed card that was not chosen into the graveyard")
	}
}

// --- hand puts ------------------------------------------------------------------

// Surprise Deployment puts a Clone onto the battlefield: the copy pick is a
// restore point, and the restored answer schedules the creature's return to
// hand at the next end step (surprise-deployment/return-at-end-step, told
// the creature that entered).
func TestSurpriseDeploymentIsARestorePointAtTheCloneQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := rooncHand(me, game.Card{Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone, Colors: []string{"G"}})

	advanceTo(t, g, game.StepBeginCombat)
	if _, err := rooncCastNow(g, rooncDeploymentCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, clone)
	if copyPrompt(g) == nil {
		t.Fatal("the put Clone did not ask what to copy")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[0]
	if len(g.DelayedTriggers) != 0 {
		t.Fatal("the restored table scheduled the return before the answer")
	}
	answerCloneCopy(t, g, bears)
	if !g.Battlefield.Contains(clone) {
		t.Fatal("the Clone did not enter after the restored answer")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after the answer, want the one return", len(g.DelayedTriggers))
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(clone) || !me.Hand.Contains(clone) {
		t.Error("the restored tail's delayed trigger did not return the creature to hand")
	}
}

// --- exile and graveyard returns ------------------------------------------------

// Essence Flux blinks a Clone: the copy pick is a restore point, and the
// restored answer, a copy of a Spirit, gets the +1/+1 counter (the words
// "Spirit" are read from the spell's text through the asking item that
// rides the tail).
func TestEssenceFluxIsARestorePointAtTheReturningClonesQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spirit := seedCopyableCreature(g, me.ID, "Wisp", "Creature — Spirit", 1, 1)
	clone := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		Power: 1, Toughness: 1, PrintedPTKnown: true, Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Essence Flux", "Instant", essenceFluxOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: clone}})
	passPriorityAroundTable(t, g)
	if copyPrompt(g) == nil {
		t.Fatal("the returning Clone did not ask what to copy")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveEntryTail)
	answerCloneCopy(t, g, spirit)
	var returned *game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wisp" && c.InstanceID != spirit {
			cc := c
			returned = &cc
		}
	}
	if returned == nil {
		t.Fatal("the Clone did not return as a copy of the Spirit")
	}
	if returned.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the returned Spirit copy has %d +1/+1 counters, want the one Essence Flux's tail adds", returned.Counters[game.CounterPlusOne])
	}
}

// It Doesn't Add Up returns a Clone from the graveyard: the copy pick is a
// restore point, and the restored answer suspects the permanent that results
// (then/suspect-each, shared with Presumed Dead and Case of the Stashed
// Skeleton).
func TestItDoesntAddUpIsARestorePointAtTheReturningClonesQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		ManaCost: "{3}{U}", Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "It Doesn't Add Up", "Instant", sbItDoesntAddUp, dgCardRef(clone))
	passPriorityAroundTable(t, g)
	if copyPrompt(g) == nil {
		t.Fatal("the returning Clone did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveEntryTail)
	answerCloneCopy(t, g, bear)
	c, ok := battlefieldCard(g, clone)
	if !ok || c.Name != "Grizzly Bears" {
		t.Fatalf("the Clone returned %v as %q, want a copy of Grizzly Bears", ok, c.Name)
	}
	if !sbSuspected(g, clone) {
		t.Error("the restored tail did not suspect the returned Clone")
	}
}

// Makeshift Mannequin returns a Clone with a mannequin counter: the copy
// pick is a restore point (the counters door), and the restored answer
// grants the sacrifice trigger while it has the counter
// (makeshift-mannequin/grant-while-counter).
func TestMakeshiftMannequinIsARestorePointAtTheReturningClonesQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Makeshift Mannequin", "Instant", rtMannequinOracle, ltCardTarget(clone))
	passPriorityAroundTable(t, g)
	if copyPrompt(g) == nil {
		t.Fatal("the returning Clone did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveEntryTail)
	me = g.Seats[0]
	answerCloneCopy(t, g, bears)
	c := layeredCard(t, g, clone)
	if c.Counters["mannequin"] != 1 {
		t.Fatalf("the Clone returned with counters %v, want a mannequin counter", c.Counters)
	}
	endo := pushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", rtEndoskeletonOracle, false)
	rtActivate(t, g, me.ID, endo, 0, ltCardTarget(clone), "C", "C")
	passPriorityAroundTable(t, g)
	if onBattlefield(g, clone) {
		t.Error("the returned Clone was targeted and is still on the battlefield: the restored tail never granted the sacrifice")
	}
}

// --- one entry of several: the together doors -----------------------------------

// Zombie Apocalypse returns a Zombie Clone: the copy pick is a restore
// point, and the restored answer, a copy of a Human, is then destroyed with
// the other Humans (zombie-apocalypse/destroy-humans, run against the
// asking item that rides the tail).
func TestZombieApocalypseIsARestorePointAtTheReturningCloneQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	human := seedCopyableCreature(g, me.ID, "My Human", "Creature — Human Soldier", 2, 2)
	clone := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Zombie Shapeshifter", OracleID: oracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	theirHuman := pushTribalCreature(g, opp.ID, "Their Human", "Creature — Human Wizard", 1, 1)
	theirElf := pushTribalCreature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)

	castCatalogSpell(t, g, "Zombie Apocalypse", "Sorcery", b29ZombieApocalypseOracle, nil)
	passPriorityAroundTable(t, g)
	if copyPrompt(g) == nil {
		t.Fatal("the returning Zombie Clone did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if !g.Battlefield.Contains(human) || !g.Battlefield.Contains(theirHuman) {
		t.Fatal("the restored table destroyed the Humans before the answer")
	}
	answerCloneCopy(t, g, human)
	if g.Battlefield.Contains(human) || g.Battlefield.Contains(theirHuman) || g.Battlefield.Contains(clone) {
		t.Error("the restored tail did not destroy every Human, the Clone's copy of one included")
	}
	if !g.Battlefield.Contains(theirElf) {
		t.Error("a non-Human was destroyed")
	}
}

// Chorale of the Void puts a Clone from a graveyard onto the battlefield
// attacking (putAttackingYourChoice, the put-attacking door): the copy pick
// is a restore point, and the restored answer leaves the copy tapped and
// attacking the defending player.
func TestChoraleOfTheVoidIsARestorePointAtTheRaisedClonesQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	aura := pushCatalogPermanent(g, me.ID, "Chorale of the Void", "Enchantment — Aura", choraleOfTheVoidOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	clone := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		Owner: opp.ID, Controller: opp.ID,
	})
	bears := seedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)

	declareAttack(t, g, opp.ID, host)
	pickCard(t, g, me.ID, clone)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	if copyPrompt(g) == nil {
		t.Fatal("the raised Clone did not ask what to copy")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if g.Battlefield.Contains(clone) {
		t.Fatal("the restored table put the Clone onto the battlefield before the answer")
	}
	answerCloneCopy(t, g, bears)
	var raised game.Card
	var found bool
	g.ReadSnapshot(func() {
		if g.Battlefield.Contains(clone) {
			raised, found = g.LookupCardForEffect(clone)
		}
	})
	if !found {
		t.Fatal("the Clone did not enter after the restored answer")
	}
	if raised.Name != "Hill Giant" || raised.Controller != me.ID || !raised.Tapped || raised.AttackingTarget == uuid.Nil {
		t.Errorf("raised %q controller=%v tapped=%v attacking=%v, want a tapped, attacking Hill Giant of mine",
			raised.Name, raised.Controller, raised.Tapped, raised.AttackingTarget)
	}
}
