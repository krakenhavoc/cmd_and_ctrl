package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b13_auras_test.go — Shielded by Faith and Songbirds' Blessing (#4834).

// socB13Shielded puts a Shielded by Faith on the battlefield attached to host.
func socB13Shielded(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Shielded by Faith", TypeLine: "Enchantment — Aura",
		ManaCost: "{1}{W}{W}", Colors: []string{"W"}, OracleID: socB13ShieldedOracle, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	return aura
}

func socB13AttachedTo(g *game.Game, aura uuid.UUID) uuid.UUID {
	c, _ := battlefieldCard(g, aura)
	return c.AttachedTo.ID
}

func TestSocB13ShieldedByFaithGrantsIndestructibleAndMovesToANewCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceTo(t, g, game.StepPrecombatMain)
		host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
		aura := socB13Shielded(t, g, me, host)
		if !hasEffectiveKeyword(t, g, host, "indestructible") {
			t.Fatal("the enchanted creature is not indestructible")
		}
		newcomer := castFromHandAndResolveCreature(t, g, me, "Newcomer")
		answerLatestTriggerPrompt(t, g, me.ID, yes)
		passPriorityAroundTable(t, g)
		if yes {
			if got := socB13AttachedTo(g, aura); got != newcomer {
				t.Fatalf("attached to %s, want the creature that entered", got)
			}
			if hasEffectiveKeyword(t, g, host, "indestructible") || !hasEffectiveKeyword(t, g, newcomer, "indestructible") {
				t.Error("indestructible did not move with the Aura")
			}
		} else if got := socB13AttachedTo(g, aura); got != host {
			t.Fatalf("attached to %s after saying no, want the original host", got)
		}
	}
}

func TestSocB13ShieldedByFaithTriggersForAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := socB13Shielded(t, g, me, host)
	for g.Turn.ActiveSeat != 1 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	newcomer := castFromHandAndResolveCreature(t, g, opp, "Their Newcomer")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := socB13AttachedTo(g, aura); got != newcomer {
		t.Fatalf("attached to %s, want the opponent's creature that entered", got)
	}
	// It is still my Aura.
	if c, _ := battlefieldCard(g, aura); c.Controller != me.ID {
		t.Error("the Aura changed controller")
	}
}

func TestSocB13ShieldedByFaithStaysPutOnACreatureWithProtection(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := socB13Shielded(t, g, me, host)
	protected := game.Card{InstanceID: uuid.New(), Name: "Protected", TypeLine: "Creature — Knight", Power: 2, Toughness: 2,
		Keywords: []string{"protection from white"}, Owner: me.ID, Controller: me.ID}
	castFromHandAndResolve(t, g, me, protected)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := socB13AttachedTo(g, aura); got != host {
		t.Errorf("the Aura moved onto a creature with protection from white (attached to %s)", got)
	}
}

func TestSocB13ShieldedByFaithIgnoresNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := socB13Shielded(t, g, me, host)
	castFromHandAndResolve(t, g, me, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if n := triggersOnStackFrom(g, aura); n != 0 {
		t.Errorf("%d triggers for an artifact entering", n)
	}
}

// castFromHandAndResolveCreature casts a vanilla 1/1 creature of that name
// for `p` and returns it once it is on the battlefield.
func castFromHandAndResolveCreature(t *testing.T, g *game.Game, p *game.Player, name string) uuid.UUID {
	t.Helper()
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID}
	castFromHandAndResolve(t, g, p, c)
	return c.InstanceID
}

// --- Songbirds' Blessing --------------------------------------------------

// socB13Library replaces `p`'s library with the named cards, bottom first, so
// the last name is the top. "Aura" is a Shielded by Faith.
func socB13Library(p *game.Player, names ...string) (aura uuid.UUID) {
	p.Library.Cards = nil
	for _, n := range names {
		c := game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID}
		if n == "Aura" {
			c = game.Card{InstanceID: uuid.New(), Name: "Shielded by Faith", TypeLine: "Enchantment — Aura", ManaCost: "{1}{W}{W}",
				Colors: []string{"W"}, OracleID: socB13ShieldedOracle, Owner: p.ID, Controller: p.ID}
			aura = c.InstanceID
		}
		p.Library.PushTop(c)
	}
	return aura
}

