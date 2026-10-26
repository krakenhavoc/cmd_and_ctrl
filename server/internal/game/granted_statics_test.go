package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// granted_statics_test.go pins ADR 0093's amendment of 2026-10-15
// (#3784): a static a layer-6 grant gives a permanent may be a layer-6
// static, which joins the layer-6 bucket at CR 613.7a's timestamp the
// moment its grant applies, or a layer-7b/7c/7e one, gathered after
// layer 6 at the same timestamp. Driven through stubbed catalog hooks so
// the engine contract is pinned without any particular card; the cards
// (Dancer's Chakrams, the Kaldheim Runes, Giant's Amulet) are pinned in
// cards/effects/granted_layer6_statics_cards_test.go.

const (
	gsGrantorOracle = "fixture-granted-statics-grantor"
	gsBundle        = "fixture/granted-statics"
	gsHostName      = "Host"
)

// gsLifelinkToOthers is the bundle's layer-6 static: "Other creatures
// you control have lifelink", read with the HOST as its source.
var gsLifelinkToOthers = StaticAbility{
	Label: "Other creatures you control have lifelink.",
	Layer: Layer6Ability,
	AppliesTo: func(target *Card, _ *Game, source *Card) bool {
		return target.InstanceID != source.InstanceID &&
			target.Controller == source.Controller && target.IsCreature()
	},
	Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
		c.Abilities = AppendKeywordAbility(c.Abilities, "lifelink")
	},
}

// gsBasePowerZero is a granted layer-7b static: "This creature has base
// power 0."
var gsBasePowerZero = StaticAbility{
	Label:     "This creature has base power 0.",
	Layer:     Layer7PT,
	SubLayer:  SubLayer7B_Set,
	AppliesTo: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID },
	Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Power = 0 },
}

// stubGrantedStaticCatalog wires a grantor whose layer-6 static gives
// every permanent named "Host" the fixture bundle, and the bundle
// itself, holding `bundle`.
func stubGrantedStaticCatalog(t *testing.T, bundle ...StaticAbility) {
	t.Helper()
	defs := map[string]*CardDef{
		gsGrantorOracle: {Static: []StaticAbility{{
			Label:          "Hosts have the fixture bundle.",
			Layer:          Layer6Ability,
			AppliesTo:      func(target *Card, _ *Game, _ *Card) bool { return target.Name == gsHostName },
			GrantAbilities: []string{gsBundle},
		}}},
		GrantKey(gsBundle): {Static: bundle, GrantText: "the fixture bundle"},
	}
	prevLookup, prevStatics := CatalogLookup, CatalogStaticAbilities
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	CatalogStaticAbilities = func(key string) []StaticAbility {
		if d := defs[key]; d != nil {
			return d.Static
		}
		return nil
	}
	t.Cleanup(func() { CatalogLookup, CatalogStaticAbilities = prevLookup, prevStatics })
}

func gsCreature(g *Game, owner uuid.UUID, name string) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: owner, Controller: owner,
	})
}

func gsGrantor(g *Game, owner uuid.UUID) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name: "Grantor", TypeLine: "Artifact", OracleID: gsGrantorOracle,
		Owner: owner, Controller: owner,
	})
}

func gsHasLifelink(t *testing.T, g *Game, id uuid.UUID) bool {
	t.Helper()
	for _, a := range scopedEffectChar(t, g, id).Abilities {
		if a == "lifelink" {
			return true
		}
	}
	return false
}

func TestAGrantedLayer6StaticAppliesThroughItsHost(t *testing.T) {
	stubGrantedStaticCatalog(t, gsLifelinkToOthers)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := gsCreature(g, me.ID, gsHostName)
	friend := gsCreature(g, me.ID, "Friend")
	enemy := gsCreature(g, opp.ID, "Enemy")
	gsGrantor(g, me.ID)

	if !gsHasLifelink(t, g, friend) {
		t.Error("the host's granted static did not give its controller's other creature lifelink")
	}
	if gsHasLifelink(t, g, host) {
		t.Error("\"other creatures\" reached the host itself")
	}
	if gsHasLifelink(t, g, enemy) {
		t.Error("\"you control\" reached an opponent's creature")
	}
}

