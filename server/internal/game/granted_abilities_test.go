package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// granted_abilities_test.go — ADR 0093's seam (the PR 1 half), pinned
// with fixture bundles against a stubbed catalog: CR 113.10 grants of
// mana and activated abilities to OTHER permanents, CR 613.6 removal
// timestamps, CR 613.8a's grantor dependency, CR 305.7, CR 707.2,
// CR 302.6, CR 602.2, and the stale-ref refusal (#544).

const (
	gaRiteOracle     = "fixture-rite"     // "Creatures you control have '{T}: Add {G}.'"
	gaPingerOracle   = "fixture-pinger"   // "Creatures you control have '{1}: this deals 1 damage'"
	gaLanternOracle  = "fixture-lantern"  // "Lands you control have '{T}: Add {C}.'"
	gaMoonOracle     = "fixture-moon"     // layer-4 CR 305.7 removal on every land
	gaSilencerOracle = "fixture-silencer" // layer-6 "loses all abilities" on the named victim
	gaSilencer2      = "fixture-silencer-2"
	gaLordOracle     = "fixture-lord" // "Other creatures have islandwalk"
	gaOwnManaOracle  = "fixture-own-mana"

	gaManaBundle   = "fixture-rite/tap-for-green"
	gaPingBundle   = "fixture-pinger/ping"
	gaColorless    = "fixture-lantern/tap-for-colorless"
	gaTriggerGrant = "fixture/dies-trigger"
)

// gaFixture holds the victims the silencer fixtures point at, so one
// catalog can serve every test in this file.
type gaFixture struct {
	victim  uuid.UUID
	victim2 uuid.UUID
	pinged  int
}

func creaturesYouControlFixture(target *Card, _ *Game, source *Card) bool {
	return target.IsCreature() && target.Controller == source.Controller
}

// stubGrantedAbilityCatalog wires a catalog of grantors and bundles and
// returns the fixture the silencers read.
func stubGrantedAbilityCatalog(t testing.TB) *gaFixture {
	t.Helper()
	fx := &gaFixture{}
	silence := func(victim *uuid.UUID) []StaticAbility {
		return []StaticAbility{{
			Layer:            Layer6Ability,
			RemovesAbilities: true,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.InstanceID == *victim
			},
		}}
	}
	defs := map[string]*CardDef{
		gaRiteOracle: {Static: []StaticAbility{{
			Layer:          Layer6Ability,
			AppliesTo:      creaturesYouControlFixture,
			GrantAbilities: []string{gaManaBundle},
		}}},
		gaPingerOracle: {Static: []StaticAbility{{
			Layer:          Layer6Ability,
			AppliesTo:      creaturesYouControlFixture,
			GrantAbilities: []string{gaPingBundle, gaTriggerGrant},
		}}},
		gaLanternOracle: {Static: []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.IsLand() && target.Controller == source.Controller
			},
			GrantAbilities: []string{gaColorless},
		}}},
		gaMoonOracle: {Static: []StaticAbility{{
			Layer:            Layer4Type,
			RemovesAbilities: true,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.IsLand()
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.SetSubtypes([]string{"Mountain"})
			},
		}}},
		gaSilencerOracle: {Static: silence(&fx.victim)},
		gaSilencer2:      {Static: silence(&fx.victim2)},
		gaLordOracle: {Static: []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.IsCreature() && target.InstanceID != source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = AppendKeywordAbility(c.Abilities, "islandwalk")
			},
		}}},
		gaOwnManaOracle: {ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{C}", Label: "Add {C}"}}},
		GrantKey(gaManaBundle): {
			ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
			GrantText:     "{T}: Add {G}.",
		},
		GrantKey(gaColorless): {
			ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{C}", Label: "Add {C}"}},
			GrantText:     "{T}: Add {C}.",
		},
		GrantKey(gaPingBundle): {
			Activated: []ActivatedAbilityShape{{
				Label: "{1}: ping",
				Effect: func(*Game, *StackItem) error {
					fx.pinged++
					return nil
				},
			}},
			GrantText: "{1}: This creature deals 1 damage to any target.",
		},
		GrantKey(gaTriggerGrant): {
			Triggered: []TriggeredAbility{{Watches: []EventKind{EventLTB}}},
			GrantText: "When this creature dies, draw a card.",
		},
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
	return fx
}

func gaPush(g *Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner,
	})
}

func gaBear(g *Game, owner uuid.UUID) uuid.UUID {
	return gaPush(g, owner, "Bear", "Creature — Bear", "")
}

