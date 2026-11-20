package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b15_cards_test.go — Adventures in the Forgotten Realms, S82 batch
// 15 (#5499). Each test plays the real catalog entries through the engine.

const (
	afrB15Mordenkainen = "6be0bf15-fdcb-4425-9bd8-3c39039cab7d"
	afrB15Lolth        = "7a8edf45-206e-4d0e-bd69-06f1ca70d17c"
	afrB15Drider       = "1db76752-31ed-4a16-986a-d0852ba7d812"
	afrB15Zariel       = "04ddb495-e9b0-47bf-b892-c7ae7b87c10c"
	afrB15GrandMaster  = "1f3dbc53-4075-40b8-b70d-ce74d3ff5f6d"
	afrB15Minsc        = "2cd10c6c-f946-457e-9175-92501322f7a5"
	afrB15Drizzt       = "1e939df0-d421-42b7-b642-88b3555fd8df"
	afrB15Minn         = "d4212d44-d6a0-4c5b-9c54-bc90f36828dd"
)

func TestAfrB15EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrB15Mordenkainen: "Mordenkainen", afrB15Lolth: "Lolth, Spider Queen",
		afrB15Drider: "Drider", afrB15Zariel: "Zariel, Archduke of Avernus",
		afrB15GrandMaster: "Grand Master of Flowers", afrB15Minsc: "Minsc, Beloved Ranger",
		afrB15Drizzt: "Drizzt Do'Urden", afrB15Minn: "Minn, Wily Illusionist",
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

// afrB15Named lists the battlefield permanents with a name under a controller.
func afrB15Named(g *game.Game, owner uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == name {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

// afrB15EnterFromHand casts a creature from hand and resolves it and its
// enters trigger.
func afrB15EnterFromHand(t *testing.T, g *game.Game, name, typeLine, oracle string) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Drider ----------------------------------------------------------------

func TestAfrB15DriderMakesASpiderWhenItDealsCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drider := effectstest.B12Push(g, me.ID, "Drider", "Creature — Elf Spider", afrB15Drider, 4, 3)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, drider)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("opponent life %d → %d, want 4 combat damage", life, opp.Life)
	}
	spiders := afrB15Named(g, me.ID, "Spider")
	if len(spiders) != 1 {
		t.Fatalf("%d Spiders, want 1", len(spiders))
	}
	c := effectstest.B12Card(t, g, spiders[0])
	if c.Power != 2 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" || !effectstest.HasKeywordOnBattlefield(t, g, spiders[0], "reach") || !effectstest.HasKeywordOnBattlefield(t, g, spiders[0], "menace") {
		t.Errorf("Spider is %+v, want a 2/1 black Spider with reach and menace", c)
	}
}

func TestAfrB15DriderMakesNoSpiderWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drider := effectstest.B12Push(g, me.ID, "Drider", "Creature — Elf Spider", afrB15Drider, 4, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, drider)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, drider); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(afrB15Named(g, me.ID, "Spider")); n != 0 {
		t.Errorf("%d Spiders after the Drider was blocked, want none", n)
	}
}

// --- Minsc, Beloved Ranger -------------------------------------------------

func TestAfrB15MinscMakesBooWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	afrB15EnterFromHand(t, g, "Minsc, Beloved Ranger", "Legendary Creature — Human Ranger", afrB15Minsc)
	boo := afrB15Named(g, me.ID, "Boo")
	if len(boo) != 1 {
		t.Fatalf("%d Boo, want 1", len(boo))
	}
	c := effectstest.B12Card(t, g, boo[0])
	if c.Power != 1 || c.Toughness != 1 || c.TypeLine != "Token Legendary Creature — Hamster" ||
		len(c.Colors) != 1 || c.Colors[0] != "R" || !effectstest.HasKeywordOnBattlefield(t, g, boo[0], "trample") || !effectstest.HasKeywordOnBattlefield(t, g, boo[0], "haste") {
		t.Errorf("Boo is %+v, want a legendary 1/1 red Hamster with trample and haste", c)
	}
}

