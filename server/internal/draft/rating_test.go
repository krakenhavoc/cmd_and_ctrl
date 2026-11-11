package draft

import (
	"reflect"
	"testing"
)

// The two orderings #5439 asks for by name, on real cards of the
// fixture: removal above a vanilla creature of the same cost, and a
// mythic bomb above a common filler.
func TestRatingOrdersTheCardsItShould(t *testing.T) {
	fx := loadBotFixture(t)
	above := func(better, worse string) {
		t.Helper()
		b, w := fx.rating(t, better), fx.rating(t, worse)
		if b <= w {
			t.Errorf("%s rates %.2f and %s %.2f: the first should rate higher", better, b, worse, w)
		}
	}
	// A removal spell above a vanilla creature of the same cost.
	above("Murder", "C2 Fixture Black Scout")   // three mana each
	above("Doom Blade", "C2 Fixture Black Cub") // two
	above("Pacifism", "C2 Fixture White Cub")   // two
	above("Lightning Strike", "C2 Fixture Red Cub")
	// A mythic bomb above a common filler.
	above("Grave Titan", "C2 Fixture Black Cub")
	above("Archangel of Thune", "C2 Fixture White Guard")
	above("Primeval Titan", "Giant Spider")
	// And the rest of what a drafter knows.
	above("Lightning Bolt", "Shock")                     // more damage for the same mana
	above("Serra Angel", "C2 Fixture White Guard")       // five mana, and one of them flies
	above("Day of Judgment", "Divination")               // a sweep above a draw spell
	above("Murder", "Cancel")                            // removal above a counterspell
	above("Divination", "Mind Rot")                      // a declared effect above an unknown one
	above("C2 Fixture Green Cub", "Giant Growth")        // a body above a trick
	above("Vampire Nighthawk", "C2 Fixture Black Scout") // keywords
	above("Mulldrifter", "C2 Fixture Blue Guard")        // a body that draws two
	above("Mind Control", "Claustrophobia")              // read from the printed text
	above("C2 Fixture Red Cub", "Mountain")              // anything above a basic land
	if got := fx.rating(t, "Plains"); got != BasicLandRating {
		t.Errorf("a basic land rates %.2f, want %.2f", got, BasicLandRating)
	}
}

// The rating is a pure function of Facts: the same facts give the same
// number, whatever was rated before, and Rate changes nothing it is
// given. This is what lets it be tabulated.
func TestRateIsAPureFunctionOfTheFacts(t *testing.T) {
	fx := loadBotFixture(t)
	for _, c := range fx.all {
		f := fx.rater.FactsOf(c)
		before := f
		before.Keywords = append([]string(nil), f.Keywords...)
		before.Colors = append([]string(nil), f.Colors...)
		first := Rate(f)
		for _, other := range fx.all[:10] {
			Rate(fx.rater.FactsOf(other))
		}
		if again := Rate(f); again != first {
			t.Errorf("%s rated %.4f, then %.4f", c.Name, first, again)
		}
		if !reflect.DeepEqual(f, before) {
			t.Errorf("Rate changed the facts of %s", c.Name)
		}
		if !c.Basic && first < RatingFloor {
			t.Errorf("%s rates %.2f, below the floor", c.Name, first)
		}
	}
}

