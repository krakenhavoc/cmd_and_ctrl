package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b229_cards_test.go — OLD batch 229 (Ixalan, #5532): twelve cards.
// Each test plays the real catalog entry through the engine.

const (
	oldB229BishopOracle      = "05058594-608f-4046-bf44-b736a7072f0a"
	oldB229CallOracle        = "94d3de8c-c8a2-4775-9096-d518f227f09a"
	oldB229AncientOracle     = "990dde55-8ee3-41ad-874b-60864ec28ff4"
	oldB229MarchOracle       = "e9e55768-ed62-4998-9922-dbcdeab12aee"
	oldB229AdaptationOracle  = "2188c03f-11b6-4651-914a-b4fdd59127e3"
	oldB229SirenOracle       = "8ed7791d-6d75-4cd7-9428-c731f623b241"
	oldB229DryadsOracle      = "aa125d1c-ab8f-48ff-a6eb-d8fc2154322c"
	oldB229CallerOracle      = "b6ec5a5c-144d-4ac3-bd05-2dc360c670fd"
	oldB229SchoonerOracle    = "3284010f-aeb2-493e-9252-3b953a192a94"
	oldB229DreadnoughtOracle = "7278f382-8249-4981-9d87-47cef7dc13dd"
	oldB229GalleonOracle     = "88b18901-50cd-461c-b1bc-be900210be8e"
	oldB229FlagshipOracle    = "e5afa88a-8bde-47fe-a890-3759f807360f"
)

func TestOldB229EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB229BishopOracle:         "Bishop of Rebirth",
		oldB229CallOracle:           "Grim Captain's Call",
		oldB229AncientOracle:        "Deathless Ancient",
		oldB229MarchOracle:          "March of the Drowned",
		oldB229AdaptationOracle:     "Arcane Adaptation",
		oldB229SirenOracle:          "Dreamcaller Siren",
		oldB229DryadsOracle:         "Old-Growth Dryads",
		oldB229CallerOracle:         "Tempest Caller",
		oldB229SchoonerOracle:       "Sleek Schooner",
		oldB229DreadnoughtOracle:    "Dusk Legion Dreadnought",
		oldB229GalleonOracle:        "Conqueror's Galleon",
		oldB229GalleonOracle + "#1": "Conqueror's Foothold",
		oldB229FlagshipOracle:       "Fell Flagship",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB229In(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestOldB229BishopOfRebirthReturnsACheapCreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bishop := effectstest.PushDiesCreatureForTest(g, me.ID, "Bishop of Rebirth", oldB229BishopOracle, "Creature — Vampire Cleric", 3, 4)
	cheap := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	pricey := effectstest.PushGraveyardPermanent(me, "Big Bear", "Creature — Bear", "{3}{G}")
	artifact := effectstest.PushGraveyardPermanent(me, "Cheap Rock", "Artifact", "{1}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")

	effectstest.DeclareAttack(t, g, opp.ID, bishop)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if !effectstest.HasID(prompt.PickTargetCards, cheap) {
		t.Error("a creature card with mana value 2 is not a legal target")
	}
	for name, id := range map[string]uuid.UUID{"mana value 4": pricey, "an artifact": artifact, "an opponent's card": theirs} {
		if effectstest.HasID(prompt.PickTargetCards, id) {
			t.Errorf("%s is offered", name)
		}
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB229In(me.Graveyard, cheap) || !oldB229In(g.Battlefield, cheap) {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestOldB229BishopOfRebirthIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bishop := effectstest.PushDiesCreatureForTest(g, me.ID, "Bishop of Rebirth", oldB229BishopOracle, "Creature — Vampire Cleric", 3, 4)
	cheap := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	effectstest.DeclareAttack(t, g, opp.ID, bishop)
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB229In(me.Graveyard, cheap) {
		t.Error("the card left the graveyard after the player declined")
	}
}

// oldB229CastCall casts Grim Captain's Call and resolves it as far as its
// first prompt.
func oldB229CastCall(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Grim Captain's Call", "Sorcery", oldB229CallOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
}

func TestOldB229GrimCaptainsCallReturnsOneOfEachType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// A Vampire Pirate is a candidate for both steps, but can be taken once.
	both := effectstest.PushGraveyardPermanent(me, "Vampire Pirate", "Creature — Vampire Pirate", "{2}{B}")
	pirate := effectstest.PushGraveyardPermanent(me, "Plain Pirate", "Creature — Human Pirate", "{1}{B}")
	vampire := effectstest.PushGraveyardPermanent(me, "Plain Vampire", "Creature — Vampire", "{1}{B}")
	dino := effectstest.PushGraveyardPermanent(me, "Plain Dinosaur", "Creature — Dinosaur", "{3}{G}")
	merfolk := effectstest.PushGraveyardPermanent(me, "Plain Merfolk", "Creature — Merfolk", "{1}{U}")
	other := effectstest.PushGraveyardPermanent(me, "Plain Bear", "Creature — Bear", "{1}{G}")

	oldB229CastCall(t, g)
	// Pirate step: the Vampire Pirate or the Plain Pirate.
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || !effectstest.HasID(c.ChooseCards, both) || !effectstest.HasID(c.ChooseCards, pirate) || effectstest.HasID(c.ChooseCards, vampire) {
		t.Fatalf("the Pirate step offers the wrong cards: %+v", c)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, both)
	// Vampire step: the Vampire Pirate is already in hand.
	c = effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || effectstest.HasID(c.ChooseCards, both) || !effectstest.HasID(c.ChooseCards, vampire) {
		t.Fatalf("the Vampire step offers the wrong cards: %+v", c)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, vampire)
	effectstest.AnswerChooseCards(t, g, me.ID, dino)
	effectstest.AnswerChooseCards(t, g, me.ID, merfolk)
	for name, id := range map[string]uuid.UUID{"Vampire Pirate": both, "Vampire": vampire, "Dinosaur": dino, "Merfolk": merfolk} {
		if !effectstest.HandHasID(me, id) {
			t.Errorf("%s is not in hand", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"Plain Pirate": pirate, "Plain Bear": other} {
		if !oldB229In(me.Graveyard, id) {
			t.Errorf("%s left the graveyard", name)
		}
	}
}

func TestOldB229GrimCaptainsCallSkipsTypesWithNoCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dino := effectstest.PushGraveyardPermanent(me, "Plain Dinosaur", "Creature — Dinosaur", "{3}{G}")
	oldB229CastCall(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, dino)
	if !effectstest.HandHasID(me, dino) {
		t.Error("the Dinosaur is not in hand")
	}
	if c := effectstest.LatestChooseCardsFor(g, me.ID); c != nil {
		t.Errorf("a prompt is still open for a type with no card: %+v", c)
	}
}

func TestOldB229MarchOfTheDrownedModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	p1 := effectstest.PushGraveyardPermanent(me, "Pirate One", "Creature — Human Pirate", "{1}{B}")
	p2 := effectstest.PushGraveyardPermanent(me, "Pirate Two", "Creature — Human Pirate", "{1}{B}")

	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB229MarchOracle, []int{0}, effectstest.DftABCardRef(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, bear) {
		t.Error("mode one did not return the creature card")
	}
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB229MarchOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: p1}, {Kind: game.TargetCard, ID: p2}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, p1) || !effectstest.HandHasID(me, p2) {
		t.Error("mode two did not return both Pirate cards")
	}
}

