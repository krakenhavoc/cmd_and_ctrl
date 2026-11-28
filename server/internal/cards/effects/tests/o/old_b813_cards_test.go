package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// old_b813_cards_test.go — S82 OLD batch 813, Mirage (#5633).

const (
	b813JunglePatrol   = "5bc8e918-7824-4ad7-a5c0-05537ce81641"
	b813SeaScryer      = "f17954dd-c8e0-4983-b81e-5440cca41afc"
	b813ManaPrism      = "f4669822-716b-45c7-a7f2-040062509fe9"
	b813Dreadnought    = "a7950edb-5670-46fd-a30e-0fe888a5a32b"
	b813Apparition     = "ced9130c-2865-49df-abdd-84069aa1b180"
	b813Shimmer        = "f557fb2c-581b-4a8d-8766-1e46e92224fa"
	b813Preferred      = "1d05d2d6-3ea1-4d8a-9f1c-6fc3f5e38ea9"
	b813CelestialDawn  = "b4d27fad-9a94-4760-80a6-6a2948109864"
	b813LeadGolem      = "5810c1e2-5a36-4a87-be4e-890ba1627b46"
	b813CrystalGolem   = "82246ef2-4267-4daf-9463-908c2e142e7d"
	b813WaitingWeeds   = "5a61d51e-1a3f-49f6-9a94-a708dd6f69df"
	b813BasaltGolem    = "9155f78e-0ccc-4e08-ae79-ef5b2db1183f"
	b813GolemTypeLine  = "Artifact Creature — Golem"
	b813ForestTypeLine = "Basic Land — Forest"
)

func TestB813CardsAreRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		b813JunglePatrol: "Jungle Patrol", b813SeaScryer: "Sea Scryer", b813ManaPrism: "Mana Prism",
		b813Dreadnought: "Phyrexian Dreadnought", b813Apparition: "Haunting Apparition", b813Shimmer: "Shimmer",
		b813Preferred: "Preferred Selection", b813CelestialDawn: "Celestial Dawn", b813LeadGolem: "Lead Golem",
		b813CrystalGolem: "Crystal Golem", b813WaitingWeeds: "Waiting in the Weeds", b813BasaltGolem: "Basalt Golem",
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func b813Pool(p *game.Player, color string) int {
	n := 0
	for _, m := range p.ManaPool {
		if m.Color == color {
			n += max(m.Count, 1)
		}
	}
	return n
}

func TestB813SeaScryerTapsForColorlessOrPaysOneForBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Sea Scryer", "Creature — Merfolk Wizard", b813SeaScryer, 1, 1)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if b813Pool(me, "C") != 1 {
		t.Fatalf("pool %v, want one {C}", me.ManaPool)
	}
	// A tapped Scryer cannot use the second ability, and an empty pool cannot pay its {1}.
	if err := g.ActivateManaAbility(me.ID, id, 1, game.ManaAbilityParams{}); err == nil {
		t.Error("a tapped Scryer paid {1},{T}")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = false })
	if err := g.ActivateManaAbility(me.ID, id, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{1}, {T}: Add {U}: %v", err)
	}
	if b813Pool(me, "U") != 1 || b813Pool(me, "C") != 0 {
		t.Errorf("pool %v, want one {U} and the {C} spent", me.ManaPool)
	}
}

func TestB813SeaScryerSecondAbilityNeedsTheGenericMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Sea Scryer", "Creature — Merfolk Wizard", b813SeaScryer, 1, 1)
	if err := g.ActivateManaAbility(me.ID, id, 1, game.ManaAbilityParams{}); err == nil {
		t.Error("{1},{T}: Add {U} worked with an empty pool")
	}
}

func TestB813ManaPrismMakesAnyColorForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Mana Prism", "Artifact", b813ManaPrism, false)
	effectstest.B28TapForMana(t, g, me.ID, id, "")
	if b813Pool(me, "C") != 1 {
		t.Fatalf("pool %v, want one {C}", me.ManaPool)
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = false })
	effectstest.B40TapForMana(t, g, me.ID, id, 1, "R")
	if b813Pool(me, "R") != 1 {
		t.Errorf("pool %v, want one {R}", me.ManaPool)
	}
}

