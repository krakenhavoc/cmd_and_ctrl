package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unblinking Observer (#5680): "{T}: Add {U}. Spend this mana only to
// pay a disturb cost or cast an instant or sorcery spell."

func observerOnTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, _ := spendTable(t)
	observer := pushCatalogPermanent(g, me.ID, "Unblinking Observer", "Creature — Homunculus", unblinkingObserverOracleID, false)
	return g, me, observer
}

func tapObserver(t *testing.T, g *game.Game, me *game.Player, observer uuid.UUID) {
	t.Helper()
	if err := g.ActivateManaAbility(me.ID, observer, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Observer: %v", err)
	}
}

func TestUnblinkingObserverIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(unblinkingObserverOracleID)
	if !ok {
		t.Fatal("Unblinking Observer is not registered")
	}
	if spec.Name != "Unblinking Observer" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestUnblinkingObserverTapsForBlueAndIsGatedBySummoningSickness(t *testing.T) {
	g, me, observer := observerOnTable(t)
	tapObserver(t, g, me, observer)
	if units := me.ManaPool.Units(); len(units) != 1 || units[0].Color != "U" {
		t.Fatalf("pool = %+v, want one U", me.ManaPool)
	}
	if c, _ := battlefieldCard(g, observer); !c.Tapped {
		t.Error("the Observer did not tap")
	}

	fresh := pushCatalogPermanent(g, me.ID, "Unblinking Observer", "Creature — Homunculus", unblinkingObserverOracleID, true)
	if err := g.ActivateManaAbility(me.ID, fresh, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Observer tapped for mana (CR 302.6)")
	}
}

// The "cast an instant or sorcery spell" half.
func TestUnblinkingObserverManaPaysForInstantsAndSorceriesOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		ok             bool
	}{
		{"instant", "Instant", true},
		{"sorcery", "Sorcery", true},
		{"creature", "Creature — Bear", false},
		{"enchantment", "Enchantment — Aura", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, observer := observerOnTable(t)
			tapObserver(t, g, me, observer)
			spell := handSpell(me, "Test "+tc.name, tc.typeLine, "{U}")
			err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true})
			if tc.ok && err != nil {
				t.Errorf("the Observer's mana could not pay for a %s: %v", tc.name, err)
			}
			if !tc.ok {
				refusedForMana(t, err, "the Observer's mana for a "+tc.name)
			}
		})
	}
}

// The "pay a disturb cost" half, on a real disturb card: Baithook
// Angler's Disturb {1}{U}, the {U} from the Observer.
func TestUnblinkingObserverManaPaysADisturbCost(t *testing.T) {
	g, me, observer := observerOnTable(t)
	id := importToGraveyard(t, g, baithookAnglerRow(), me)
	tapObserver(t, g, me, observer)
	// One mana short: the Observer's {U} is spent only if the whole
	// cost is paid.
	err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb", Strict: true})
	refusedForMana(t, err, "disturb {1}{U} with one mana")
	if len(me.ManaPool) != 1 {
		t.Fatalf("the refused cast spent mana: %+v", me.ManaPool)
	}

	fillPool(me, 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb", Strict: true}); err != nil {
		t.Fatalf("disturb paid in part with the Observer's mana: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the disturb cast left %+v floating", me.ManaPool)
	}
	passPriorityAroundTable(t, g)
	if zone, c := cardWhere(g, me, id); zone != "battlefield" || c.Name != "Hook-Haunt Drifter" {
		t.Errorf("the disturbed card is in %s as %q, want Hook-Haunt Drifter on the battlefield", zone, c.Name)
	}
}

// Both mana of a disturb cost may be the Observer's: the generic part of
// the cost is part of the disturb cost too.
func TestUnblinkingObserverManaPaysTheWholeDisturbCost(t *testing.T) {
	g, me, first := observerOnTable(t)
	second := pushCatalogPermanent(g, me.ID, "Unblinking Observer", "Creature — Homunculus", unblinkingObserverOracleID, false)
	id := importToGraveyard(t, g, baithookAnglerRow(), me)
	tapObserver(t, g, me, first)
	tapObserver(t, g, me, second)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb", Strict: true}); err != nil {
		t.Fatalf("disturb {1}{U} with two Observers: %v", err)
	}
}

// The auto-tapper reaches for the Observer on a disturb cast, and not
// on the same card's mana cost.
func TestUnblinkingObserverIsAutoTappedForADisturbCost(t *testing.T) {
	g, me, first := observerOnTable(t)
	second := pushCatalogPermanent(g, me.ID, "Unblinking Observer", "Creature — Homunculus", unblinkingObserverOracleID, false)

	// From hand, for its mana cost, the card is a creature spell.
	angler := deck.ToGameCard(baithookAnglerRow(), false)
	angler.Owner, angler.Controller = me.ID, me.ID
	me.Hand.PushTop(angler)
	err := g.CastSpell(me.ID, angler.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true})
	refusedForMana(t, err, "Baithook Angler's mana cost with two Observers")
	for _, o := range []uuid.UUID{first, second} {
		if c, _ := battlefieldCard(g, o); c.Tapped {
			t.Fatal("a refused cast left an Observer tapped")
		}
	}

	id := importToGraveyard(t, g, baithookAnglerRow(), me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb", Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped disturb: %v", err)
	}
	for _, o := range []uuid.UUID{first, second} {
		if c, _ := battlefieldCard(g, o); !c.Tapped {
			t.Error("an Observer was not tapped for the disturb cost")
		}
	}
}

// What the mana does not pay: an activation, and a graveyard cast under
// another alternative cost (escape is not a disturb cost).
func TestUnblinkingObserverManaRefusesEverythingElse(t *testing.T) {
	g, me, observer := observerOnTable(t)
	tapObserver(t, g, me, observer)
	cost, _ := game.ParseCost("{U}")

	if c, _ := battlefieldCard(g, observer); me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(c)) {
		t.Error("the mana paid for an activated ability")
	}
	creature := game.Card{Name: "Test Escapee", TypeLine: "Creature — Giant", ManaCost: "{U}", Owner: me.ID, Controller: me.ID}
	for _, tc := range []struct {
		name string
		ctx  game.ManaSpendContext
		want bool
	}{
		{"a creature cast from the graveyard for its mana cost", game.ManaSpendForCastFrom(creature, game.ZoneGraveyard), false},
		{"an escaped creature", func() game.ManaSpendContext {
			ctx := game.ManaSpendForCastFrom(creature, game.ZoneGraveyard)
			ctx.AltCost = Escape("{U}", 1).Key
			return ctx
		}(), false},
		{"a disturbed creature", func() game.ManaSpendContext {
			ctx := game.ManaSpendForCastFrom(creature, game.ZoneGraveyard)
			ctx.AltCost = Disturb("{U}").Key
			return ctx
		}(), true},
		{"a payment that names no purpose", game.ManaSpendContext{}, false},
	} {
		if got := me.ManaPool.CanPayFor(cost, 0, tc.ctx); got != tc.want {
			t.Errorf("%s: can pay = %v, want %v", tc.name, got, tc.want)
		}
	}
}
