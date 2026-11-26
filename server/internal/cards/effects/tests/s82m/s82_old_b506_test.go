package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b506_test.go — Conflux, batch 506 (#5587). Every test plays the real
// catalog entry through the engine.

const (
	oldB506Fleshformer = "67b8b67d-fda3-4a62-9bfb-51a171638502"
	oldB506Charge      = "7b052ce1-ee82-4335-bc1d-9563bc59b948"
	oldB506Sunsinger   = "dcc82c14-d632-4333-be35-4a206a10b184"
	oldB506Knight      = "f8fb5a40-4e32-4817-bf56-484add976475"
	oldB506Counsel     = "079b1f07-6910-491b-83ba-5ec41260f32d"
	oldB506View        = "3a128587-c751-4b27-b608-d4e47646b80e"
	oldB506Telemin     = "4014f665-08a6-4ae0-b4d7-5a69d043af6a"
	oldB506Borders     = "1a19fc9e-a822-49ca-ade0-d58054f3caf6"
	oldB506Shard       = "c51fc4f2-76a6-4337-ab8f-1b3e84f82019"
	oldB506Fracture    = "e2376660-20f2-4474-b6fd-d270e0e6a31e"
	oldB506Roots       = "0c6f343a-0088-477a-a6b5-256783ad0c27"
	oldB506Instincts   = "7d56d2e2-e83b-4505-af8e-1fd207e161af"
	oldB506Bear        = "Creature — Bear"
)

func TestOldB506EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB506Fleshformer: "Fleshformer", oldB506Charge: "Suicidal Charge", oldB506Sunsinger: "Cylian Sunsinger",
		oldB506Knight: "Dragonsoul Knight", oldB506Counsel: "Worldly Counsel", oldB506View: "View from Above",
		oldB506Telemin: "Telemin Performance", oldB506Borders: "Exploding Borders", oldB506Shard: "Shard Convergence",
		oldB506Fracture: "Filigree Fracture", oldB506Roots: "Corrupted Roots", oldB506Instincts: "Controlled Instincts",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB506Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	g.WithWriteLock(func() { who.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, who, mana)
	params.Strict = true
	err := g.ActivateCatalogAbility(who.ID, id, 0, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB506FleshformerPumpsItselfAndShrinksATargetOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flesh := effectstest.B12Push(g, me.ID, "Fleshformer", "Creature — Human Wizard", oldB506Fleshformer, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB506Bear, 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := oldB506Activate(t, g, me, flesh, "WUBR", game.ActivateAbilityParams{Targets: effectstest.CardTargets(victim)}); err == nil {
		t.Fatal("activated with four colours of mana")
	}
	if err := oldB506Activate(t, g, me, flesh, "WUBRG", game.ActivateAbilityParams{Targets: effectstest.CardTargets(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p, tt := effectstest.PtOf(t, g, flesh); p != 4 || tt != 4 {
		t.Errorf("Fleshformer is %d/%d, want 4/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, flesh, "fear") {
		t.Error("Fleshformer has no fear")
	}
	if p, tt := effectstest.PtOf(t, g, victim); p != 1 || tt != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tt)
	}
}

func TestOldB506FleshformerCantBeActivatedOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flesh := effectstest.B12Push(g, me.ID, "Fleshformer", "Creature — Human Wizard", oldB506Fleshformer, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB506Bear, 3, 3)
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := oldB506Activate(t, g, me, flesh, "WUBRG", game.ActivateAbilityParams{Targets: effectstest.CardTargets(victim)}); err == nil {
		t.Error("activated on the opponent's turn")
	}
}

func TestOldB506DragonsoulKnightBecomesAFlyingTramplingDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	knight := effectstest.B12Push(g, me.ID, "Dragonsoul Knight", "Creature — Human Knight", oldB506Knight, 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := oldB506Activate(t, g, me, knight, "WUBRG", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p, tt := effectstest.PtOf(t, g, knight); p != 7 || tt != 5 {
		t.Errorf("the Knight is %d/%d, want 7/5", p, tt)
	}
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, knight, kw) {
			t.Errorf("the Knight lacks %s", kw)
		}
	}
	c, _ := g.LookupCardForEffect(knight)
	if !c.HasSubtypeWord("Dragon") || c.HasSubtypeWord("Human") || c.HasSubtypeWord("Knight") {
		t.Errorf("subtypes after the change: %v, want only Dragon", c.Effective().Subtypes)
	}
}

func TestOldB506CylianSunsingerPumpsEveryCreatureWithTheSameName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sun := effectstest.B12Push(g, me.ID, "Cylian Sunsinger", "Creature — Elf Shaman", oldB506Sunsinger, 2, 2)
	twin := effectstest.B12Creature(g, opp.ID, "Cylian Sunsinger", "Creature — Elf Shaman", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Elf", "Creature — Elf Shaman", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := oldB506Activate(t, g, me, sun, "RGW", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range []uuid.UUID{sun, twin} {
		if p, tt := effectstest.PtOf(t, g, id); p != 5 || tt != 5 {
			t.Errorf("a Cylian Sunsinger is %d/%d, want 5/5", p, tt)
		}
	}
	if p, tt := effectstest.PtOf(t, g, other); p != 2 || tt != 2 {
		t.Errorf("a differently named creature is %d/%d, want 2/2", p, tt)
	}
}

func TestOldB506SuicidalChargeShrinksAndForcesOnlyTheCreaturesThereWhenItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Charge belongs to the non-active player, so the active player's
	// creatures are its "creatures your opponents control".
	charge := effectstest.PushCatalogPermanent(g, opp.ID, "Suicidal Charge", "Enchantment", oldB506Charge, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB506Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB506Bear, 2, 2)
	if err := g.ActivateCatalogAbility(opp.ID, charge, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(charge) {
		t.Error("the Charge was not sacrificed")
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 1 || tt != 1 {
		t.Errorf("the opponent's creature is %d/%d, want 1/1", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("the Charge controller's own creature is %d/%d, want 2/2", p, tt)
	}
	late := effectstest.B12Creature(g, me.ID, "Late", oldB506Bear, 2, 2)
	if p, tt := effectstest.PtOf(t, g, late); p != 2 || tt != 2 {
		t.Errorf("a creature that arrived later is %d/%d, want 2/2", p, tt)
	}
	if !effectstest.FaMustAttack(t, g) {
		t.Error("the shrunken creature is not required to attack")
	}
}

func oldB506CastCounsel(t *testing.T, lands int) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tl := range []string{"Basic Land — Plains", "Basic Land — Island"}[:lands] {
		effectstest.SeedLand(g, me.ID, tl, tl, "")
	}
	effectstest.CastCatalogSpell(t, g, "Worldly Counsel", "Instant", oldB506Counsel, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me
}

func TestOldB506WorldlyCounselLooksAtOneCardPerBasicLandType(t *testing.T) {
	g, me := oldB506CastCounsel(t, 2)
	ids := make([]uuid.UUID, 0, 4)
	n := len(me.Library.Cards)
	for i := 0; i < 4; i++ {
		ids = append(ids, me.Library.Cards[n-1-i].InstanceID)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("prompt over %v, want the top two cards (two basic land types)", pick)
	}
	hand := len(me.Hand.Cards)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || len(me.Hand.Cards) != hand+1 {
		t.Error("the chosen card did not reach the hand")
	}
	bottom := me.Library.Cards[0].InstanceID
	if bottom != ids[0] && me.Library.Cards[1].InstanceID != ids[0] {
		t.Error("the other looked-at card is not on the bottom")
	}
}

func TestOldB506WorldlyCounselWithNoBasicLandTypesDoesNothing(t *testing.T) {
	g, me := oldB506CastCounsel(t, 0)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to choose a card with no basic land type among your lands")
	}
}

func oldB506CastView(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "View from Above", "Instant", oldB506View, effectstest.CardTargets(target))
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB506ViewFromAboveReturnsToHandOnlyWithAWhitePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB506Bear, 2, 2)
	spell := oldB506CastView(t, g, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	if effectstest.ZoneOfCard(g, spell) != game.ZoneGraveyard {
		t.Errorf("without a white permanent the spell is in zone %v, want the graveyard", effectstest.ZoneOfCard(g, spell))
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	bear = effectstest.B12Creature(g, me.ID, "Bear", oldB506Bear, 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Knight", TypeLine: oldB506Bear,
		Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	spell = oldB506CastView(t, g, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	if !effectstest.HandHas(me, spell) {
		t.Errorf("with a white permanent the spell is in zone %v, want the hand", effectstest.ZoneOfCard(g, spell))
	}
}

func TestOldB506TeleminPerformanceMillsTheNoncreaturesAndStealsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mk := func(name, tl string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: tl, Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID}
	}
	below, hit, land, spell := mk("Below", oldB506Bear), mk("Hit", "Creature — Elf"), mk("Land", "Basic Land — Forest"), mk("Spell", "Instant")
	g.WithWriteLock(func() { opp.Library.Cards = []game.Card{below, hit, land, spell} }) // the last card is the top
	belowLen := 1
	effectstest.CastCatalogSpell(t, g, "Telemin Performance", "Sorcery", oldB506Telemin,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []game.Card{land, spell} {
		if z := g.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneGraveyard {
			t.Errorf("%s is in %v, want the opponent's graveyard", c.Name, z)
		}
	}
	if !g.Battlefield.Contains(hit.InstanceID) {
		t.Fatal("the creature card is not on the battlefield")
	}
	if got := effectstest.ControllerOf(t, g, hit.InstanceID); got != me.ID {
		t.Error("the creature is not under the caster's control")
	}
	if len(opp.Library.Cards) != belowLen {
		t.Errorf("library has %d cards, want %d", len(opp.Library.Cards), belowLen)
	}
}

func TestOldB506TeleminPerformanceWithNoCreatureMillsTheWholeLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		opp.Library.Cards = []game.Card{
			{InstanceID: uuid.New(), Name: "Land", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID},
			{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID},
		}
	})
	gy := len(opp.Graveyard.Cards)
	effectstest.CastCatalogSpell(t, g, "Telemin Performance", "Sorcery", oldB506Telemin,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Library.Cards) != 0 || len(opp.Graveyard.Cards) != gy+2 {
		t.Errorf("library %d, graveyard %d, want an empty library and %d in the graveyard", len(opp.Library.Cards), len(opp.Graveyard.Cards), gy+2)
	}
}

func TestOldB506ExplodingBordersFetchesALandTappedThenDealsDomainDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: oldB506Bear})
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Exploding Borders", "Sorcery", oldB506Borders,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	var forest *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Forest" {
			forest = &g.Battlefield.Cards[i]
		}
	}
	if forest == nil || !forest.Tapped {
		t.Fatalf("the Forest is %v, want a tapped Forest on the battlefield", forest)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before-2 {
		t.Errorf("life %d, want %d (Mountain and Forest: 2 basic land types)", got, before-2)
	}
}

