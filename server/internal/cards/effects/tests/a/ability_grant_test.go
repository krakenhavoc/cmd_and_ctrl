package testsa

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestEveryGrantKeyResolves is the TestEveryTokenKeyResolves twin
// (ADR 0093 Decision 9): every key a layer-6 grant names resolves to a
// registered bundle, and every static in those bundles is one the layer
// pass can apply once granted (#2562, #3784).
// Walked over `defs`, so an emblem's statics are covered too.
func TestEveryGrantKeyResolves(t *testing.T) {
	problems, _ := effectstest.GrantKeyProblems(effects.Defs)
	for _, p := range problems {
		t.Error(p)
	}
	// The checker itself: both failure shapes are caught.
	fixture := map[string]*game.CardDef{
		"card-a":                    {Static: []game.StaticAbility{effects.GrantAbilities(nil, "nobody/registered")}},
		"card-b":                    {Static: []game.StaticAbility{effects.GrantAbilities(nil, "has/static")}},
		"card-c":                    {Static: []game.StaticAbility{effects.GrantAbilities(nil, "fine/mana")}},
		"card-d":                    {Static: []game.StaticAbility{effects.GrantAbilities(nil, "fine/pump")}},
		game.GrantKey("has/static"): {Static: []game.StaticAbility{{Layer: game.Layer6Ability}}},
		game.GrantKey("fine/mana"):  {ManaAbilities: []game.ManaAbilityShape{{TapCost: true, Produced: "{G}"}}},
		game.GrantKey("fine/pump"):  {Static: []game.StaticAbility{effects.PumpAttached(3, 3)}},
	}
	got, grants := effectstest.GrantKeyProblems(fixture)
	if grants != 4 || len(got) != 2 ||
		!strings.Contains(got[0], "nobody/registered") || !strings.Contains(got[1], "has/static") {
		t.Errorf("fixture problems = %v (grants %d), want the unregistered key and the static bundle", got, grants)
	}
}

func TestGrantAbilitiesIsALayer6Declaration(t *testing.T) {
	s := effects.GrantAbilities(func(*game.Card, *game.Game, *game.Card) bool { return true }, "a/one", "grant:a/two")
	if s.Layer != game.Layer6Ability || s.Apply != nil || s.RemovesAbilities {
		t.Errorf("static = %+v, want a bare layer-6 declaration", s)
	}
	if len(s.GrantAbilities) != 2 || s.GrantAbilities[0] != "a/one" {
		t.Errorf("GrantAbilities = %v", s.GrantAbilities)
	}
}

