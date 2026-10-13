package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// borrowed_abilities_test.go — ADR 0155's seam (PR 1, #1557), pinned
// with fixture lenders against a stubbed catalog: the lent rows reach the
// borrower as a fourth half, are activated as the borrower's (CR 302.6,
// CR 602.2), resolve without their lender (CR 113.7a) and restore through
// catalog/activated; the borrow step's CR 613.8 ordering and CR 613.6
// timestamps; the riders; and the stale ref (#544).

const (
	bwElfOracle       = "fixture-bw-elf"       // creature: {T}: Add {G}; {T}: gain 1 life; a loyalty row
	bwOozeOracle      = "fixture-bw-ooze"      // has all activated abilities of creature cards in graveyards
	bwSafehouseOracle = "fixture-bw-safehouse" // has all activated abilities of land cards in graveyards
	bwMarvinOracle    = "fixture-bw-marvin"    // has all activated abilities of other creatures you control
	bwCauldronOracle  = "fixture-bw-cauldron"  // creatures you control have all activated abilities of the fixture's lenders
	bwKrajOracle      = "fixture-bw-kraj"      // has all activated abilities of every other Kraj
	bwMairsilOracle   = "fixture-bw-mairsil"   // the Ooze, each lent row once each turn, with an any-colour rider
	bwSharkeyOracle   = "fixture-bw-sharkey"   // creature cards in graveyards, except mana abilities
	bwFenceOracle     = "fixture-bw-fence"     // creature cards in graveyards, except loyalty abilities
	bwSilencerOracle  = "fixture-bw-silencer"  // layer-6 "loses all abilities" on the fixture's victim
	bwRiteOracle      = "fixture-bw-rite"      // grants the ping bundle to the fixture's grant target
	bwDranaOracle     = "fixture-bw-drana"     // the fixture's forbidden object's activated abilities can't be activated

	bwPingBundle = "fixture-bw/ping"

	bwElfName      = "Mana Elf"
	bwGainLabel    = "{T}: Mana Elf gains you 1 life."
	bwLoyaltyLabel = "+1: Mana Elf studies."
	bwKrajLabel    = "{2}: Kraj hums."
	bwPingLabel    = "{1}: ping"
	bwOozeClause   = "Has all activated abilities of all creature cards in all graveyards."
	bwMairsilRider = "You may spend mana as though it were mana of any color to activate those abilities."
)

// bwFixture is the mutable half of the fixture catalog: who the silencer
// silences, who the Rite grants to, whose abilities Drana forbids, and
// what the Cauldron lends.
type bwFixture struct {
	victim      uuid.UUID
	grantTarget uuid.UUID
	forbidden   uuid.UUID
	lenders     []uuid.UUID
	gained      int
}

func bwSelf(target *Card, _ *Game, source *Card) bool {
	return source != nil && target.InstanceID == source.InstanceID
}

// bwGraveyardCards lists every card of a type in every graveyard.
func bwGraveyardCards(cardType string) func(g *Game, _ *Card) []*Card {
	return func(g *Game, _ *Card) []*Card {
		var out []*Card
		for _, p := range g.Seats {
			if p.Graveyard == nil {
				continue
			}
			for i := range p.Graveyard.Cards {
				if hasCardType(&p.Graveyard.Cards[i], cardType) {
					out = append(out, &p.Graveyard.Cards[i])
				}
			}
		}
		return out
	}
}

func bwBorrow(lenders func(*Game, *Card) []*Card, label string) StaticAbility {
	return StaticAbility{
		Layer:     Layer6Ability,
		AppliesTo: bwSelf,
		Borrows:   &AbilityBorrow{Lenders: lenders, Label: label},
	}
}

