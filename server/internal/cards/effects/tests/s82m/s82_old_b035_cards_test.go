package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 older sets, batch 35 (Modern Horizons 2: removal, #5501).

const (
	oldB035BreakTheIce = "22ec98d2-0ac5-48a2-b48b-c6a98f35ecf0"
	oldB035Sanctifier  = "810d3afe-c644-444f-a4d9-88b323f0b581"
	oldB035Rays        = "8960a213-cb72-4d67-bb4e-bcb1f657b944"
	oldB035Curator     = "8a20707f-9764-4236-9aa9-858a2419f012"
	oldB035Gorilla     = "9f53b691-da24-4264-a2a0-d0b95a385963"
	oldB035Charmaw     = "c995bc01-f278-4088-9ae2-1b50f87b2a61"
)

func TestOldB035CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB035BreakTheIce: "Break the Ice",
		oldB035Sanctifier:  "Sanctifier en-Vec",
		oldB035Rays:        "Piercing Rays",
		oldB035Curator:     "Loathsome Curator",
		oldB035Gorilla:     "Gorilla Shaman",
		oldB035Charmaw:     "Obsidian Charmaw",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB035Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var out game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			out = z.Kind
		}
	})
	return out
}

func oldB035Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- Break the Ice ------------------------------------------------------------

func TestOldB035BreakTheIceHitsSnowAndColorlessLandsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wastes := effectstest.SeedManaLand(g, opp.ID, "Wastes", "Basic Land", "C")
	forest := effectstest.SeedManaLand(g, opp.ID, "Forest", "Basic Land — Forest", "G")
	snow := effectstest.SeedManaLand(g, opp.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "G")
	_ = me
	for _, id := range []uuid.UUID{forest} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Break the Ice", "Sorcery", oldB035BreakTheIce, []game.TargetRef{oldB035Card(id)}); err == nil {
			t.Error("a Forest is neither snow nor able to produce {C}")
		}
	}
	effectstest.CastCatalogSpell(t, g, "Break the Ice", "Sorcery", oldB035BreakTheIce, []game.TargetRef{oldB035Card(wastes)})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB035Zone(g, wastes) != game.ZoneGraveyard {
		t.Errorf("Wastes is in %q, want the graveyard", oldB035Zone(g, wastes))
	}
	effectstest.CastCatalogSpell(t, g, "Break the Ice", "Sorcery", oldB035BreakTheIce, []game.TargetRef{oldB035Card(snow)})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB035Zone(g, snow) != game.ZoneGraveyard {
		t.Errorf("the snow land is in %q, want the graveyard", oldB035Zone(g, snow))
	}
	if oldB035Zone(g, forest) != game.ZoneBattlefield {
		t.Error("the plain Forest was destroyed")
	}
}

func TestOldB035BreakTheIceOverloadDestroysEveryMatchingLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.SeedManaLand(g, me.ID, "Wastes", "Basic Land", "C")
	theirs := effectstest.SeedManaLand(g, opp.ID, "Wastes", "Basic Land", "C")
	forest := effectstest.SeedManaLand(g, opp.ID, "Forest", "Basic Land — Forest", "G")
	effectstest.CastWithAltCost(t, g, "Break the Ice", "Sorcery", oldB035BreakTheIce, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	if oldB035Zone(g, mine) != game.ZoneGraveyard || oldB035Zone(g, theirs) != game.ZoneGraveyard {
		t.Error("overload spared a colorless land (each land, yours included)")
	}
	if oldB035Zone(g, forest) != game.ZoneBattlefield {
		t.Error("overload destroyed a Forest")
	}
}

// --- Sanctifier en-Vec -----------------------------------------------------------

func TestOldB035SanctifierExilesBlackAndRedFromEveryGraveyardOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	var ids = map[string]uuid.UUID{}
	add := func(p *game.Player, name string, colors ...string) {
		id := uuid.New()
		ids[name] = id
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", Colors: colors, Owner: p.ID, Controller: p.ID})
	}
	add(me, "Mine Red", "R")
	add(opp, "Their Black", "B")
	add(opp, "Their Blue", "U")
	add(opp, "Their Red-White", "R", "W")
	add(opp, "Their Colorless")
	effectstest.CastCatalogSpell(t, g, "Sanctifier en-Vec", "Creature — Human Cleric", oldB035Sanctifier, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, want := range map[string]game.ZoneKind{
		"Mine Red": game.ZoneExile, "Their Black": game.ZoneExile, "Their Red-White": game.ZoneExile,
		"Their Blue": game.ZoneGraveyard, "Their Colorless": game.ZoneGraveyard,
	} {
		if got := oldB035Zone(g, ids[name]); got != want {
			t.Errorf("%s is in %q, want %q", name, got, want)
		}
	}
}

