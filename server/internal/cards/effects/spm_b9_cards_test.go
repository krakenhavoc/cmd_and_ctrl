package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b9_cards_test.go — Marvel's Spider-Man, batch 9 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spmB9BrooklynOracle   = "2ca85872-9436-4515-af50-c13e59605627"
	spmB9SlayerOracle     = "ca68f436-2522-4007-87bf-2669c7743db0"
	spmB9SuitOracle       = "ba08b4cf-9a13-40fd-8522-9504d2ff1a19"
	spmB9SpidersManOracle = "d21d7557-9702-46c1-8277-42a3442084c2"
	spmB9WreckingOracle   = "3a07e22c-71cd-4723-b07e-0ff02009a878"
	spmB9WillOracle       = "0389b48a-e64c-44fb-8a41-fd571cba237c"
	spmB9SanctuaryOracle  = "6ef30340-a26d-49aa-bc86-0b8aa5252f87"
	spmB9TrainOracle      = "6d4472ad-3fbf-4a5d-94de-dfee04fda9c3"
	spmB9SunSpiderOracle  = "2d961997-4f5f-4ce2-b062-11411147e2f7"
	spmB9ParentsOracle    = "b1266feb-516e-4b44-b750-7988ca6e6e55"
)

func TestSpmB9EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB9BrooklynOracle:   "Spider-Man, Brooklyn Visionary",
		spmB9SlayerOracle:     "Spider-Slayer, Hatred Honed",
		spmB9SuitOracle:       "Spider-Suit",
		spmB9SpidersManOracle: "Spiders-Man, Heroic Horde",
		spmB9WreckingOracle:   "Steel Wrecking Ball",
		spmB9WillOracle:       "Strength of Will",
		spmB9SanctuaryOracle:  "Suburban Sanctuary",
		spmB9TrainOracle:      "Subway Train",
		spmB9SunSpiderOracle:  "Sun-Spider, Nimble Webber",
		spmB9ParentsOracle:    "Supportive Parents",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// spmB9Cast puts a card in the active player's hand and casts it for free
// (the sandbox cast), returning its id and the cast's error.
func spmB9Cast(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, colors []string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(active, name, typeLine, cost, oracle, colors)
	for i := range active.Hand.Cards {
		if active.Hand.Cards[i].InstanceID == id {
			active.Hand.Cards[i].Power, active.Hand.Cards[i].Toughness = power, toughness
		}
	}
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// spmB9Named lists the battlefield cards a player controls with a name.
func spmB9Named(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

// --- Spider-Man, Brooklyn Visionary ---

func TestSpmB9BrooklynFetchesATappedBasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Wasteland", TypeLine: "Land"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	spmB9Cast(t, g, "Spider-Man, Brooklyn Visionary", "Legendary Creature — Spider Human Hero", spmB9BrooklynOracle, "{4}{G}", []string{"G"}, 4, 3)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[2])
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, ids[2])
	if !ok {
		t.Fatal("the Island was not put onto the battlefield")
	}
	if !card.Tapped {
		t.Error("the fetched land should enter tapped")
	}
	if !me.Library.Contains(ids[0]) || !me.Library.Contains(ids[1]) {
		t.Error("the other cards should stay in the library")
	}
}

func TestSpmB9BrooklynOffersNoNonbasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Wasteland", TypeLine: "Land"})
	spmB9Cast(t, g, "Spider-Man, Brooklyn Visionary", "Legendary Creature — Spider Human Hero", spmB9BrooklynOracle, "{4}{G}", []string{"G"}, 4, 3)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil && len(c.SearchCards) > 0 {
		t.Errorf("a nonbasic land was offered: %+v", c.SearchCards)
	}
}

func TestSpmB9BrooklynWebSlingsForTwoGreenAndATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spider := handCardFull(me, "Spider-Man, Brooklyn Visionary", "Legendary Creature — Spider Human Hero", "{4}{G}", spmB9BrooklynOracle, []string{"G"})
	untapped := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{G}")

	if err := g.CastSpell(me.ID, spider, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the {4}{G} price was accepted with three mana")
	}
	if err := g.CastSpell(me.ID, spider, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{untapped},
	}); err == nil {
		t.Fatal("web-slinging accepted an untapped creature")
	}
	if err := g.CastSpell(me.ID, spider, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{tapped},
	}); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	if !me.Hand.Contains(tapped) {
		t.Error("the tapped creature did not return to hand")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want {2}{G} spent", me.ManaPool.Total())
	}
}

// --- Spiders-Man, Heroic Horde ---