// stubBorrowedAbilityCatalog wires the fixture catalog for one test.
func stubBorrowedAbilityCatalog(t testing.TB) *bwFixture {
	t.Helper()
	fx := &bwFixture{}
	loyalty := 1
	gain := ActivatedAbilityShape{
		Label: bwGainLabel,
		Cost:  AbilityCost{Tap: true},
		Effect: func(g *Game, item *StackItem) error {
			fx.gained++
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
		},
	}
	elfLoyalty := ActivatedAbilityShape{
		Label:  bwLoyaltyLabel,
		Cost:   AbilityCost{Loyalty: &loyalty},
		Effect: func(*Game, *StackItem) error { return nil },
	}
	ooze := bwBorrow(bwGraveyardCards("creature"), bwOozeClause)
	mairsil := bwBorrow(bwGraveyardCards("creature"), "Has all activated abilities of creature cards in graveyards; each once each turn.")
	mairsil.Borrows.EachOncePerTurn = true
	sharkey := bwBorrow(bwGraveyardCards("creature"), "Has all activated abilities of creature cards in graveyards except mana abilities.")
	sharkey.Borrows.Except = BorrowExceptMana
	fence := bwBorrow(bwGraveyardCards("creature"), "Has all activated abilities of creature cards in graveyards except loyalty abilities.")
	fence.Borrows.Except = BorrowExceptLoyalty
	defs := map[string]*CardDef{
		bwElfOracle: {
			ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
			Activated:     []ActivatedAbilityShape{gain, elfLoyalty},
		},
		bwOozeOracle:      {Static: []StaticAbility{ooze}},
		bwSafehouseOracle: {Static: []StaticAbility{bwBorrow(bwGraveyardCards("land"), "Has all activated abilities of all land cards in all graveyards.")}},
		bwMarvinOracle: {Static: []StaticAbility{bwBorrow(func(g *Game, source *Card) []*Card {
			var out []*Card
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				if c.InstanceID != source.InstanceID && c.IsCreature() && c.Controller == source.Controller {
					out = append(out, c)
				}
			}
			return out
		}, "Has all activated abilities of other creatures you control.")}},
		bwCauldronOracle: {Static: []StaticAbility{{
			Layer:     Layer6Ability,
			AppliesTo: creaturesYouControlFixture,
			Borrows: &AbilityBorrow{Label: "Creatures you control have all activated abilities of the lenders.", Lenders: func(g *Game, _ *Card) []*Card {
				var out []*Card
				for _, id := range fx.lenders {
					if c := g.findCardByIDLocked(id); c != nil {
						out = append(out, c)
					}
				}
				return out
			}},
		}}},
		bwKrajOracle: {
			Activated: []ActivatedAbilityShape{{Label: bwKrajLabel, Effect: func(*Game, *StackItem) error { return nil }}},
			Static: []StaticAbility{bwBorrow(func(g *Game, source *Card) []*Card {
				var out []*Card
				for i := range g.Battlefield.Cards {
					c := &g.Battlefield.Cards[i]
					if c.InstanceID != source.InstanceID && c.OracleID == bwKrajOracle {
						out = append(out, c)
					}
				}
				return out
			}, "Has all activated abilities of each other Kraj.")},
		},
		bwMairsilOracle: {
			Static: []StaticAbility{mairsil},
			AnyColorSpend: []AnyColorSpendStatic{{
				Label: bwMairsilRider,
				Covers: func(ctx ManaSpendContext) bool {
					return ctx.AbilityOrigin.LentByBorrow(mairsil.Borrows.Label)
				},
			}},
		},
		bwSharkeyOracle: {Static: []StaticAbility{sharkey}},
		bwFenceOracle:   {Static: []StaticAbility{fence}},
		bwSilencerOracle: {Static: []StaticAbility{{
			Layer:            Layer6Ability,
			RemovesAbilities: true,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.InstanceID == fx.victim
			},
		}}},
		bwRiteOracle: {Static: []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.InstanceID == fx.grantTarget
			},
			GrantAbilities: []string{bwPingBundle},
		}}},
		bwDranaOracle: {ActivationRestrictions: []ActivationRestriction{{
			Label: "Activated abilities of the forbidden object can't be activated.",
			Forbids: func(q ActivationQuery) bool {
				return q.Card.InstanceID == fx.forbidden
			},
		}}},
		GrantKey(bwPingBundle): {
			Activated: []ActivatedAbilityShape{{Label: bwPingLabel, Effect: func(*Game, *StackItem) error { return nil }}},
			GrantText: "{1}: ping.",
		},
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
	return fx
}

// bwToGraveyard puts a fixture card straight into a seat's graveyard and
// invalidates the layers, as the graveyard bump (#1117) does for a move.
func bwToGraveyard(g *Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p := g.playerByIDLocked(owner)
		p.Graveyard.PushTop(Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: owner, Controller: owner})
	})
	g.BumpLayerVersionForTest()
	return id
}

func bwElfInGraveyard(g *Game, owner uuid.UUID) uuid.UUID {
	return bwToGraveyard(g, owner, bwElfName, "Creature — Elf", bwElfOracle)
}

// bwExile moves a card from a graveyard to exile, through the engine.
func bwExile(t *testing.T, g *Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneGraveyard, Owner: owner}, ZoneRef{Kind: ZoneExile}, id); err != nil {
		t.Fatalf("exile %s: %v", id, err)
	}
}

