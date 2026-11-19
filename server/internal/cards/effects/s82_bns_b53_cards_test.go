package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b53_cards_test.go — Foundations Jumpstart, S82 BNS batch 53 (#5929):
// Ajani, Adversary of Tyrants; Garruk, Caller of Beasts; Jiang Yanggu,
// Wildcrafter; Mu Yanling, Sky Dancer; Sarkhan, Fireblood; Adventure Awaits;
// Adventurous Impulse; Aggressive Instinct; Agonizing Syphon; Bathe in
// Dragonfire; Blood Feud.

const (
	b53Ajani      = "b25bea73-7444-4ee2-8022-611461506117"
	b53Garruk     = "8e2fbc26-5373-4fd7-8e9e-41af5bb9ff18"
	b53Jiang      = "04f2c320-dfcc-440a-9e24-3c2083d74e7c"
	b53MuYanling  = "1e825298-7c99-40cc-81c9-280cc7ed98d3"
	b53Sarkhan    = "074cbe14-a0c8-4772-955a-b4052333d6ce"
	b53Awaits     = "70235727-e330-46fa-b057-ddbb55cd1ac3"
	b53Impulse    = "5bf49f39-b3d5-4148-8dc2-b5b5011769d2"
	b53Instinct   = "d1c8f14b-7977-488a-8bd9-cefd6687f017"
	b53Syphon     = "11669fe5-9298-4e1a-a4f7-6a4598dbc8dd"
	b53Dragonfire = "72d793e8-1d19-4ca1-894c-8851a2ebecde"
	b53BloodFeud  = "fc401d42-80fa-4a17-937a-c230e9317543"
)

func TestBnsB53CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		b53Ajani:      "Ajani, Adversary of Tyrants",
		b53Garruk:     "Garruk, Caller of Beasts",
		b53Jiang:      "Jiang Yanggu, Wildcrafter",
		b53MuYanling:  "Mu Yanling, Sky Dancer",
		b53Sarkhan:    "Sarkhan, Fireblood",
		b53Awaits:     "Adventure Awaits",
		b53Impulse:    "Adventurous Impulse",
		b53Instinct:   "Aggressive Instinct",
		b53Syphon:     "Agonizing Syphon",
		b53Dragonfire: "Bathe in Dragonfire",
		b53BloodFeud:  "Blood Feud",
	} {
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

// --- The six sorceries ---------------------------------------------------------

func TestBnsB53BatheInDragonfireDealsFour(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 5, 5)
	dmuB09Cast(t, g, "Bathe in Dragonfire", "Sorcery", b53Dragonfire, false, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 4 {
		t.Errorf("damage %d, want 4", got)
	}
}

func TestBnsB53BloodFeudMakesTwoCreaturesFight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Ogre", 3, 6)
	b := b12Creature(g, me.ID, "B", "Creature — Ogre", 2, 6)
	dmuB09Cast(t, g, "Blood Feud", "Sorcery", b53BloodFeud, false,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, a); got != 2 {
		t.Errorf("A took %d, want 2", got)
	}
	if got := damageOn(t, g, b); got != 3 {
		t.Errorf("B took %d, want 3", got)
	}
}

func TestBnsB53BloodFeudNeedsTwoDifferentCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Ogre", 3, 6)
	advanceToMain(t, g)
	id := handCardForTest(me, "Blood Feud", "Sorcery", b53BloodFeud)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: a}}}); err == nil {
		t.Error("Blood Feud was cast with one creature in both slots")
	}
}

func TestBnsB53AggressiveInstinctIsOneSided(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Ogre", 4, 4)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 9, 9)
	dmuB09Cast(t, g, "Aggressive Instinct", "Sorcery", b53Instinct, false,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 4 {
		t.Errorf("theirs took %d, want 4", got)
	}
	if got := damageOn(t, g, mine); got != 0 {
		t.Errorf("mine took %d, want 0", got)
	}
}

func TestBnsB53AggressiveInstinctRefusesYourOwnSecondTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Ogre", 4, 4)
	other := b12Creature(g, me.ID, "Other", "Creature — Ogre", 1, 1)
	advanceToMain(t, g)
	id := handCardForTest(me, "Aggressive Instinct", "Sorcery", b53Instinct)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Error("Aggressive Instinct was cast at a creature you control")
	}
}