func TestSpmB9SpidersManWebSlungGainsLifeAndMakesTwoReachSpiders(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	card := handCardFull(me, "Spiders-Man, Heroic Horde", "Legendary Creature — Spider Hero", "{1}{G}", spmB9SpidersManOracle, []string{"G"})
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{C}{C}{G}{G}")
	life := me.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{tapped},
	}); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card) {
		t.Fatal("Spiders-Man did not resolve")
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
	spiders := spmB9Named(g, me.ID, "Spider")
	if len(spiders) != 2 {
		t.Fatalf("%d Spider tokens, want 2", len(spiders))
	}
	for _, s := range spiders {
		if s.Power != 2 || s.Toughness != 1 || !game.HasKeyword(&s, "reach") || !s.IsToken() {
			t.Errorf("token %+v, want a 2/1 reach Spider token", s)
		}
	}
}

func TestSpmB9SpidersManHardCastDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	spmB9Cast(t, g, "Spiders-Man, Heroic Horde", "Legendary Creature — Spider Hero", spmB9SpidersManOracle, "{1}{G}", []string{"G"}, 2, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a hard cast gained life: %d -> %d", life, me.Life)
	}
	if n := len(spmB9Named(g, me.ID, "Spider")); n != 0 {
		t.Errorf("a hard cast made %d Spider tokens", n)
	}
}

func TestSpmB9SpidersManWebSlingingRefusesAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	card := handCardFull(me, "Spiders-Man, Heroic Horde", "Legendary Creature — Spider Hero", "{1}{G}", spmB9SpidersManOracle, []string{"G"})
	untapped := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{C}{C}{G}{G}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{untapped},
	}); err == nil {
		t.Fatal("web-slinging accepted an untapped creature")
	}
}

// --- Spider-Slayer, Hatred Honed ---

func TestSpmB9SlayerDestroysASpiderItDamages(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slayer := b12Push(g, me.ID, "Spider-Slayer, Hatred Honed", "Legendary Artifact Creature — Human Villain", spmB9SlayerOracle, 2, 1)
	spider := b12Creature(g, opp.ID, "Big Spider", "Creature — Spider Hero", 1, 9)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(slayer, spider, 1) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spider) {
		t.Error("the damaged Spider is still on the battlefield")
	}
}

func TestSpmB9SlayerLeavesNonSpidersAndOtherSourcesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slayer := b12Push(g, me.ID, "Spider-Slayer, Hatred Honed", "Legendary Artifact Creature — Human Villain", spmB9SlayerOracle, 2, 1)
	other := b12Creature(g, me.ID, "Pinger", "Creature — Bear", 2, 2)
	bear := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 1, 9)
	spider := b12Creature(g, opp.ID, "Big Spider", "Creature — Spider", 1, 9)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(slayer, bear, 1) })
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(other, spider, 1) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a non-Spider was destroyed")
	}
	if !g.Battlefield.Contains(spider) {
		t.Error("a Spider damaged by another source was destroyed")
	}
}

func TestSpmB9SlayerDoesNotMisfireOnASpiderThatDiesToTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slayer := b12Push(g, me.ID, "Spider-Slayer, Hatred Honed", "Legendary Artifact Creature — Human Villain", spmB9SlayerOracle, 2, 1)
	// A Spider that dies to the damage is not destroyed a second time; the
	// trigger is a no-op rather than an error.
	spider := b12Creature(g, opp.ID, "Small Spider", "Creature — Spider", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(slayer, spider, 2) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spider) {
		t.Error("the lethally damaged Spider survived")
	}
}

func TestSpmB9SlayerGraveyardAbilityMakesTwoTappedFlyingRobots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Spider-Slayer, Hatred Honed", TypeLine: "Legendary Artifact Creature — Human Villain",
		OracleID: spmB9SlayerOracle, Owner: me.ID, Controller: me.ID,
	})
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {6} with five mana")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("a refused activation moved the card")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("the card was not exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	robots := spmB9Named(g, me.ID, "Robot")
	if len(robots) != 2 {
		t.Fatalf("%d Robot tokens, want 2", len(robots))
	}
	for _, r := range robots {
		if !r.Tapped || r.Power != 1 || r.Toughness != 1 || !r.IsArtifact() || !game.HasKeyword(&r, "flying") {
			t.Errorf("token %+v, want a tapped 1/1 flying artifact Robot", r)
		}
	}
}

func TestSpmB9SlayerAbilityDoesNotWorkFromTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	slayer := b12Push(g, me.ID, "Spider-Slayer, Hatred Honed", "Legendary Artifact Creature — Human Villain", spmB9SlayerOracle, 2, 1)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, slayer, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the graveyard ability was activated from the battlefield")
	}
}

// --- Spider-Suit ---

