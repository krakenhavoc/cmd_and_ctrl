package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4834, Secrets of Strixhaven Commander cards, batch 1. Every test drives
// the registered catalog row through the engine.

// socB1Oracle is the registered oracle id of `name`.
func socB1Oracle(t *testing.T, name string) string {
	t.Helper()
	return hobOracle(t, name)
}

// socB1Count is how many permanents named `name` the player controls.
func socB1Count(g *game.Game, controller uuid.UUID, name string) int {
	return hobB6BattlefieldCount(g, controller, name)
}

// socB1GraveyardCard puts a card with a mana cost into the player's
// graveyard.
func socB1GraveyardCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func TestSocB1CardsAreRegisteredFull(t *testing.T) {
	for _, name := range []string{
		"Abstract Performance", "Advanced Reconstruction", "Ajani's Chosen", "Alchemist's Refuge",
		"Angelic Destiny", "Angel of Indemnity", "Atsushi, the Blazing Sky", "Augusta, Order Returned",
		"Balefire Liege", "Biomass Mutation",
	} {
		spec, ok := Lookup(socB1Oracle(t, name))
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, completeness %v, caveats %v", name, ok, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Balefire Liege ---------------------------------------------------

func TestSocB1BalefireLiegeBuffsOtherRedAndWhiteCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := hobPush(t, g, me.ID, "Balefire Liege", "Creature — Spirit Horror", 2, 4)
	red := b16Creature(g, me.ID, "Red Bear", "Creature — Bear", 2, 2, "R")
	white := b16Creature(g, me.ID, "White Bear", "Creature — Bear", 2, 2, "W")
	green := b16Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	both := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Boros Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"R", "W"}, Owner: me.ID, Controller: me.ID})
	theirs := b16Creature(g, opp.ID, "Their Red Bear", "Creature — Bear", 2, 2, "R")
	for name, want := range map[string]struct {
		id uuid.UUID
		p  int
	}{
		"red": {red, 3}, "white": {white, 3}, "green": {green, 2}, "red-white": {both, 4},
		"opponent's red": {theirs, 2}, "the Liege itself": {liege, 2},
	} {
		if got := effectivePower(t, g, want.id); got != want.p {
			t.Errorf("%s creature power %d, want %d", name, got, want.p)
		}
	}
}

func TestSocB1BalefireLiegeRedSpellDeals3ToTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hobPush(t, g, me.ID, "Balefire Liege", "Creature — Spirit Horror", 2, 4)
	life, mine := hobB6Life(g, opp), hobB6Life(g, me)
	floatForTest(g, me, "R")
	castFromHandForTest(t, g, me, "Spell", "Instant", "{R}", "", strict)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := hobB6Life(g, opp); got != life-3 {
		t.Errorf("the opponent's life %d -> %d, want -3", life, got)
	}
	if got := hobB6Life(g, me); got != mine {
		t.Errorf("a red spell changed my life %d -> %d", mine, got)
	}
}

func TestSocB1BalefireLiegeWhiteSpellGains3AndAGreenSpellNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hobPush(t, g, me.ID, "Balefire Liege", "Creature — Spirit Horror", 2, 4)
	mine := hobB6Life(g, me)
	floatForTest(g, me, "WG")
	castFromHandForTest(t, g, me, "White Spell", "Instant", "{W}", "", strict)
	passPriorityAroundTable(t, g)
	if got := hobB6Life(g, me); got != mine+3 {
		t.Fatalf("a white spell: life %d -> %d, want +3", mine, got)
	}
	castFromHandForTest(t, g, me, "Green Spell", "Instant", "{G}", "", strict)
	passPriorityAroundTable(t, g)
	if got := hobB6Life(g, me); got != mine+3 {
		t.Errorf("a green spell changed my life to %d", got)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a green spell asked for a damage target")
	}
}

// --- Alchemist's Refuge -----------------------------------------------

func TestSocB1AlchemistsRefugeLetsSpellsBeCastWithFlashThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushPermanentForTest(g, me.ID, "Alchemist's Refuge", socB1Oracle(t, "Alchemist's Refuge"), "Land")

	// Without the ability, a creature spell in the end step is refused.
	advanceTo(t, g, game.StepEnd)
	early := ctHandCard(me, "Early Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, early, game.CastSpellParams{}); err == nil {
		t.Fatal("a creature spell was cast in the end step without the Refuge's ability")
	}
	g2 := newCatalogGame(t)
	me = g2.Seats[g2.Turn.ActiveSeat]
	advanceToMain(t, g2)
	refuge := pushPermanentForTest(g2, me.ID, "Alchemist's Refuge", socB1Oracle(t, "Alchemist's Refuge"), "Land")
	b06AddMana(me, "G", "U")
	if err := g2.ActivateCatalogAbility(me.ID, refuge, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Refuge: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(refuge) {
		t.Error("the Refuge is not sacrificed")
	}
	advanceTo(t, g2, game.StepEnd)
	bear := ctHandCard(me, "Proof Bear", "Creature — Bear")
	if err := g2.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("a creature spell in the end step under the Refuge: %v", err)
	}
}