// socB13Songbirds attacks with a Songbirds'-Blessed bear and resolves the
// trigger up to the point it asks about the Aura.
func socB13Songbirds(t *testing.T, g *game.Game, me *game.Player, bearKeywords ...string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Keywords: bearKeywords, Owner: me.ID, Controller: me.ID})
	songbirds := pushCatalogPermanent(g, me.ID, "Songbirds' Blessing", "Enchantment — Aura", socB13SongbirdsOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(songbirds, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	return bear
}

func TestSocB13SongbirdsBlessingPutsTheAuraOntoTheBattlefieldAndTheRestOnTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := socB13Library(me, "Z1", "Z2", "Aura", "Top1", "Top2")
	bear := socB13Songbirds(t, g, me)
	answerChooseCards(t, g, me.ID, aura)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, aura)
	if !ok {
		t.Fatal("the Aura card is not on the battlefield")
	}
	if c.AttachedTo.ID != bear {
		t.Errorf("the Aura is attached to %s, want the only creature", c.AttachedTo.ID)
	}
	if me.Library.Size() != 4 {
		t.Fatalf("library %d, want 4", me.Library.Size())
	}
	if top, err := me.Library.Top(); err != nil || top.Name != "Z2" {
		t.Errorf("top of library %v, want Z2 (the cards above the Aura went to the bottom)", top.Name)
	}
	bottom := map[string]bool{}
	for _, c := range me.Library.Cards[:2] {
		bottom[c.Name] = true
	}
	if !bottom["Top1"] || !bottom["Top2"] {
		t.Errorf("the bottom of the library is %v, want the two revealed cards", bottom)
	}
	if me.Hand.Contains(aura) {
		t.Error("the Aura is also in the hand")
	}
}

func TestSocB13SongbirdsBlessingPutsTheAuraInHandIfYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := socB13Library(me, "Z1", "Aura", "Top1")
	handBefore := me.Hand.Size()
	socB13Songbirds(t, g, me)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) || me.Hand.Size() != handBefore+1 {
		t.Fatalf("the Aura is not in the hand (hand %d, was %d)", me.Hand.Size(), handBefore)
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura entered after declining")
	}
	if me.Library.Size() != 2 {
		t.Fatalf("library %d, want 2", me.Library.Size())
	}
	if top, _ := me.Library.Top(); top.Name != "Z1" {
		t.Errorf("top of library %q, want Z1", top.Name)
	}
	if bottom, _ := me.Library.Bottom(); bottom.Name != "Top1" {
		t.Errorf("bottom of library %q, want the revealed Top1", bottom.Name)
	}
}

func TestSocB13SongbirdsBlessingPutsTheAuraInHandWhenNothingCanBeEnchanted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := socB13Library(me, "Z1", "Aura", "Top1")
	// The only creature has protection from white: the white Aura has no legal
	// host (CR 303.4g), so it can't be put onto the battlefield and goes to
	// the hand.
	socB13Songbirds(t, g, me, "protection from white")
	if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
		answerChooseCards(t, g, me.ID, aura)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Fatal("the Aura entered with no legal host")
	}
	if !me.Hand.Contains(aura) {
		t.Fatal("the Aura that could not enter is not in the hand")
	}
	if bottom, _ := me.Library.Bottom(); bottom.Name != "Top1" {
		t.Errorf("bottom of library %q, want the revealed Top1", bottom.Name)
	}
}

func TestSocB13SongbirdsBlessingWithNoAuraPutsEverythingOnTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	socB13Library(me, "Z1", "Z2", "Z3")
	handBefore := me.Hand.Size()
	socB13Songbirds(t, g, me)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 3 || me.Hand.Size() != handBefore {
		t.Errorf("library %d hand %d, want 3 and %d", me.Library.Size(), me.Hand.Size(), handBefore)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a prompt was opened with no Aura to find")
	}
}