func TestBnsB53AgonizingSyphonDamagesAndGains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := lifeOf(g, me.ID)
	dmuB09Cast(t, g, "Agonizing Syphon", "Sorcery", b53Syphon, false, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != 37 {
		t.Errorf("opponent life %d, want 37", got)
	}
	if got := lifeOf(g, me.ID); got != before+3 {
		t.Errorf("my life %d, want %d", got, before+3)
	}
}

// --- Adventurous Impulse / Adventure Awaits --------------------------------------

func b53Stack(p *game.Player, lines ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: lines[i], TypeLine: lines[i], Owner: p.ID, Controller: p.ID})
		ids[i] = id
	}
	return ids
}

func TestBnsB53AdventurousImpulseOffersCreaturesAndLandsFromTheTopThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := b53Stack(me, "Instant", "Basic Land — Forest", "Creature — Bear", "Creature — Wolf")
	dmuB09Cast(t, g, "Adventurous Impulse", "Sorcery", b53Impulse, false, nil)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("offered %d cards, want the land and the bear (the wolf is fourth)", len(prompt.ChooseCards))
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds [%d,%d], want [0,1]", prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Error("the land is not in hand")
	}
}

func TestBnsB53AdventureAwaitsTakesACreatureAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := b53Stack(me, "Instant", "Creature — Bear", "Sorcery", "Basic Land — Forest", "Artifact")
	dmuB09Cast(t, g, "Adventure Awaits", "Sorcery", b53Awaits, false, nil)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil || len(prompt.ChooseCards) != 1 {
		t.Fatalf("want one creature offered, got %v", prompt)
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !me.Hand.Contains(top[1]) || me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want the creature and no draw", hand, me.Hand.Size())
	}
}

func TestBnsB53AdventureAwaitsDrawsWhenNoCardIsTaken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := b53Stack(me, "Instant", "Creature — Bear", "Sorcery", "Basic Land — Forest", "Artifact", "Enchantment")
	dmuB09Cast(t, g, "Adventure Awaits", "Sorcery", b53Awaits, false, nil)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	// The five went to the bottom, so the sixth card is the one drawn.
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(top[5]) {
		t.Errorf("hand %d → %d, want the card under the five drawn", hand, me.Hand.Size())
	}
}

func TestBnsB53AdventureAwaitsDrawsWhenNoCreatureIsAmongThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := b53Stack(me, "Instant", "Sorcery", "Basic Land — Forest", "Artifact", "Enchantment", "Creature — Bear")
	dmuB09Cast(t, g, "Adventure Awaits", "Sorcery", b53Awaits, false, nil)
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, me.ID) != nil {
		t.Error("a prompt was raised with no creature among the five")
	}
	if !me.Hand.Contains(top[5]) {
		t.Error("the card under the five was not drawn")
	}
}

// --- Planeswalkers ------------------------------------------------------------------

func TestBnsB53AjaniPlusOneCountersUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ajani := pushWalkerForTest(g, me.ID, "Ajani, Adversary of Tyrants", b53Ajani, 4)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, g.Seats[1].ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, ajani, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}}); err == nil {
		t.Fatal("three targets were accepted for \"up to two\"")
	}
	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	passPriorityAroundTable(t, g)
	if countersOn(g, a, game.CounterPlusOne) != 1 || countersOn(g, b, game.CounterPlusOne) != 1 {
		t.Error("each target should have a +1/+1 counter")
	}
	if countersOn(g, c, game.CounterPlusOne) != 0 {
		t.Error("an untargeted creature got a counter")
	}
	if got := loyaltyOf(g, ajani); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestBnsB53AjaniMinusTwoReturnsASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ajani := pushWalkerForTest(g, me.ID, "Ajani, Adversary of Tyrants", b53Ajani, 4)
	small := uuid.New()
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Ogre", ManaCost: "{2}{G}",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, ajani, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err == nil {
		t.Fatal("a mana value 3 creature card was targeted")
	}
	b16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the small creature did not return to the battlefield")
	}
	if got := loyaltyOf(g, ajani); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestBnsB53AjaniEmblemMakesThreeLifelinkCatsEachEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ajani := pushWalkerForTest(g, me.ID, "Ajani, Adversary of Tyrants", b53Ajani, 7)
	b16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := len(emblemsOf(g, me.ID)); n != 1 {
		t.Fatalf("%d emblems, want 1", n)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Cat"); n != 3 {
		t.Fatalf("%d Cats, want 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat" && !hasEffectiveKeyword(t, g, c.InstanceID, "lifelink") {
			t.Error("a Cat has no lifelink")
		}
	}
}

