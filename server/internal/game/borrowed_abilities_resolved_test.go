package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// borrowed_abilities_resolved_test.go — ADR 0155 PR 4 (#3202) on the
// fixture catalog: the borrowAbilities mod (Decision 9), the one-colour
// spend widening, and the library-top watch (DependsOnLibraryTop).

// bwRegisterBorrow captures the elf's rows and lends them to `to` with a
// resolved borrow at timestamp `ts`.
func bwRegisterBorrow(t *testing.T, g *Game, lender, to uuid.UUID, ts int64) {
	t.Helper()
	g.WithWriteLock(func() {
		var rows []LentRow
		if c := findBattlefieldCard(g, lender); c != nil {
			rows = LentRowsForEffect(c)
		} else if c, ok := g.LookupCardForEffect(lender); ok {
			rows = PrintedLentRowsForEffect(c)
		}
		if len(rows) == 0 {
			t.Fatal("the lender has no rows to capture")
		}
		if !g.registerScopedEffectLocked(to, g.PinnedObjectsLocked(to), []Mod{BorrowAbilitiesMod("", rows...)},
			g.UntilEndOfTurnDuration(), "gains all activated abilities of Mana Elf until end of turn", ts) {
			t.Fatal("nothing registered")
		}
	})
}

// The captured rows are lent until the duration ends, survive the lender
// leaving (CR 608.2h), and name the lender's catalog rows.
func TestResolvedBorrowKeepsItsRowsAfterTheLenderLeaves(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	elf := gaPush(g, me, bwElfName, "Creature — Elf", bwElfOracle)
	bear := gaBear(g, me)
	bwRegisterBorrow(t, g, elf, bear, timeNowUnixNano())
	c := layeredBattlefieldCard(t, g, bear)
	if n := len(borrowedActivated(c)); n != 2 {
		t.Fatalf("lent activated rows = %d, want the elf's 2", n)
	}
	mana := borrowedMana(c)
	if len(mana) != 1 || mana[0].LentBy != elf || mana[0].Ref != "borrow:"+bwElfOracle+":0:0" {
		t.Fatalf("lent mana = %+v", mana)
	}
	if info := GrantedAbilitiesOf(c); len(info) != 1 || !strings.Contains(info[0].Text, "Mana Elf") {
		t.Errorf("granted_abilities = %+v, want the resolved borrow's label", info)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me}, elf); err != nil {
		t.Fatalf("move the elf: %v", err)
	}
	if n := len(borrowedMana(layeredBattlefieldCard(t, g, bear))); n != 1 {
		t.Errorf("the lender left and %d mana rows are lent, want 1", n)
	}
}

// CR 613.6: a removal on the recipient takes the rows when it is later
// and leaves them when it is earlier — the record's timestamp slot, not
// the borrow step.
func TestResolvedBorrowFollowsTimestampsAgainstARemoval(t *testing.T) {
	for _, tc := range []struct {
		name  string
		delta int64
		want  int
	}{{"borrow first, removal later: gone", -1, 0}, {"removal first, borrow later: kept", +1, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			fx := stubBorrowedAbilityCatalog(t)
			g := newActiveGame(t)
			me := g.Seats[0].ID
			elf := gaPush(g, me, bwElfName, "Creature — Elf", bwElfOracle)
			fx.victim = gaBear(g, me)
			silencer := gaPush(g, me, "Silencer", "Enchantment", bwSilencerOracle)
			at := layeredBattlefieldCard(t, g, silencer).EnteredBattlefieldAt
			bwRegisterBorrow(t, g, elf, fx.victim, at+tc.delta)
			c := layeredBattlefieldCard(t, g, fx.victim)
			if !c.HasLostAllAbilities() {
				t.Fatal("fixture: the bear should have lost its abilities")
			}
			if n := len(borrowedActivated(c)); n != tc.want {
				t.Errorf("lent activated rows = %d, want %d", n, tc.want)
			}
		})
	}
}

