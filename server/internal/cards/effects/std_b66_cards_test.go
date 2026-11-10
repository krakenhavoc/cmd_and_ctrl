package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b66_cards_test.go — Lorwyn Eclipsed leftovers, batch 66 (#4458).

const (
	stdB66InfernoOracle     = "649e4969-71d4-434f-85b0-205ec2c6b24f"
	stdB66EndBlazeOracle    = "00cfbea0-e862-468b-90d5-7478eb9847c0"
	stdB66FlamebraiderOracl = "c552bd73-ce75-40fd-9d80-9e90b6f45274"
	stdB66EntranceOracle    = "b77764fe-bc79-410a-995a-ac392f5b46b8"
	stdB66ZealotOracle      = "d18ca75c-3572-4851-9580-27bafe10edd6"
	stdB66MeekAttackOracle  = "b9fff42d-31cf-4edf-8353-4a3b996bb77b"
	stdB66RansackingOracle  = "ebcd198e-7fbd-46e0-a6d9-ff499e541978"
	stdB66ChangelingOracle  = "60ad7aaa-8ce6-4449-8e8d-4ccf925010db"
	stdB66SquawkOracle      = "6b79404b-2fae-41fb-994c-d4cd895e8f09"
	stdB66TweezeOracle      = "9a294cbd-df18-4ee0-acaa-e31390a7f508"
	stdB66DruidOracle       = "5cb6ebf7-6a75-4464-9e39-e171fcb52de3"
)

func TestStdB66EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		stdB66InfernoOracle:     "Collective Inferno",
		stdB66EndBlazeOracle:    "End-Blaze Epiphany",
		stdB66FlamebraiderOracl: "Flamebraider",
		stdB66EntranceOracle:    "Impolite Entrance",
		stdB66ZealotOracle:      "Kulrath Zealot",
		stdB66MeekAttackOracle:  "Meek Attack",
		stdB66RansackingOracle:  "Reckless Ransacking",
		stdB66ChangelingOracle:  "Sizzling Changeling",
		stdB66SquawkOracle:      "Squawkroaster",
		stdB66TweezeOracle:      "Tweeze",
		stdB66DruidOracle:       "Great Forest Druid",
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
	if spec, _ := Lookup(stdB66InfernoOracle); spec.TapCost == nil {
		t.Error("Collective Inferno has no convoke tap cost")
	}
}

func TestStdB66CollectiveInfernoDoublesDamageFromSourcesOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	goblin := b12Creature(g, me.ID, "Goblin", "Creature — Goblin", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Collective Inferno", "Enchantment", stdB66InfernoOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")
	passPriorityAroundTable(t, g)
	life := opp.Life
	attackWith(t, g, opp.ID, goblin, bear)
	passPriorityAroundTable(t, g)
	// The Goblin deals 2 doubled to 4, the Bear 3 as printed.
	if got := life - opp.Life; got != 7 {
		t.Errorf("the opponent lost %d, want 7 (4 from the Goblin, 3 from the Bear)", got)
	}
}

func TestStdB66CollectiveInfernoIgnoresAnOpponentsSourceOfTheType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	castCatalogSpell(t, g, "Collective Inferno", "Enchantment", stdB66InfernoOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")
	passPriorityAroundTable(t, g)
	theirs := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 2, 2)
	_ = theirs
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 9, 9)
	g.WithWriteLock(func() {
		_ = (DealDamage{Source: theirs, Target: victim, Amount: 2}).Apply(NewContext(g, &game.StackItem{Controller: opp.ID, SourceCardID: theirs}))
	})
	if got := findBattlefieldCardByID(g, victim).DamageMarked; got != 2 {
		t.Errorf("their Goblin dealt %d, want 2 (not doubled)", got)
	}
}