func TestSocB1AlchemistsRefugeNeedsGreenAndBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	refuge := pushPermanentForTest(g, me.ID, "Alchemist's Refuge", socB1Oracle(t, "Alchemist's Refuge"), "Land")
	b06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, refuge, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was paid with {G}{G}")
	}
}

// --- Angelic Destiny --------------------------------------------------

func TestSocB1AngelicDestinyMakesAnAngelAndReturnsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Angelic Destiny", socB1Oracle(t, "Angelic Destiny"), bear)
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("enchanted power %d, want 6", got)
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("enchanted creature lacks %s", kw)
		}
	}
	var subtypes []string
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(bear)
		subtypes = c.Effective().Subtypes
	})
	if !slices.Contains(subtypes, "Angel") || !slices.Contains(subtypes, "Bear") {
		t.Errorf("subtypes %v, want an Angel in addition to a Bear", subtypes)
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestSocB1AngelicDestinyStaysWhenTheCreatureIsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Angelic Destiny", socB1Oracle(t, "Angelic Destiny"), bear)
	exileCreature(t, g, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(aura) {
		t.Error("the Aura returned although the creature did not die")
	}
}

// --- Biomass Mutation -------------------------------------------------

func TestSocB1BiomassMutationSetsBasePowerAndToughnessOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	big := b12Creature(g, me.ID, "Big", "Creature — Beast", 6, 6)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(small, game.CounterPlusOne, 1) })
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	floatForTest(g, me, "GGUGG")
	castFromHandForTest(t, g, me, "Biomass Mutation", "Instant", "{X}{G/U}{G/U}", socB1Oracle(t, "Biomass Mutation"),
		game.CastSpellParams{Strict: true, XValue: 3})
	passPriorityAroundTable(t, g)
	if got := effectivePowerWithCounters(t, g, big); got != 3 {
		t.Errorf("Big's power %d, want 3", got)
	}
	if got := effectivePowerWithCounters(t, g, small); got != 4 {
		t.Errorf("Small's power %d, want 4 (base 3 and a counter)", got)
	}
	if got := effectivePowerWithCounters(t, g, theirs); got != 1 {
		t.Errorf("an opponent's creature changed to %d", got)
	}
	// The set is fixed as it resolves (CR 611.2c): a later creature keeps
	// its own base.
	later := b12Creature(g, me.ID, "Later", "Creature — Beast", 5, 5)
	if got := effectivePowerWithCounters(t, g, later); got != 5 {
		t.Errorf("a creature that entered afterwards is %d, want 5", got)
	}
	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepCleanup)
	if got := effectivePowerWithCounters(t, g, big); got != 6 {
		t.Errorf("after the turn Big's power %d, want 6", got)
	}
}

// --- Angel of Indemnity -----------------------------------------------

func TestSocB1AngelOfIndemnityReturnsAPermanentCardWithManaValue4OrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := socB1GraveyardCard(me, "Rock", "Artifact", "{4}")
	castCatalogSpell(t, g, "Angel of Indemnity", "Creature — Angel Warrior", socB1Oracle(t, "Angel of Indemnity"), nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("the artifact card did not return to the battlefield")
	}
}

func TestSocB1AngelOfIndemnityCantReturnManaValue5OrASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := socB1GraveyardCard(me, "Big Rock", "Artifact", "{5}")
	bolt := socB1GraveyardCard(me, "Bolt", "Instant", "{R}")
	castCatalogSpell(t, g, "Angel of Indemnity", "Creature — Angel Warrior", socB1Oracle(t, "Angel of Indemnity"), nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for with no permanent card of mana value 4 or less")
	}
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(bolt) {
		t.Error("an illegal card returned")
	}
}