// Rate on facts written by hand: each term of the rating moves it the
// way its weight says.
func TestRateTable(t *testing.T) {
	bear := Facts{Creature: true, ManaValue: 2, Power: 2, Toughness: 2, Rarity: "common"}
	with := func(change func(*Facts)) Facts {
		f := bear
		change(&f)
		return f
	}
	near := func(what string, got, want float64) {
		t.Helper()
		if d := got - want; d > 1e-9 || d < -1e-9 {
			t.Errorf("%s: %.3f, want %.3f", what, got, want)
		}
	}
	near("a two-mana 2/2", Rate(bear), CreatureBase+4*PerStatPoint-2*PerManaCreature)
	near("a point of toughness", Rate(with(func(f *Facts) { f.Toughness = 3 }))-Rate(bear), PerStatPoint)
	near("a point of mana value", Rate(with(func(f *Facts) { f.ManaValue = 3 }))-Rate(bear), -PerManaCreature)
	near("flying", Rate(with(func(f *Facts) { f.Keywords = []string{"flying"} }))-Rate(bear), KeywordFlying)
	near("uncommon", Rate(with(func(f *Facts) { f.Rarity = "uncommon" }))-Rate(bear), RarityUncommon)
	near("rare", Rate(with(func(f *Facts) { f.Rarity = "rare" }))-Rate(bear), RarityRare)
	near("mythic", Rate(with(func(f *Facts) { f.Rarity = "mythic" }))-Rate(bear), RarityMythic)
	near("a card drawn on a body", Rate(with(func(f *Facts) { f.Draws = 1 }))-Rate(bear), OnABody*PerDraw)

	// Flying on a creature that deals no damage is worth nothing, and
	// the creature loses for dealing none.
	wall := Facts{Creature: true, ManaValue: 0, Power: 0, Toughness: 2, Keywords: []string{"flying"}}
	near("a 0/2 flier", Rate(wall), CreatureBase+2*PerStatPoint-NoPowerPenalty)

	spell := func(mv int, change func(*Facts)) float64 {
		f := Facts{Instant: true, ManaValue: mv}
		change(&f)
		return Rate(f)
	}
	near("a spell that declares nothing", spell(3, func(*Facts) {}), SpellBase)
	near("removal", spell(3, func(f *Facts) { f.Removal = true }), RemovalValue-3*PerManaSpell)
	near("three damage", spell(1, func(f *Facts) { f.Damage = 3 }), DamageBase+3*PerDamage-PerManaSpell)
	near("a sweep", spell(4, func(f *Facts) { f.Sweep = true }), SweepValue-4*PerManaSpell)
	near("a counterspell", spell(2, func(f *Facts) { f.Counterspell = true }), CounterspellValue-2*PerManaSpell)
	near("two cards", spell(3, func(f *Facts) { f.Draws = 2 }), SpellBase+2*PerDraw-3*PerManaSpell)
	near("a land", spell(2, func(f *Facts) { f.Ramp = 1 }), SpellBase+PerRamp-2*PerManaSpell)
	near("two tokens", spell(2, func(f *Facts) { f.Tokens = 2 }), SpellBase+2*PerToken-2*PerManaSpell)
	near("a trick", spell(1, func(f *Facts) { f.Trick = true }), TrickValue-PerManaSpell)
	near("removal that draws", spell(3, func(f *Facts) { f.Removal, f.Draws = true, 1 }), RemovalValue+PerDraw/2-3*PerManaSpell)
	near("a nonbasic land", Rate(Facts{Land: true}), NonbasicLandRating)
	near("a planeswalker", Rate(Facts{Planeswalker: true, ManaValue: 4}), PlaneswalkerValue-4*PerManaSpell)
}