func TestAfrB15MinscXSetsBasePTAndAddsGiantUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	minsc := effectstest.B12Push(g, me.ID, "Minsc, Beloved Ranger", "Legendary Creature — Human Ranger", afrB15Minsc, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 1) })
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, minsc, 0, game.ActivateAbilityParams{XValue: 5, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Base 5/5 with the +1/+1 counter on top (CR 613.4b, 613.4d).
	if p, tg := effectstest.EffectivePowerWithCounters(t, g, bear), afrB15ToughnessWithCounters(t, g, bear); p != 6 || tg != 6 {
		t.Errorf("Bear is %d/%d, want 6/6 (base 5/5 plus its counter)", p, tg)
	}
	c, _ := effectstest.BattlefieldCard(g, bear)
	if !c.HasSubtype("Giant") || !c.HasSubtype("Bear") {
		t.Errorf("Bear's subtypes %v, want Bear and Giant", c.Effective().Subtypes)
	}
	// It lasts until end of turn only.
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, bear); p != 3 {
		t.Errorf("Bear's power %d after the turn ended, want 3", p)
	}
}

func TestAfrB15MinscXIsSorceryOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	minsc := effectstest.B12Push(g, me.ID, "Minsc, Beloved Ranger", "Legendary Creature — Human Ranger", afrB15Minsc, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	err := g.ActivateCatalogAbility(me.ID, minsc, 0, game.ActivateAbilityParams{XValue: 5, Targets: effectstest.CardTarget(bear), Strict: true})
	if err == nil {
		t.Error("Minsc's {X} ability was activated during combat")
	}
}

func TestAfrB15MinscXCanTargetOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	minsc := effectstest.B12Push(g, me.ID, "Minsc, Beloved Ranger", "Legendary Creature — Human Ranger", afrB15Minsc, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, minsc, 0, game.ActivateAbilityParams{XValue: 5, Targets: effectstest.CardTarget(theirs), Strict: true})
	if err == nil {
		t.Error("Minsc's {X} ability targeted an opponent's creature")
	}
}

// --- Drizzt Do'Urden ---------------------------------------------------------

func TestAfrB15DrizztMakesGuenhwyvarWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	afrB15EnterFromHand(t, g, "Drizzt Do'Urden", "Legendary Creature — Elf Ranger", afrB15Drizzt)
	cat := afrB15Named(g, me.ID, "Guenhwyvar")
	if len(cat) != 1 {
		t.Fatalf("%d Guenhwyvar, want 1", len(cat))
	}
	c := effectstest.B12Card(t, g, cat[0])
	if c.Power != 4 || c.Toughness != 1 || c.TypeLine != "Token Legendary Creature — Cat" ||
		len(c.Colors) != 1 || c.Colors[0] != "G" || !effectstest.HasKeywordOnBattlefield(t, g, cat[0], "trample") {
		t.Errorf("Guenhwyvar is %+v, want a legendary 4/1 green Cat with trample", c)
	}
}

func TestAfrB15DrizztGrowsByTheDifferenceWhenAStrongerCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drizzt := effectstest.B12Push(g, me.ID, "Drizzt Do'Urden", "Legendary Creature — Elf Ranger", afrB15Drizzt, 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, opp.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, drizzt)
	if c.Counters["+1/+1"] != 4 {
		t.Errorf("Drizzt has %d +1/+1 counters, want 4 (7 − 3)", c.Counters["+1/+1"])
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, drizzt); p != 7 {
		t.Errorf("Drizzt's power %d, want 7", p)
	}
}

func TestAfrB15DrizztIgnoresCreaturesWithNoMorePowerAndCountsYoursToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drizzt := effectstest.B12Push(g, me.ID, "Drizzt Do'Urden", "Legendary Creature — Elf Ranger", afrB15Drizzt, 3, 3)
	equal := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 5, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, opp.ID, equal)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, drizzt); c.Counters["+1/+1"] != 0 {
		t.Errorf("an equal-power creature gave Drizzt %d counters", c.Counters["+1/+1"])
	}
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, drizzt); c.Counters["+1/+1"] != 2 {
		t.Errorf("Drizzt has %d counters after his own controller's 5-power creature died, want 2", c.Counters["+1/+1"])
	}
}

func TestAfrB15DrizztDoesNotCountItselfDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drizzt := effectstest.B12Push(g, me.ID, "Drizzt Do'Urden", "Legendary Creature — Elf Ranger", afrB15Drizzt, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, drizzt)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(g.StackMeta); n != 0 {
		t.Errorf("%d items on the stack after Drizzt died", n)
	}
}

// --- Minn, Wily Illusionist ----------------------------------------------------

func TestAfrB15MinnMakesAnIllusionOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Minn, Wily Illusionist", "Legendary Creature — Gnome Wizard", afrB15Minn, 1, 3)
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryTopForTest(me, "Island")
	}
	effectstest.AdvanceToMain(t, g)
	// The turn's draw step was the first card drawn this turn.
	if got := g.TurnTallyFor(me.ID).CardsDrawn; got != 1 {
		t.Fatalf("%d cards drawn so far this turn, want 1", got)
	}
	if n := len(afrB15Named(g, me.ID, "Illusion")); n != 0 {
		t.Fatalf("%d Illusions after the first draw of the turn", n)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(afrB15Named(g, me.ID, "Illusion")); n != 1 {
		t.Fatalf("%d Illusions after the second draw, want 1", n)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(afrB15Named(g, me.ID, "Illusion")); n != 1 {
		t.Errorf("%d Illusions after the third draw, want still 1", n)
	}
}

func TestAfrB15IllusionGetsPlusOnePowerForEachOtherIllusion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.AfrB15IllusionToken(), 3) })
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.AfrB15IllusionToken(), 1) })
	mine := afrB15Named(g, me.ID, "Illusion")
	if len(mine) != 3 {
		t.Fatalf("%d Illusions, want 3", len(mine))
	}
	for _, id := range mine {
		if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tg != 1 {
			t.Errorf("Illusion is %d/%d with two others, want 3/1 (the opponent's does not count)", p, tg)
		}
	}
	for _, id := range afrB15Named(g, opp.ID, "Illusion") {
		if p := effectstest.EffectivePower(t, g, id); p != 1 {
			t.Errorf("the opponent's lone Illusion has power %d, want 1", p)
		}
	}
	c := effectstest.B12Card(t, g, mine[0])
	if c.TypeLine != "Token Creature — Illusion" || len(c.Colors) != 1 || c.Colors[0] != "U" {
		t.Errorf("Illusion is %+v", c)
	}
}

func TestAfrB15MinnPutsAPermanentFromHandWhenAnIllusionDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Minn, Wily Illusionist", "Legendary Creature — Gnome Wizard", afrB15Minn, 1, 3)
	illusion := effectstest.B12Creature(g, me.ID, "Illusion", "Token Creature — Illusion", 3, 1)
	cheap := uuid.New()
	pricey := uuid.New()
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2})
	me.Hand.PushTop(game.Card{InstanceID: pricey, Name: "Pricey Ogre", TypeLine: "Creature — Ogre", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID, Power: 4, Toughness: 4})
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, illusion)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt for the put")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range prompt.ChooseCards {
		offered[id] = true
	}
	if !offered[cheap] || offered[pricey] || offered[spell] {
		t.Errorf("offered %v: want only the mana value 2 permanent card (power 3 allows at most 3)", prompt.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, cheap) != game.ZoneBattlefield {
		t.Errorf("the chosen card is in %v, want the battlefield", effectstest.B12ZoneOf(g, cheap))
	}
}

func TestAfrB15MinnIgnoresADyingNonIllusion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Minn, Wily Illusionist", "Legendary Creature — Gnome Wizard", afrB15Minn, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 1)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if prompt := effectstest.ChooseCardsChoiceFor(g, me.ID); prompt != nil {
		t.Error("a dying Bear offered Minn's put")
	}
}