// borrowedRows returns the borrowed origins of a card's activated list.
func borrowedActivated(c Card) []AbilityOrigin {
	_, origins := ActivatedAbilitiesWithOrigins(c)
	var out []AbilityOrigin
	for _, o := range origins {
		if o.Borrowed {
			out = append(out, o)
		}
	}
	return out
}

func borrowedMana(c Card) []AbilityOrigin {
	_, origins := ManaAbilitiesWithOrigins(c)
	var out []AbilityOrigin
	for _, o := range origins {
		if o.Borrowed {
			out = append(out, o)
		}
	}
	return out
}

func bwStackItemFrom(g *Game, source uuid.UUID) *StackItem {
	var out *StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it.SourceCardID == source && it.Kind == StackItemActivated {
				out = it
			}
		}
	})
	return out
}

// --- the lent rows reach the borrower -----------------------------------

// A borrower gets every activated and mana row of a graveyard lender, as
// a fourth half after its own, named by the LENDER's catalog row, with
// one granted_abilities line for the borrow.
func TestBorrowerHasALendersRowsAfterItsOwn(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	elf := bwElfInGraveyard(g, me)
	ooze := gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)

	c := layeredBattlefieldCard(t, g, ooze)
	abs, origins := ActivatedAbilitiesWithOrigins(c)
	if len(abs) != 2 || abs[0].Label != bwGainLabel || abs[1].Label != bwLoyaltyLabel {
		t.Fatalf("activated rows = %+v, want the elf's two", abs)
	}
	o := origins.At(0)
	if !o.Borrowed || o.LentBy != elf || o.GrantedBy != ooze || o.Granted() {
		t.Errorf("origin = %+v, want lent by the elf through the ooze's borrow", o)
	}
	if want := "borrow:" + bwElfOracle + ":0:0"; o.Ref != want {
		t.Errorf("ref = %q, want %q", o.Ref, want)
	}
	if o.Lent.Key != bwElfOracle || o.Lent.Index != 0 || o.Lent.LenderName != bwElfName {
		t.Errorf("entry = %+v", o.Lent)
	}
	manas, mo := ManaAbilitiesWithOrigins(c)
	if len(manas) != 1 || manas[0].Produced != "{G}" || !mo.At(0).Borrowed {
		t.Errorf("mana rows = %+v / %+v, want the elf's {G}", manas, mo)
	}
	info := GrantedAbilitiesOf(c)
	if len(info) != 1 || info[0].Text != bwOozeClause || info[0].Source != ooze {
		t.Errorf("GrantedAbilitiesOf = %+v, want the borrow's clause", info)
	}
	// Not composed into the ability key: the lender's triggers, statics
	// and cost modifiers are not lent.
	if got := CatalogAbilityKey(c); got != bwOozeOracle {
		t.Errorf("CatalogAbilityKey = %q, want the ooze's own key", got)
	}
}

// Two identical lenders lend the row twice (CR 113.2c), counted by <n>.
func TestTwoIdenticalLendersLendTwoInstances(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	bwElfInGraveyard(g, me)
	bwElfInGraveyard(g, them)
	ooze := gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
	rows := borrowedMana(layeredBattlefieldCard(t, g, ooze))
	if len(rows) != 2 {
		t.Fatalf("lent mana rows = %d, want 2", len(rows))
	}
	if rows[0].Ref != "borrow:"+bwElfOracle+":0:0" || rows[1].Ref != "borrow:"+bwElfOracle+":0:1" {
		t.Errorf("refs = %q, %q", rows[0].Ref, rows[1].Ref)
	}
}

// Except leaves out a kind of row when the entry is written.
func TestBorrowExceptLeavesOutAKindOfRow(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bwElfInGraveyard(g, me)
	sharkey := gaPush(g, me, "Sharkey", "Creature — Halfling", bwSharkeyOracle)
	fence := gaPush(g, me, "Fence", "Creature — Rogue", bwFenceOracle)
	s := layeredBattlefieldCard(t, g, sharkey)
	if n := len(borrowedMana(s)); n != 0 {
		t.Errorf("except mana: %d mana rows lent", n)
	}
	if n := len(borrowedActivated(s)); n != 2 {
		t.Errorf("except mana: %d activated rows lent, want 2", n)
	}
	f := layeredBattlefieldCard(t, g, fence)
	abs := ActivatedAbilitiesForCard(f)
	if len(abs) != 1 || abs[0].Label != bwGainLabel {
		t.Errorf("except loyalty: rows = %+v, want the gain row alone", abs)
	}
	if n := len(borrowedMana(f)); n != 1 {
		t.Errorf("except loyalty: %d mana rows, want 1", n)
	}
}

