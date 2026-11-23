package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	cardsg "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/g"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	cardsmi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/mi"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	cardsto "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/to"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		{effectstest.DisturbRow(cardsb.BelovedBeggarOracleID, []string{"W"},
			effectstest.DisturbFace{Name: "Beloved Beggar", TypeLine: "Creature — Human Peasant", Cost: "{1}{W}", Power: "0", Toughness: "4"},
			effectstest.DisturbFace{Name: "Generous Soul", TypeLine: "Creature — Spirit", Cost: "", Power: "4", Toughness: "4"}),
			"Generous Soul", 4, 4, []string{"flying", "vigilance"}, nil, 2},
		{effectstest.DisturbRow(cardsg.GaledrifterOracleID, []string{"U"},
			effectstest.DisturbFace{Name: "Galedrifter", TypeLine: "Creature — Hippogriff", Cost: "{3}{U}", Power: "3", Toughness: "2"},
			effectstest.DisturbFace{Name: "Waildrifter", TypeLine: "Creature — Hippogriff Spirit", Cost: "", Power: "2", Toughness: "2"}),
			"Waildrifter", 2, 2, []string{"flying"}, nil, 4},
		{effectstest.DisturbRow(cardsmi.MourningPatrolOracleID, []string{"W"},
			effectstest.DisturbFace{Name: "Mourning Patrol", TypeLine: "Creature — Human Soldier", Cost: "{2}{W}", Power: "2", Toughness: "3"},
			effectstest.DisturbFace{Name: "Morning Apparition", TypeLine: "Creature — Spirit Soldier", Cost: "", Power: "2", Toughness: "1"}),
			"Morning Apparition", 2, 1, []string{"flying", "vigilance"}, nil, 3},
	}
	for _, tc := range cases {
		t.Run(tc.back, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.ImportToGraveyard(t, g, tc.row, me)
			disturb(t, g, me, id)
			effectstest.PassPriorityAroundTable(t, g)
			zone, c := effectstest.CardWhere(g, me, id)
			if zone != "battlefield" || c.Name != tc.back || c.ActiveFace != 1 {
				t.Fatalf("disturbed card is in %s as %q (face %d)", zone, c.Name, c.ActiveFace)
			}
			if p, tt := effectstest.PtOf(t, g, id); p != tc.power || tt != tc.tough {
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
			effectstest.Destroy(t, g, id)
			if zone, _ := effectstest.CardWhere(g, me, id); zone != "exile" {
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
		{effectstest.DisturbRow(cardsj.KindlyAncestorOracleID, []string{"W"},
			effectstest.DisturbFace{Name: "Kindly Ancestor", TypeLine: "Creature — Spirit", Cost: "{2}{W}", Power: "2", Toughness: "3"},
			effectstest.DisturbFace{Name: "Ancestor's Embrace", TypeLine: "Enchantment — Aura", Cost: "", Power: "", Toughness: ""}),
			"Ancestor's Embrace", 2, 2, "lifelink"},
		{effectstest.DisturbRow(cardskr.LanternBearerOracleID, []string{"U"},
			effectstest.DisturbFace{Name: "Lantern Bearer", TypeLine: "Creature — Spirit", Cost: "{U}", Power: "1", Toughness: "1"},
			effectstest.DisturbFace{Name: "Lanterns' Lift", TypeLine: "Enchantment — Aura", Cost: "", Power: "", Toughness: ""}),
			"Lanterns' Lift", 3, 3, "flying"},
		{effectstest.DisturbRow(cardsto.TwinbladeGeistOracleID, []string{"W"},
			effectstest.DisturbFace{Name: "Twinblade Geist", TypeLine: "Creature — Spirit Warrior", Cost: "{1}{W}", Power: "1", Toughness: "1"},
			effectstest.DisturbFace{Name: "Twinblade Invocation", TypeLine: "Enchantment — Aura", Cost: "", Power: "", Toughness: ""}),
			"Twinblade Invocation", 2, 2, "double strike"},
		{effectstest.DisturbRow(cardsb.BindingGeistOracleID, []string{"U"},
			effectstest.DisturbFace{Name: "Binding Geist", TypeLine: "Creature — Spirit", Cost: "{2}{U}", Power: "3", Toughness: "1"},
			effectstest.DisturbFace{Name: "Spectral Binding", TypeLine: "Enchantment — Aura", Cost: "", Power: "", Toughness: ""}),
			"Spectral Binding", 0, 2, ""},
	}
	for _, tc := range cases {
		t.Run(tc.back, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.ImportToGraveyard(t, g, tc.row, me)
			bear := effectstest.AuraBear(g, me.ID)
			if effectstest.OpenPickTarget(g) {
				t.Fatal("setup left a prompt open")
			}
			disturb(t, g, me, id, game.TargetRef{Kind: game.TargetCard, ID: bear})
			effectstest.PassPriorityAroundTable(t, g)
			zone, aura := effectstest.CardWhere(g, me, id)
			if zone != "battlefield" || aura.Name != tc.back || aura.AttachedTo.ID != bear {
				t.Fatalf("aura is in %s as %q attached to %v", zone, aura.Name, aura.AttachedTo.ID)
			}
			if p, tt := effectstest.PtOf(t, g, bear); p != tc.power || tt != tc.tough {
				t.Errorf("enchanted bear is %d/%d, want %d/%d", p, tt, tc.power, tc.tough)
			}
			if tc.keyword != "" {
				_, host := effectstest.CardWhere(g, me, bear)
				if !game.HasKeyword(&host, tc.keyword) {
					t.Errorf("enchanted bear lacks %s", tc.keyword)
				}
			}
			effectstest.KillCreature(t, g, me.ID, bear)
			if zone, _ := effectstest.CardWhere(g, me, id); zone != "exile" {
				t.Errorf("fallen-off %s went to %s, want exile", tc.back, zone)
			}
		})
	}
}

// Binding Geist's attack trigger shrinks the creature an opponent
// controls that it targets, until end of turn.
func TestBindingGeistAttackShrinksAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := effectstest.PushDiesCreatureForTest(g, me.ID, "Binding Geist", cardsb.BindingGeistOracleID, "Creature — Spirit", 3, 1)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, geist)
	if effectstest.OpenPickTarget(g) {
		effectstest.AnswerPickTarget(t, g, theirs)
	}
	item := effectstest.TriggerOnStack(g, geist)
	if item == nil {
		t.Fatal("Binding Geist's attack trigger is not on the stack")
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != theirs {
		t.Fatalf("trigger targets %v, want only the opponent's bear (not %s)", item.Targets, mine)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 0 || tt != 2 {
		t.Errorf("the opponent's bear is %d/%d, want 0/2", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, mine); p != 2 {
		t.Errorf("my own bear was shrunk to %d power", p)
	}
}

func overwhelmedArchivistRow() cards.Card {
	return effectstest.DisturbRow(cardso.OverwhelmedArchivistOracleID, []string{"U"},
		effectstest.DisturbFace{Name: "Overwhelmed Archivist", TypeLine: "Creature — Human Wizard", Cost: "{2}{U}", Power: "3", Toughness: "2"},
		effectstest.DisturbFace{Name: "Archive Haunt", TypeLine: "Creature — Spirit Wizard", Cost: "", Power: "2", Toughness: "1"})
}

// Overwhelmed Archivist loots as it enters.
func TestOverwhelmedArchivistLootsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.ImportToHand(overwhelmedArchivistRow(), me)
	effectstest.ToMainPhase(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after the ETB, want %d (one drawn, the discard still to choose)", me.Hand.Size(), hand+1)
	}
	if !(effectstest.DiscardOwed(g, me.ID) == 1) {
		t.Error("no discard prompt after the draw")
	}
}

// Archive Haunt loots when it attacks; a disturbed Archive Haunt does
// not run the front face's entry loot.
func TestArchiveHauntLootsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	cast := effectstest.ImportToGraveyard(t, g, overwhelmedArchivistRow(), me)
	disturb(t, g, me, cast)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, me.ID) == 1 {
		t.Fatal("a disturbed Archive Haunt ran the front face's entry loot")
	}

	haunt := effectstest.BackFaceOnBattlefield(g, overwhelmedArchivistRow(), me)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, haunt)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || !(effectstest.DiscardOwed(g, me.ID) == 1) {
		t.Errorf("after the attack trigger: hand %d (want %d), discard prompt %v", me.Hand.Size(), hand+1, (effectstest.DiscardOwed(g, me.ID) == 1))
	}
}