func TestSocB1AngelOfIndemnityEncoreMakesAFlyingTokenPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Angel of Indemnity", "Creature — Angel Warrior",
		socB1Oracle(t, "Angel of Indemnity"), 5, 5, "{C}{C}{C}{C}{C}{C}{W}{W}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatal("encore's cost exiles the card")
	}
	passPriorityAroundTable(t, g)
	var toks []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel of Indemnity" && c.Controller == me.ID && c.IsToken() {
			toks = append(toks, c.InstanceID)
		}
	}
	if len(toks) != 3 {
		t.Fatalf("%d tokens, want one per opponent", len(toks))
	}
	for _, kw := range []string{"flying", "lifelink", "haste"} {
		if !hasEffectiveKeyword(t, g, toks[0], kw) {
			t.Errorf("the token lacks %s", kw)
		}
	}
}

// --- Atsushi, the Blazing Sky -----------------------------------------

func socB1KillAtsushi(t *testing.T, g *game.Game, owner uuid.UUID) *game.PendingChoice {
	t.Helper()
	atsushi := hobPush(t, g, owner, "Atsushi, the Blazing Sky", "Legendary Creature — Dragon Spirit", 4, 4)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(atsushi); err != nil {
			t.Fatalf("destroy Atsushi: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, owner)
	if c == nil || len(c.ModeOptionIndex) != 2 {
		t.Fatalf("the dies trigger asks one of two modes: %+v", c)
	}
	return c
}

func TestSocB1AtsushiDiesForThreeTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := socB1KillAtsushi(t, g, me.ID)
	answerMode(t, g, c, me.ID, 1)
	passPriorityAroundTable(t, g)
	if got := socB1Count(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("%d Treasures, want 3", got)
	}
}

func TestSocB1AtsushiDiesToExileTheTopTwoYouMayPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := sunbirdLibrary(me,
		game.Card{Name: "Top One", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "Top Two", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Third", TypeLine: "Sorcery", ManaCost: "{R}"})
	c := socB1KillAtsushi(t, g, me.ID)
	answerMode(t, g, c, me.ID, 0)
	passPriorityAroundTable(t, g)
	for _, id := range top[:2] {
		if !inExile(g, id) {
			t.Fatalf("card %s was not exiled", id)
		}
		perm := g.CastPermissionOnCardByIDForEffect(id)
		if !perm.Granted() || perm.Player != me.ID || perm.CastOnly {
			t.Errorf("permission on %s = %+v, want a play grant to the controller", id, perm)
		}
	}
	if inExile(g, top[2]) {
		t.Error("a third card was exiled")
	}
}

// --- Ajani's Chosen ---------------------------------------------------

func TestSocB1AjanisChosenMakesACatAndMayMoveAnAuraOntoIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobPush(t, g, me.ID, "Ajani's Chosen", "Creature — Cat Soldier", 3, 3)
	bear := auraBear(g, me.ID)
	aura := castCatalogSpell(t, g, "Angelic Destiny", auraTypeLine, socB1Oracle(t, "Angelic Destiny"),
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	var cat uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat" && c.IsToken() && c.Controller == me.ID {
			cat = c.InstanceID
		}
	}
	if cat == uuid.Nil {
		t.Fatal("no Cat token")
	}
	var host uuid.UUID
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(aura)
		host = c.AttachedTo.ID
	})
	if host != cat {
		t.Fatalf("the Aura is attached to %s, want the Cat", host)
	}
	if got := effectivePower(t, g, cat); got != 6 {
		t.Errorf("the enchanted Cat has power %d, want 6", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("the bear kept the bonus: power %d", got)
	}
}

func TestSocB1AjanisChosenDecliningLeavesTheAura(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobPush(t, g, me.ID, "Ajani's Chosen", "Creature — Cat Soldier", 3, 3)
	bear := auraBear(g, me.ID)
	castCatalogSpell(t, g, "Angelic Destiny", auraTypeLine, socB1Oracle(t, "Angelic Destiny"),
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("the bear is %d after declining, want 6", got)
	}
	if got := socB1Count(g, me.ID, "Cat"); got != 1 {
		t.Errorf("%d Cats, want 1", got)
	}
}

func TestSocB1AjanisChosenANonAuraEnchantmentMakesACatAndAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hobPush(t, g, me.ID, "Ajani's Chosen", "Creature — Cat Soldier", 3, 3)
	floatForTest(g, me, "W")
	castFromHandForTest(t, g, me, "Plain Enchantment", "Enchantment", "{W}", "", strict)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("a non-Aura enchantment asked to attach")
	}
	if got := socB1Count(g, me.ID, "Cat"); got != 1 {
		t.Errorf("%d Cats, want 1", got)
	}
	// A creature entering makes nothing.
	floatForTest(g, me, "G")
	castFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{G}", "", strict)
	passPriorityAroundTable(t, g)
	if got := socB1Count(g, me.ID, "Cat"); got != 1 {
		t.Errorf("a creature made a Cat: %d", got)
	}
}