func TestBnsB53SarkhanRummagesAndMakesDragons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sarkhan := pushWalkerForTest(g, me.ID, "Sarkhan, Fireblood", b53Sarkhan, 7)
	fodder := handCard(me, "Fodder", "Sorcery")
	b16Activate(t, g, me.ID, sarkhan, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if pick := latestChooseCardsFor(g, me.ID); pick != nil {
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{fodder}); err != nil {
			t.Fatalf("discard: %v", err)
		}
	}
	if me.Hand.Contains(fodder) {
		t.Error("the discarded card is still in hand")
	}
	// Loyalty abilities are once a turn per planeswalker (CR 606.3): the −7
	// comes from a second copy.
	sarkhan = pushWalkerForTest(g, me.ID, "Sarkhan, Fireblood", b53Sarkhan, 7)
	b16Activate(t, g, me.ID, sarkhan, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Dragon"); n != 4 {
		t.Fatalf("%d Dragons, want 4", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Dragon" {
			continue
		}
		if p, tough := ptOf(t, g, c.InstanceID); p != 5 || tough != 5 {
			t.Errorf("Dragon is %d/%d, want 5/5", p, tough)
		}
		if !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
			t.Error("a Dragon has no flying")
		}
	}
}

func TestBnsB53SarkhanManaPaysOnlyForDragonSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sarkhan := pushWalkerForTest(g, me.ID, "Sarkhan, Fireblood", b53Sarkhan, 3)
	b16Activate(t, g, me.ID, sarkhan, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		var open *game.PendingChoice
		for _, ch := range g.PendingChoices {
			if ch != nil && ch.Kind == game.PendingChoiceMana && ch.Chooser == me.ID {
				open = ch
				break
			}
		}
		if open == nil {
			break
		}
		if err := g.ResolveManaChoice(open.ID, me.ID, "R"); err != nil {
			t.Fatalf("mana choice: %v", err)
		}
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool has %d mana, want 2", got)
	}
	for _, m := range me.ManaPool.Units() {
		if len(m.Restrictions) == 0 {
			t.Errorf("mana %v carries no spend restriction", m)
		}
	}
	bear := handCardOf(g, me, "Bear", "Creature — Bear", "{R}{R}", "")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the Dragon-only mana cast a Bear")
	}
	drake := handCardOf(g, me, "Drake", "Creature — Dragon", "{R}{R}", "")
	if err := g.CastSpell(me.ID, drake, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the Dragon-only mana could not cast a Dragon: %v", err)
	}
}

func TestBnsB53GarrukPlusOneTakesEveryCreatureOfTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	garruk := pushWalkerForTest(g, me.ID, "Garruk, Caller of Beasts", b53Garruk, 4)
	top := b53Stack(me, "Creature — Bear", "Instant", "Creature — Wolf", "Sorcery", "Basic Land — Forest", "Creature — Ape")
	b16Activate(t, g, me.ID, garruk, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[0]) || !me.Hand.Contains(top[2]) {
		t.Error("both creatures of the top five should be in hand")
	}
	if me.Hand.Contains(top[5]) {
		t.Error("the sixth card's creature was taken")
	}
}