func TestOldB229MarchOfTheDrownedModeTwoNeedsTwoPirates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	p1 := effectstest.PushGraveyardPermanent(me, "Pirate One", "Creature — Human Pirate", "{1}{B}")
	bear := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "March of the Drowned", TypeLine: "Sorcery", OracleID: oldB229MarchOracle, ManaCost: "{B}", Owner: me.ID, Controller: me.ID})
	march := me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID
	effectstest.FloatMana(t, g, me, "{B}")
	err := g.CastSpell(me.ID, march, game.CastSpellParams{Strict: true, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: p1}}})
	if err == nil {
		t.Error("mode two was accepted with one target")
	}
	err = g.CastSpell(me.ID, march, game.CastSpellParams{Strict: true, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: p1}, {Kind: game.TargetCard, ID: bear}}})
	if err == nil {
		t.Error("mode two was accepted with a non-Pirate target")
	}
}

func TestOldB229DeathlessAncientReturnsByTappingThreeVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ancient := effectstest.PushGraveyardPermanent(me, "Deathless Ancient", "Creature — Vampire Knight", "{4}{B}{B}")
	setOracle(me, ancient, oldB229AncientOracle)
	v1 := effectstest.B12Creature(g, me.ID, "Vampire One", "Creature — Vampire", 1, 1)
	v2 := effectstest.B12Creature(g, me.ID, "Vampire Two", "Creature — Vampire", 1, 1)
	v3 := effectstest.B12Creature(g, me.ID, "Vampire Three", "Creature — Vampire", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, ancient, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{v1, v2}}); err == nil {
		t.Error("two Vampires paid for a cost of three")
	}
	if err := g.ActivateCatalogAbility(me.ID, ancient, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{v1, v2, bear}}); err == nil {
		t.Error("a Bear paid for a Vampire")
	}
	if err := g.ActivateCatalogAbility(me.ID, ancient, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{v1, v2, v3}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, ancient) || oldB229In(me.Graveyard, ancient) {
		t.Error("the card did not return to hand")
	}
	for _, id := range []uuid.UUID{v1, v2, v3} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, id); !tapped {
			t.Error("a Vampire was not tapped")
		}
	}
}