// Mirran Safehouse: a land card in a graveyard lends its CR 305.6
// intrinsic mana ability, read off its printed land type.
func TestIntrinsicLandRowIsLentFromAGraveyard(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	forest := bwToGraveyard(g, me, "Forest", "Basic Land — Forest", "")
	safe := gaPush(g, me, "Safehouse", "Artifact Land", bwSafehouseOracle)
	gaSettle(g, safe)
	c := layeredBattlefieldCard(t, g, safe)
	rows := borrowedMana(c)
	if len(rows) != 1 || rows[0].Ref != "borrow:land:G:0" || rows[0].LentBy != forest || rows[0].Lent.Kind != LentLand {
		t.Fatalf("lent rows = %+v, want the Forest's {G}", rows)
	}
	idx := len(ManaAbilitiesForCard(c)) - 1
	if err := g.ActivateManaAbility(me, safe, idx, ManaAbilityParams{Ref: rows[0].Ref}); err != nil {
		t.Fatalf("activate the lent land row: %v", err)
	}
	if n := len(g.Seats[0].ManaPool); n != 1 || g.Seats[0].ManaPool[0].Color != "G" {
		t.Errorf("pool = %+v, want one {G}", g.Seats[0].ManaPool)
	}
}

// A face-down lender lends nothing (CR 708.2a), and an uncatalogued one
// has nothing the engine knows to lend.
func TestFaceDownAndUncataloguedLendersLendNothing(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	id := bwElfInGraveyard(g, me)
	g.WithWriteLock(func() {
		g.findCardByIDLocked(id).FaceDown = true
	})
	bwToGraveyard(g, me, "Vanilla", "Creature — Bear", "")
	ooze := gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
	c := layeredBattlefieldCard(t, g, ooze)
	if n := len(borrowedActivated(c)) + len(borrowedMana(c)); n != 0 {
		t.Errorf("%d rows lent by a face-down and an uncatalogued lender", n)
	}
}

// --- activation ----------------------------------------------------------

// CR 302.6 and CR 602.2 read the BORROWER: a lent {T} row taps it, is
// refused while it is summoning sick and allowed with haste; its
// controller activates it, and the lender's owner cannot.
func TestLentTapRowIsTheBorrowers(t *testing.T) {
	fx := stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	bwElfInGraveyard(g, them)
	sick := gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
	hasty := pushTypedTestCard(g, Card{Name: "Hasty Ooze", TypeLine: "Creature — Ooze", OracleID: bwOozeOracle,
		Owner: me, Controller: me, Keywords: []string{"haste"}})
	g.BumpLayerVersionForTest()

	if err := g.ActivateCatalogAbility(me, sick, 0, ActivateAbilityParams{}); !errors.Is(err, ErrSummoningSick) {
		t.Errorf("sick borrower: err = %v, want ErrSummoningSick", err)
	}
	if err := g.ActivateCatalogAbility(them, hasty, 0, ActivateAbilityParams{}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Errorf("the lender's owner: err = %v, want ErrCardCallerMismatch", err)
	}
	_, origins := ActivatedAbilitiesWithOrigins(layeredBattlefieldCard(t, g, hasty))
	if err := g.ActivateCatalogAbility(me, hasty, 0, ActivateAbilityParams{Ref: origins.Ref(0)}); err != nil {
		t.Fatalf("hasty borrower: %v", err)
	}
	if !layeredBattlefieldCard(t, g, hasty).Tapped {
		t.Error("the lent {T} did not tap the borrower")
	}
	item := bwStackItemFrom(g, hasty)
	if item == nil {
		t.Fatal("no stack item sourced from the borrower")
	}
	if item.Controller != me {
		t.Errorf("item controller = %v, want the borrower's controller", item.Controller)
	}
	// CR 201.5b: the label names the borrower where it named the lender.
	if want := strings.ReplaceAll(bwGainLabel, bwElfName, "Hasty Ooze"); item.Label != want {
		t.Errorf("item label = %q, want %q", item.Label, want)
	}
	before := g.Seats[0].Life
	resolveRefAbility(g)
	if fx.gained != 1 || g.Seats[0].Life != before+1 {
		t.Errorf("the lent row resolved %d times, life %d -> %d", fx.gained, before, g.Seats[0].Life)
	}
}