func TestOldB506ShardConvergenceTakesOneLandOfEachTypeAndRefusesDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Scalding Tarn Dual", "Land — Plains Island"))
	effectstest.CastCatalogSpell(t, g, "Shard Convergence", "Sorcery", oldB506Shard, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("the search offered a Forest")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	picks := []uuid.UUID{
		effectstest.SearchOptionNamed(g, c, "Scalding Tarn Dual"),
		effectstest.SearchOptionNamed(g, c, "Plains"),
		effectstest.SearchOptionNamed(g, c, "Island"),
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, picks); err == nil {
		t.Fatal("a dual land and both its types were accepted as three slots")
	}
	hand := len(me.Hand.Cards)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains", "Island", "Swamp", "Mountain")
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+4 {
		t.Errorf("hand grew by %d, want 4", len(me.Hand.Cards)-hand)
	}
}

func oldB506Permanent(g *game.Game, owner uuid.UUID, name, typeLine string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Colors: colors, Owner: owner, Controller: owner})
}

func TestOldB506FiligreeFractureDrawsOnlyForABlueOrBlackPermanent(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		colors         []string
		draw           bool
	}{
		{"Blue Idol", "Enchantment", []string{"U"}, true},
		{"Black Relic", "Artifact", []string{"B"}, true},
		{"Green Totem", "Enchantment", []string{"G"}, false},
		{"Plain Rock", "Artifact", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			target := oldB506Permanent(g, opp.ID, tc.name, tc.typeLine, tc.colors...)
			hand := len(me.Hand.Cards)
			effectstest.CastCatalogSpell(t, g, "Filigree Fracture", "Instant", oldB506Fracture, effectstest.CardTargets(target))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(target) {
				t.Error("the permanent was not destroyed")
			}
			want := hand
			if tc.draw {
				want++
			}
			if len(me.Hand.Cards) != want {
				t.Errorf("hand %d, want %d", len(me.Hand.Cards), want)
			}
		})
	}
}