// gaSettle clears the summoning sickness the entry listener stamped, so
// a {T} fixture can be activated this turn.
func gaSettle(g *Game, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			for _, id := range ids {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].SummonedThisTurn = false
				}
			}
		}
	})
}

// grantedManaRows returns the granted rows of a card's mana list.
func grantedManaRows(c Card) []AbilityOrigin {
	_, origins := ManaAbilitiesWithOrigins(c)
	var out []AbilityOrigin
	for _, o := range origins {
		if o.Granted() {
			out = append(out, o)
		}
	}
	return out
}

// --- the grant reaches the recipient ----------------------------------

// An uncatalogued bear — the common case, an imported vanilla creature —
// gets the granted mana ability, labelled with its grantor, under a
// composite ability key with an EMPTY base (ADR 0093 Decision 2 §1),
// while its identity key does not move (§3).
func TestGrantReachesAnUncataloguedCreature(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := gaBear(g, me)
	rite := gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)

	c := layeredBattlefieldCard(t, g, bear)
	abs, origins := ManaAbilitiesWithOrigins(c)
	if len(abs) != 1 || abs[0].Produced != "{G}" {
		t.Fatalf("bear's mana abilities = %+v, want the granted {G}", abs)
	}
	o := origins.At(0)
	if !o.Granted() || o.GrantedBy != rite || o.Ref != GrantedAbilityRef(gaManaBundle, 0, 0) {
		t.Errorf("origin = %+v, want the Rite's grant", o)
	}
	if got := CatalogAbilityKey(c); got != "|"+GrantKey(gaManaBundle) {
		t.Errorf("CatalogAbilityKey = %q, want the grant-only composite", got)
	}
	if got := CatalogKey(c); got != "" {
		t.Errorf("CatalogKey moved to %q: a granted ability is not a catalog card", got)
	}
	if IsCatalogCard != nil && IsCatalogCard(CatalogKey(c)) {
		t.Error("the bear reads as a catalog card because an enchantment is on the table")
	}
	info := GrantedAbilitiesOf(c)
	if len(info) != 1 || info[0].Text != "{T}: Add {G}." || info[0].Source != rite {
		t.Errorf("GrantedAbilitiesOf = %+v", info)
	}

	// An opponent's creature is not "a creature you control".
	theirs := gaBear(g, g.Seats[1].ID)
	if rows := grantedManaRows(layeredBattlefieldCard(t, g, theirs)); len(rows) != 0 {
		t.Errorf("the opponent's bear got the grant: %+v", rows)
	}
}

// CR 708.2a silences the card's TEXT, not the effects on it: a
// face-down 2/2 under the grant has the granted ability.
func TestGrantReachesAFaceDownCreature(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Hidden", TypeLine: "Creature — Elf", OracleID: gaOwnManaOracle,
		Owner: me, Controller: me, FaceDown: true, FaceDownKind: FaceDownManifested,
	})
	gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
	c := layeredBattlefieldCard(t, g, id)
	abs := ManaAbilitiesForCard(c)
	if len(abs) != 1 || abs[0].Produced != "{G}" {
		t.Fatalf("face-down creature's mana abilities = %+v, want only the granted {G} (its own {C} is silenced)", abs)
	}
}

// Own abilities come first and keep their indexes; the granted row is
// appended LAST, so a grant appearing renumbers nothing (Decision 5).
func TestGrantedRowsComeAfterOwnRows(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	id := gaPush(g, me, "Mana Elf", "Creature — Elf", gaOwnManaOracle)
	gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
	abs, origins := ManaAbilitiesWithOrigins(layeredBattlefieldCard(t, g, id))
	if len(abs) != 2 || abs[0].Produced != "{C}" || abs[1].Produced != "{G}" {
		t.Fatalf("rows = %+v, want own {C} then granted {G}", abs)
	}
	if origins.Ref(0) != OwnAbilityRef(0) || origins.At(0).Granted() {
		t.Errorf("row 0 origin = %+v, want own:0", origins.At(0))
	}
	if !origins.At(1).Granted() {
		t.Errorf("row 1 origin = %+v, want the grant", origins.At(1))
	}
}