// afrB15ToughnessWithCounters is a permanent's toughness with its counters.
func afrB15ToughnessWithCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	out, found := 0, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.CurrentToughness(), true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return out
}

// afrB15Walker seeds a planeswalker with the given loyalty under seat 0 and
// walks to the main phase.
func afrB15Walker(t *testing.T, g *game.Game, name, oracle string, loyalty int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, name, "Legendary Planeswalker — "+name, oracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	effectstest.AdvanceToMain(t, g)
	return id
}

// --- Lolth, Spider Queen -------------------------------------------------------

func TestAfrB15LolthGainsLoyaltyWhenACreatureYouControlDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lolth := afrB15Walker(t, g, "Lolth, Spider Queen", afrB15Lolth, 4)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, lolth); got != 4 {
		t.Errorf("an opponent's creature dying changed Lolth's loyalty to %d", got)
	}
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, lolth); got != 5 {
		t.Errorf("Lolth's loyalty %d after your creature died, want 5", got)
	}
}

func TestAfrB15LolthZeroDrawsAndLosesOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryTopForTest(me, "Island")
	lolth := afrB15Walker(t, g, "Lolth, Spider Queen", afrB15Lolth, 4)
	hand, life := me.Hand.Size(), me.Life
	effectstest.B16Activate(t, g, me.ID, lolth, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d → %d, life %d → %d; want one card and one life", hand, me.Hand.Size(), life, me.Life)
	}
	if got := effectstest.LoyaltyOf(g, lolth); got != 4 {
		t.Errorf("0 ability changed loyalty to %d", got)
	}
}

func TestAfrB15LolthMinusThreeMakesTwoSpiders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lolth := afrB15Walker(t, g, "Lolth, Spider Queen", afrB15Lolth, 4)
	effectstest.B16Activate(t, g, me.ID, lolth, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(afrB15Named(g, me.ID, "Spider")); n != 2 {
		t.Fatalf("%d Spiders, want 2", n)
	}
	if got := effectstest.LoyaltyOf(g, lolth); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestAfrB15LolthMinusThreeIsRefusedBelowThreeLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lolth := afrB15Walker(t, g, "Lolth, Spider Queen", afrB15Lolth, 2)
	if err := g.ActivateCatalogAbility(me.ID, lolth, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("−3 was activated with 2 loyalty")
	}
}

// lolthEmblemTable gives seat 0 Lolth's emblem and a creature that can attack.
func lolthEmblemTable(t *testing.T, power int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lolth := afrB15Walker(t, g, "Lolth, Spider Queen", afrB15Lolth, 8)
	effectstest.B16Activate(t, g, me.ID, lolth, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("%d emblems, want 1", me.Emblems.Size())
	}
	attacker := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", power, power)
	return g, me, opp, attacker
}

func TestAfrB15LolthEmblemDrainsTheDifferenceFromEightLifeLostThisTurn(t *testing.T) {
	g, me, opp, ogre := lolthEmblemTable(t, 5)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-8 {
		t.Errorf("opponent life %d → %d, want 8 lost in all (5 combat damage and 3 more)", life, opp.Life)
	}
	_ = me
}

func TestAfrB15LolthEmblemCountsLifeAlreadyLostThisTurn(t *testing.T) {
	g, me, opp, ogre := lolthEmblemTable(t, 5)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(me.ID, opp.ID, -2) })
	effectstest.AttackWith(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	// 2 lost already, 5 combat damage: 7 lost, so 1 more.
	if opp.Life != life-8 {
		t.Errorf("opponent life %d → %d, want 8 lost in all", life, opp.Life)
	}
}

func TestAfrB15LolthEmblemDoesNothingWhenEightOrMoreWasLost(t *testing.T) {
	g, _, opp, ogre := lolthEmblemTable(t, 9)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-9 {
		t.Errorf("opponent life %d → %d, want only the 9 combat damage", life, opp.Life)
	}
}