func TestOldB506CorruptedRootsDrainsTheLandsControllerWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mountain := effectstest.SeedLand(g, opp.ID, "Mountain", "Basic Land — Mountain", "")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Corrupted Roots", effectstest.AuraTypeLine, oldB506Roots, effectstest.CardTargets(mountain)); err == nil {
		t.Fatal("enchanted a Mountain")
	}
	forest := effectstest.SeedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
	effectstest.Enchant(t, g, "Corrupted Roots", oldB506Roots, forest)
	before, mine := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, me.ID)
	effectstest.TapForTest(t, g, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-2 {
		t.Errorf("the land's controller has %d life, want %d", got, before-2)
	}
	if got := effectstest.LifeOf(g, me.ID); got != mine {
		t.Errorf("the Aura's controller lost life: %d -> %d", mine, got)
	}
}

func TestOldB506ControlledInstinctsKeepsARedOrGreenCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	blue := oldB506Permanent(g, opp.ID, "Blue Bear", oldB506Bear, "U")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Controlled Instincts", effectstest.AuraTypeLine, oldB506Instincts, effectstest.CardTargets(blue)); err == nil {
		t.Fatal("enchanted a blue creature")
	}
	red := oldB506Permanent(g, opp.ID, "Red Bear", oldB506Bear, "R")
	effectstest.TapForTest(t, g, red)
	effectstest.TapForTest(t, g, blue)
	effectstest.Enchant(t, g, "Controlled Instincts", oldB506Instincts, red)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if !effectstest.TappedNow(g, red) {
		t.Error("the enchanted creature untapped")
	}
	if effectstest.TappedNow(g, blue) {
		t.Error("the unenchanted creature stayed tapped")
	}
}

func TestOldB506FiligreeFractureStillDrawsForAnIndestructibleBluePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Blue Anvil", TypeLine: "Artifact",
		Colors: []string{"U"}, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Filigree Fracture", "Instant", oldB506Fracture, effectstest.CardTargets(target))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(target) {
		t.Error("the indestructible permanent was destroyed")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d (it was blue, whether or not it was destroyed)", len(me.Hand.Cards), hand+1)
	}
}