// Two grantors give two instances (CR 113.2c), each with its own ref.
func TestTwoGrantorsGiveTwoInstances(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := gaBear(g, me)
	r1 := gaPush(g, me, "Rite A", "Enchantment", gaRiteOracle)
	r2 := gaPush(g, me, "Rite B", "Enchantment", gaRiteOracle)
	rows := grantedManaRows(layeredBattlefieldCard(t, g, bear))
	if len(rows) != 2 {
		t.Fatalf("granted rows = %+v, want two", rows)
	}
	if rows[0].GrantedBy != r1 || rows[1].GrantedBy != r2 {
		t.Errorf("grantors = %v, %v; want the two Rites in timestamp order", rows[0].GrantedBy, rows[1].GrantedBy)
	}
	if rows[0].Ref == rows[1].Ref {
		t.Errorf("both instances share the ref %q", rows[0].Ref)
	}
}

// --- removal (CR 613.6, CR 613.8a, CR 305.7) -------------------------

// A removal on the RECIPIENT takes a grant that sorted before it, and
// a grant with a later timestamp survives (CR 613.6).
func TestRemovalOnTheRecipientFollowsTimestamps(t *testing.T) {
	t.Run("grant first, removal later: gone", func(t *testing.T) {
		fx := stubGrantedAbilityCatalog(t)
		g := newActiveGame(t)
		me := g.Seats[0].ID
		fx.victim = gaBear(g, me)
		gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
		gaPush(g, me, "Silencer", "Enchantment", gaSilencerOracle)
		c := layeredBattlefieldCard(t, g, fx.victim)
		if rows := grantedManaRows(c); len(rows) != 0 {
			t.Errorf("an earlier grant survived a later removal: %+v", rows)
		}
		if got := CatalogAbilityKey(c); got != "" {
			t.Errorf("CatalogAbilityKey = %q, want empty", got)
		}
	})
	t.Run("removal first, grant later: kept", func(t *testing.T) {
		fx := stubGrantedAbilityCatalog(t)
		g := newActiveGame(t)
		me := g.Seats[0].ID
		fx.victim = gaPush(g, me, "Mana Elf", "Creature — Elf", gaOwnManaOracle)
		gaPush(g, me, "Silencer", "Enchantment", gaSilencerOracle)
		gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
		c := layeredBattlefieldCard(t, g, fx.victim)
		if !c.HasLostAllAbilities() {
			t.Fatal("fixture: the elf should have lost its own abilities")
		}
		abs := ManaAbilitiesForCard(c)
		if len(abs) != 1 || abs[0].Produced != "{G}" {
			t.Errorf("mana = %+v, want the later grant alone — own {C} removed, granted {G} kept", abs)
		}
	})
}

// A removal on the GRANTOR strips its grant from every recipient,
// whatever the timestamps (CR 613.8a(b)).
func TestRemovalOnTheGrantorWinsWhateverTheTimestamps(t *testing.T) {
	for _, order := range []string{"silencer first", "silencer last"} {
		t.Run(order, func(t *testing.T) {
			fx := stubGrantedAbilityCatalog(t)
			g := newActiveGame(t)
			me := g.Seats[0].ID
			bear := gaBear(g, me)
			if order == "silencer first" {
				gaPush(g, me, "Silencer", "Enchantment", gaSilencerOracle)
				fx.victim = gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
			} else {
				fx.victim = gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
				gaPush(g, me, "Silencer", "Enchantment", gaSilencerOracle)
			}
			g.BumpLayerVersionForTest()
			if rows := grantedManaRows(layeredBattlefieldCard(t, g, bear)); len(rows) != 0 {
				t.Errorf("a silenced grantor still grants: %+v", rows)
			}
		})
	}
}

// The same dependency fixes keyword lords (ADR 0093 Decision 3): a
// Lord of Atlantis under a LATER removal used to keep granting
// islandwalk because the lord's effect sorted first.
func TestRemovalOnAKeywordLordWinsWhateverTheTimestamps(t *testing.T) {
	for _, order := range []string{"silencer first", "silencer last"} {
		t.Run(order, func(t *testing.T) {
			fx := stubGrantedAbilityCatalog(t)
			g := newActiveGame(t)
			me := g.Seats[0].ID
			merfolk := gaBear(g, me)
			if order == "silencer first" {
				gaPush(g, me, "Silencer", "Enchantment", gaSilencerOracle)
				fx.victim = gaPush(g, me, "Lord", "Creature — Merfolk", gaLordOracle)
			} else {
				fx.victim = gaPush(g, me, "Lord", "Creature — Merfolk", gaLordOracle)
				gaPush(g, me, "Silencer", "Enchantment", gaSilencerOracle)
			}
			g.BumpLayerVersionForTest()
			if hasKeywordOnBattlefield(t, g, merfolk, "islandwalk") {
				t.Error("a silenced lord still grants islandwalk")
			}
		})
	}
	// And without the silencer the lord does grant it — the fixture
	// is not vacuous.
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	merfolk := gaBear(g, me)
	gaPush(g, me, "Lord", "Creature — Merfolk", gaLordOracle)
	if !hasKeywordOnBattlefield(t, g, merfolk, "islandwalk") {
		t.Fatal("fixture: the lord should grant islandwalk")
	}
}