func TestAfrB15LolthEmblemTriggersOncePerOpponentPerDamageStep(t *testing.T) {
	g, me, opp, ogre := lolthEmblemTable(t, 2)
	second := effectstest.B12Creature(g, me.ID, "Imp", "Creature — Imp", 1, 1)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, ogre, second)
	effectstest.PassPriorityAroundTable(t, g)
	// 3 combat damage, then one trigger for 5 more: 8 lost, not 8 + 5.
	if opp.Life != life-8 {
		t.Errorf("opponent life %d → %d, want 8 lost in all from one trigger", life, opp.Life)
	}
}

// --- Zariel, Archduke of Avernus ---------------------------------------------

func TestAfrB15ZarielPlusOnePumpsAndGivesHasteUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	zariel := afrB15Walker(t, g, "Zariel, Archduke of Avernus", afrB15Zariel, 4)
	effectstest.B16Activate(t, g, me.ID, zariel, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("my Bear has power %d, want 3", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "haste") {
		t.Error("my Bear has no haste")
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 || effectstest.HasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Errorf("the opponent's Bear is affected: power %d", p)
	}
	if got := effectstest.LoyaltyOf(g, zariel); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestAfrB15ZarielZeroMakesADevilThatPingsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zariel := afrB15Walker(t, g, "Zariel, Archduke of Avernus", afrB15Zariel, 4)
	effectstest.B16Activate(t, g, me.ID, zariel, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	devils := afrB15Named(g, me.ID, "Devil")
	if len(devils) != 1 {
		t.Fatalf("%d Devils, want 1", len(devils))
	}
	c := effectstest.B12Card(t, g, devils[0])
	if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "R" {
		t.Errorf("Devil is %+v, want a 1/1 red Devil", c)
	}
	life := opp.Life
	effectstest.KillCreature(t, g, me.ID, devils[0])
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d → %d, want 1 damage from the dying Devil", life, opp.Life)
	}
}

func TestAfrB15ZarielEmblemUntapsAndAddsACombatAtTheEndOfTheFirstCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zariel := afrB15Walker(t, g, "Zariel, Archduke of Avernus", afrB15Zariel, 6)
	effectstest.B16Activate(t, g, me.ID, zariel, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("%d emblems, want 1", me.Emblems.Size())
	}
	ogre := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.AnswerPickTarget(t, g, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, ogre); tapped {
		t.Error("the emblem's trigger did not untap the Ogre")
	}
	// The added combat: attack again with the untapped Ogre.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ogre, opp.ID); err != nil {
		t.Fatalf("the Ogre could not attack in the additional combat: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Errorf("opponent life %d → %d, want 6 (two combats of 3)", life, opp.Life)
	}
	// The additional combat's end does not trigger it again.
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := afrB15PickOpen(g); pick {
		t.Error("the emblem triggered again at the end of the additional combat")
	}
}

func TestAfrB15ZarielEmblemWithNoCreatureAddsNoCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zariel := afrB15Walker(t, g, "Zariel, Archduke of Avernus", afrB15Zariel, 6)
	effectstest.B16Activate(t, g, me.ID, zariel, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.StackMeta) != 0 || afrB15PickOpen(g) {
		t.Error("the emblem's trigger went on the stack with no creature to target")
	}
	for i := 0; i < 12 && g.Turn.Step != game.StepCleanup; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepBeginCombat {
			t.Fatal("an additional combat phase began")
		}
	}
}

// afrB15PickOpen reports whether a pick_target prompt is outstanding.
func afrB15PickOpen(g *game.Game) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return true
		}
	}
	return false
}

// --- Grand Master of Flowers ------------------------------------------------