// --- Augusta, Order Returned ------------------------------------------

func TestSocB1AugustaCountsTheNonlandCardsEachPlayerExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	augusta := hobPush(t, g, me.ID, "Augusta, Order Returned", "Legendary Creature — Spirit Advisor", 1, 3)
	mine := socB1GraveyardCard(me, "My Spell", "Sorcery", "{1}")
	theirLand := socB1GraveyardCard(opp, "Their Land", "Land", "")
	_ = socB1GraveyardCard(opp, "Their Spell", "Instant", "{U}")
	thirdSpell := socB1GraveyardCard(third, "Third Spell", "Instant", "{U}")
	declareAttack(t, g, opp.ID, augusta)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, mine)
	answerChooseCards(t, g, opp.ID, theirLand)
	answerChooseCards(t, g, third.ID, thirdSpell)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirLand, thirdSpell} {
		if !inExile(g, id) {
			t.Errorf("chosen card %s was not exiled", id)
		}
	}
	answerPickTarget(t, g, augusta)
	passPriorityAroundTable(t, g)
	if got := effectivePowerWithCounters(t, g, augusta); got != 3 {
		t.Errorf("Augusta's power %d, want 3 (two nonland cards)", got)
	}
}

func TestSocB1AugustaOnlyLandsExiledMakesNoTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	augusta := hobPush(t, g, me.ID, "Augusta, Order Returned", "Legendary Creature — Spirit Advisor", 1, 3)
	land := socB1GraveyardCard(opp, "Their Land", "Land", "")
	declareAttack(t, g, opp.ID, augusta)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, land)
	passPriorityAroundTable(t, g)
	if !inExile(g, land) {
		t.Fatal("the land was not exiled")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a land alone made the counter trigger")
	}
	if got := effectivePowerWithCounters(t, g, augusta); got != 1 {
		t.Errorf("Augusta's power %d, want 1", got)
	}
}

// --- Advanced Reconstruction ------------------------------------------

func TestSocB1AdvancedReconstructionMillsThenExilesAtRandomToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushClass(g, me.ID, "Advanced Reconstruction", "Enchantment — Class", socB1Oracle(t, "Advanced Reconstruction"))
	top := sunbirdLibrary(me, game.Card{Name: "Milled", TypeLine: "Sorcery", ManaCost: "{R}"})
	gyBefore := me.Graveyard.Size()
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(top[0]) == inExile(g, top[0]) {
		t.Fatalf("the milled card is neither only in the graveyard nor only exiled")
	}
	if got := me.Graveyard.Size(); got != gyBefore {
		t.Errorf("graveyard %d -> %d, want one in (mill) and one out (exile)", gyBefore, got)
	}
	var exiled uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Owner == me.ID {
			exiled = c.InstanceID
		}
	}
	if exiled == uuid.Nil {
		t.Fatal("nothing was exiled")
	}
	perm := g.CastPermissionOnCardByIDForEffect(exiled)
	if !perm.Granted() || perm.Player != me.ID || perm.CastOnly {
		t.Errorf("permission = %+v, want a play grant", perm)
	}
}

func TestSocB1AdvancedReconstructionLevelTwoBurnsWhenCardsLeaveYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	class := pushClass(g, me.ID, "Advanced Reconstruction", "Enchantment — Class", socB1Oracle(t, "Advanced Reconstruction"))
	a := socB1GraveyardCard(me, "A", "Sorcery", "{1}")
	b := socB1GraveyardCard(me, "B", "Sorcery", "{1}")
	life, myLife := hobB6Life(g, opp), hobB6Life(g, me)
	// Level 1: nothing happens.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, a); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := hobB6Life(g, opp); got != life {
		t.Fatalf("level 1 dealt damage: %d -> %d", life, got)
	}
	if err := levelUpTo(t, g, me.ID, class, 0, "{R}{R}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if levelOf(g, class) != 2 {
		t.Fatalf("level %d, want 2", levelOf(g, class))
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, b); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if got := hobB6Life(g, p); got != life-2 {
			t.Errorf("seat %s life %d, want %d", p.Name, got, life-2)
		}
	}
	if got := hobB6Life(g, me); got != myLife {
		t.Errorf("my life changed: %d", got)
	}
}