// Two removals that each apply to the other's source are a dependency
// loop (CR 613.8b): timestamp order decides, so the EARLIER one wins
// and the later one is silenced before it applies.
func TestMutualRemovalsFallBackToTimestampOrder(t *testing.T) {
	fx := stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	first := gaPush(g, me, "Silencer A", "Enchantment", gaSilencerOracle)
	second := gaPush(g, me, "Silencer B", "Enchantment", gaSilencer2)
	fx.victim = second // A silences B
	fx.victim2 = first // B silences A
	g.BumpLayerVersionForTest()
	if !layeredBattlefieldCard(t, g, second).HasLostAllAbilities() {
		t.Error("the earlier silencer did not silence the later one")
	}
	if layeredBattlefieldCard(t, g, first).HasLostAllAbilities() {
		t.Error("the later, silenced silencer still silenced the earlier one")
	}
}

// CR 305.7: a layer-4 basic-land-type change removes the land's own
// abilities before any layer-6 grant, so a Lantern-style grant survives
// a Blood Moon-style type change.
func TestLandTypeRemovalKeepsALayer6Grant(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	land := gaPush(g, me, "Utility Land", "Land", gaOwnManaOracle)
	gaPush(g, me, "Moon", "Enchantment", gaMoonOracle)
	gaPush(g, me, "Lantern", "Artifact", gaLanternOracle)
	abs, origins := ManaAbilitiesWithOrigins(layeredBattlefieldCard(t, g, land))
	var produced []string
	for _, a := range abs {
		produced = append(produced, a.Produced)
	}
	if len(abs) != 2 || abs[0].Produced != "{R}" || abs[1].Produced != "{C}" {
		t.Fatalf("mana = %v, want the Mountain's {R} then the granted {C}", produced)
	}
	if origins.Ref(0) != IntrinsicLandAbilityRef("R") || !origins.At(1).Granted() {
		t.Errorf("refs = %q, %q", origins.Ref(0), origins.Ref(1))
	}
}

// CR 707.2: a layer-6 grant is not a copiable value.
func TestGrantIsNotCopied(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := gaBear(g, me)
	gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
	c := layeredBattlefieldCard(t, g, bear)
	if len(grantedManaRows(c)) != 1 {
		t.Fatal("fixture: the bear should have the grant")
	}
	if v := CopiableValuesOf(c); len(v.GrantedAbilities) != 0 {
		t.Errorf("a copy would take the layered grant: %v", v.GrantedAbilities)
	}
}

// --- activation --------------------------------------------------------

// CR 302.6 reads the HOST: a creature that entered this turn cannot use
// a granted {T} ability, and one with haste can.
func TestGrantedTapAbilityRespectsTheHostsSickness(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	sick := pushTypedTestCard(g, Card{Name: "Sick", TypeLine: "Creature — Bear", Owner: me, Controller: me, SummonedThisTurn: true})
	hasty := pushTypedTestCard(g, Card{Name: "Hasty", TypeLine: "Creature — Bear", Owner: me, Controller: me, SummonedThisTurn: true, Keywords: []string{"haste"}})
	gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
	g.BumpLayerVersionForTest()

	if err := g.ActivateManaAbility(me, sick, 0, ManaAbilityParams{}); !errors.Is(err, ErrSummoningSick) {
		t.Errorf("sick host: err = %v, want ErrSummoningSick", err)
	}
	if err := g.ActivateManaAbility(me, hasty, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("hasty host: %v", err)
	}
	if !layeredBattlefieldCard(t, g, hasty).Tapped {
		t.Error("the granted {T} did not tap the HOST")
	}
	if n := len(g.Seats[0].ManaPool); n != 1 {
		t.Errorf("pool has %d tokens, want the granted {G}", n)
	}
}

