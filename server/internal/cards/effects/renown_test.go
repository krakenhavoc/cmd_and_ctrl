package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// renown_test.go — #2049, CR 702.112 (ADR 0071 amendment 2026-10-09).
// The keyword's rules are in game/renown_test.go; this file plays each
// renown card through a real combat, and the cards that read the
// designation both ways.

const (
	enshroudingMistOracle    = "e513c05c-14c0-4d2e-a5b3-e4f29cb4c1a2"
	goblinGloryChaserOracle  = "cad1cd10-4810-495e-9305-0ae812596f6a"
	valeronWardensOracle     = "d1202573-2105-4d4d-ae7d-7164974b6d89"
	kytheonsIrregularsOracle = "ccfddb1a-1b7f-44f9-89c4-1907e27e1c4a"
	undercityTrollOracle     = "b7851b17-faa2-4767-9716-86997b882690"
	outlandColossusOracle    = "5b034e04-70e5-4783-8ac8-04d8e5a9f961"
	consulsLieutenantOracle  = "a38b672f-4739-4b7e-8958-2a455510656e"
	scabClanBerserkerOracle  = "d4f659ba-c277-47a2-8b95-ce81ddac34fb"
)

// pushRenownCard puts a catalog card (or, with oracle "", a plain
// creature carrying `keywords`) on the battlefield, ready to attack.
func pushRenownCard(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Human Soldier", OracleID: oracle,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: owner, Controller: owner,
	})
}

// makeRenowned gives the permanent the designation the way the engine
// does: its counters and the event that switches the gate on.
func makeRenowned(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, id)
		c.Renowned = true
		g.EmitEvent(game.Event{Kind: game.EventBecameRenowned, CardID: id, Source: id, Target: id, Actor: c.Controller, Amount: 1})
	})
	passPriorityAroundTable(t, g)
}

func renownedNow(g *game.Game, id uuid.UUID) bool {
	var r bool
	g.ReadSnapshot(func() { r = g.IsRenowned(id) })
	return r
}

// A keyword-only renown creature (Topan Freeblade) needs no card file:
// it hits a player, gets its counter and becomes renowned.
func TestRenownKeywordOnlyCreatureBecomesRenowned(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	freeblade := pushRenownCard(g, me.ID, "Topan Freeblade", "", 2, 2, "vigilance", "renown 1")
	attackWith(t, g, opp.ID, freeblade)
	passPriorityAroundTable(t, g)
	if !renownedNow(g, freeblade) || plusOneCounters(g, freeblade) != 1 {
		t.Errorf("renowned %v, %d counters; want renowned with 1", renownedNow(g, freeblade), plusOneCounters(g, freeblade))
	}
}

// Goblin Glory Chaser: no menace until it is renowned, menace after.
func TestGoblinGloryChaserHasMenaceOnlyOnceRenowned(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chaser := pushRenownCard(g, me.ID, "Goblin Glory Chaser", goblinGloryChaserOracle, 1, 1)
	if game.HasKeyword(mustBattlefieldCard(t, g, chaser), "menace") {
		t.Fatal("the Chaser has menace before it is renowned")
	}
	attackWith(t, g, opp.ID, chaser)
	passPriorityAroundTable(t, g)
	if !renownedNow(g, chaser) || plusOneCounters(g, chaser) != 1 {
		t.Fatalf("renowned %v, %d counters; want renowned with 1", renownedNow(g, chaser), plusOneCounters(g, chaser))
	}
	if !game.HasKeyword(mustBattlefieldCard(t, g, chaser), "menace") {
		t.Error("a renowned Chaser has no menace")
	}
}