func setOracle(p *game.Player, id uuid.UUID, oracle string) {
	for i := range p.Graveyard.Cards {
		if p.Graveyard.Cards[i].InstanceID == id {
			p.Graveyard.Cards[i].OracleID = oracle
		}
	}
}

func TestOldB229ArcaneAdaptationAddsTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	inHand := effectstest.HandCardFull(me, "Hand Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	effectstest.CastCatalogSpell(t, g, "Arcane Adaptation", "Enchantment", oldB229AdaptationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Pirate")
	c := effectstest.B12Card(t, g, bear)
	if !c.HasSubtype("Pirate") || !c.HasSubtype("Bear") {
		t.Errorf("the Bear is %q, want a Bear and a Pirate", c.TypeLine)
	}
	if hc, ok := effectstest.HandCardByID(g, me, inHand); !ok || !hc.HasSubtype("Pirate") {
		t.Error("a creature card in hand did not become a Pirate")
	}
}

func TestOldB229DreamcallerSirenTapsTwoPermanentsWithAnotherPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Pirate Friend", "Creature — Human Pirate", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Elk A", "Creature — Elk", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Elk B", "Creature — Elk", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Elk C", "Creature — Elk", 2, 2)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Dreamcaller Siren", TypeLine: "Creature — Siren Pirate", OracleID: oldB229SirenOracle,
		ManaCost: "{2}{U}{U}", Power: 3, Toughness: 3})
	prompt := effectstest.FirstPendingTarget(g)
	if prompt == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, want := range map[string]struct {
		id  uuid.UUID
		tap bool
	}{"A": {a, true}, "B": {b, true}, "C": {c, false}} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, want.id); tapped != want.tap {
			t.Errorf("Elk %s tapped = %v, want %v", name, tapped, want.tap)
		}
	}
}

func TestOldB229DreamcallerSirenWithoutAnotherPirateDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "Elk A", "Creature — Elk", 2, 2)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Dreamcaller Siren", TypeLine: "Creature — Siren Pirate", OracleID: oldB229SirenOracle,
		ManaCost: "{2}{U}{U}", Power: 3, Toughness: 3})
	if p := effectstest.FirstPendingTarget(g); p != nil {
		t.Fatal("the trigger asked for targets with no other Pirate")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, a); tapped {
		t.Error("a permanent was tapped")
	}
}

func TestOldB229OldGrowthDryadsLetsEachOpponentFetchABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	effectstest.DftABCast(t, g, me, game.Card{Name: "Old-Growth Dryads", TypeLine: "Creature — Dryad", OracleID: oldB229DryadsOracle,
		ManaCost: "{G}", Power: 3, Toughness: 3})
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok || c.Controller != opp.ID || !c.Tapped {
		t.Errorf("the Forest is on the battlefield=%v, controller %v, tapped=%v; want tapped under the opponent", ok, c.Controller, c.Tapped)
	}
}

func TestOldB229TempestCallerTapsEveryCreatureTheTargetOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "Elk A", "Creature — Elk", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Elk B", "Creature — Elk", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Elk", 2, 2)
	theirs := effectstest.B12Creature(g, other.ID, "Other", "Creature — Elk", 2, 2)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Tempest Caller", TypeLine: "Creature — Merfolk Wizard", OracleID: oldB229CallerOracle,
		ManaCost: "{2}{U}{U}", Power: 2, Toughness: 3})
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for name, want := range map[string]struct {
		id  uuid.UUID
		tap bool
	}{"A": {a, true}, "B": {b, true}, "mine": {mine, false}, "other": {theirs, false}} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, want.id); tapped != want.tap {
			t.Errorf("%s tapped = %v, want %v", name, tapped, want.tap)
		}
	}
}