// CR 602.2: only the host's controller activates its granted ability;
// the grantor's controller gets nothing.
func TestOnlyTheHostsControllerActivatesAGrant(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	theirBear := gaBear(g, them)
	gaSettle(g, theirBear)
	gaPush(g, them, "Their Rite", "Enchantment", gaRiteOracle)
	g.BumpLayerVersionForTest()
	if err := g.ActivateManaAbility(me, theirBear, 0, ManaAbilityParams{}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Errorf("err = %v, want ErrCardCallerMismatch", err)
	}
	if err := g.ActivateManaAbility(them, theirBear, 0, ManaAbilityParams{}); err != nil {
		t.Errorf("the host's controller could not activate it: %v", err)
	}
}

// A granted ACTIVATED ability goes on the stack with the host as its
// source, and the ref it was announced with is checked.
func TestGrantedActivatedAbilityActivates(t *testing.T) {
	fx := stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := gaBear(g, me)
	gaPush(g, me, "Pinger", "Enchantment", gaPingerOracle)
	g.BumpLayerVersionForTest()

	abs, origins := ActivatedAbilitiesWithOrigins(layeredBattlefieldCard(t, g, bear))
	if len(abs) != 1 || !origins.At(0).Granted() {
		t.Fatalf("activated rows = %+v / %+v, want the granted ping", abs, origins)
	}
	if err := g.ActivateCatalogAbility(me, bear, 0, ActivateAbilityParams{Ref: "grant:somebody-else:0:0"}); !errors.Is(err, ErrStaleAbilityRef) {
		t.Fatalf("wrong ref: err = %v, want ErrStaleAbilityRef", err)
	}
	if err := g.ActivateCatalogAbility(me, bear, 0, ActivateAbilityParams{Ref: origins.Ref(0)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	var item StackItem
	found := false
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it.SourceCardID == bear {
				item, found = *it, true
			}
		}
	})
	if !found {
		t.Fatal("no stack item sourced from the HOST")
	}
	if item.Controller != me {
		t.Errorf("item controller = %v, want the host's controller", item.Controller)
	}
	_ = fx
}

// --- the stale ref (ADR 0093 Decision 5, #544) --------------------------

// A grant vanishing between the view and the announcement: the stale
// ref is refused before anything is paid, the row that moved into its
// index is not fired, and an absent ref is still accepted.
func TestStaleGrantedRefIsRefusedAndCostsNothing(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	elf := gaPush(g, me, "Mana Elf", "Creature — Elf", gaOwnManaOracle)
	gaSettle(g, elf)
	r1 := gaPush(g, me, "Rite A", "Enchantment", gaRiteOracle)
	gaPush(g, me, "Rite B", "Enchantment", gaRiteOracle)
	g.BumpLayerVersionForTest()

	_, origins := ManaAbilitiesWithOrigins(layeredBattlefieldCard(t, g, elf))
	// Rows: own {C}, Rite A's {G}, Rite B's {G}. The view says index 2
	// is Rite B's second instance.
	staleRef := origins.Ref(2)
	if staleRef != GrantedAbilityRef(gaManaBundle, 0, 1) {
		t.Fatalf("fixture: row 2 ref = %q", staleRef)
	}
	// Rite A leaves: the elf's list is now own {C}, Rite B's {G}, and
	// Rite B's instance is the FIRST of its bundle — index 2 does not
	// exist and row 1's ref is n=0.
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me}, r1); err != nil {
		t.Fatalf("move Rite A: %v", err)
	}
	if err := g.ActivateManaAbility(me, elf, 2, ManaAbilityParams{Ref: staleRef}); !errors.Is(err, ErrStaleAbilityRef) {
		t.Fatalf("err = %v, want ErrStaleAbilityRef", err)
	}
	if err := g.ActivateManaAbility(me, elf, 1, ManaAbilityParams{Ref: staleRef}); !errors.Is(err, ErrStaleAbilityRef) {
		t.Fatalf("index 1 with the old ref: err = %v, want ErrStaleAbilityRef", err)
	}
	c := layeredBattlefieldCard(t, g, elf)
	if c.Tapped || len(g.Seats[0].ManaPool) != 0 {
		t.Fatal("a refused stale activation paid something")
	}
	// No ref at all: accepted, as every pre-ref client sends.
	if err := g.ActivateManaAbility(me, elf, 1, ManaAbilityParams{}); err != nil {
		t.Fatalf("absent ref: %v", err)
	}
}