func TestBnsB53GarrukMinusThreePutsAGreenCreatureOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	garruk := pushWalkerForTest(g, me.ID, "Garruk, Caller of Beasts", b53Garruk, 4)
	green := handCardFull(me, "Green Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	red := handCardFull(me, "Red Bear", "Creature — Bear", "{1}{R}", "", []string{"R"})
	b16Activate(t, g, me.ID, garruk, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != green {
		t.Fatalf("offered %v, want only the green creature (red %s)", prompt.ChooseCards, red)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{green}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !g.Battlefield.Contains(green) {
		t.Error("the green creature is not on the battlefield")
	}
}

func TestBnsB53GarrukEmblemSearchesOnACreatureCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	garruk := pushWalkerForTest(g, me.ID, "Garruk, Caller of Beasts", b53Garruk, 7)
	b16Activate(t, g, me.ID, garruk, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := len(emblemsOf(g, me.ID)); n != 1 {
		t.Fatalf("%d emblems, want 1", n)
	}
	fetch := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: fetch, Name: "Fetched", TypeLine: "Creature — Wurm", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	id := handCardForTest(me, "Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fetch) {
		t.Error("the searched creature is not on the battlefield")
	}
}

func TestBnsB53JiangGivesCounteredCreaturesAnyColorMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jiang := pushWalkerForTest(g, me.ID, "Jiang Yanggu, Wildcrafter", b53Jiang, 3)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	countered := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Countered", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 1}})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterPlusOne: 1}})
	if n := grantedManaRowsOf(g, countered); n != 1 {
		t.Errorf("countered creature has %d granted rows, want 1", n)
	}
	if n := grantedManaRowsOf(g, plain); n != 0 {
		t.Errorf("a creature with no counter has %d granted rows, want 0", n)
	}
	if n := grantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's countered creature has %d granted rows, want 0", n)
	}
	b16Activate(t, g, me.ID, jiang, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}}})
	passPriorityAroundTable(t, g)
	if countersOn(g, plain, game.CounterPlusOne) != 1 {
		t.Error("the −1 put no counter on the target")
	}
	if n := grantedManaRowsOf(g, plain); n != 1 {
		t.Errorf("the creature that just got a counter has %d granted rows, want 1", n)
	}
	if got := loyaltyOf(g, jiang); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestBnsB53MuYanlingShrinksAndGroundsUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mu := pushWalkerForTest(g, me.ID, "Mu Yanling, Sky Dancer", b53MuYanling, 2)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird",
		Power: 3, Toughness: 3, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	b16Activate(t, g, me.ID, mu, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: flier}}})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, flier); p != 1 || tough != 3 {
		t.Errorf("flier is %d/%d, want 1/3", p, tough)
	}
	if hasEffectiveKeyword(t, g, flier, "flying") {
		t.Error("the flier kept flying")
	}
	advanceToNextSeatsTurn(t, g)
	if p, _ := ptOf(t, g, flier); p != 1 {
		t.Errorf("flier power %d on another player's turn, want it still 1", p)
	}
	advanceToNextSeatsTurn(t, g)
	advanceToNextSeatsTurn(t, g)
	advanceToNextSeatsTurn(t, g)
	if p, _ := ptOf(t, g, flier); p != 3 {
		t.Errorf("flier power %d after my next turn began, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, flier, "flying") {
		t.Error("the flier did not get flying back")
	}
}

func TestBnsB53MuYanlingMakesABirdAndAnEmblemThatDrawsFromIslands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mu := pushWalkerForTest(g, me.ID, "Mu Yanling, Sky Dancer", b53MuYanling, 8)
	b16Activate(t, g, me.ID, mu, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Elemental Bird"); n != 1 {
		t.Fatalf("%d Elemental Birds, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental Bird" {
			if p, tough := ptOf(t, g, c.InstanceID); p != 4 || tough != 4 {
				t.Errorf("Bird is %d/%d, want 4/4", p, tough)
			}
			if !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("the Bird has no flying")
			}
		}
	}
	island := b12Creature(g, me.ID, "Island", "Basic Land — Island", 0, 0)
	forest := b12Creature(g, me.ID, "Forest", "Basic Land — Forest", 0, 0)
	mu = pushWalkerForTest(g, me.ID, "Mu Yanling, Sky Dancer", b53MuYanling, 8)
	b16Activate(t, g, me.ID, mu, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := len(emblemsOf(g, me.ID)); n != 1 {
		t.Fatalf("%d emblems, want 1", n)
	}
	// Islands have the granted ability; Forests do not. Tap the Island.
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, forest, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a Forest drew a card")
	}
	b16Activate(t, g, me.ID, island, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn from the Island", hand, me.Hand.Size())
	}
}