func TestStdB66EndBlazeEpiphanyExilesCardsEqualToItsPowerAndLetsYouPlayOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 2)
	ids := seedLibrary(me, "One", "Two", "Three", "Four")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "End-Blaze Epiphany", TypeLine: "Instant", ManaCost: "{X}{R}",
		OracleID: stdB66EndBlazeOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i := 0; i < 6 && !stdB66HasChoose(g, me.ID); i++ {
		passPriorityAroundTable(t, g)
	}
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no choice among the exiled cards")
	}
	if len(c.ChooseCards) != 3 {
		t.Errorf("exiled %d cards, want the dead creature's power (3)", len(c.ChooseCards))
	}
	pick := c.ChooseCards[0]
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	_ = ids
	if !g.Exile.Contains(pick) {
		t.Error("the chosen card is not in exile")
	}
	if !g.CastPermissionOnCardByIDForEffect(pick).Granted() {
		t.Error("the chosen card got no play permission")
	}
}

func stdB66HasChoose(g *game.Game, p uuid.UUID) bool { return latestChooseCardsFor(g, p) != nil }

func TestStdB66FlamebraiderAddsTwoRestrictedMana(t *testing.T) {
	spec, _ := Lookup(stdB66FlamebraiderOracl)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != AnyCombinationOfColors(2) ||
		len(spec.ManaAbilities[0].Restrictions) != 1 {
		t.Errorf("Flamebraider's mana ability is %+v", spec.ManaAbilities)
	}
}

func TestStdB66GreatForestDruidAddsOneManaOfAnyColor(t *testing.T) {
	spec, _ := Lookup(stdB66DruidOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("Great Forest Druid's mana ability is %+v", spec.ManaAbilities)
	}
}

func TestStdB66ImpoliteEntranceGrantsTrampleAndHasteAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Impolite Entrance", "Sorcery", stdB66EntranceOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "trample") || !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the creature did not gain trample and haste")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (cast one, drew one)", me.Hand.Size(), hand+1)
	}
}

func TestStdB66KulrathZealotExilesTheTopCardForYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedLibrary(me, "Top")
	stdB63CastCreature(t, g, "Kulrath Zealot", "Creature — Elemental Warrior", "{5}{R}", stdB66ZealotOracle, []string{"R"}, 6, 5)
	top := ids[0]
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if !g.CastPermissionOnCardByIDForEffect(top).Granted() {
		t.Error("the exiled card has no play permission")
	}
	spec, _ := Lookup(stdB66ZealotOracle)
	if len(spec.Activated) != 1 {
		t.Errorf("want basic landcycling, got %d activated abilities", len(spec.Activated))
	}
}

func TestStdB66MeekAttackPutsASmallCreatureInWithHasteAndSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	attack := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Meek Attack", TypeLine: "Enchantment", OracleID: stdB66MeekAttackOracle,
		Owner: me.ID, Controller: me.ID,
	})
	small := rfdHandCard(me, "Small", "Creature — Goblin", "", "", 2, 3)
	big := rfdHandCard(me, "Big", "Creature — Giant", "", "", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, attack, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the creature card")
	}
	if hasID(pick.ChooseCards, big) || !hasID(pick.ChooseCards, small) {
		t.Errorf("offered %v: want only the creature with total power and toughness 5 or less", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Fatal("the creature did not enter")
	}
	if !hasEffectiveKeyword(t, g, small, "haste") {
		t.Error("the creature did not gain haste")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the creature was not sacrificed at the next end step")
	}
}

func TestStdB66RecklessRansackingPumpsAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Reckless Ransacking", "Instant", stdB66RansackingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 5 || tt != 4 {
		t.Errorf("the creature is %d/%d, want 5/4", p, tt)
	}
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
}

func TestStdB66SizzlingChangelingExilesTheTopCardWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ids := seedLibrary(me, "Top")
	ch := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sizzling Changeling", TypeLine: "Creature — Shapeshifter",
		OracleID: stdB66ChangelingOracle, Power: 3, Toughness: 2, Keywords: []string{game.KeywordChangeling},
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ch) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.CastPermissionOnCardByIDForEffect(ids[0]).Granted() {
		t.Error("the top card was not exiled with a play permission")
	}
}