// What the facts say comes from the printing and the catalog's declared
// Purpose (ADR 0126), read here off real cards.
func TestFactsComeFromThePrintingAndTheCatalog(t *testing.T) {
	fx := loadBotFixture(t)
	facts := func(name string) Facts { return fx.rater.FactsOf(fx.card(t, name, "x")) }

	if f := facts("Serra Angel"); !f.Creature || f.Power != 4 || f.Toughness != 4 || f.ManaValue != 5 ||
		!reflect.DeepEqual(f.Keywords, []string{"flying", "vigilance"}) || !reflect.DeepEqual(f.Colors, []string{"W"}) {
		t.Errorf("Serra Angel: %+v", f)
	}
	for name, check := range map[string]func(Facts) bool{
		"Murder":               func(f Facts) bool { return f.Removal && f.Instant && !f.Counterspell },
		"Swords to Plowshares": func(f Facts) bool { return f.Removal },
		"Pacifism":             func(f Facts) bool { return f.Removal && f.Aura },
		"Terminate":            func(f Facts) bool { return f.Removal }, // from the printed text
		"Royal Assassin":       func(f Facts) bool { return f.Removal && f.Creature },
		"Day of Judgment":      func(f Facts) bool { return f.Sweep && !f.Removal },
		"Pyroclasm":            func(f Facts) bool { return f.Sweep && f.Damage == 0 },
		"Lightning Bolt":       func(f Facts) bool { return f.Damage == 3 && !f.Removal },
		"Shock":                func(f Facts) bool { return f.Damage == 2 && !f.Removal },
		"Fireball":             func(f Facts) bool { return f.Damage == DamageForX },
		"Counterspell":         func(f Facts) bool { return f.Counterspell && !f.Removal },
		"Cancel":               func(f Facts) bool { return f.Counterspell && !f.Removal },
		"Naturalize":           func(f Facts) bool { return !f.Removal },
		"Act of Treason":       func(f Facts) bool { return !f.Removal },
		"Divination":           func(f Facts) bool { return f.Draws == 2 },
		"Sign in Blood":        func(f Facts) bool { return f.Draws == 2 },
		"Mulldrifter":          func(f Facts) bool { return f.Draws == 2 && f.Creature },
		"Rampant Growth":       func(f Facts) bool { return f.Ramp == 1 },
		"Llanowar Elves":       func(f Facts) bool { return f.Ramp == 1 && f.Creature },
		"Dragon Fodder":        func(f Facts) bool { return f.Tokens == 2 },
		"Raise the Alarm":      func(f Facts) bool { return f.Tokens == 2 },
		"Giant Growth":         func(f Facts) bool { return f.Trick },
		"Bonesplitter":         func(f Facts) bool { return f.Equipment && f.Artifact },
		"Forest":               func(f Facts) bool { return f.Land },
	} {
		if f := facts(name); !check(f) {
			t.Errorf("%s: the facts are wrong: %+v", name, f)
		}
	}
}

// A card the index does not hold, and a rater with no index at all, are
// rated from the draft Card's own fields.
func TestARaterWithNoIndexRatesFromTheCard(t *testing.T) {
	fx := loadBotFixture(t)
	c := fx.card(t, "Serra Angel", "x")
	got := Rater{}.FactsOf(c)
	if !got.Creature || got.ManaValue != 5 || got.Power != 5 || got.Toughness != 5 || got.Rarity != "uncommon" {
		t.Errorf("CardFacts(Serra Angel) = %+v", got)
	}
	if !reflect.DeepEqual(CatalogFacts(nil)(c), CardFacts(c)) {
		t.Error("CatalogFacts(nil) is not CardFacts")
	}
	if r := (Rater{}).Rating(c); r <= RatingFloor {
		t.Errorf("rated %.2f with no index", r)
	}
}

// The seam for a per-set ratings file: an override replaces the computed
// rating, and no file ships.
func TestRatingOverridesAreASeamAndEmptyToday(t *testing.T) {
	fx := loadBotFixture(t)
	c := fx.card(t, "Giant Growth", "x")
	if _, ok := (NoOverrides{}).Rating(c.Set, c); ok {
		t.Fatal("NoOverrides has an override")
	}
	computed := fx.rater.Rating(c)
	with := Rater{Facts: fx.rater.Facts, Overrides: overrideMap{c.OracleID.String(): 9.5}}
	if got := with.Rating(c); got != 9.5 {
		t.Errorf("with an override Giant Growth rates %.2f, want 9.5", got)
	}
	other := fx.card(t, "Titanic Growth", "y")
	if got, want := with.Rating(other), fx.rater.Rating(other); got != want {
		t.Errorf("a card with no override rates %.2f, want the computed %.2f", got, want)
	}
	if got := (Rater{Facts: fx.rater.Facts, Overrides: NoOverrides{}}).Rating(c); got != computed {
		t.Errorf("NoOverrides changed a rating: %.2f, want %.2f", got, computed)
	}
	// A basic land is never worth a pick, whatever a file says.
	if got := (Rater{Overrides: overrideMap{}}).Rating(fx.card(t, "Plains", "z")); got != BasicLandRating {
		t.Errorf("a basic land rates %.2f", got)
	}
}

type overrideMap map[string]float64

func (m overrideMap) Rating(_ string, c Card) (float64, bool) {
	v, ok := m[c.OracleID.String()]
	return v, ok
}