func TestBuildGrantDefCarriesManaAndText(t *testing.T) {
	d := effects.BuildGrantDef(effects.AbilityGrant{
		Key:  "fixture/any-color",
		Mana: []effects.ManaAbility{{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"}},
		Text: "{T}: Add one mana of any color.",
	})
	if len(d.ManaAbilities) != 1 || !d.ManaAbilities[0].TapCost || d.ManaAbilities[0].Produced != "{W|U|B|R|G}" {
		t.Errorf("mana = %+v", d.ManaAbilities)
	}
	if d.GrantText != "{T}: Add one mana of any color." {
		t.Errorf("GrantText = %q", d.GrantText)
	}
}

func TestCheckGrantsRefusesWhatABundleCannotSay(t *testing.T) {
	tap := []effects.ManaAbility{{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{G}"}}
	effectstest.ExpectRegisterPanic(t, "no Text", func() {
		effects.CheckGrants("Fixture", []effects.AbilityGrant{{Key: "fixture/no-text", Mana: tap}})
	})
	effectstest.ExpectRegisterPanic(t, "ActiveWhen", func() {
		effects.CheckGrants("Fixture", []effects.AbilityGrant{{Key: "fixture/gated", Text: "x",
			Activated: []effects.ActivatedAbility{{Label: "x", ActiveWhen: effects.Level(2)}}}})
	})
	effectstest.ExpectRegisterPanic(t, "outside the battlefield", func() {
		effects.CheckGrants("Fixture", []effects.AbilityGrant{{Key: "fixture/hand", Text: "x",
			Mana: []effects.ManaAbility{{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{G}", Zones: []game.ZoneKind{game.ZoneHand}}}}})
	})
	effectstest.ExpectRegisterPanic(t, "with no abilities", func() {
		effects.CheckGrants("Fixture", []effects.AbilityGrant{{Key: "fixture/empty", Text: "x"}})
	})
}

func TestCheckStaticGrantsRefusesTheWrongLayer(t *testing.T) {
	s := effects.GrantAbilities(nil, "fixture/x")
	s.Layer = game.Layer7PT
	effectstest.ExpectRegisterPanic(t, "layer-6 effect", func() {
		effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{s})
	})
	effectstest.ExpectRegisterPanic(t, "no key", func() {
		effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{effects.GrantAbilities(nil, "")})
	})
	// The right layer is fine.
	effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{effects.GrantAbilities(nil, "fixture/x")})
}

// The whole path through the real catalog plumbing: a card whose
// static is GrantAbilities, a bundle registered the way Register files
// one, and a vanilla creature that can then tap for the granted mana.
func TestGrantAbilitiesGivesACreatureTheBundlesManaAbility(t *testing.T) {
	const grantor = "fixture-oracle-grant-rite"
	bundle := effects.AbilityGrant{
		Key:  "fixture-rite/tap-for-green",
		Mana: []effects.ManaAbility{{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{G}", Label: "Add {G}"}},
		Text: "{T}: Add {G}.",
	}
	effects.Defs[grantor] = effects.BuildDef(effects.Spec{
		OracleID: grantor,
		Name:     "Fixture Rite",
		Static: []game.StaticAbility{effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, bundle.Key)},
	})
	effects.Defs[game.GrantKey(bundle.Key)] = effects.BuildGrantDef(bundle)
	t.Cleanup(func() {
		delete(effects.Defs, grantor)
		delete(effects.Defs, game.GrantKey(bundle.Key))
	})

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me, Controller: me}
	effectstest.PushBattlefieldCardWithTimestamp(g, bear)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fixture Rite", TypeLine: "Enchantment", OracleID: grantor, Owner: me, Controller: me})

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

// ADR 0093 amendment 2026-11-10 (#5402): Register lets a bundle static
// grant only by the explicit opt-in, and only one bundle the same card
// declares, with no static of its own.
func TestCheckFixedBundleGrantsRefusesEveryOtherShape(t *testing.T) {
	yours := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller
	}
	mana := effects.TapForManaGrant("fixture/inner-mana", "{G}", "Add {G}", "{T}: Add {G}.")
	withStatic := effects.AbilityGrant{Key: "fixture/inner-static", Text: "x", Static: []game.StaticAbility{effects.PumpAttached(1, 1)}}
	outer := func(s game.StaticAbility) effects.AbilityGrant {
		return effects.AbilityGrant{Key: "fixture/outer", Text: "x", Static: []game.StaticAbility{s}}
	}
	ok := outer(effects.GrantedFixedBundle("x", yours, mana.Key))
	effects.CheckFixedBundleGrants("Fixture", ok, []effects.AbilityGrant{ok, mana})
	effects.CheckGrantAbilities("Fixture", ok)

	effectstest.ExpectRegisterPanic(t, "grants abilities itself", func() {
		effects.CheckGrantAbilities("Fixture", outer(effects.GrantAbilities(yours, mana.Key)))
	})
	two := effects.GrantedFixedBundle("x", yours, mana.Key)
	two.GrantAbilities = append(two.GrantAbilities, mana.Key)
	effectstest.ExpectRegisterPanic(t, "exactly one bundle", func() {
		effects.CheckFixedBundleGrants("Fixture", outer(two), []effects.AbilityGrant{outer(two), mana})
	})
	effectstest.ExpectRegisterPanic(t, "does not declare", func() {
		effects.CheckFixedBundleGrants("Fixture", ok, []effects.AbilityGrant{ok})
	})
	self := outer(effects.GrantedFixedBundle("x", yours, "fixture/outer"))
	effectstest.ExpectRegisterPanic(t, "its own bundle", func() {
		effects.CheckFixedBundleGrants("Fixture", self, []effects.AbilityGrant{self})
	})
	nested := outer(effects.GrantedFixedBundle("x", yours, withStatic.Key))
	effectstest.ExpectRegisterPanic(t, "a bundle with a static", func() {
		effects.CheckFixedBundleGrants("Fixture", nested, []effects.AbilityGrant{nested, withStatic})
	})
	layer7 := effects.GrantedFixedBundle("x", yours, mana.Key)
	layer7.Layer = game.Layer7PT
	effectstest.ExpectRegisterPanic(t, "layer-6 static", func() {
		effects.CheckFixedBundleGrants("Fixture", outer(layer7), []effects.AbilityGrant{outer(layer7), mana})
	})
}

// The catalog's one fixed-bundle grant passes the resolution-time check,
// and its outer bundle's static is accepted by the engine.
func TestRoarOfTheFifthPeopleChapterIIBundleIsAcceptedOnlyByOptIn(t *testing.T) {
	if err := effects.CheckGrantMods([]game.Mod{game.GrantAbilitiesMod(cardshu.RoarChapterII)}); err != nil {
		t.Fatalf("Roar of the Fifth People's chapter II bundle was refused: %v", err)
	}
	def := effects.Defs[game.GrantKey(cardshu.RoarChapterII)]
	if def == nil || len(def.Static) != 1 || !def.Static[0].GrantsFixedBundle {
		t.Fatalf("chapter II bundle = %+v, want one opted-in static", def)
	}
	s := def.Static[0]
	s.GrantsFixedBundle = false
	if game.GrantedStaticProblem(s) == "" {
		t.Error("the same static without the opt-in was accepted")
	}
}