func TestAfrB15GrandMasterIsAPlaneswalkerBelowSevenAndADragonGodAtSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	gm := afrB15Walker(t, g, "Grand Master of Flowers", afrB15GrandMaster, 6)
	c, _ := effectstest.BattlefieldCard(g, gm)
	if c.HasCardType("creature") {
		t.Fatal("a six-loyalty Grand Master is a creature")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(gm, game.CounterLoyalty, 1) })
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, _ = effectstest.BattlefieldCard(g, gm)
	if !c.HasCardType("creature") || !c.HasCardType("planeswalker") || !c.HasSubtype("Dragon") || !c.HasSubtype("God") {
		t.Errorf("Grand Master at seven loyalty has types %v subtypes %v", c.Effective().Types, c.Effective().Subtypes)
	}
	if p, tg := effectstest.EffectivePower(t, g, gm), effectstest.EffectiveToughness(t, g, gm); p != 7 || tg != 7 {
		t.Errorf("Grand Master is %d/%d, want 7/7", p, tg)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, gm, "flying") || !effectstest.HasKeywordOnBattlefield(t, g, gm, "indestructible") {
		t.Error("Grand Master lacks flying or indestructible at seven loyalty")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(gm, game.CounterLoyalty, -1) })
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, _ = effectstest.BattlefieldCard(g, gm)
	if c.HasCardType("creature") || effectstest.HasKeywordOnBattlefield(t, g, gm, "flying") {
		t.Error("Grand Master stayed a creature after dropping back to six loyalty")
	}
}

func TestAfrB15GrandMasterPlusOneStopsTheTargetBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := afrB15Walker(t, g, "Grand Master of Flowers", afrB15GrandMaster, 3)
	target := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(target)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, gm); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
	// It can't block on this turn...
	attacker := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(target, attacker); err == nil {
		t.Error("the restricted creature blocked")
	}
}

func TestAfrB15GrandMasterPlusOneRefusesACreatureWithVigilanceFirstStrikeOrDoubleStrike(t *testing.T) {
	for _, kw := range []string{"vigilance", "first strike", "double strike"} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		gm := afrB15Walker(t, g, "Grand Master of Flowers", afrB15GrandMaster, 3)
		target := effectstest.B12Creature(g, opp.ID, "Knight", "Creature — Knight", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == target {
					g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, kw)
				}
			}
		})
		if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(target), Strict: true}); err == nil {
			t.Errorf("the +1 targeted a creature with %s", kw)
		}
	}
}

func TestAfrB15GrandMasterPlusOneFindsMonkOfTheOpenHandInLibraryOrGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	monk := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Monk of the Open Hand", TypeLine: "Creature — Human Monk", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryTopForTest(me, "Island")
	gm := afrB15Walker(t, g, "Grand Master of Flowers", afrB15GrandMaster, 3)
	effectstest.B16Activate(t, g, me.ID, gm, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// The zones prompt, then the pick.
	if !me.Hand.Contains(monk) {
		effectstest.AnswerSearchByID(t, g, me.ID, monk)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !me.Hand.Contains(monk) {
		t.Fatalf("Monk of the Open Hand is in %v, want the hand", effectstest.B12ZoneOf(g, monk))
	}
	if got := effectstest.LoyaltyOf(g, gm); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

// --- Mordenkainen ------------------------------------------------------------------

func TestAfrB15MordenkainenPlusTwoDrawsTwoThenPutsACardOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryTopForTest(me, "Island")
	}
	mord := afrB15Walker(t, g, "Mordenkainen", afrB15Mordenkainen, 5)
	keep := effectstest.HandCardForTest(me, "Keep", "Sorcery", "")
	hand, library := me.Hand.Size(), me.Library.Size()
	effectstest.B16Activate(t, g, me.ID, mord, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt for the card to put on the bottom")
	}
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d → %d, want two cards drawn", hand, me.Hand.Size())
	}
	effectstest.AnswerChooseCards(t, g, me.ID, keep)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Library.Size() != library-1 {
		t.Errorf("hand %d, library %d; want a net of one more card in hand and one fewer in the library", me.Hand.Size(), me.Library.Size())
	}
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != keep {
		t.Error("the chosen card is not on the bottom of the library")
	}
	if got := effectstest.LoyaltyOf(g, mord); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestAfrB15MordenkainenMinusTwoMakesADogIllusionSizedByTwiceTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mord := afrB15Walker(t, g, "Mordenkainen", afrB15Mordenkainen, 5)
	for i := 0; i < 3; i++ {
		effectstest.HandCardForTest(me, "Card", "Sorcery", "")
	}
	effectstest.B16Activate(t, g, me.ID, mord, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	dogs := afrB15Named(g, me.ID, "Dog Illusion")
	if len(dogs) != 1 {
		t.Fatalf("%d Dog Illusions, want 1", len(dogs))
	}
	size := me.Hand.Size()
	if p, tg := effectstest.EffectivePower(t, g, dogs[0]), effectstest.EffectiveToughness(t, g, dogs[0]); p != 2*size || tg != 2*size {
		t.Errorf("Dog Illusion is %d/%d with %d cards in hand, want %d/%d", p, tg, size, 2*size, 2*size)
	}
	effectstest.PushLibraryTopForTest(me, "Island")
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	if p := effectstest.EffectivePower(t, g, dogs[0]); p != 2*me.Hand.Size() {
		t.Errorf("Dog Illusion's power %d after another card, want %d", p, 2*me.Hand.Size())
	}
	c := effectstest.B12Card(t, g, dogs[0])
	if c.TypeLine != "Token Creature — Dog Illusion" || len(c.Colors) != 1 || c.Colors[0] != "U" {
		t.Errorf("Dog Illusion is %+v", c)
	}
}