func TestB813CrystalGolemPhasesOutAtYourEndStepAndBackInBeforeYouUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Crystal Golem", b813GolemTypeLine, b813CrystalGolem, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	var out bool
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(id) })
	if !out || g.Battlefield.Contains(id) {
		t.Fatalf("Crystal Golem phased out = %v, on battlefield = %v", out, g.Battlefield.Contains(id))
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(id) })
	if out || !g.Battlefield.Contains(id) {
		t.Errorf("Crystal Golem is still phased out in its controller's next turn")
	}
}

func TestB813LeadGolemDoesNotUntapAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Lead Golem", b813GolemTypeLine, b813LeadGolem, 3, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, id)
	if !effectstest.B16Tapped(t, g, id) {
		t.Fatal("the Golem did not tap to attack")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if !effectstest.B16Tapped(t, g, id) {
		t.Error("the Golem untapped in its controller's next untap step")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if effectstest.B16Tapped(t, g, id) {
		t.Error("the Golem stayed tapped for a second untap step")
	}
}

func TestB813WaitingInTheWeedsMakesACatPerUntappedForestForEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", b813ForestTypeLine)
	}
	tapped := effectstest.B12Permanent(g, me.ID, "Forest", b813ForestTypeLine)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
	effectstest.B12Permanent(g, opp.ID, "Forest", b813ForestTypeLine)
	effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	effectstest.CastCatalogSpell(t, g, "Waiting in the Weeds", "Sorcery", b813WaitingWeeds, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B28TokensNamed(g, me.ID, "Cat"); n != 3 {
		t.Errorf("caster has %d Cats, want 3 (one per untapped Forest)", n)
	}
	if n := effectstest.B28TokensNamed(g, opp.ID, "Cat"); n != 1 {
		t.Errorf("opponent has %d Cats, want 1", n)
	}
	if n := effectstest.B28TokensNamed(g, g.Seats[2].ID, "Cat"); n != 0 {
		t.Errorf("a player with no Forest has %d Cats", n)
	}
}

func b813Wood(g *game.Game, controller uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Wood" {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestB813JunglePatrolMakesWoodAndSacrificesItForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	patrol := effectstest.B12Push(g, me.ID, "Jungle Patrol", "Creature — Human Soldier", b813JunglePatrol, 3, 2)
	effectstest.ApaMana(me, "G", "G")
	effectstest.P7Activate(t, g, me, patrol, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B28TokensNamed(g, me.ID, "Wood"); n != 1 {
		t.Fatalf("%d Wood tokens, want 1", n)
	}
	wood := b813Wood(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, patrol, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{wood}}); err != nil {
		t.Fatalf("Sacrifice a token named Wood: %v", err)
	}
	if n := effectstest.B28TokensNamed(g, me.ID, "Wood"); n != 0 {
		t.Errorf("the Wood was not sacrificed (%d left)", n)
	}
	if b813Pool(me, "R") != 1 {
		t.Errorf("pool %v, want one {R}", me.ManaPool)
	}
	if err := g.ActivateManaAbility(me.ID, patrol, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{wood}}); err == nil {
		t.Error("sacrificed a Wood that was not there")
	}
}

func TestB813JunglePatrolCannotSacrificeAnotherToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	patrol := effectstest.B12Push(g, me.ID, "Jungle Patrol", "Creature — Human Soldier", b813JunglePatrol, 3, 2)
	cat := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 green Cat"))
	if err := g.ActivateManaAbility(me.ID, patrol, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{cat}}); err == nil {
		t.Error("a Cat paid a cost that wants a token named Wood")
	}
}

func TestB813BasaltGolemCannotBeBlockedByArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.B12Push(g, me.ID, "Basalt Golem", b813GolemTypeLine, b813BasaltGolem, 2, 4)
	artifact := effectstest.B12Creature(g, opp.ID, "Their Golem", b813GolemTypeLine, 1, 1)
	plain := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	effectstest.Pr7bAttack(t, g, opp.ID, golem)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(artifact, golem); err == nil {
		t.Error("an artifact creature blocked the Basalt Golem")
	}
	if err := g.DeclareBlocker(plain, golem); err != nil {
		t.Errorf("a nonartifact creature could not block it: %v", err)
	}
}

