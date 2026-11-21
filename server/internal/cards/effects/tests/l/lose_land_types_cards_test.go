package testsl

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// llPushArbor is a Dryad Arbor, a 1/1 Land Creature — Forest Dryad, so
// the state-based actions leave it alone.
func llPushArbor(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dryad Arbor",
		TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1, Owner: owner, Controller: owner})
}

// TestUltimaBlightsALandForAsLongAsItHasTheCounter is the card's
// headline: the land gets a blight counter, and while it has one it has
// no land types and no abilities and taps only for {C}. Dryad Arbor
// keeps its creature type (CR 205.1a). The moment the last counter goes,
// so does the effect (CR 611.2b).
func TestUltimaBlightsALandForAsLongAsItHasTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbor := llPushArbor(g, opp.ID)
	effectstest.LlAttackWithUltima(t, g, me.ID, opp.ID, arbor)

	land := effectstest.LayeredCard(t, g, arbor)
	if got := land.Counters["blight"]; got != 1 {
		t.Fatalf("blight counters %d, want 1", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Dryad"}) {
		t.Errorf("subtypes %v, want [Dryad]: no land types, the creature type stays", got)
	}
	if got := effectstest.LtProduced(t, g, arbor); !reflect.DeepEqual(got, []string{"{C}"}) {
		t.Errorf("taps for %v, want only the granted {C}", got)
	}
	if !land.Effective().AbilitiesRemoved {
		t.Error("the land kept its abilities")
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(arbor, "blight", -1) })
	if got := effectstest.EffectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Forest", "Dryad"}) && !reflect.DeepEqual(got, []string{"Dryad", "Forest"}) {
		t.Errorf("the counter went: subtypes %v, want the Forest back", got)
	}
	if got := effectstest.LtProduced(t, g, arbor); !reflect.DeepEqual(got, []string{"{G}"}) {
		t.Errorf("the counter went: taps for %v, want {G}", got)
	}
}

// TestUltimaTakesANonbasicLandsOwnAbilities: Unstable Frontier's own
// {C} and its activated ability go with the rest of its abilities, and
// the {C} it taps for is the granted one.
func TestUltimaTakesANonbasicLandsOwnAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frontier := effectstest.PushCatalogPermanent(g, opp.ID, "Unstable Frontier", "Land", effectstest.LtUnstableFrontierOracle, false)
	effectstest.LlAttackWithUltima(t, g, me.ID, opp.ID, frontier)

	rows := game.ManaAbilitiesForCard(effectstest.LayeredCard(t, g, frontier))
	if len(rows) != 1 || rows[0].Produced != "{C}" {
		t.Fatalf("mana abilities %+v, want the one granted {C}", rows)
	}
	if err := g.ActivateCatalogAbility(opp.ID, frontier, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(frontier)}); err == nil {
		t.Error("the Frontier's own activated ability survived \"loses all abilities\"")
	}
}

// TestUltimaAddsAnAdditionalColorless is the third line: a land you tap
// for {C} adds one more {C}, without the stack (CR 605.1b). A land
// tapped for a colour does not, and neither does an opponent's land.
func TestUltimaAddsAnAdditionalColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ultima, Origin of Oblivion", "Legendary Creature — God", effectstest.LlUltimaOracle, false)
	frontier := effectstest.PushCatalogPermanent(g, me.ID, "Unstable Frontier", "Land", effectstest.LtUnstableFrontierOracle, false)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Unstable Frontier", "Land", effectstest.LtUnstableFrontierOracle, false)

	effectstest.TapForMana(t, g, me.ID, frontier)
	if got := effectstest.SortedPool(me); !reflect.DeepEqual(got, []string{"C", "C"}) {
		t.Errorf("Unstable Frontier for {C}: pool %v, want {C}{C}", got)
	}
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); !reflect.DeepEqual(got, []string{"C", "C", "G"}) {
		t.Errorf("then a Forest: pool %v, want {C}{C}{G}", got)
	}
	effectstest.TapForMana(t, g, opp.ID, theirs)
	if got := effectstest.SortedPool(opp); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("an opponent's land: pool %v, want {C}", got)
	}
}

// TestUltimaShowsTheBlightOnTheLand is ADR 0109 §2 decision 5: the chip
// says what the land has lost and gained, and for how long.
func TestUltimaShowsTheBlightOnTheLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.LlAttackWithUltima(t, g, me.ID, opp.ID, forest)
	var got []game.LandTypeEffect
	g.WithWriteLock(func() { got = g.LandTypeEffectsForEffect(forest) })
	want := []game.LandTypeEffect{{
		LosesAll: true, LosesAbilities: true, Gains: []string{"{T}: Add {C}."},
		Until: "for as long as it has a blight counter on it", Source: "Ultima, Origin of Oblivion",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("land type effects %+v, want %+v", got, want)
	}
}