func TestOldB035SanctifierExilesBlackAndRedInsteadOfTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Sanctifier en-Vec", "Creature — Human Cleric", oldB035Sanctifier, false)
	redDies := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear", Colors: []string{"R"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	greenDies := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	for _, id := range []uuid.UUID{redDies, greenDies} {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	if got := oldB035Zone(g, redDies); got != game.ZoneExile {
		t.Errorf("the red creature is in %q, want exile instead of the graveyard", got)
	}
	if got := oldB035Zone(g, greenDies); got != game.ZoneGraveyard {
		t.Errorf("the green creature is in %q, want the graveyard", got)
	}
	// A black card discarded from a hand is exiled too.
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Pitched", TypeLine: "Sorcery", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, card); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if got := oldB035Zone(g, card); got != game.ZoneExile {
		t.Errorf("the black card is in %q, want exile", got)
	}
}

func TestOldB035SanctifierHasProtectionFromBlackAndRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s := effectstest.PushCatalogPermanent(g, me.ID, "Sanctifier en-Vec", "Creature — Human Cleric", oldB035Sanctifier, false)
	for _, kw := range []string{"protection from black", "protection from red"} {
		if !effectstest.HasEffectiveKeyword(t, g, s, kw) {
			t.Errorf("missing %q", kw)
		}
	}
}

// --- Piercing Rays ------------------------------------------------------------------

func TestOldB035PiercingRaysExilesOnlyTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tapped := effectstest.PushTappedVanillaCreature(g, opp.ID, "Tapped", 2, 2)
	untapped := effectstest.PushVanillaCreature(g, opp.ID, "Untapped", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Piercing Rays", "Sorcery", oldB035Rays, []game.TargetRef{oldB035Card(untapped)}); err == nil {
		t.Error("an untapped creature is not a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Piercing Rays", "Sorcery", oldB035Rays, []game.TargetRef{oldB035Card(tapped)})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB035Zone(g, tapped) != game.ZoneExile {
		t.Errorf("the tapped creature is in %q, want exile", oldB035Zone(g, tapped))
	}
}

func TestOldB035PiercingRaysForecastTapsAnUntappedCreatureInYourUpkeepOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rays := effectstest.HandCardForTest(me, "Piercing Rays", "Sorcery", oldB035Rays)
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.FillPoolColored(me, "W", 3)
	if err := g.ActivateCatalogAbility(me.ID, rays, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB035Card(a)}}); err != nil {
		t.Fatalf("forecast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedOnBattlefield(t, g, a) {
		t.Error("the targeted creature is not tapped")
	}
	if oldB035Zone(g, rays) != game.ZoneHand {
		t.Errorf("the card is in %q, want it still in the hand", oldB035Zone(g, rays))
	}
	effectstest.FillPoolColored(me, "W", 3)
	if err := g.ActivateCatalogAbility(me.ID, rays, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB035Card(b)}}); err == nil {
		t.Error("forecast may be activated only once each turn")
	}
}

func TestOldB035PiercingRaysForecastIsRefusedOutsideYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rays := effectstest.HandCardForTest(me, "Piercing Rays", "Sorcery", oldB035Rays)
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "W", 3)
	if err := g.ActivateCatalogAbility(me.ID, rays, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB035Card(a)}}); err == nil {
		t.Error("forecast is for your upkeep only")
	}
}

// --- Loathsome Curator ----------------------------------------------------------------

func TestOldB035CuratorExploitDestroysASmallCreatureYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{3}{G}",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Loathsome Curator", "Creature — Gorgon Wizard", oldB035Curator, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrPick(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.PassUntilPickTarget(t, g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, big) {
			t.Error("a mana value 4 creature is offered as a target")
		}
		effectstest.PickCard(t, g, me.ID, small)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if oldB035Zone(g, fodder) != game.ZoneGraveyard {
		t.Errorf("the exploited creature is in %q", oldB035Zone(g, fodder))
	}
	if oldB035Zone(g, small) != game.ZoneGraveyard {
		t.Errorf("the small creature is in %q, want the graveyard", oldB035Zone(g, small))
	}
	if oldB035Zone(g, big) != game.ZoneBattlefield {
		t.Error("the big creature was destroyed")
	}
}

// --- Gorilla Shaman ---------------------------------------------------------------------

