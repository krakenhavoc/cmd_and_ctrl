package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_batch_a_test.go — ADR 0107 §4 (#1855): the disturb cards
// whose faces are keywords, Auras that pump or grant, and simple
// triggers. disturb_test.go has the helpers and the seam's own tests.

// backFaceOnBattlefield imports a row and puts it onto the battlefield
// back face up, as a disturbed permanent that has been there a while
// (no summoning sickness), so a back-face attack trigger can be tested
// without waiting a turn.
func backFaceOnBattlefield(g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	c.SetFace(1)
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// openPickTarget reports whether a pick_target prompt is waiting.
func openPickTarget(g *game.Game) bool {
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoicePickTarget {
			return true
		}
	}
	return false
}

// The disturb creatures whose back face is keywords and a body: each
// lands back face up with exactly the back face's keywords and stats,
// and the front face's mana value (CR 712.8e).
func TestDisturbKeywordBackFaces(t *testing.T) {
	cases := []struct {
		row      cards.Card
		back     string
		power    int
		tough    int
		keywords []string
		absent   []string
		mv       int
	}{
		{disturbRow(belovedBeggarOracleID, []string{"W"},
			disturbFace{"Beloved Beggar", "Creature — Human Peasant", "{1}{W}", "0", "4"},
			disturbFace{"Generous Soul", "Creature — Spirit", "", "4", "4"}),
			"Generous Soul", 4, 4, []string{"flying", "vigilance"}, nil, 2},
		{disturbRow(galedrifterOracleID, []string{"U"},
			disturbFace{"Galedrifter", "Creature — Hippogriff", "{3}{U}", "3", "2"},
			disturbFace{"Waildrifter", "Creature — Hippogriff Spirit", "", "2", "2"}),
			"Waildrifter", 2, 2, []string{"flying"}, nil, 4},
		{disturbRow(mourningPatrolOracleID, []string{"W"},
			disturbFace{"Mourning Patrol", "Creature — Human Soldier", "{2}{W}", "2", "3"},
			disturbFace{"Morning Apparition", "Creature — Spirit Soldier", "", "2", "1"}),
			"Morning Apparition", 2, 1, []string{"flying", "vigilance"}, nil, 3},
	}
	for _, tc := range cases {
		t.Run(tc.back, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := importToGraveyard(t, g, tc.row, me)
			disturb(t, g, me, id)
			passPriorityAroundTable(t, g)
			zone, c := cardWhere(g, me, id)
			if zone != "battlefield" || c.Name != tc.back || c.ActiveFace != 1 {
				t.Fatalf("disturbed card is in %s as %q (face %d)", zone, c.Name, c.ActiveFace)
			}
			if p, tt := ptOf(t, g, id); p != tc.power || tt != tc.tough {
				t.Errorf("%s is %d/%d, want %d/%d", tc.back, p, tt, tc.power, tc.tough)
			}
			for _, kw := range tc.keywords {
				if !game.HasKeyword(&c, kw) {
					t.Errorf("%s lacks %s", tc.back, kw)
				}
			}
			if c.ManaValue() != tc.mv {
				t.Errorf("%s mana value = %d, want %d", tc.back, c.ManaValue(), tc.mv)
			}
			destroy(t, g, id)
			if zone, _ := cardWhere(g, me, id); zone != "exile" {
				t.Errorf("destroyed %s went to %s, want exile", tc.back, zone)
			}
		})
	}
}

// The disturb Auras: cast onto a creature through the back face's
// enchant clause, each pumps or grants what it prints.
func TestDisturbAuraBackFaces(t *testing.T) {
	cases := []struct {
		row          cards.Card
		back         string
		power, tough int
		keyword      string
	}{
		{disturbRow(kindlyAncestorOracleID, []string{"W"},
			disturbFace{"Kindly Ancestor", "Creature — Spirit", "{2}{W}", "2", "3"},
			disturbFace{"Ancestor's Embrace", "Enchantment — Aura", "", "", ""}),
			"Ancestor's Embrace", 2, 2, "lifelink"},
		{disturbRow(lanternBearerOracleID, []string{"U"},
			disturbFace{"Lantern Bearer", "Creature — Spirit", "{U}", "1", "1"},
			disturbFace{"Lanterns' Lift", "Enchantment — Aura", "", "", ""}),
			"Lanterns' Lift", 3, 3, "flying"},
		{disturbRow(twinbladeGeistOracleID, []string{"W"},
			disturbFace{"Twinblade Geist", "Creature — Spirit Warrior", "{1}{W}", "1", "1"},
			disturbFace{"Twinblade Invocation", "Enchantment — Aura", "", "", ""}),
			"Twinblade Invocation", 2, 2, "double strike"},
		{disturbRow(bindingGeistOracleID, []string{"U"},
			disturbFace{"Binding Geist", "Creature — Spirit", "{2}{U}", "3", "1"},
			disturbFace{"Spectral Binding", "Enchantment — Aura", "", "", ""}),
			"Spectral Binding", 0, 2, ""},
	}
	for _, tc := range cases {
		t.Run(tc.back, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := importToGraveyard(t, g, tc.row, me)
			bear := auraBear(g, me.ID)
			if openPickTarget(g) {
				t.Fatal("setup left a prompt open")
			}
			disturb(t, g, me, id, game.TargetRef{Kind: game.TargetCard, ID: bear})
			passPriorityAroundTable(t, g)
			zone, aura := cardWhere(g, me, id)
			if zone != "battlefield" || aura.Name != tc.back || aura.AttachedTo.ID != bear {
				t.Fatalf("aura is in %s as %q attached to %v", zone, aura.Name, aura.AttachedTo.ID)
			}
			if p, tt := ptOf(t, g, bear); p != tc.power || tt != tc.tough {
				t.Errorf("enchanted bear is %d/%d, want %d/%d", p, tt, tc.power, tc.tough)
			}
			if tc.keyword != "" {
				_, host := cardWhere(g, me, bear)
				if !game.HasKeyword(&host, tc.keyword) {
					t.Errorf("enchanted bear lacks %s", tc.keyword)
				}
			}
			killCreature(t, g, me.ID, bear)
			if zone, _ := cardWhere(g, me, id); zone != "exile" {
				t.Errorf("fallen-off %s went to %s, want exile", tc.back, zone)
			}
		})
	}
}