// Valeron Wardens: a card for each creature you control that becomes
// renowned, its own renown included — and none for an opponent's.
func TestValeronWardensDrawsForEachCreatureThatBecomesRenowned(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wardens := pushRenownCard(g, me.ID, "Valeron Wardens", valeronWardensOracle, 1, 3)
	freeblade := pushRenownCard(g, me.ID, "Topan Freeblade", "", 2, 2, "renown 1")
	theirs := pushRenownCard(g, opp.ID, "Their Freeblade", "", 2, 2, "renown 1")
	hand := len(me.Hand.Cards)
	attackWith(t, g, opp.ID, wardens, freeblade)
	// The two renown triggers are ordered (CR 603.3b): they are from
	// two sources, and each one's draw lands between them.
	for range 4 {
		passPriorityAroundTable(t, g)
		if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
			break
		}
	}
	if !renownedNow(g, wardens) || plusOneCounters(g, wardens) != 2 {
		t.Fatalf("Wardens renowned %v, %d counters; want renowned with 2", renownedNow(g, wardens), plusOneCounters(g, wardens))
	}
	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (the Wardens and the Freeblade became renowned)", got)
	}
	hand = len(me.Hand.Cards)
	makeRenowned(t, g, theirs)
	if got := len(me.Hand.Cards) - hand; got != 0 {
		t.Errorf("an opponent's creature becoming renowned drew %d cards, want 0", got)
	}
}

// Enshrouding Mist: +1/+1 and a shield either way; it untaps the
// creature only if it is renowned.
func TestEnshroudingMistUntapsOnlyARenownedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := pushRenownCard(g, me.ID, "Plain", "", 2, 2)
	famous := pushRenownCard(g, me.ID, "Famous", "", 2, 2, "renown 1")
	makeRenowned(t, g, famous)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	for _, id := range []uuid.UUID{plain, famous} {
		findBattlefieldCardForTest(g, id).Tapped = true
	}

	castCatalogSpell(t, g, "Enshrouding Mist", "Instant", enshroudingMistOracle, []game.TargetRef{pr7aCard(plain)})
	passPriorityAroundTable(t, g)
	if !findBattlefieldCardForTest(g, plain).Tapped {
		t.Error("Enshrouding Mist untapped a creature that is not renowned")
	}
	if p := mustBattlefieldCard(t, g, plain).CurrentPower(); p != 3 {
		t.Errorf("the target's power is %d, want 3", p)
	}
	pr6Damage(t, g, src, plain, 3)
	if got := pr6Marked(g, plain); got != 0 {
		t.Errorf("the target has %d damage, want 0 — the damage is prevented", got)
	}

	castCatalogSpell(t, g, "Enshrouding Mist", "Instant", enshroudingMistOracle, []game.TargetRef{pr7aCard(famous)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, famous).Tapped {
		t.Error("Enshrouding Mist did not untap a renowned creature")
	}
}