func TestB813BasaltGolemMakesTheBlockersControllerSacrificeItAndGetAWall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.B12Push(g, me.ID, "Basalt Golem", b813GolemTypeLine, b813BasaltGolem, 2, 4)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	effectstest.Pr7bAttack(t, g, opp.ID, golem)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: golem})
	if !g.Battlefield.Contains(wall) {
		t.Fatal("the blocker was sacrificed before the end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the blocker was not sacrificed at end of combat")
	}
	if n := effectstest.B28TokensNamed(g, opp.ID, "Wall"); n != 1 {
		t.Errorf("the blocker's controller has %d Wall tokens, want 1", n)
	}
	if n := effectstest.B28TokensNamed(g, me.ID, "Wall"); n != 0 {
		t.Errorf("the Golem's controller got %d Wall tokens", n)
	}
}

func TestB813DreadnoughtCanOnlyBePaidForByItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := b813Cast(t, g, me, "Phyrexian Dreadnought", b813GolemTypeLine, b813Dreadnought, 12, 12)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
	if c == nil {
		t.Fatalf("no pay prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, me.ID, true, []uuid.UUID{bear}); err == nil {
		t.Fatal("two power paid a cost of twelve")
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, me.ID, true, []uuid.UUID{id, bear}); err != nil {
		t.Fatalf("sacrificing the Dreadnought itself with the Bear: %v", err)
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(id) || g.Battlefield.Contains(bear) {
		t.Error("the sacrificed creatures stayed")
	}
}

func TestB813DreadnoughtStaysWhenYouSacrificeTwelvePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	seven := effectstest.B12Creature(g, me.ID, "Seven", "Creature — Giant", 7, 7)
	five := effectstest.B12Creature(g, me.ID, "Five", "Creature — Giant", 5, 5)
	spare := effectstest.B12Creature(g, me.ID, "Spare", "Creature — Bear", 2, 2)
	id := b813Cast(t, g, me, "Phyrexian Dreadnought", b813GolemTypeLine, b813Dreadnought, 12, 12)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
	if c == nil {
		t.Fatalf("no pay prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, me.ID, true, []uuid.UUID{seven, spare}); err == nil {
		t.Fatal("nine power paid a cost of twelve")
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, me.ID, true, []uuid.UUID{seven, five}); err != nil {
		t.Fatalf("pay with twelve power: %v", err)
	}
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(id) || g.Battlefield.Contains(seven) || g.Battlefield.Contains(five) || !g.Battlefield.Contains(spare) {
		t.Errorf("after paying: Dreadnought %v, seven %v, five %v, spare %v",
			g.Battlefield.Contains(id), g.Battlefield.Contains(seven), g.Battlefield.Contains(five), g.Battlefield.Contains(spare))
	}
}

func TestB813DreadnoughtDeclinedIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Twelve", "Creature — Giant", 12, 12)
	id := b813Cast(t, g, me, "Phyrexian Dreadnought", b813GolemTypeLine, b813Dreadnought, 12, 12)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
	if c == nil {
		t.Fatalf("no pay prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(id) {
		t.Error("a declined Dreadnought stayed")
	}
}

func b813Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestB813HauntingApparitionCountsGreenCreatureCardsInTheChosenGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	for _, c := range []game.Card{
		{Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"}},
		{Name: "Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}},
		{Name: "Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"}},
		{Name: "Growth", TypeLine: "Sorcery", Colors: []string{"G"}},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), opp.ID, opp.ID
		g.WithWriteLock(func() { opp.Graveyard.PushTop(c) })
	}
	other.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"}, Owner: other.ID, Controller: other.ID})
	id := b813Cast(t, g, me, "Haunting Apparition", "Creature — Spirit", b813Apparition, 1, 2)
	effectstest.AnswerChosenPlayer(t, g, me.ID, opp.ID)
	if p, tt := effectstest.EffectivePTOf(t, g, id); p != 3 || tt != 2 {
		t.Fatalf("Apparition is %d/%d, want 3/2 (1 + two green creature cards)", p, tt)
	}
	wurm := effectstest.B12Push(g, opp.ID, "Wurm", "Creature — Wurm", "", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, wurm).Colors = []string{"G"} })
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{wurm}, game.DestroyOptions{}) })
	g.RunStateChecksForTest()
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != 4 {
		t.Errorf("after a third green creature card, power = %d, want 4", p)
	}
}

func b813Phasing(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c := effectstest.LayeredCard(t, g, id)
	return game.HasKeyword(&c, game.KeywordPhasing)
}

func TestB813ShimmerGivesPhasingToLandsOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myForest := effectstest.B12Permanent(g, me.ID, "Forest", b813ForestTypeLine)
	theirForest := effectstest.B12Permanent(g, opp.ID, "Forest", b813ForestTypeLine)
	island := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b813Cast(t, g, me, "Shimmer", "Enchantment", b813Shimmer, 0, 0)
	effectstest.AnswerAnchorWord(t, g, me.ID, "Forest")
	if !b813Phasing(t, g, myForest) || !b813Phasing(t, g, theirForest) {
		t.Error("a Forest does not have phasing")
	}
	if b813Phasing(t, g, island) {
		t.Error("an Island has phasing")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	var out bool
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(theirForest) })
	if !out {
		t.Error("their Forest did not phase out in their untap step")
	}
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(myForest) })
	if out {
		t.Error("my Forest phased out in their untap step")
	}
}

func TestB813CelestialDawnMakesLandsPlainsAndPermanentsWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	theirs := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	theirBear := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	dawn := effectstest.PushCatalogPermanent(g, me.ID, "Celestial Dawn", "Enchantment", b813CelestialDawn, false)
	if c := effectstest.LayeredCard(t, g, mine); !c.HasSubtype("Plains") || c.HasSubtype("Island") {
		t.Errorf("my land is %q, want a Plains", c.TypeLine)
	}
	if c := effectstest.LayeredCard(t, g, theirs); !c.HasSubtype("Island") {
		t.Errorf("their land changed: %q", c.TypeLine)
	}
	if cs := effectstest.EffectiveColors(t, g, bear); len(cs) != 1 || cs[0] != "W" {
		t.Errorf("my Bear is %v, want white", cs)
	}
	if cs := effectstest.EffectiveColors(t, g, dawn); len(cs) != 1 || cs[0] != "W" {
		t.Errorf("Celestial Dawn is %v, want white", cs)
	}
	if cs := effectstest.EffectiveColors(t, g, theirBear); len(cs) != 1 || cs[0] != "G" {
		t.Errorf("their Bear is %v, want green", cs)
	}
	if cs := effectstest.EffectiveColors(t, g, mine); len(cs) != 0 {
		t.Errorf("my land is %v, want colorless", cs)
	}
}

func TestB813CelestialDawnSpendsOnlyWhiteAsAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Celestial Dawn", "Enchantment", b813CelestialDawn, false)
	spec, ok := effects.Lookup(b813CelestialDawn)
	if !ok || len(spec.AnyColorSpend) != 1 || spec.AnyColorSpend[0].From != "W" || !spec.AnyColorSpend[0].OthersOnlyColorless {
		t.Fatalf("spend rule %+v", spec.AnyColorSpend)
	}
}

// b813Library puts n named cards on top of me's library, the last one on top.
func b813Library(me *game.Player, names ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range names {
		ids = append(ids, effectstest.B02TopOfLibrary(me, n, "Sorcery"))
	}
	return ids
}

// b813AtMyUpkeepWithSelection puts Preferred Selection out, walks to my upkeep,
// stacks the library as the trigger waits, and resolves it.
func b813AtMyUpkeepWithSelection(t *testing.T, g *game.Game, me *game.Player, pool []string, names ...string) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	id := effectstest.PushCatalogPermanent(g, me.ID, "Preferred Selection", "Enchantment", b813Preferred, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.AwbPadLibrary(me)
	effectstest.B10AddMana(me, pool...)
	ids := b813Library(me, names...)
	effectstest.PassPriorityAroundTable(t, g)
	return id, ids
}

func TestB813PreferredSelectionDeclinedPutsOneOfTheTopTwoOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sel, ids := b813AtMyUpkeepWithSelection(t, g, me, nil, "Lower", "Upper")
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
	if c == nil {
		t.Fatalf("no pay prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sel) {
		t.Error("the enchantment was sacrificed although the payment was declined")
	}
	if b, _ := me.Library.Bottom(); b.InstanceID != ids[1] {
		t.Errorf("bottom of the library is %q, want the chosen card", b.Name)
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[0] {
		t.Errorf("top of the library is %q, want the other of the two", top.Name)
	}
}

func TestB813PreferredSelectionPaidSacrificesItAndTakesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sel, ids := b813AtMyUpkeepWithSelection(t, g, me, []string{"G", "G", "G", "G"}, "Lower", "Upper")
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
	if c == nil {
		t.Fatalf("no pay prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sel) {
		t.Error("the enchantment was not sacrificed")
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the chosen card did not reach the hand")
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[1] {
		t.Errorf("top of the library is %q, want the card that was not taken", top.Name)
	}
}