// An own ability's ref is its index in the full declared list, and an
// intrinsic land ability's is its colour; both are accepted when right.
func TestOwnAndIntrinsicRefsAreAccepted(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	forest := pushTypedTestCard(g, Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me, Controller: me})
	g.BumpLayerVersionForTest()
	_, origins := ManaAbilitiesWithOrigins(layeredBattlefieldCard(t, g, forest))
	if origins.Ref(0) != IntrinsicLandAbilityRef("G") {
		t.Fatalf("forest ref = %q", origins.Ref(0))
	}
	if err := g.ActivateManaAbility(me, forest, 0, ManaAbilityParams{Ref: "land:G"}); err != nil {
		t.Fatalf("activate with the right ref: %v", err)
	}
}

// --- triggers and last-known information --------------------------------

// A granted trigger is found through the composed key (the ADR's PR 3
// ships the cards; the seam is here).
func TestGrantedTriggerIsFoundThroughTheComposedKey(t *testing.T) {
	stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := gaBear(g, me)
	gaPush(g, me, "Pinger", "Enchantment", gaPingerOracle)
	c := layeredBattlefieldCard(t, g, bear)
	if n := len(TriggersForCard(c)); n != 1 {
		t.Errorf("TriggersForCard = %d triggers, want the granted one", n)
	}
	lki := c.Effective()
	if got := AbilityKeyFromLKI(c, lki); got != CatalogAbilityKey(c) {
		t.Errorf("AbilityKeyFromLKI = %q, CatalogAbilityKey = %q — the two compositions disagree", got, CatalogAbilityKey(c))
	}
	info := GrantedAbilitiesOf(c)
	if len(info) != 2 {
		t.Errorf("GrantedAbilitiesOf = %+v, want the ping and the trigger", info)
	}
}

// The merged definition is memoised and the memo is validated: a
// swapped catalog is never answered from a stale entry.
func TestMergedDefMemoIsValidated(t *testing.T) {
	a := &CardDef{Activated: []ActivatedAbilityShape{{Label: "a"}}}
	b := &CardDef{Activated: []ActivatedAbilityShape{{Label: "b"}}}
	current := a
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef {
		if key == GrantKey("memo/probe") {
			return current
		}
		return nil
	}
	t.Cleanup(func() { CatalogLookup = prev })
	key := "|" + GrantKey("memo/probe")
	first := catalogDef(key)
	if first == nil || first.Activated[0].Label != "a" {
		t.Fatalf("first = %+v", first)
	}
	if again := catalogDef(key); again != first {
		t.Error("an unchanged catalog was not answered from the memo")
	}
	current = b
	if swapped := catalogDef(key); swapped == nil || swapped.Activated[0].Label != "b" {
		t.Errorf("a swapped catalog was answered from the memo: %+v", swapped)
	}
}

// BenchmarkGrantedAbilityReaders is ADR 0093 Decision 2 §4's benchmark:
// forty creatures, with and without one grant over all of them. Three
// costs, per board:
//
//   - recompute: one full layer pass;
//   - harvest: TriggersForCard on every creature, what the trigger
//     harvest asks per event;
//   - rows: the mana and activated rows with their refs, what the view
//     and the enumerator ask per frame.
//
// Run with `go test ./internal/game -run XXX -bench GrantedAbilityReaders -benchmem`.
func BenchmarkGrantedAbilityReaders(b *testing.B) {
	for _, withGrant := range []bool{false, true} {
		name := "no-grant"
		if withGrant {
			name = "one-grant"
		}
		setup := func(b *testing.B) (*Game, []uuid.UUID) {
			stubGrantedAbilityCatalog(b)
			g := newActiveGameWithSeats(b, 2)
			me := g.Seats[0].ID
			ids := make([]uuid.UUID, 40)
			for i := range ids {
				ids[i] = gaBear(g, me)
			}
			if withGrant {
				gaPush(g, me, "Rite", "Enchantment", gaRiteOracle)
				gaPush(g, me, "Pinger", "Enchantment", gaPingerOracle)
			}
			g.WithWriteLock(g.RecomputeLayersIfStaleLocked)
			return g, ids
		}
		b.Run(name+"/recompute", func(b *testing.B) {
			g, _ := setup(b)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.WithWriteLock(g.recomputeLayersLocked)
			}
		})
		b.Run(name+"/harvest", func(b *testing.B) {
			g, _ := setup(b)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.ReadSnapshot(func() {
					for _, c := range g.Battlefield.Cards {
						_ = TriggersForCard(c)
					}
				})
			}
		})
		b.Run(name+"/rows", func(b *testing.B) {
			g, _ := setup(b)
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