// CR 613.8a: a removal on the HOST that sorts after the grant ends the
// granted static's existence, so it waits for that removal even though
// its own timestamp is earlier.
func TestALaterRemovalOnTheHostEndsItsGrantedLayer6Static(t *testing.T) {
	stubGrantedStaticCatalog(t, gsLifelinkToOthers)
	g := newActiveGame(t)
	me := g.Seats[0]
	host := gsCreature(g, me.ID, gsHostName)
	friend := gsCreature(g, me.ID, "Friend")
	gsGrantor(g, me.ID)
	if !gsHasLifelink(t, g, friend) {
		t.Fatal("setup: no lifelink before the removal")
	}
	registerScopedEffectForTest(t, g, host, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	if gsHasLifelink(t, g, friend) {
		t.Error("the host lost all abilities after the grant, but its granted static still gives lifelink")
	}
	if n := len(scopedEffectChar(t, g, host).GrantedAbilities); n != 0 {
		t.Errorf("the host still has %d grants", n)
	}
}

// CR 613.6: a removal on the host that sorts BEFORE the grant leaves it,
// and the static with it.
func TestAnEarlierRemovalOnTheHostLeavesItsGrantedLayer6Static(t *testing.T) {
	stubGrantedStaticCatalog(t, gsLifelinkToOthers)
	g := newActiveGame(t)
	me := g.Seats[0]
	host := gsCreature(g, me.ID, gsHostName)
	friend := gsCreature(g, me.ID, "Friend")
	registerScopedEffectForTest(t, g, host, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	gsGrantor(g, me.ID)
	if !gsHasLifelink(t, g, friend) {
		t.Error("the grant came after the removal and survives it, but its static does not apply")
	}
}

// CR 613.7: inside layer 6 the granted static sorts against a removal on
// its RECIPIENT by timestamp. A removal before the grant is overridden
// by the lifelink; one after it takes the lifelink away. Applying the
// granted static after the whole bucket would get the second wrong.
func TestAGrantedLayer6StaticSortsAgainstARemovalOnItsRecipient(t *testing.T) {
	t.Run("removal first", func(t *testing.T) {
		stubGrantedStaticCatalog(t, gsLifelinkToOthers)
		g := newActiveGame(t)
		me := g.Seats[0]
		gsCreature(g, me.ID, gsHostName)
		friend := gsCreature(g, me.ID, "Friend")
		registerScopedEffectForTest(t, g, friend, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
		gsGrantor(g, me.ID)
		if !gsHasLifelink(t, g, friend) {
			t.Error("a static granted after the recipient lost its abilities should still give it lifelink")
		}
	})
	t.Run("removal last", func(t *testing.T) {
		stubGrantedStaticCatalog(t, gsLifelinkToOthers)
		g := newActiveGame(t)
		me := g.Seats[0]
		gsCreature(g, me.ID, gsHostName)
		friend := gsCreature(g, me.ID, "Friend")
		gsGrantor(g, me.ID)
		registerScopedEffectForTest(t, g, friend, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
		if gsHasLifelink(t, g, friend) {
			t.Error("the recipient lost all abilities after the grant, but kept the granted lifelink")
		}
	})
}

// CR 613.7a: a granted static's timestamp is the grant's when the grant
// is later than the host. A layer-7b set registered between the host's
// entry and the grant is overridden by the granted set.
func TestAGrantedLayer7bStaticTakesTheGrantsLaterTimestamp(t *testing.T) {
	stubGrantedStaticCatalog(t, gsBasePowerZero)
	g := newActiveGame(t)
	me := g.Seats[0]
	host := gsCreature(g, me.ID, gsHostName)
	registerScopedEffectForTest(t, g, host, []Mod{SetBasePowerMod(5)}, IndefiniteDuration())
	if p := scopedEffectChar(t, g, host).Power; p != 5 {
		t.Fatalf("setup: power %d, want 5", p)
	}
	gsGrantor(g, me.ID)
	if p := scopedEffectChar(t, g, host).Power; p != 0 {
		t.Errorf("power %d, want 0: the granted set has the grant's later timestamp", p)
	}
}

func TestGrantedStaticProblemNamesWhatAGrantCannotCarry(t *testing.T) {
	ok := func(target *Card, _ *Game, _ *Card) bool { return true }
	apply := func(*Characteristic, *Card, *Game, *Card) {}
	base := func(l Layer, sub SubLayer) StaticAbility {
		return StaticAbility{Layer: l, SubLayer: sub, AppliesTo: ok, Apply: apply}
	}
	for _, s := range []StaticAbility{
		base(Layer6Ability, 0),
		base(Layer7PT, SubLayer7B_Set),
		base(Layer7PT, SubLayer7C_Modify),
		base(Layer7PT, SubLayer7E_Switch),
	} {
		if p := GrantedStaticProblem(s); p != "" {
			t.Errorf("layer %v/%v refused: %s", s.Layer, s.SubLayer, p)
		}
	}
	removes := base(Layer6Ability, 0)
	removes.RemovesAbilities = true
	grants := base(Layer6Ability, 0)
	grants.GrantAbilities = []string{"x"}
	borrows := base(Layer6Ability, 0)
	borrows.Borrows = &AbilityBorrow{}
	for name, tc := range map[string]struct {
		s    StaticAbility
		want string
	}{
		"layer 4":    {base(Layer4Type, 0), "only in layer 6"},
		"layer 1":    {base(Layer1Copy, 0), "only in layer 6"},
		"7a CDA":     {base(Layer7PT, SubLayer7A_CDA), "only in layer 6"},
		"7d":         {base(Layer7PT, SubLayer7D_Counters), "only in layer 6"},
		"no apply":   {StaticAbility{Layer: Layer6Ability, AppliesTo: ok}, "needs AppliesTo and Apply"},
		"removal":    {removes, "remove, grant or borrow"},
		"nested":     {grants, "remove, grant or borrow"},
		"borrow":     {borrows, "remove, grant or borrow"},
		"zone":       {StaticAbility{Layer: Layer6Ability, AppliesTo: ok, Apply: apply, GrantZone: ZoneHand}, "battlefield"},
		"hint":       {StaticAbility{Layer: Layer6Ability, AppliesTo: ok, Apply: apply, DependsOnLifeTotal: true}, "invalidation hint"},
		"designated": {StaticAbility{Layer: Layer6Ability, AppliesTo: ok, Apply: apply, ActiveWhen: ClassLevel(2)}, "ActiveWhen"},
	} {
		if p := GrantedStaticProblem(tc.s); !strings.Contains(p, tc.want) {
			t.Errorf("%s: problem %q, want one mentioning %q", name, p, tc.want)
		}
	}
}