func TestSpmB9SuitGivesPlusTwoAndTheSpiderHeroTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	bear := seedBear(g, me.ID)
	suit := seedEquipment(g, me.ID, "Spider-Suit", spmB9SuitOracle)
	floatMana(t, g, me, "{C}{C}{C}")
	equipTo(t, g, me.ID, suit, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped creature is %d/%d, want 4/4", p, tough)
	}
	subs := effectiveSubtypes(t, g, bear)
	for _, want := range []string{"Spider", "Hero"} {
		if !slices.Contains(subs, want) {
			t.Errorf("subtypes %v lack %s", subs, want)
		}
	}
	if len(subs) < 3 {
		t.Errorf("subtypes %v lost the creature's own types", subs)
	}
}

func TestSpmB9SuitEquipCostsThreeAndTheBonusLeavesWithIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	bear := seedBear(g, me.ID)
	other := seedBear(g, me.ID)
	suit := seedEquipment(g, me.ID, "Spider-Suit", spmB9SuitOracle)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, suit, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("equipped for two mana")
	}
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, suit, bear)
	floatMana(t, g, me, "{C}{C}{C}")
	equipTo(t, g, me.ID, suit, other)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the old host is %d/%d, want 2/2", p, tough)
	}
	if slices.Contains(effectiveSubtypes(t, g, bear), "Spider") {
		t.Error("the old host is still a Spider")
	}
	if p, tough := effectivePTOf(t, g, other); p != 4 || tough != 4 {
		t.Errorf("the new host is %d/%d, want 4/4", p, tough)
	}
}

// --- Steel Wrecking Ball ---

func TestSpmB9WreckingBallDealsFiveToATargetCreatureWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	spmB9Cast(t, g, "Steel Wrecking Ball", "Artifact", spmB9WreckingOracle, "{5}", nil, 0, 0)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("a 0/5 survived five damage")
	}
}

func TestSpmB9WreckingBallDiscardDestroysAnArtifactFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ball := handCardFull(me, "Steel Wrecking Ball", "Artifact", "{5}", spmB9WreckingOracle, nil)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Plain Land", "Land")

	floatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Fatal("activated {1}{R} with one mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err == nil {
		t.Fatal("targeted a land with Destroy target artifact")
	}
	if !me.Hand.Contains(ball) {
		t.Fatal("a refused activation discarded the card")
	}
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(ball) {
		t.Error("the card was not discarded as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

// --- Strength of Will ---

func TestSpmB9StrengthOfWillMakesItIndestructibleAndGrowsByTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pinger := b12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)
	castCatalogSpell(t, g, "Strength of Will", "Instant", spmB9WillOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, bear, "indestructible")

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(pinger, bear, 3) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the indestructible Bear died to lethal damage")
	}
	if got := plusOnes(g, bear); got != 3 {
		t.Errorf("%d +1/+1 counters after 3 damage, want 3", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(pinger, bear, 1) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 4 {
		t.Errorf("%d counters after another 1 damage, want 4", got)
	}
}

func TestSpmB9StrengthOfWillAddsNothingToAnotherCreatureOrAfterTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	friend := b12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 9)
	pinger := b12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)
	castCatalogSpell(t, g, "Strength of Will", "Instant", spmB9WillOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(pinger, friend, 2) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, friend); got != 0 {
		t.Errorf("another creature got %d counters", got)
	}
}

func TestSpmB9StrengthOfWillRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	toMainForCost(t, g)
	card := handCardFull(me, "Strength of Will", "Instant", "{1}{G}", spmB9WillOracle, []string{"G"})
	floatMana(t, g, me, "{C}{G}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("targeted a creature the caster does not control")
	}
}

// --- Suburban Sanctuary ---

func TestSpmB9SanctuaryEntersTappedAndMakesGreenOrWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := playLandFromHand(t, g, "Suburban Sanctuary", spmB9SanctuaryOracle)
	if !b16Tapped(t, g, land) {
		t.Fatal("Suburban Sanctuary enters untapped")
	}
	// Untap it by hand and tap it for each colour.
	for _, color := range []string{"G", "W"} {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == land {
					g.Battlefield.Cards[i].Tapped = false
				}
			}
			me.ManaPool = game.ManaPool{}
		})
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{color}}); err != nil {
			t.Fatalf("tap for %s: %v", color, err)
		}
		units := me.ManaPool.Units()
		if len(units) != 1 || units[0].Color != color {
			t.Errorf("tapped for %s, pool %v", color, me.ManaPool)
		}
	}
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Error("tapped for red")
	}
}

func TestSpmB9SanctuarySurveilsForFourAndTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := seedPermanentWithOracle(g, me.ID, "Suburban Sanctuary", "Land", spmB9SanctuaryOracle)
	top := seedLibraryTop(me, "Top Card", "Sorcery", "{1}")

	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {4},{T} with three mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !b16Tapped(t, g, land) {
		t.Error("the land did not tap to pay")
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatal(err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not go to the graveyard")
	}
}