// TestLithoformBlightStripsTheEnchantedLand: the Aura draws a card, and
// the enchanted land has no land types and no abilities of its own, and
// taps for {C} or, for 1 life, any colour. Dryad Arbor keeps its Dryad.
func TestLithoformBlightStripsTheEnchantedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbor := llPushArbor(g, opp.ID)
	hand := len(me.Hand.Cards)
	effectstest.Enchant(t, g, "Lithoform Blight", effectstest.LlLithoformBlightOracle, arbor)
	effectstest.PassPriorityAroundTable(t, g) // the draw trigger
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d, want %d: the Aura draws a card as it enters", got, hand+1)
	}
	if got := effectstest.EffectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Dryad"}) {
		t.Errorf("subtypes %v, want [Dryad]", got)
	}
	if got := effectstest.LtProduced(t, g, arbor); !reflect.DeepEqual(got, []string{"{C}", "{W|U|B|R|G}"}) {
		t.Errorf("taps for %v, want {C} and any colour", got)
	}

	// "{T}, Pay 1 life: Add one mana of any color." is a cost: the life
	// is paid as the land taps.
	life := opp.Life
	if err := g.ActivateManaAbility(opp.ID, arbor, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility (any colour): %v", err)
	}
	effectstest.B10ResolveAllManaPicks(t, g, opp.ID, "B")
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if got := effectstest.SortedPool(opp); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("pool %v, want {B}", got)
	}
}

// TestLithoformBlightTakesANonbasicLandsOwnAbilities: Unstable
// Frontier's own {C} and activated ability are gone; the two the Aura
// gives are all it has.
func TestLithoformBlightTakesANonbasicLandsOwnAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	frontier := effectstest.PushCatalogPermanent(g, me.ID, "Unstable Frontier", "Land", effectstest.LtUnstableFrontierOracle, false)
	effectstest.Enchant(t, g, "Lithoform Blight", effectstest.LlLithoformBlightOracle, frontier)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LtProduced(t, g, frontier); !reflect.DeepEqual(got, []string{"{C}", "{W|U|B|R|G}"}) {
		t.Errorf("taps for %v, want the Aura's {C} and any colour", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, frontier, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(frontier)}); err == nil {
		t.Error("the Frontier's own activated ability survived")
	}
}

// TestAlpineMoonStripsOpponentsLandsWithTheChosenName: lands your
// opponents control with the chosen name lose all land types and
// abilities and tap for any colour. Yours with the name, and theirs with
// another name, are untouched.
func TestAlpineMoonStripsOpponentsLandsWithTheChosenName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Unstable Frontier", "Land", effectstest.LtUnstableFrontierOracle, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Unstable Frontier", "Land", effectstest.LtUnstableFrontierOracle, false)
	arbor := llPushArbor(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Alpine Moon", "Enchantment", effectstest.LlAlpineMoonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// Before the name is chosen, nothing is named and nothing changes.
	if got := effectstest.LtProduced(t, g, theirs); !reflect.DeepEqual(got, []string{"{C}"}) {
		t.Fatalf("before the choice their Frontier taps for %v, want its own {C}", got)
	}
	effectstest.ActNameFor(t, g, me.ID, "Unstable Frontier")

	if got := effectstest.LtProduced(t, g, theirs); !reflect.DeepEqual(got, []string{"{W|U|B|R|G}"}) {
		t.Errorf("their Frontier taps for %v, want only any colour", got)
	}
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(theirs)}); err == nil {
		t.Error("their Frontier kept its activated ability")
	}
	if got := effectstest.LtProduced(t, g, mine); !reflect.DeepEqual(got, []string{"{C}"}) {
		t.Errorf("your own Frontier taps for %v, want its own {C}: the Moon names your opponents' lands", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Forest", "Dryad"}) && !reflect.DeepEqual(got, []string{"Dryad", "Forest"}) {
		t.Errorf("a land with another name: subtypes %v, want its Forest", got)
	}
}

// TestAlpineMoonLosesTheLandTypes: a named nonbasic land with a land
// type (Dryad Arbor) loses it, keeps its creature type, and taps for any
// colour instead of {G}.
func TestAlpineMoonLosesTheLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbor := llPushArbor(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Alpine Moon", "Enchantment", effectstest.LlAlpineMoonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "dryad arbor")
	if got := effectstest.EffectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Dryad"}) {
		t.Errorf("subtypes %v, want [Dryad]", got)
	}
	if got := effectstest.LtProduced(t, g, arbor); !reflect.DeepEqual(got, []string{"{W|U|B|R|G}"}) {
		t.Errorf("taps for %v, want only any colour", got)
	}
}

// TestAlpineMoonNamesOnlyANonbasicLand is the choice's restriction, "a
// nonbasic land card name": naming a basic land changes nothing, which
// is what a legal choice no opponent's land shares would do. The engine
// takes a card name as free text (game/choose_card_name.go), so the
// restriction is enforced where the name is read.
func TestAlpineMoonNamesOnlyANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Alpine Moon", "Enchantment", effectstest.LlAlpineMoonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Forest")
	if got := effectstest.LtProduced(t, g, forest); !reflect.DeepEqual(got, []string{"{G}"}) {
		t.Errorf("a basic Forest under an Alpine Moon naming Forest taps for %v, want {G}", got)
	}
}