func TestOldB035GorillaShamanDestroysAnArtifactWithManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gor := effectstest.PushCatalogPermanent(g, me.ID, "Gorilla Shaman", "Creature — Ape Shaman", oldB035Gorilla, false)
	two := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Signet", TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID,
	})
	three := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Idol", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID,
	})
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{2}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.FillPoolColored(me, "R", 5)
	// X = 2 costs {X}{X}{1} = five mana. A mana value 3 artifact and a creature are not legal.
	for _, bad := range []uuid.UUID{three, golem} {
		if err := g.ActivateCatalogAbility(me.ID, gor, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: []game.TargetRef{oldB035Card(bad)}}); err == nil {
			t.Errorf("target %s should be illegal for X=2", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, gor, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: []game.TargetRef{oldB035Card(two)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB035Zone(g, two) != game.ZoneGraveyard {
		t.Errorf("the mana value 2 artifact is in %q, want the graveyard", oldB035Zone(g, two))
	}
	if oldB035Zone(g, three) != game.ZoneBattlefield || oldB035Zone(g, golem) != game.ZoneBattlefield {
		t.Error("another permanent was destroyed")
	}
}

func TestOldB035GorillaShamanXCostsTwiceXPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gor := effectstest.PushCatalogPermanent(g, me.ID, "Gorilla Shaman", "Creature — Ape Shaman", oldB035Gorilla, false)
	two := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Signet", TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.FillPoolColored(me, "R", 4)
	if err := g.ActivateCatalogAbility(me.ID, gor, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: []game.TargetRef{oldB035Card(two)}}); err == nil {
		t.Error("X=2 needs five mana, four were not enough")
	}
}

// --- Obsidian Charmaw ----------------------------------------------------------------------

func oldB035Price(t *testing.T, g *game.Game, caster uuid.UUID) string {
	t.Helper()
	card := game.Card{
		InstanceID: uuid.New(), Name: "Obsidian Charmaw", TypeLine: "Creature — Dragon", ManaCost: "{3}{R}{R}",
		OracleID: oldB035Charmaw, Owner: caster, Controller: caster,
	}
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	var out game.ParsedCost
	g.WithWriteLock(func() {
		var err error
		out, err = g.ApplyCostModifiersForEffect(base, game.CostQuery{Game: g, Card: card, Controller: caster, FromZone: game.ZoneHand})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
	})
	return out.String()
}

func TestOldB035CharmawCostsLessForEachOpponentLandThatCouldProduceColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := oldB035Price(t, g, me.ID); got != "{3}{R}{R}" {
		t.Fatalf("base price %s", got)
	}
	effectstest.SeedManaLand(g, opp.ID, "Wastes", "Basic Land", "C")
	effectstest.SeedManaLand(g, opp.ID, "Wastes", "Basic Land", "C")
	effectstest.SeedManaLand(g, opp.ID, "Forest", "Basic Land — Forest", "G")
	effectstest.SeedManaLand(g, me.ID, "Wastes", "Basic Land", "C")
	if got := oldB035Price(t, g, me.ID); got != "{1}{R}{R}" {
		t.Errorf("price %s, want {1}{R}{R} (two opposing colorless lands; yours and the Forest don't count)", got)
	}
	effectstest.SeedManaLand(g, opp.ID, "Wastes", "Basic Land", "C")
	effectstest.SeedManaLand(g, opp.ID, "Wastes", "Basic Land", "C")
	if got := oldB035Price(t, g, me.ID); got != "{R}{R}" {
		t.Errorf("price %s, want {R}{R}: the reduction never touches the coloured part", got)
	}
}

func TestOldB035CharmawEntersAndDestroysANonbasicLandAnOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	basic := effectstest.SeedManaLand(g, opp.ID, "Forest", "Basic Land — Forest", "G")
	nonbasic := effectstest.SeedManaLand(g, opp.ID, "Barren Moor", "Land", "B")
	mine := effectstest.SeedManaLand(g, me.ID, "Barren Moor", "Land", "B")
	effectstest.CastCatalogSpell(t, g, "Obsidian Charmaw", "Creature — Dragon", oldB035Charmaw, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.PassUntilPickTarget(t, g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, basic) || effectstest.HasID(p.PickTargetCards, mine) {
			t.Errorf("%v offers a land that is not a nonbasic land an opponent controls", p.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, nonbasic)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if oldB035Zone(g, nonbasic) != game.ZoneGraveyard {
		t.Errorf("the nonbasic land is in %q, want the graveyard", oldB035Zone(g, nonbasic))
	}
	if oldB035Zone(g, basic) != game.ZoneBattlefield || oldB035Zone(g, mine) != game.ZoneBattlefield {
		t.Error("another land was destroyed")
	}
}