// --- Subway Train ---

func TestSpmB9TrainPaysGenericGreenToFetchABasicLandIntoHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Wasteland", TypeLine: "Land"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	spmB9Cast(t, g, "Subway Train", "Artifact — Vehicle", spmB9TrainOracle, "{2}", nil, 3, 1)
	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{G}")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the Forest was not put into hand")
	}
	if !me.Library.Contains(ids[0]) {
		t.Error("the Wasteland should stay in the library")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want the {G} spent", me.ManaPool.Total())
	}
}

func TestSpmB9TrainDecliningTheGreenSearchesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	spmB9Cast(t, g, "Subway Train", "Artifact — Vehicle", spmB9TrainOracle, "{2}", nil, 3, 1)
	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{G}")
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("a search was offered after declining")
	}
	if me.Hand.Contains(ids[0]) {
		t.Error("the Forest reached the hand without paying")
	}
}

func TestSpmB9TrainCrewTwoMakesItAThreeOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	train := pushVehicleForTest(g, owner.ID, "Subway Train", spmB9TrainOracle, 3, 1)
	small := pushCrewerForTest(g, owner.ID, "Small", 1)
	if err := g.ActivateCatalogAbility(owner.ID, train, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("crewed Crew 2 with total power 1")
	}
	big := pushCrewerForTest(g, owner.ID, "Big", 2)
	if err := g.ActivateCatalogAbility(owner.ID, train, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{big}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	v, _ := battlefieldCardByID(g, train)
	if !v.IsCreature() || v.CurrentPower() != 3 || v.CurrentToughness() != 1 {
		t.Errorf("crewed Train = %+v, want a 3/1 creature", v)
	}
}

// --- Sun-Spider, Nimble Webber ---

func TestSpmB9SunSpiderFliesOnlyDuringItsControllersTurn(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Push(g, active.ID, "Sun-Spider, Nimble Webber", "Legendary Creature — Spider Human Hero", spmB9SunSpiderOracle, 3, 2)
	theirs := b12Push(g, other.ID, "Sun-Spider, Nimble Webber", "Legendary Creature — Spider Human Hero", spmB9SunSpiderOracle, 3, 2)
	assertKeywords(t, g, mine, "flying")
	if effectiveAbilitiesContain(t, g, theirs, "flying") {
		t.Error("Sun-Spider flies during another player's turn")
	}
}

func TestSpmB9SunSpiderFetchesAnAuraOrEquipmentIntoHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bonesplitter", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Pacifism", TypeLine: "Enchantment — Aura"},
	)
	spmB9Cast(t, g, "Sun-Spider, Nimble Webber", "Legendary Creature — Spider Human Hero", spmB9SunSpiderOracle, "{3}{W/U}", []string{"W", "U"}, 3, 2)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.SearchCards {
		offered[id] = true
	}
	if offered[ids[0]] || !offered[ids[1]] || !offered[ids[2]] {
		t.Errorf("offered %v, want only the Equipment and the Aura", offered)
	}
	answerSearchByID(t, g, me.ID, ids[2])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[2]) {
		t.Error("the Aura was not put into hand")
	}
}

// --- Supportive Parents ---

func TestSpmB9ParentsTapTwoUntappedCreaturesForOneManaOfAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	parents := seedPermanentWithOracle(g, me.ID, "Supportive Parents", "Creature — Human Citizen", spmB9ParentsOracle)
	friend := b12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, parents, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{friend}}); err == nil {
		t.Fatal("paid with only one creature")
	}
	if err := g.ActivateManaAbility(me.ID, parents, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{friend, foe}}); err == nil {
		t.Fatal("tapped an opponent's creature")
	}
	assertTapped(t, g, false, parents, friend, foe)
	if err := g.ActivateManaAbility(me.ID, parents, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{parents, friend}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	assertTapped(t, g, true, parents, friend)
	if manaPickFor(g, me.ID) == nil {
		t.Fatal("the any-color ability did not ask for a color")
	}
}

func TestSpmB9ParentsTapTwoOtherCreaturesAndRefuseAnAlreadyTappedOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	parents := seedPermanentWithOracle(g, me.ID, "Supportive Parents", "Creature — Human Citizen", spmB9ParentsOracle)
	one := b12Creature(g, me.ID, "One", "Creature — Bear", 2, 2)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	spent := pushTappedVanillaCreature(g, me.ID, "Spent", 2, 2)

	if err := g.ActivateManaAbility(me.ID, parents, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{one, spent}}); err == nil {
		t.Fatal("tapped an already tapped creature")
	}
	if err := g.ActivateManaAbility(me.ID, parents, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{one, two}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	assertTapped(t, g, true, one, two)
	assertTapped(t, g, false, parents)
}