func TestOldB229SleekSchoonerAndDreadnoughtCrew(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	schooner := effectstest.DftABPush(g, me.ID, "Sleek Schooner", "Artifact — Vehicle", oldB229SchoonerOracle, 4, 3)
	dread := effectstest.DftABPush(g, me.ID, "Dusk Legion Dreadnought", "Artifact — Vehicle", oldB229DreadnoughtOracle, 4, 6)
	one := effectstest.PushCrewerForTest(g, me.ID, "One", 1)
	two := effectstest.PushCrewerForTest(g, me.ID, "Two", 2)
	if err := g.ActivateCatalogAbility(me.ID, dread, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Error("power 1 crewed Crew 2")
	}
	effectstest.CrewForTest(t, g, me.ID, schooner, one)
	effectstest.DftABCrew(t, g, me, dread, 0, two)
}

func TestOldB229FellFlagshipPumpsPiratesAndMakesThePlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	flag := effectstest.DftABPush(g, me.ID, "Fell Flagship", "Artifact — Vehicle", oldB229FlagshipOracle, 3, 3)
	pirate := effectstest.B12Creature(g, me.ID, "Pirate", "Creature — Human Pirate", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Pirate", "Creature — Human Pirate", 2, 2)
	if pw := effectstest.EffectivePower(t, g, pirate); pw != 3 {
		t.Errorf("my Pirate has power %d, want 3", pw)
	}
	if pw := effectstest.EffectivePower(t, g, bear); pw != 2 {
		t.Errorf("my Bear has power %d, want 2", pw)
	}
	if pw := effectstest.EffectivePower(t, g, theirs); pw != 2 {
		t.Errorf("their Pirate has power %d, want 2", pw)
	}
	effectstest.DftABCrew(t, g, me, flag, 0, effectstest.PushCrewerForTest(g, me.ID, "Crew", 3))
	card := effectstest.HandCardFull(opp, "Their Card", "Sorcery", "{1}", "", nil)
	handBefore := opp.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, flag)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	t.Logf("choices %d", len(g.PendingChoices))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, card)
	if opp.Hand.Size() != handBefore-1 || !oldB229In(opp.Graveyard, card) {
		t.Errorf("their hand is %d, want %d with their card discarded", opp.Hand.Size(), handBefore-1)
	}
}

func TestOldB229ConquerorsGalleonBecomesAFoothold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	row := effectstest.TransformRow(oldB229GalleonOracle, "Conqueror's Galleon", "Artifact — Vehicle", "{4}",
		"Conqueror's Foothold", "Land", "", "", nil)
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = "2", "10"
	galleon := effectstest.PushCraftCard(g, me, row)
	crew := effectstest.PushCrewerForTest(g, me.ID, "Crew", 4)
	if err := g.ActivateCatalogAbility(me.ID, galleon, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{effectstest.PushCrewerForTest(g, me.ID, "Small", 3)}}); err == nil {
		t.Error("power 3 crewed Crew 4")
	}
	effectstest.DftABCrew(t, g, me, galleon, 0, crew)
	effectstest.DeclareAttack(t, g, opp.ID, galleon)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, galleon); !ok || c.IsLand() {
		t.Fatal("the Galleon changed before the end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	foothold := effectstest.CraftedPermanent(t, g, "Conqueror's Foothold")
	if foothold.Controller != me.ID || !foothold.IsLand() || foothold.IsCreature() {
		t.Errorf("Conqueror's Foothold is %q under %v, want a land under me", foothold.TypeLine, foothold.Controller)
	}
	if foothold.InstanceID != galleon {
		t.Log("the Foothold is a new object, as exile and return makes it")
	}
}

func TestOldB229ConquerorsFootholdAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	row := effectstest.TransformRow(oldB229GalleonOracle, "Conqueror's Galleon", "Artifact — Vehicle", "{4}",
		"Conqueror's Foothold", "Land", "", "", nil)
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = "2", "10"
	id := effectstest.PushCraftCard(g, me, row)
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("transform: %v", err)
	}
	dead := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 2, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.DftABCardRef(dead)}); err != nil {
		t.Fatalf("the {6} ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, dead) {
		t.Error("the card did not return to hand")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second tap ability activated on a tapped land")
	}
}