func TestStdB66SquawkroasterPowerIsTheColorsAmongYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	sq := rfdCast(t, g, "Squawkroaster", "Creature — Elemental", stdB66SquawkOracle, 0, 4)
	if p, tt := ptOf(t, g, sq); p != 0 || tt != 4 {
		// Colorless in this harness (no colours on the card): 0/4.
		t.Errorf("a colorless Squawkroaster alone is %d/%d, want 0/4", p, tt)
	}
	b16Creature(g, me.ID, "Red", "Creature — Goblin", 1, 1, "R")
	b16Creature(g, me.ID, "Blue", "Creature — Merfolk", 1, 1, "U")
	b16Creature(g, opp.ID, "Their Green", "Creature — Elf", 1, 1, "G")
	if p, tt := ptOf(t, g, sq); p != 2 || tt != 4 {
		t.Errorf("with red and blue among your permanents it is %d/%d, want 2/4", p, tt)
	}
}

func TestStdB66TweezeDamagesThenMayRummage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	handBefore := me.Hand.Size()
	castCatalogSpell(t, g, "Tweeze", "Instant", stdB66TweezeOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-3)
	}
	// The rummage is optional: a prompt for up to one card.
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no optional discard prompt")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("discard bounds %d..%d, want 0..1", c.ChooseMin, c.ChooseMax)
	}
	pick := me.Hand.Cards[0].InstanceID
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d, want %d (cast one, discarded one, drew one)", me.Hand.Size(), handBefore)
	}
}

const stdB66AwakenerOracle = "cf9d13e2-bdcc-4c66-95ad-654ef795dbd3"

func TestStdB66AuroraAwakenerRevealsUntilXPermanentCardsAndPutsAnyNumber(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b16Creature(g, me.ID, "Red Guy", "Creature — Goblin", 1, 1, "R")
	// Library, top first: Bolt (instant), Forest, Bear, Rock, Deeper.
	deeper := seedLibraryTop(me, "Deeper", "Creature — Elf", "{1}{G}")
	rock := seedLibraryTop(me, "Rock", "Artifact", "{2}")
	bear := seedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
	forest := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	bolt := seedLibraryTop(me, "Bolt", "Instant", "{R}")
	stdB63CastCreature(t, g, "Aurora Awakener", "Creature — Giant Druid", "{6}{G}", stdB66AwakenerOracle, []string{"G"}, 7, 7)
	// Green (itself) and red: X = 2, so Bolt, Forest and Bear are revealed.
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no choice of permanent cards to put onto the battlefield")
	}
	if !hasID(c.ChooseCards, forest) || !hasID(c.ChooseCards, bear) || hasID(c.ChooseCards, bolt) || hasID(c.ChooseCards, rock) {
		t.Errorf("offered %v: want the Forest and the Bear only (the Rock was never revealed, the Bolt is not a permanent card)", c.ChooseCards)
	}
	if c.ChooseMin != 0 {
		t.Errorf("floor %d: \"any number\" includes none", c.ChooseMin)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Error("the chosen card did not enter")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("an unchosen card entered")
	}
	// The unchosen Bear and the Bolt are on the bottom; the Rock and Deeper stay on top.
	var bottomTwo []uuid.UUID
	g.WithWriteLock(func() {
		n := len(me.Library.Cards)
		bottomTwo = []uuid.UUID{me.Library.Cards[0].InstanceID, me.Library.Cards[1].InstanceID}
		_ = n
	})
	if !(hasID(bottomTwo, bear) && hasID(bottomTwo, bolt)) {
		t.Errorf("the bottom two cards are %v, want the Bear and the Bolt", bottomTwo)
	}
	if me.Library.Contains(deeper) == false || me.Library.Contains(rock) == false {
		t.Error("an unrevealed card left the library")
	}
}