// Consul's Lieutenant: its attack trigger does nothing until it is
// renowned (it does not even trigger), then pumps the OTHER attackers.
func TestConsulsLieutenantPumpsOtherAttackersOnlyWhenRenowned(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lieutenant := pushRenownCard(g, me.ID, "Consul's Lieutenant", consulsLieutenantOracle, 2, 1)
	bear := pushRenownCard(g, me.ID, "Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, lieutenant, bear)
	if n := triggersOnStackFrom(g, lieutenant); n != 0 {
		t.Fatalf("an unrenowned Lieutenant's attack put %d triggers on the stack, want 0", n)
	}

	g2 := newCatalogGame(t)
	me, opp = g2.Seats[0], g2.Seats[1]
	lieutenant = pushRenownCard(g2, me.ID, "Consul's Lieutenant", consulsLieutenantOracle, 2, 1)
	bear = pushRenownCard(g2, me.ID, "Bear", "", 2, 2)
	makeRenowned(t, g2, lieutenant)
	declareAttack(t, g2, opp.ID, lieutenant, bear)
	passPriorityAroundTable(t, g2)
	if p := mustBattlefieldCard(t, g2, bear).CurrentPower(); p != 3 {
		t.Errorf("the other attacker's power is %d, want 3", p)
	}
	if p := mustBattlefieldCard(t, g2, lieutenant).CurrentPower(); p != 2 {
		t.Errorf("the Lieutenant pumped itself: power %d, want 2", p)
	}
}

// CR 603.4 on Consul's Lieutenant: a trigger whose source stopped being
// renowned — here, by leaving and coming back as a new object — is
// read from the object that triggered it, which was renowned.
func TestConsulsLieutenantRechecksAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lieutenant := pushRenownCard(g, me.ID, "Consul's Lieutenant", consulsLieutenantOracle, 2, 1)
	bear := pushRenownCard(g, me.ID, "Bear", "", 2, 2)
	makeRenowned(t, g, lieutenant)
	declareAttack(t, g, opp.ID, lieutenant, bear)
	if n := triggersOnStackFrom(g, lieutenant); n != 1 {
		t.Fatalf("%d Lieutenant triggers on the stack, want 1", n)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(lieutenant); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if p := mustBattlefieldCard(t, g, bear).CurrentPower(); p != 3 {
		t.Errorf("the other attacker's power is %d, want 3 — the Lieutenant was renowned as it last existed", p)
	}
}

// Scab-Clan Berserker: 2 damage to an opponent who casts a noncreature
// spell, but only once it is renowned.
func TestScabClanBerserkerPunishesNoncreatureSpellsOnlyWhenRenowned(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Berserker is the opponent's; the active player casts.
	berserker := pushRenownCard(g, opp.ID, "Scab-Clan Berserker", scabClanBerserkerOracle, 2, 2)
	target := pushRenownCard(g, me.ID, "Bear", "", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Enshrouding Mist", "Instant", enshroudingMistOracle, []game.TargetRef{pr7aCard(target)})
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("an unrenowned Berserker dealt %d damage", life-me.Life)
	}

	makeRenowned(t, g, berserker)
	castCatalogSpell(t, g, "Enshrouding Mist", "Instant", enshroudingMistOracle, []game.TargetRef{pr7aCard(target)})
	passPriorityAroundTable(t, g)
	if life-me.Life != 2 {
		t.Errorf("a renowned Berserker dealt %d damage, want 2", life-me.Life)
	}
}

// Outland Colossus: renown 6 on a hit, and never more than one blocker.
func TestOutlandColossusRenownSixAndOneBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := pushRenownCard(g, me.ID, "Outland Colossus", outlandColossusOracle, 6, 6)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	brAttack(t, g, colossus)
	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: colossus}, {Blocker: b, Attacker: colossus},
	}), game.BlockReasonTooManyBlockers)

	g2 := newCatalogGame(t)
	me, opp = g2.Seats[0], g2.Seats[1]
	colossus = pushRenownCard(g2, me.ID, "Outland Colossus", outlandColossusOracle, 6, 6)
	attackWith(t, g2, opp.ID, colossus)
	passPriorityAroundTable(t, g2)
	if !renownedNow(g2, colossus) || plusOneCounters(g2, colossus) != 6 {
		t.Errorf("renowned %v, %d counters; want renowned with 6", renownedNow(g2, colossus), plusOneCounters(g2, colossus))
	}
}

// Kytheon's Irregulars: {W}{W} taps target creature.
func TestKytheonsIrregularsTapsTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	irregulars := pushRenownCard(g, me.ID, "Kytheon's Irregulars", kytheonsIrregularsOracle, 4, 3)
	victim := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	pr7Activate(t, g, me.ID, irregulars, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{pr7aCard(victim)}})
	if !findBattlefieldCardForTest(g, victim).Tapped {
		t.Error("the target is not tapped")
	}
	if findBattlefieldCardForTest(g, irregulars).Tapped {
		t.Error("the Irregulars tapped themselves; the cost has no {T}")
	}
}

// Undercity Troll: {2}{G} regenerates it.
func TestUndercityTrollRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	troll := pushRenownCard(g, me.ID, "Undercity Troll", undercityTrollOracle, 2, 2)
	pr7Activate(t, g, me.ID, troll, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(troll); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if findBattlefieldCardForTest(g, troll) == nil {
		t.Error("the regenerated Troll was destroyed")
	}
}