// Registration refuses a malformed mod — a programming error the first
// test of the card catches.
func TestBorrowAbilitiesModRefusesAMalformedMod(t *testing.T) {
	good := LentRow{Slot: AbilitySlotActivated, Kind: LentCatalog, Key: "k", Index: 0}
	for _, tc := range []struct {
		name string
		mod  Mod
	}{
		{"no row", BorrowAbilitiesMod("")},
		{"no key", BorrowAbilitiesMod("", LentRow{Slot: AbilitySlotActivated})},
		{"bad slot", BorrowAbilitiesMod("", LentRow{Slot: "loyalty", Key: "k"})},
		{"land row outside the mana slot", BorrowAbilitiesMod("", LentRow{Slot: AbilitySlotActivated, Kind: LentLand, Color: "G"})},
		{"land row of no colour", BorrowAbilitiesMod("", LentRow{Slot: BorrowSlotMana, Kind: LentLand, Color: "Q"})},
		{"unknown kind", BorrowAbilitiesMod("", LentRow{Slot: AbilitySlotActivated, Kind: LentKind(9), Key: "k"})},
		{"spend from no colour", BorrowAbilitiesMod("X", good)},
		{"rows on another kind", Mod{Kind: ModAddKeywords, Keywords: []string{"flying"}, Borrows: []LentRow{good}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if borrowModProblem(tc.mod) == "" {
				t.Errorf("borrowModProblem(%+v) accepted it", tc.mod)
			}
		})
	}
	if p := borrowModProblem(BorrowAbilitiesMod("U", good, LentRow{Slot: BorrowSlotMana, Kind: LentLand, Color: "G"})); p != "" {
		t.Errorf("a well-formed mod was refused: %s", p)
	}
}

// A one-colour grant widens each coloured symbol to that colour too, and
// leaves {C}, generic and an already-widened symbol alone.
func TestWidenFromColorsAddsTheColourToEachColouredSymbol(t *testing.T) {
	cost, err := ParseCost("{2}{G}{G/W}{C}{B/P}")
	if err != nil {
		t.Fatal(err)
	}
	got := widenFromColors(cost, []string{"U"})
	if got.Generic != cost.Generic {
		t.Errorf("generic changed: %d → %d", cost.Generic, got.Generic)
	}
	for i, r := range got.Required {
		blue := r.Admits("U")
		switch {
		case requiresColorless(cost.Required[i]):
			if blue {
				t.Errorf("blue pays the colourless symbol %v", cost.Required[i])
			}
		default:
			if !blue {
				t.Errorf("blue does not pay %v", cost.Required[i])
			}
			if r.Admits("R") {
				t.Errorf("red pays %v after a blue-only widening", cost.Required[i])
			}
			if r.Phyrexian != cost.Required[i].Phyrexian {
				t.Errorf("%v lost its Phyrexian half", cost.Required[i])
			}
		}
	}
	if cost.Required[1].Admits("U") {
		t.Error("the widening wrote through to the printed cost")
	}
	blue, _ := ParseCost("{2}{U}")
	if same := widenFromColors(blue, []string{"U"}); len(same.Required) != 1 || len(same.Required[0].Options) != 1 {
		t.Errorf("a {U} was widened: %+v", same.Required)
	}
}

// DependsOnLibraryTop: the top of a library changing with no event at all
// re-runs the pass, and a table with no such static arms no watch.
func TestLibraryTopWatchReReadsTheTopCard(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	snoop := bwBorrow(func(g *Game, source *Card) []*Card {
		p := g.playerByIDLocked(source.Controller)
		if p == nil || len(p.Library.Cards) == 0 {
			return nil
		}
		return []*Card{&p.Library.Cards[len(p.Library.Cards)-1]}
	}, "Has all activated abilities of the top card of your library.")
	snoop.DependsOnLibraryTop = true
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef {
		if key == "fixture-bw-snoop" {
			return &CardDef{Static: []StaticAbility{snoop}}
		}
		return prev(key)
	}
	t.Cleanup(func() { CatalogLookup = prev })

	layeredBattlefieldCard(t, g, gaBear(g, me))
	if g.libraryTopWatch != nil {
		t.Fatal("a table with no DependsOnLibraryTop static armed the watch")
	}
	borrower := gaPush(g, me, "Snoop", "Creature — Goblin", "fixture-bw-snoop")
	if n := len(borrowedMana(layeredBattlefieldCard(t, g, borrower))); n != 0 {
		t.Fatalf("a filler on top lent %d rows", n)
	}
	g.WithWriteLock(func() {
		g.playerByIDLocked(me).Library.PushTop(Card{InstanceID: uuid.New(), Name: bwElfName, TypeLine: "Creature — Elf", OracleID: bwElfOracle, Owner: me, Controller: me})
	})
	if n := len(borrowedMana(layeredBattlefieldCard(t, g, borrower))); n != 1 {
		t.Fatalf("the elf on top lent %d mana rows, want 1", n)
	}
	g.WithWriteLock(func() { _, _ = g.playerByIDLocked(me).Library.PopTop() })
	if n := len(borrowedMana(layeredBattlefieldCard(t, g, borrower))); n != 0 {
		t.Errorf("the elf left the top and %d rows are lent", n)
	}
}