// A lent mana row produces mana for the borrower.
func TestLentManaRowProducesMana(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bwElfInGraveyard(g, me)
	ooze := gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
	gaSettle(g, ooze)
	_, mo := ManaAbilitiesWithOrigins(layeredBattlefieldCard(t, g, ooze))
	if err := g.ActivateManaAbility(me, ooze, 0, ManaAbilityParams{Ref: mo.Ref(0)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := len(g.Seats[0].ManaPool); n != 1 || g.Seats[0].ManaPool[0].Color != "G" {
		t.Errorf("pool = %+v, want one {G}", g.Seats[0].ManaPool)
	}
	if !layeredBattlefieldCard(t, g, ooze).Tapped {
		t.Error("the lent mana row did not tap the borrower")
	}
}

// A lender leaving between the view and the announcement: the stale ref
// is refused and nothing is paid. A second identical lender keeps a
// first-instance ref valid — the same ability.
func TestLenderLeavingStalesTheRef(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	elfA := bwElfInGraveyard(g, me)
	elfB := bwElfInGraveyard(g, me)
	ooze := gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
	gaSettle(g, ooze)
	rows := borrowedMana(layeredBattlefieldCard(t, g, ooze))
	if len(rows) != 2 {
		t.Fatalf("fixture: %d lent rows", len(rows))
	}
	first, second := rows[0].Ref, rows[1].Ref
	bwExile(t, g, me, elfA)
	if err := g.ActivateManaAbility(me, ooze, 1, ManaAbilityParams{Ref: second}); !errors.Is(err, ErrStaleAbilityRef) {
		t.Fatalf("second instance gone: err = %v, want ErrStaleAbilityRef", err)
	}
	c := layeredBattlefieldCard(t, g, ooze)
	if c.Tapped || len(g.Seats[0].ManaPool) != 0 {
		t.Fatal("a refused stale activation paid something")
	}
	if err := g.ActivateManaAbility(me, ooze, 0, ManaAbilityParams{Ref: first}); err != nil {
		t.Fatalf("the remaining identical lender: %v", err)
	}
	// And with no lender left at all the row is gone.
	bwExile(t, g, me, elfB)
	if n := len(borrowedMana(layeredBattlefieldCard(t, g, ooze))); n != 0 {
		t.Errorf("%d rows lent by an empty graveyard", n)
	}
	if err := g.ActivateCatalogAbility(me, ooze, 0, ActivateAbilityParams{Ref: "borrow:" + bwElfOracle + ":0:0"}); !errors.Is(err, ErrStaleAbilityRef) {
		t.Errorf("no lender: err = %v, want ErrStaleAbilityRef", err)
	}
}

// CR 113.7a: an activated borrowed ability exists apart from its source
// and its lender — it resolves after the lender has left, and a restore
// point taken while it waits rebuilds it from the lender's catalog row.
func TestLentAbilityOnTheStackResolvesAndRestoresWithoutItsLender(t *testing.T) {
	fx := stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	elf := bwElfInGraveyard(g, me)
	ooze := gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
	gaSettle(g, ooze)
	_, origins := ActivatedAbilitiesWithOrigins(layeredBattlefieldCard(t, g, ooze))
	if err := g.ActivateCatalogAbility(me, ooze, 0, ActivateAbilityParams{Ref: origins.Ref(0)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	item := bwStackItemFrom(g, ooze)
	want := AbilityRef{Key: bwElfOracle, Slot: AbilitySlotActivated, Ref: "own:0", Name: bwGainLabel}
	if item == nil || item.Body != CatalogActivatedBodyKey || item.Params.Ability == nil || *item.Params.Ability != want {
		t.Fatalf("item stamp = %+v, want the lender's row %+v", item, want)
	}
	bwExile(t, g, me, elf)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a lent ability on the stack blocks the restore point: %+v", snap.Continuations)
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	back := restoredStackItem(t, restored)
	if back.Effect == nil {
		t.Fatal("restore did not re-derive the lent row")
	}
	before := restored.Seats[0].Life
	resolveRefAbility(restored)
	if restored.Seats[0].Life != before+1 {
		t.Errorf("the restored lent ability resolved to life %d, want %d", restored.Seats[0].Life, before+1)
	}
	// And the live one resolves too, its lender long gone.
	resolveRefAbility(g)
	if fx.gained != 2 {
		t.Errorf("resolutions = %d, want 2 (restored + live)", fx.gained)
	}
}

// --- layer 6 ordering (ADR 0155 Decision 4) -----------------------------

// A removal on the borrower's SOURCE ends the borrow whatever the
// timestamps (CR 613.8a): Darksteel Mutation on Necrotic Ooze.
func TestRemovalOnTheBorrowersSourceWinsWhateverTheTimestamps(t *testing.T) {
	for _, order := range []string{"silencer first", "silencer last"} {
		t.Run(order, func(t *testing.T) {
			fx := stubBorrowedAbilityCatalog(t)
			g := newActiveGame(t)
			me := g.Seats[0].ID
			bwElfInGraveyard(g, me)
			if order == "silencer first" {
				gaPush(g, me, "Silencer", "Enchantment", bwSilencerOracle)
				fx.victim = gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
			} else {
				fx.victim = gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
				gaPush(g, me, "Silencer", "Enchantment", bwSilencerOracle)
			}
			g.BumpLayerVersionForTest()
			c := layeredBattlefieldCard(t, g, fx.victim)
			if n := len(borrowedActivated(c)) + len(borrowedMana(c)); n != 0 {
				t.Errorf("a silenced borrower still borrows %d rows", n)
			}
		})
	}
}

// A removal on a recipient that is NOT the source keeps timestamp order
// (CR 613.6): a later "loses all abilities" takes the lent rows, an
// earlier one does not.
func TestRemovalOnARecipientFollowsTimestamps(t *testing.T) {
	t.Run("borrow first, removal later: gone", func(t *testing.T) {
		fx := stubBorrowedAbilityCatalog(t)
		g := newActiveGame(t)
		me := g.Seats[0].ID
		fx.lenders = []uuid.UUID{bwElfInGraveyard(g, me)}
		fx.victim = gaBear(g, me)
		gaPush(g, me, "Cauldron", "Artifact", bwCauldronOracle)
		gaPush(g, me, "Silencer", "Enchantment", bwSilencerOracle)
		c := layeredBattlefieldCard(t, g, fx.victim)
		if n := len(borrowedActivated(c)) + len(borrowedMana(c)); n != 0 {
			t.Errorf("an earlier borrow survived a later removal: %d rows", n)
		}
	})
	t.Run("removal first, borrow later: kept", func(t *testing.T) {
		fx := stubBorrowedAbilityCatalog(t)
		g := newActiveGame(t)
		me := g.Seats[0].ID
		fx.lenders = []uuid.UUID{bwElfInGraveyard(g, me)}
		fx.victim = gaBear(g, me)
		gaPush(g, me, "Silencer", "Enchantment", bwSilencerOracle)
		gaPush(g, me, "Cauldron", "Artifact", bwCauldronOracle)
		c := layeredBattlefieldCard(t, g, fx.victim)
		if !c.HasLostAllAbilities() {
			t.Fatal("fixture: the bear should have lost its own abilities")
		}
		if n := len(borrowedActivated(c)); n != 2 {
			t.Errorf("a later borrow lent %d activated rows past an earlier removal, want 2", n)
		}
	})
}

// A grant and a removal on a battlefield LENDER are visible to the borrow
// whatever the timestamps (CR 613.8a): the borrow reads finished rows.
func TestGrantAndRemovalOnALenderAreSeenWhateverTheTimestamps(t *testing.T) {
	for _, order := range []string{"borrower first", "borrower last"} {
		t.Run("grant/"+order, func(t *testing.T) {
			fx := stubBorrowedAbilityCatalog(t)
			g := newActiveGame(t)
			me := g.Seats[0].ID
			var marvin uuid.UUID
			if order == "borrower first" {
				marvin = gaPush(g, me, "Marvin", "Creature — Toy", bwMarvinOracle)
			}
			fx.grantTarget = gaBear(g, me)
			gaPush(g, me, "Rite", "Enchantment", bwRiteOracle)
			if order == "borrower last" {
				marvin = gaPush(g, me, "Marvin", "Creature — Toy", bwMarvinOracle)
			}
			rows := borrowedActivated(layeredBattlefieldCard(t, g, marvin))
			if len(rows) != 1 || rows[0].Lent.Key != GrantKey(bwPingBundle) || rows[0].LentBy != fx.grantTarget {
				t.Errorf("lent rows = %+v, want the bear's granted ping", rows)
			}
		})
		t.Run("removal/"+order, func(t *testing.T) {
			fx := stubBorrowedAbilityCatalog(t)
			g := newActiveGame(t)
			me := g.Seats[0].ID
			var marvin uuid.UUID
			if order == "borrower first" {
				marvin = gaPush(g, me, "Marvin", "Creature — Toy", bwMarvinOracle)
			}
			fx.victim = gaPush(g, me, bwElfName, "Creature — Elf", bwElfOracle)
			gaPush(g, me, "Silencer", "Enchantment", bwSilencerOracle)
			if order == "borrower last" {
				marvin = gaPush(g, me, "Marvin", "Creature — Toy", bwMarvinOracle)
			}
			c := layeredBattlefieldCard(t, g, marvin)
			if n := len(borrowedActivated(c)) + len(borrowedMana(c)); n != 0 {
				t.Errorf("a silenced lender still lends %d rows", n)
			}
		})
	}
}

// Two mutual borrowers are a CR 613.8b loop, settled in timestamp order:
// the earlier takes the other's own row; the later takes the earlier's
// rows, its own lent back included (CR 113.2c, owner decision 4). Nothing
// grows across passes.
func TestMutualBorrowersSettleInTimestampOrder(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	a := gaPush(g, me, "Kraj A", "Creature — Ooze", bwKrajOracle)
	b := gaPush(g, me, "Kraj B", "Creature — Ooze", bwKrajOracle)
	check := func(pass string) {
		ra := borrowedActivated(layeredBattlefieldCard(t, g, a))
		rb := borrowedActivated(layeredBattlefieldCard(t, g, b))
		if len(ra) != 1 || ra[0].LentBy != b {
			t.Errorf("%s: earlier Kraj lent rows = %+v, want B's own row", pass, ra)
		}
		if len(rb) != 2 {
			t.Errorf("%s: later Kraj lent rows = %d, want A's own and B's own lent back", pass, len(rb))
		}
	}
	check("first pass")
	g.BumpLayerVersionForTest()
	g.WithWriteLock(g.RecomputeLayersIfStaleLocked)
	g.BumpLayerVersionForTest()
	check("third pass")
}

// A borrow ordered among borrows: B before A when B's recipients include
// A's lender, whatever the timestamps. Marvin borrowing from a Necrotic
// Ooze sees the Ooze's lent rows even when Marvin is older.
func TestABorrowerLendsWhatItBorrowed(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	marvin := gaPush(g, me, "Marvin", "Creature — Toy", bwMarvinOracle)
	elf := bwElfInGraveyard(g, me)
	gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
	rows := borrowedActivated(layeredBattlefieldCard(t, g, marvin))
	if len(rows) != 2 || rows[0].LentBy != elf || rows[0].Lent.Key != bwElfOracle {
		t.Errorf("Marvin's lent rows = %+v, want the elf's two, re-lent by the Ooze", rows)
	}
}

// CR 707.2: a copy takes the static, never the lent rows.
func TestACopyOfABorrowerCopiesTheStaticNotTheRows(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bwElfInGraveyard(g, me)
	ooze := gaPush(g, me, "Necrotic Ooze", "Creature — Ooze", bwOozeOracle)
	c := layeredBattlefieldCard(t, g, ooze)
	v := CopiableValuesOf(c)
	if v.OracleID != bwOozeOracle {
		t.Fatalf("copiable oracle = %q", v.OracleID)
	}
	// The copy borrows through its own static, from the same lenders.
	copyID := pushTypedTestCard(g, Card{Name: "Ooze Token", TypeLine: "Creature — Ooze", OracleID: v.OracleID, Owner: me, Controller: me})
	if n := len(borrowedActivated(layeredBattlefieldCard(t, g, copyID))); n != 2 {
		t.Errorf("the copy borrows %d rows through its own static, want 2", n)
	}
}

// --- the riders (ADR 0155 Decision 7) ------------------------------------

// "You may spend mana as though it were mana of any color to activate
// those abilities" covers the rows that borrow lent, and nothing else.
func TestAnyColorRiderCoversOnlyTheBorrowsRows(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bwElfInGraveyard(g, me)
	mairsil := gaPush(g, me, "Mairsil", "Creature — Vampire", bwMairsilOracle)
	c := layeredBattlefieldCard(t, g, mairsil)
	_, origins := ActivatedAbilitiesWithOrigins(c)
	cost, err := ParseCost("{G}")
	if err != nil {
		t.Fatal(err)
	}
	lent := g.CostAsPaidBy(me, ManaSpendForAbilityRow(c, origins.At(0)), cost, 0)
	own := g.CostAsPaidBy(me, ManaSpendForAbility(c), cost, 0)
	if len(lent.Required) != 1 || !lent.Required[0].AnyMana {
		t.Errorf("a lent row's {G} is not widened: %+v", lent.Required)
	}
	if len(own.Required) != 1 || own.Required[0].AnyMana {
		t.Errorf("an unlent payment was widened: %+v", own.Required)
	}
}

// "Activated abilities of [lenders] can't be activated" looks at the
// SOURCE object, the lender: the borrower's copy is not stopped.
func TestLendersActivationGateDoesNotStopTheBorrower(t *testing.T) {
	fx := stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	marvin := gaPush(g, me, "Marvin", "Creature — Toy", bwMarvinOracle)
	fx.forbidden = gaPush(g, me, bwElfName, "Creature — Elf", bwElfOracle)
	gaPush(g, me, "Drana", "Creature — Vampire", bwDranaOracle)
	gaSettle(g, marvin, fx.forbidden)
	if err := g.ActivateManaAbility(me, fx.forbidden, 0, ManaAbilityParams{}); err == nil {
		t.Error("the lender's own mana ability was activated under the gate")
	}
	_, mo := ManaAbilitiesWithOrigins(layeredBattlefieldCard(t, g, marvin))
	if err := g.ActivateManaAbility(me, marvin, 0, ManaAbilityParams{Ref: mo.Ref(0)}); err != nil {
		t.Errorf("the borrower's lent copy was stopped: %v", err)
	}
}

// Mairsil's "each of those abilities only once each turn" counts two
// identical lent rows as one ability, and the turn resets it.
func TestEachOncePerTurnCountsIdenticalRowsAsOne(t *testing.T) {
	stubBorrowedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bwElfInGraveyard(g, me)
	bwElfInGraveyard(g, me)
	mairsil := gaPush(g, me, "Mairsil", "Creature — Vampire", bwMairsilOracle)
	gaSettle(g, mairsil)
	_, mo := ManaAbilitiesWithOrigins(layeredBattlefieldCard(t, g, mairsil))
	if len(mo) != 2 {
		t.Fatalf("fixture: %d lent mana rows", len(mo))
	}
	if err := g.ActivateManaAbility(me, mairsil, 0, ManaAbilityParams{Ref: mo.Ref(0)}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	g.WithWriteLock(func() {
		g.findCardByIDLocked(mairsil).Tapped = false
	})
	if err := g.ActivateManaAbility(me, mairsil, 1, ManaAbilityParams{Ref: mo.Ref(1)}); !errors.Is(err, ErrBorrowedAbilityUsed) {
		t.Fatalf("the identical row: err = %v, want ErrBorrowedAbilityUsed", err)
	}
	var spent bool
	g.ReadSnapshot(func() {
		spent = g.BorrowedRowSpentLocked(mairsil, mo.At(1))
	})
	if !spent {
		t.Error("the gate the view and the enumerator read says the row is unspent")
	}
	g.WithWriteLock(g.resetActivationTurnTallyLocked)
	if err := g.ActivateManaAbility(me, mairsil, 1, ManaAbilityParams{Ref: mo.Ref(1)}); err != nil {
		t.Errorf("next turn: %v", err)
	}
}

// --- the cost on a board with no borrow, and with one --------------------

// BenchmarkBorrowedAbilityLayerPass is ADR 0067's measurement for this
// seam: a 40-permanent board, recomputed and read, with and without a
// battlefield borrower (Marvin over forty creatures, ten of them lenders).
//
// Run with `go test ./internal/game -run XXX -bench BorrowedAbilityLayerPass -benchmem`.
func BenchmarkBorrowedAbilityLayerPass(b *testing.B) {
	for _, withBorrower := range []bool{false, true} {
		name := "no-borrower"
		if withBorrower {
			name = "battlefield-borrower"
		}
		setup := func(b *testing.B) *Game {
			stubBorrowedAbilityCatalog(b)
			g := newActiveGameWithSeats(b, 2)
			me := g.Seats[0].ID
			for i := 0; i < 40; i++ {
				oracle := ""
				if i%4 == 0 {
					oracle = bwElfOracle
				}
				pushTypedTestCard(g, Card{Name: "Body", TypeLine: "Creature — Elf", OracleID: oracle, Owner: me, Controller: me})
			}
			if withBorrower {
				pushTypedTestCard(g, Card{Name: "Marvin", TypeLine: "Creature — Toy", OracleID: bwMarvinOracle, Owner: me, Controller: me})
			}
			g.WithWriteLock(g.RecomputeLayersIfStaleLocked)
			return g
		}
		b.Run(name+"/recompute", func(b *testing.B) {
			g := setup(b)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.WithWriteLock(g.recomputeLayersLocked)
			}
		})
		b.Run(name+"/rows", func(b *testing.B) {
			g := setup(b)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.ReadSnapshot(func() {
					for _, c := range g.Battlefield.Cards {
						_, _ = ManaAbilitiesWithOrigins(c)
						_, _ = ActivatedAbilitiesWithOrigins(c)
					}
				})
			}
		})
	}
}