func lunarchVeteranRow() cards.Card {
	return effectstest.DisturbRow(effects.LunarchVeteranOracleID, []string{"W"},
		effectstest.DisturbFace{Name: "Lunarch Veteran", TypeLine: "Creature — Human Cleric", Cost: "{W}", Power: "1", Toughness: "1"},
		effectstest.DisturbFace{Name: "Luminous Phantom", TypeLine: "Creature — Spirit Cleric", Cost: "", Power: "1", Toughness: "1"})
}

// enterCreatureFromHand puts a vanilla creature onto the battlefield
// through the entry pipeline, so "enters" triggers see it.
func enterCreatureFromHand(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	c := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID}
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainPhase(t, g)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Lunarch Veteran", effects.LunarchVeteranOracleID, "Creature — Human Cleric", 1, 1)
	life := me.Life
	enterCreatureFromHand(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after a creature entered, want %d", me.Life, life+1)
	}
	enterCreatureFromHand(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("an opponent's creature entering gained life (now %d)", me.Life)
	}
}

// Luminous Phantom gains 1 life whenever another creature its
// controller controls leaves the battlefield — dying or otherwise.
func TestLuminousPhantomGainsLifeWhenAnotherCreatureLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainPhase(t, g)
	effectstest.BackFaceOnBattlefield(g, lunarchVeteranRow(), me)
	dies := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	exiled := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	life := me.Life

	effectstest.Destroy(t, g, dies)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life %d after a creature died, want %d", me.Life, life+1)
	}
	effectstest.ExileCreature(t, g, exiled)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d after a creature was exiled, want %d", me.Life, life+2)
	}
}
