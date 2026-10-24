package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_permission_riders_test.go — what a cast permission puts on the
// permanent its cast produces (#3768): Edgar, Master Machinist's "that
// artifact enters tapped" and Strago and Relm's "it gains haste and 'At
// the beginning of the end step, sacrifice this creature.'" (deck request
// #3655).

const (
	edgarMasterMachinistOracle = "75348d55-99cd-4dab-97be-1f818c43a69d"
	stragoAndRelmOracle        = "316e7c28-2b7a-40ae-961b-909644109fe4"
)

// edgarGame is a game in seat 0's precombat main with Edgar on the
// battlefield and two free artifact cards in seat 0's graveyard.
func edgarGame(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	pushCatalogPermanent(g, me.ID, "Edgar, Master Machinist", "Legendary Creature — Human Artificer Noble", edgarMasterMachinistOracle, false)
	a := uuid.New()
	b := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "Gear A", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "Gear B", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	return g, me, a, b
}

// TestEdgarCastsOneArtifactFromTheGraveyardThatEntersTapped: the
// artifact enters tapped — through the entry, so no tap event — and the
// permission is spent for the turn.
func TestEdgarCastsOneArtifactFromTheGraveyardThatEntersTapped(t *testing.T) {
	g, me, a, b := edgarGame(t)
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(a)
	if !ok || !g.Battlefield.Contains(a) {
		t.Fatal("the artifact did not reach the battlefield")
	}
	if !c.Tapped {
		t.Fatal("the artifact cast through Edgar entered untapped")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventTapCard && ev.CardID == a {
			t.Fatal("the artifact was tapped after it entered; it must ENTER tapped")
		}
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Fatal("a second artifact was cast from the graveyard in the same turn")
	}
}

// TestEdgarDoesNotTapAnArtifactCastFromHand: the rider is the
// permission's; the same artifact hard-cast from hand enters untapped.
func TestEdgarDoesNotTapAnArtifactCastFromHand(t *testing.T) {
	g, me, _, _ := edgarGame(t)
	id := castCatalogSpell(t, g, "Gear C", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(id); !ok || c.Tapped {
		t.Fatalf("an artifact cast from hand entered tapped (found %v)", ok)
	}
	_ = me
}

// TestEdgarToolsPumpsByTheGreatestArtifactManaValue: +X/+0 where X is the
// greatest mana value among artifacts you control.
func TestEdgarToolsPumpsByTheGreatestArtifactManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	edgar := pushCatalogPermanent(g, me.ID, "Edgar, Master Machinist", "Legendary Creature — Human Artificer Noble", edgarMasterMachinistOracle, false)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Gear", TypeLine: "Artifact", ManaCost: "{5}", Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small Gear", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	before := effectivePower(t, g, edgar)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: edgar, Actor: me.ID, Target: g.Seats[1].ID})
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, edgar); got != before+5 {
		t.Fatalf("Tools: power %d -> %d, want +5", before, got)
	}
}

// stragoGame is seat 0 in its precombat main with an untapped Strago and
// Relm and {2}{R} floating; seat 1's library has a land on top of a
// creature card. Returns the game and the two library cards.
func stragoGame(t *testing.T) (g *game.Game, me, opp *game.Player, strago, land, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	strago = pushCatalogPermanent(g, me.ID, "Strago and Relm", "Legendary Creature — Human Wizard", stragoAndRelmOracle, false)
	bear, land = uuid.New(), uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: bear, Name: "Stolen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: land, Name: "Some Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, strago, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("Sketch and Lore: %v", err)
	}
	passPriorityAroundTable(t, g)
	return g, me, opp, strago, land, bear
}

// TestStragoAndRelmCastCreatureGainsHasteAndIsSacrificedAtEndStep: the
// run exiles down to the creature card, the free cast makes it yours
// with haste, and the granted trigger sacrifices it at the end step into
// its owner's graveyard. The land exiled on the way stays in exile.
func TestStragoAndRelmCastCreatureGainsHasteAndIsSacrificedAtEndStep(t *testing.T) {
	g, me, opp, _, land, bear := stragoGame(t)
	if z := g.FindCardZoneForEffect(land); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the land on the way was not exiled")
	}
	if z := g.FindCardZoneForEffect(bear); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the run did not stop on the creature card in exile")
	}
	// #4250: the free cast is made during the ability's resolution.
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bear)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("free cast of the exiled creature: %v", err)
	}
	castInsideTheResolution(t, g, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the cast creature did not resolve")
	}
	c, _ := g.LookupCardForEffect(bear)
	if c.Controller != me.ID {
		t.Fatalf("the creature is controlled by %s, want the caster", c.Controller)
	}
	if !effectiveAbilitiesContain(t, g, bear, "haste") {
		t.Fatal("the creature cast this way has no haste")
	}

	// An effect with no duration outlives a control change.
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, bear, opp.ID, game.IndefiniteDuration(), "test — control change")
	})
	g.RunStateChecksForTest()
	if !effectiveAbilitiesContain(t, g, bear, "haste") {
		t.Fatal("the haste ended when control changed; it has no duration")
	}

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the granted end-step trigger did not sacrifice the creature")
	}
	if !opp.Graveyard.Contains(bear) {
		t.Fatal("the sacrificed creature is not in its owner's graveyard")
	}
}

// TestStragoAndRelmHardCastCreatureGainsNothing: the gains belong to the
// permission. Any other creature you cast has neither haste nor the
// sacrifice.
func TestStragoAndRelmHardCastCreatureGainsNothing(t *testing.T) {
	g, me, _, _, _, bear := stragoGame(t)
	// #4250: declining the offer leaves the exiled creature in exile.
	declineResolutionCast(t, g, me.ID)
	if z := g.FindCardZoneForEffect(bear); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the declined creature card left exile")
	}
	other := castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if effectiveAbilitiesContain(t, g, other, "haste") {
		t.Fatal("a hard-cast creature gained haste")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(other) {
		t.Fatal("a hard-cast creature was sacrificed at the end step")
	}
	_ = me
}
