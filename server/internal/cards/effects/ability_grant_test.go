package effects

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_grant_test.go — the catalog half of ADR 0093's seam: the
// AbilityGrant bundle's mana slot and text, the GrantAbilities
// constructor, the boot-time refusals, and TestEveryGrantKeyResolves.

// grantKeyProblems walks an engine-facing def map and reports every
// layer-6 grant that names a bundle nobody registered, or a bundle
// with a Static slot (ADR 0093 Decision 10: a static that only exists
// after layer 6 is never gathered). Separate from the test so the test
// can prove it reports both on a fixture.
func grantKeyProblems(all map[string]*game.CardDef) (problems []string, grants int) {
	for owner, d := range all {
		for _, s := range d.Static {
			for _, k := range s.GrantAbilities {
				grants++
				bundle := all[game.GrantKey(k)]
				switch {
				case bundle == nil:
					problems = append(problems, owner+" grants "+k+", which no card registers in Spec.Grants")
				case len(bundle.Static) > 0:
					problems = append(problems, owner+" grants "+k+", a bundle with a Static slot — a layer-6 grant of a static is not modelled")
				}
			}
		}
	}
	sort.Strings(problems)
	return problems, grants
}

// TestEveryGrantKeyResolves is the TestEveryTokenKeyResolves twin
// (ADR 0093 Decision 9): every key a layer-6 grant names resolves to a
// registered bundle, and none of those bundles has a Static slot.
// Walked over `defs`, so an emblem's statics are covered too.
func TestEveryGrantKeyResolves(t *testing.T) {
	problems, _ := grantKeyProblems(defs)
	for _, p := range problems {
		t.Error(p)
	}
	// The checker itself: both failure shapes are caught.
	fixture := map[string]*game.CardDef{
		"card-a":                    {Static: []game.StaticAbility{GrantAbilities(nil, "nobody/registered")}},
		"card-b":                    {Static: []game.StaticAbility{GrantAbilities(nil, "has/static")}},
		"card-c":                    {Static: []game.StaticAbility{GrantAbilities(nil, "fine/mana")}},
		game.GrantKey("has/static"): {Static: []game.StaticAbility{{Layer: game.Layer7PT}}},
		game.GrantKey("fine/mana"):  {ManaAbilities: []game.ManaAbilityShape{{TapCost: true, Produced: "{G}"}}},
	}
	got, grants := grantKeyProblems(fixture)
	if grants != 3 || len(got) != 2 ||
		!strings.Contains(got[0], "nobody/registered") || !strings.Contains(got[1], "has/static") {
		t.Errorf("fixture problems = %v (grants %d), want the unregistered key and the static bundle", got, grants)
	}
}

func TestGrantAbilitiesIsALayer6Declaration(t *testing.T) {
	s := GrantAbilities(func(*game.Card, *game.Game, *game.Card) bool { return true }, "a/one", "grant:a/two")
	if s.Layer != game.Layer6Ability || s.Apply != nil || s.RemovesAbilities {
		t.Errorf("static = %+v, want a bare layer-6 declaration", s)
	}
	if len(s.GrantAbilities) != 2 || s.GrantAbilities[0] != "a/one" {
		t.Errorf("GrantAbilities = %v", s.GrantAbilities)
	}
}

func TestBuildGrantDefCarriesManaAndText(t *testing.T) {
	d := buildGrantDef(AbilityGrant{
		Key:  "fixture/any-color",
		Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"}},
		Text: "{T}: Add one mana of any color.",
	})
	if len(d.ManaAbilities) != 1 || !d.ManaAbilities[0].TapCost || d.ManaAbilities[0].Produced != "{W|U|B|R|G}" {
		t.Errorf("mana = %+v", d.ManaAbilities)
	}
	if d.GrantText != "{T}: Add one mana of any color." {
		t.Errorf("GrantText = %q", d.GrantText)
	}
}

func expectRegisterPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("no panic; want one mentioning %q", want)
		}
		if msg, _ := r.(string); !strings.Contains(msg, want) {
			t.Fatalf("panic %v; want one mentioning %q", r, want)
		}
	}()
	fn()
}

func TestCheckGrantsRefusesWhatABundleCannotSay(t *testing.T) {
	tap := []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{G}"}}
	expectRegisterPanic(t, "no Text", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/no-text", Mana: tap}})
	})
	expectRegisterPanic(t, "ActiveWhen", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/gated", Text: "x",
			Activated: []ActivatedAbility{{Label: "x", ActiveWhen: Level(2)}}}})
	})
	expectRegisterPanic(t, "outside the battlefield", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/hand", Text: "x",
			Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{G}", Zones: []game.ZoneKind{game.ZoneHand}}}}})
	})
	expectRegisterPanic(t, "with no abilities", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/empty", Text: "x"}})
	})
}

func TestCheckStaticGrantsRefusesTheWrongLayer(t *testing.T) {
	s := GrantAbilities(nil, "fixture/x")
	s.Layer = game.Layer7PT
	expectRegisterPanic(t, "layer-6 effect", func() {
		checkStaticGrants("Fixture", "static", []game.StaticAbility{s})
	})
	expectRegisterPanic(t, "no key", func() {
		checkStaticGrants("Fixture", "static", []game.StaticAbility{GrantAbilities(nil, "")})
	})
	// The right layer is fine.
	checkStaticGrants("Fixture", "static", []game.StaticAbility{GrantAbilities(nil, "fixture/x")})
}

// The whole path through the real catalog plumbing: a card whose
// static is GrantAbilities, a bundle registered the way Register files
// one, and a vanilla creature that can then tap for the granted mana.
func TestGrantAbilitiesGivesACreatureTheBundlesManaAbility(t *testing.T) {
	const grantor = "fixture-oracle-grant-rite"
	bundle := AbilityGrant{
		Key:  "fixture-rite/tap-for-green",
		Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{G}", Label: "Add {G}"}},
		Text: "{T}: Add {G}.",
	}
	defs[grantor] = buildDef(Spec{
		OracleID: grantor,
		Name:     "Fixture Rite",
		Static: []game.StaticAbility{GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, bundle.Key)},
	})
	defs[game.GrantKey(bundle.Key)] = buildGrantDef(bundle)
	t.Cleanup(func() {
		delete(defs, grantor)
		delete(defs, game.GrantKey(bundle.Key))
	})

	g := newCatalogGame(t)
	me := g.Seats[0].ID
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me, Controller: me}
	pushBattlefieldCardWithTimestamp(g, bear)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fixture Rite", TypeLine: "Enchantment", OracleID: grantor, Owner: me, Controller: me})

	var abs []game.ManaAbilityShape
	var origins game.AbilityOrigins
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(bear.InstanceID)
		abs, origins = game.ManaAbilitiesWithOrigins(c)
	})
	if len(abs) != 1 || abs[0].Produced != "{G}" || !origins.At(0).Granted() {
		t.Fatalf("bear mana = %+v / %+v, want the granted {G}", abs, origins)
	}
	if err := g.ActivateManaAbility(me, bear.InstanceID, 0, game.ManaAbilityParams{Ref: origins.Ref(0)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	if n := len(g.Seats[0].ManaPool); n != 1 {
		t.Errorf("pool has %d tokens, want one {G}", n)
	}
}