func TestAfrB15MordenkainenMinusTenExchangesHandAndLibraryAndGivesTheEmblem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var library []uuid.UUID
	for i := 0; i < 5; i++ {
		library = append(library, effectstest.PushLibraryTopForTest(me, "Island"))
	}
	mord := afrB15Walker(t, g, "Mordenkainen", afrB15Mordenkainen, 10)
	// The turn's draw put one library card in hand; note what is there now.
	var hand []uuid.UUID
	for _, c := range me.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	for i := 0; i < 2; i++ {
		hand = append(hand, effectstest.HandCardForTest(me, "Spell", "Sorcery", ""))
	}
	var libNow []uuid.UUID
	for _, c := range me.Library.Cards {
		libNow = append(libNow, c.InstanceID)
	}
	effectstest.B16Activate(t, g, me.ID, mord, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != len(libNow) || me.Library.Size() != len(hand) {
		t.Fatalf("hand %d and library %d; want %d and %d", me.Hand.Size(), me.Library.Size(), len(libNow), len(hand))
	}
	for _, id := range libNow {
		if !me.Hand.Contains(id) {
			t.Errorf("library card %s did not reach the hand", id)
		}
	}
	for _, id := range hand {
		if !me.Library.Contains(id) {
			t.Errorf("hand card %s did not reach the library", id)
		}
	}
	if me.Emblems.Size() != 1 {
		t.Errorf("%d emblems, want 1", me.Emblems.Size())
	}
	if got := effectstest.LoyaltyOf(g, mord); got != 0 && got != -1 {
		t.Errorf("loyalty %d after −10 from 10", got)
	}
	_ = library
}

func TestAfrB15MordenkainenMinusTenWithAnEmptyLibraryStillMovesTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mord := afrB15Walker(t, g, "Mordenkainen", afrB15Mordenkainen, 10)
	g.WithWriteLock(func() {
		for me.Library.Size() > 0 {
			me.Library.Cards = me.Library.Cards[:0]
		}
	})
	effectstest.HandCardForTest(me, "Spell", "Sorcery", "")
	effectstest.HandCardForTest(me, "Spell", "Sorcery", "")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, mord, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 || me.Library.Size() != hand {
		t.Errorf("hand %d, library %d; want 0 and %d (CR 701.12f: an empty zone is still exchanged)", me.Hand.Size(), me.Library.Size(), hand)
	}
}

func TestAfrB15MordenkainenMinusTenNeedsTenLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mord := afrB15Walker(t, g, "Mordenkainen", afrB15Mordenkainen, 9)
	if err := g.ActivateCatalogAbility(me.ID, mord, 2, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("−10 was activated with 9 loyalty")
	}
}