func TestSocB1AdvancedReconstructionLevelThreeDiscountsSpellsFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	class := pushClass(g, me.ID, "Advanced Reconstruction", "Enchantment — Class", socB1Oracle(t, "Advanced Reconstruction"))
	if err := levelUpTo(t, g, me.ID, class, 0, "{R}{R}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := levelUpTo(t, g, me.ID, class, 1, "{R}{R}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	if levelOf(g, class) != 3 {
		t.Fatalf("level %d, want 3", levelOf(g, class))
	}
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: spell, Name: "Exiled Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{R}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		g.GrantCastPermissionOverCardForEffect(spell, game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Label: "test"})
	})
	// From hand a {2}{R} spell is not discounted: {R} can't pay it.
	hand := ctHandCard(me, "Hand Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == hand {
				me.Hand.Cards[i].ManaCost = "{2}{R}"
			}
		}
	})
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, hand, strict); err == nil {
		t.Error("a spell from hand was discounted")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("the exiled {2}{R} spell for {R}: %v", err)
	}
}

// --- Abstract Performance ---------------------------------------------

// socB1AbstractPerformance casts the sorcery over a seeded library and
// has the controller name `opp` as the chooser. It returns the face-down
// pile and the face-up pile, top first.
func socB1AbstractPerformance(t *testing.T, g *game.Game, opp *game.Player) (down, up []uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	ids := sunbirdLibrary(me,
		game.Card{Name: "Down Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Down Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Down Rock", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Down Imp", TypeLine: "Creature — Imp", ManaCost: "{B}", Power: 1, Toughness: 1},
		game.Card{Name: "Up Wurm", TypeLine: "Creature — Wurm", ManaCost: "{6}{G}", Power: 7, Toughness: 7},
		game.Card{Name: "Up Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Up Elf", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1},
		game.Card{Name: "Up Spell", TypeLine: "Sorcery", ManaCost: "{U}"},
		game.Card{Name: "Stays", TypeLine: "Sorcery", ManaCost: "{U}"},
	)
	castCatalogSpell(t, g, "Abstract Performance", "Sorcery", socB1Oracle(t, "Abstract Performance"), nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	down, up = ids[:4], ids[4:8]
	for _, id := range append(append([]uuid.UUID(nil), down...), up...) {
		if !inExile(g, id) {
			t.Fatalf("card %s is not exiled", id)
		}
	}
	if inExile(g, ids[8]) {
		t.Fatal("a ninth card was exiled")
	}
	for _, id := range down {
		for _, s := range g.Seats {
			if knownBy(g, id, s.ID) {
				t.Fatalf("face-down card %s is known by %s", id, s.Name)
			}
		}
	}
	for _, id := range up {
		if !knownBy(g, id, opp.ID) || !knownBy(g, id, me.ID) {
			t.Fatalf("face-up card %s is not public", id)
		}
	}
	return down, up
}

func TestSocB1AbstractPerformanceOpponentBinsTheFaceDownPile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	down, up := socB1AbstractPerformance(t, g, opp)
	answerOptionPick(t, g, opp.ID, 0)
	for _, id := range down {
		if !me.Graveyard.Contains(id) {
			t.Errorf("face-down card %s is not in the graveyard", id)
		}
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, up[0], up[2], up[3])
	if err := castThroughResolutionCast(t, g, me.ID, up[0], game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast of the Wurm: %v", err)
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a second free cast is offered: the card casts one spell")
	}
	for _, id := range up[1:] {
		if !inHandOf(me, id) {
			t.Errorf("card %s of the kept pile is not in hand", id)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(up[0]) {
		t.Error("the free Wurm did not resolve")
	}
}

func TestSocB1AbstractPerformanceKeptFaceDownPileIsSeenOnlyByItsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	down, up := socB1AbstractPerformance(t, g, opp)
	answerOptionPick(t, g, opp.ID, 1)
	for _, id := range up {
		if !me.Graveyard.Contains(id) {
			t.Errorf("face-up card %s is not in the graveyard", id)
		}
	}
	for _, id := range down {
		if !knownBy(g, id, me.ID) {
			t.Errorf("the controller does not see kept face-down card %s", id)
		}
		if knownBy(g, id, opp.ID) {
			t.Errorf("the opponent sees kept face-down card %s", id)
		}
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, down[0], down[2], down[3])
	declineResolutionCast(t, g, me.ID)
	for _, id := range down {
		if !inHandOf(me, id) {
			t.Errorf("card %s is not in hand after declining", id)
		}
	}
}
