package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b13_prepare_test.go — Stensian Sanguinist // Exsanguinate and Striding
// Shotcaller // Run the Play (#4834).

func socB13CastSanguinist(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB13SanguinistOracle, nil, "Stensian Sanguinist", "Creature — Vampire Cleric", "{1}{B}", 2, 2,
		"Exsanguinate", "Sorcery", "{X}{B}{B}")
	castFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

func socB13CastShotcaller(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB13ShotcallerOracle, []string{"reach"}, "Striding Shotcaller", "Creature — Troll Druid", "{G}{U}", 0, 4,
		"Run the Play", "Sorcery", "{X}{G}{U}")
	castFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

// socB13AttackAndTarget declares `attacker` against `defender`, answers the
// "whenever you attack" trigger's target pick, resolves it and walks into the
// combat damage step.
func socB13AttackAndTarget(t *testing.T, g *game.Game, defender, attacker, target uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	answerPickTargetIfOpen(t, g, target)
	passPriorityAroundTable(t, g)
}

func TestSocB13SanguinistGivesDeathtouchAndPreparesWhenThatCreatureHitsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sang := socB13CastSanguinist(t, g, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if g.IsPreparedForEffect(sang) {
		t.Fatal("prepared before anything happened")
	}
	socB13AttackAndTarget(t, g, opp.ID, bear, bear)
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Fatal("the target did not gain deathtouch")
	}
	if g.IsPreparedForEffect(sang) {
		t.Fatal("prepared before the creature dealt damage")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(sang) {
		t.Fatal("the Sanguinist is not prepared after the targeted creature hit a player")
	}
	if _, ok := prepareCopyOf(g, "Exsanguinate"); !ok {
		t.Fatal("no Exsanguinate copy in exile")
	}
}

func TestSocB13SanguinistIgnoresAnotherCreaturesDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sang := socB13CastSanguinist(t, g, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 2)
	socB13AttackAndTarget(t, g, opp.ID, bear, bystander)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life >= 40 {
		t.Fatalf("the bear did not connect (life %d)", opp.Life)
	}
	if g.IsPreparedForEffect(sang) {
		t.Error("prepared by a creature that was not the target")
	}
}

func TestSocB13SanguinistDoesNotTriggerOnAnOpponentsAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sang := socB13CastSanguinist(t, g, me)
	endTurn(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if n := triggersOnStackFrom(g, sang); n != 0 {
		t.Errorf("%d Sanguinist triggers on an opponent's turn", n)
	}
}

func TestSocB13ExsanguinateDrainsEachOpponentForX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sang := socB13CastSanguinist(t, g, me)
	g.WithWriteLock(func() {
		if _, err := g.BecomePreparedForEffect(sang); err != nil {
			t.Fatalf("prepare: %v", err)
		}
	})
	before := map[uuid.UUID]int{}
	opponents := 0
	for _, s := range g.Seats {
		before[s.ID] = s.Life
		if s.ID != me.ID && !s.Eliminated {
			opponents++
		}
	}
	cp, ok := prepareCopyOf(g, "Exsanguinate")
	if !ok {
		t.Fatal("no Exsanguinate copy in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), XValue: 3}); err != nil {
		t.Fatalf("cast Exsanguinate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := before[s.ID] - 3
		if s.ID == me.ID {
			want = before[s.ID] + 3*opponents
		}
		if s.Eliminated {
			continue
		}
		if s.Life != want {
			t.Errorf("player %s life %d, want %d", s.ID, s.Life, want)
		}
	}
	if g.IsPreparedForEffect(sang) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSocB13ShotcallerPreparesOncePerCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shot := socB13CastShotcaller(t, g, me)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, a, b)
	if n := triggersOnStackFrom(g, shot); n != 1 {
		t.Fatalf("%d triggers for two creatures hitting the same player, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(shot) {
		t.Fatal("not prepared after combat damage to a player")
	}
}

func TestSocB13ShotcallerDoesNotPrepareWhenNothingHitsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shot := socB13CastShotcaller(t, g, me)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEndCombat)
	if g.IsPreparedForEffect(shot) || triggersOnStackFrom(g, shot) != 0 {
		t.Error("prepared with no combat damage dealt")
	}
}

func TestSocB13RunThePlayCountersFliesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shot := socB13CastShotcaller(t, g, me)
	g.WithWriteLock(func() {
		if _, err := g.BecomePreparedForEffect(shot); err != nil {
			t.Fatalf("prepare: %v", err)
		}
	})
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	cp, ok := prepareCopyOf(g, "Run the Play")
	if !ok {
		t.Fatal("no Run the Play copy in exile")
	}
	targets := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}
	// X = 2 takes at most two targets.
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), XValue: 2, Strict: true,
		Targets: targets(a, b, c)}); err == nil {
		t.Fatal("three targets were accepted for X = 2")
	}
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), XValue: 2,
		Targets: targets(a, b)}); err != nil {
		t.Fatalf("cast Run the Play: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("a target has %d +1/+1 counters, want 1", got)
		}
		if !hasEffectiveKeyword(t, g, id, "flying") {
			t.Error("a target has no flying")
		}
	}
	if got := counterCount(g, c, game.CounterPlusOne); got != 0 || hasEffectiveKeyword(t, g, c, "flying") {
		t.Error("the untargeted creature was changed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if g.IsPreparedForEffect(shot) {
		t.Error("still prepared after casting the copy")
	}
	// Flying ends with the turn; the counters stay.
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, a, "flying") {
		t.Error("flying lasted past the turn")
	}
	if counterCount(g, a, game.CounterPlusOne) != 1 {
		t.Error("the counter did not stay")
	}
}

func TestSocB13RunThePlayWithNoTargetsStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shot := socB13CastShotcaller(t, g, me)
	g.WithWriteLock(func() {
		if _, err := g.BecomePreparedForEffect(shot); err != nil {
			t.Fatalf("prepare: %v", err)
		}
	})
	cp, _ := prepareCopyOf(g, "Run the Play")
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), XValue: 0}); err != nil {
		t.Fatalf("cast Run the Play for X = 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}