// Binding Geist's attack trigger shrinks the creature an opponent
// controls that it targets, until end of turn.
func TestBindingGeistAttackShrinksAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := pushDiesCreatureForTest(g, me.ID, "Binding Geist", bindingGeistOracleID, "Creature — Spirit", 3, 1)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	declareAttack(t, g, opp.ID, geist)
	if openPickTarget(g) {
		answerPickTarget(t, g, theirs)
	}
	item := triggerOnStack(g, geist)
	if item == nil {
		t.Fatal("Binding Geist's attack trigger is not on the stack")
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != theirs {
		t.Fatalf("trigger targets %v, want only the opponent's bear (not %s)", item.Targets, mine)
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, theirs); p != 0 || tt != 2 {
		t.Errorf("the opponent's bear is %d/%d, want 0/2", p, tt)
	}
	if p, _ := ptOf(t, g, mine); p != 2 {
		t.Errorf("my own bear was shrunk to %d power", p)
	}
}

func overwhelmedArchivistRow() cards.Card {
	return disturbRow(overwhelmedArchivistOracleID, []string{"U"},
		disturbFace{"Overwhelmed Archivist", "Creature — Human Wizard", "{2}{U}", "3", "2"},
		disturbFace{"Archive Haunt", "Creature — Spirit Wizard", "", "2", "1"})
}

// Overwhelmed Archivist loots as it enters.
func TestOverwhelmedArchivistLootsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToHand(overwhelmedArchivistRow(), me)
	toMainPhase(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after the ETB, want %d (one drawn, the discard still to choose)", me.Hand.Size(), hand+1)
	}
	if !(discardOwed(g, me.ID) == 1) {
		t.Error("no discard prompt after the draw")
	}
}

// Archive Haunt loots when it attacks; a disturbed Archive Haunt does
// not run the front face's entry loot.
func TestArchiveHauntLootsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	cast := importToGraveyard(t, g, overwhelmedArchivistRow(), me)
	disturb(t, g, me, cast)
	passPriorityAroundTable(t, g)
	if discardOwed(g, me.ID) == 1 {
		t.Fatal("a disturbed Archive Haunt ran the front face's entry loot")
	}

	haunt := backFaceOnBattlefield(g, overwhelmedArchivistRow(), me)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, haunt)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || !(discardOwed(g, me.ID) == 1) {
		t.Errorf("after the attack trigger: hand %d (want %d), discard prompt %v", me.Hand.Size(), hand+1, (discardOwed(g, me.ID) == 1))
	}
}

func lunarchVeteranRow() cards.Card {
	return disturbRow(lunarchVeteranOracleID, []string{"W"},
		disturbFace{"Lunarch Veteran", "Creature — Human Cleric", "{W}", "1", "1"},
		disturbFace{"Luminous Phantom", "Creature — Spirit Cleric", "", "1", "1"})
}

// enterCreatureFromHand puts a vanilla creature onto the battlefield
// through the entry pipeline, so "enters" triggers see it.
func enterCreatureFromHand(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	c := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID}
	p.Hand.PushTop(c)
	var err error
	g.WithWriteLock(func() {
		_, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{})
	})
	if err != nil {
		t.Fatalf("put the bear onto the battlefield: %v", err)
	}
	return c.InstanceID
}

// Lunarch Veteran gains 1 life for another creature entering under its
// controller's control, and none for an opponent's.
func TestLunarchVeteranGainsLifeOnAnotherEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	pushDiesCreatureForTest(g, me.ID, "Lunarch Veteran", lunarchVeteranOracleID, "Creature — Human Cleric", 1, 1)
	life := me.Life
	enterCreatureFromHand(t, g, me)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after a creature entered, want %d", me.Life, life+1)
	}
	enterCreatureFromHand(t, g, opp)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("an opponent's creature entering gained life (now %d)", me.Life)
	}
}

// Luminous Phantom gains 1 life whenever another creature its
// controller controls leaves the battlefield — dying or otherwise.
func TestLuminousPhantomGainsLifeWhenAnotherCreatureLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	backFaceOnBattlefield(g, lunarchVeteranRow(), me)
	dies := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	exiled := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	life := me.Life

	destroy(t, g, dies)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life %d after a creature died, want %d", me.Life, life+1)
	}
	exileCreature(t, g, exiled)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d after a creature was exiled, want %d", me.Life, life+2)
	}
}
