package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b15_cards_test.go — Standard set leftovers, batches 15 and 16
// (#4452). Each test plays the real catalog entry and asserts what its
// text says.

const (
	stdB15RiverchurnOracle  = "71c714a2-baa9-4082-bc8e-ceadbea231e0"
	stdB15CurseclothOracle  = "d9c4996a-35eb-4d38-9680-57d9d039ed82"
	stdB15CoalstokeOracle   = "2cdbaa4e-8907-49dc-9fa9-0df70db3bf23"
	stdB15KetramoseOracle   = "de41ad94-b9d7-4c2b-bfa6-394d18737228"
	stdB15ValkyrieOracle    = "27400927-9aba-4fb3-90c4-490d01969e60"
	stdB15FaebloomOracle    = "903d114b-1899-4e67-bee3-af0673850388"
	stdB15KaitoOracle       = "f6d3f145-00e0-47cd-91d8-0381c19a000d"
	stdB15VoyageOracle      = "1d6796ce-9316-430a-8126-8b0414cf181c"
	stdB15PlundererOracle   = "31265357-5ecb-4d83-80ff-1795b1dfdecb"
	stdB15VesselOracle      = "0e13c283-f462-441f-a8a5-0c75a59ff1d9"
	stdB15SnackOracle       = "3ddcaee9-a11e-477a-bd19-f5ee10512549"
	stdB15RatsOracle        = "29e3a3d9-4839-44d0-ad48-159c01645a6a"
	stdB15StrongboxOracle   = "ba5a2689-3578-4b1a-99d6-6b6d2356330e"
	stdB15CeratopsOracle    = "e56629ba-82d4-470f-9a6b-40c29d835af2"
	stdB15GreatwurmOracle   = "01635d21-aeda-4754-9d94-68fcd0cd6556"
	stdB15PandaOracle       = "b5d64271-4478-4c55-8c71-ba8f02b47fac"
	stdB15ArtistOracle      = "a3643f66-dfa2-4275-a093-23e162496354"
	stdB15FishingPoleOracle = "2c1d3e9e-213e-492a-ad5a-91139d70bc0f"
	stdB15KatanaOracle      = "604ef241-df8f-48f8-8680-f9ed77ff30ae"
	stdB15AmuletOracle      = "87424e20-00ab-4f6d-a905-1955897b358e"
)

func TestStdB15EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB15RiverchurnOracle:  "Riverchurn Monument",
		stdB15CurseclothOracle:  "Cursecloth Wrappings",
		stdB15CoalstokeOracle:   "Coalstoke Gearhulk",
		stdB15KetramoseOracle:   "Ketramose, the New Dawn",
		stdB15ValkyrieOracle:    "Valkyrie's Call",
		stdB15FaebloomOracle:    "Faebloom Trick",
		stdB15KaitoOracle:       "Kaito, Cunning Infiltrator",
		stdB15VoyageOracle:      "Uncharted Voyage",
		stdB15PlundererOracle:   "Gutless Plunderer",
		stdB15VesselOracle:      "Infernal Vessel",
		stdB15SnackOracle:       "Midnight Snack",
		stdB15RatsOracle:        "Revenge of the Rats",
		stdB15StrongboxOracle:   "Strongbox Raider",
		stdB15CeratopsOracle:    "Quakestrider Ceratops",
		stdB15PandaOracle:       "Fiendish Panda",
		stdB15ArtistOracle:      "Perforating Artist",
		stdB15FishingPoleOracle: "Fishing Pole",
		stdB15KatanaOracle:      "Quick-Draw Katana",
		stdB15AmuletOracle:      "Ravenous Amulet",
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

// stdB15Library puts n filler cards on top of p's library and returns the
// ids, top first.
func stdB15Library(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{InstanceID: ids[i], Name: names[i], TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	}
	return ids
}

// --- Quick-Draw Katana ------------------------------------------------------------

func TestStdB15QuickDrawKatanaArmsTheCreatureOnYourTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	katana := dftABPush(g, me.ID, "Quick-Draw Katana", "Artifact — Equipment", stdB15KatanaOracle, 0, 0)
	bear := dftABPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, katana, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, bear); p != 4 || th != 2 {
		t.Errorf("on your turn the bear is %d/%d, want 4/2", p, th)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("on your turn the bear has no first strike")
	}
	advanceToNextSeatsTurn(t, g)
	if p, th := effectivePTOf(t, g, bear); p != 2 || th != 2 {
		t.Errorf("on another turn the bear is %d/%d, want 2/2", p, th)
	}
	if hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("on another turn the bear has first strike")
	}
}

// --- Quakestrider Ceratops ----------------------------------------------------------

func TestStdB15QuakestriderCeratopsIsATwelveEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := dftABPush(g, me.ID, "Quakestrider Ceratops", "Creature — Dinosaur", stdB15CeratopsOracle, 12, 8)
	if p, th := effectivePTOf(t, g, id); p != 12 || th != 8 {
		t.Errorf("%d/%d, want 12/8", p, th)
	}
}

// --- Uncharted Voyage ---------------------------------------------------------------

func TestStdB15UnchartedVoyageTucksTheCreatureThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	stdB15Library(me, "Mine A", "Mine B")
	castCatalogSpell(t, g, "Uncharted Voyage", "Instant", stdB15VoyageOracle, cardTarget(victim))
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if surveilChoiceFor(g, me.ID) != nil {
		t.Fatal("the surveil came before the owner answered")
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{victim}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(victim) || !opp.Library.Contains(victim) {
		t.Fatal("the creature is not in its owner's library")
	}
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("the caster was not asked to surveil 1")
	}
}

// --- Revenge of the Rats ------------------------------------------------------------

func TestStdB15RevengeOfTheRatsMakesTappedRatsPerCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	pushCatalogGraveyardCard(me, "Dead Elk", "Creature — Elk", "", 2, 2)
	pushCatalogGraveyardCard(me, "Dead Forest", "Land", "", 0, 0)
	castCatalogSpell(t, g, "Revenge of the Rats", "Sorcery", stdB15RatsOracle, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Rat" && c.Controller == me.ID {
			n++
			if !c.Tapped {
				t.Error("a Rat entered untapped")
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Rats, want 2", n)
	}
}

// --- Faebloom Trick -------------------------------------------------------------------

func TestStdB15FaebloomTrickMakesTwoFaeriesAndTapsAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Faebloom Trick", "Instant", stdB15FaebloomOracle, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if len(p.PickTargetCards) == 0 || slices.Contains(p.PickTargetCards, mine) {
			t.Errorf("offered %v, my own creature must not be a target", p.PickTargetCards)
		}
		pickTriggerTarget(t, g, me.ID, theirs)
		passPriorityAroundTable(t, g)
	}
	if got := countOnBattlefield(g, "Faerie", me.ID); got != 2 {
		t.Errorf("%d Faeries, want 2", got)
	}
	if c, _ := battlefieldCard(g, theirs); !c.Tapped {
		t.Error("the opposing creature is untapped")
	}
	if c, _ := battlefieldCard(g, mine); c.Tapped {
		t.Error("my creature was tapped")
	}
}

// --- Infernal Vessel ------------------------------------------------------------------

func TestStdB15InfernalVesselReturnsOnceAsADemonWithTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vessel := dftABPush(g, me.ID, "Infernal Vessel", "Creature — Human Cleric", stdB15VesselOracle, 2, 1)
	killCreature(t, g, me.ID, vessel)
	passPriorityAroundTable(t, g)
	bf := findBattlefieldByName(g, "Infernal Vessel")
	if bf == uuid.Nil {
		t.Fatal("the Vessel did not return")
	}
	card := ringBFCard(t, g, bf)
	if !card.HasSubtype("Demon") || !card.HasSubtype("Cleric") {
		t.Errorf("returned types %q, want Demon in addition to Human Cleric", card.TypeLine)
	}
	if got := countersOn(g, bf, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if p, th := card.CurrentPower(), card.CurrentToughness(); p != 4 || th != 3 {
		t.Errorf("returned as %d/%d, want 4/3", p, th)
	}
	killCreature(t, g, me.ID, bf)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Infernal Vessel") != uuid.Nil {
		t.Error("a Vessel that was a Demon returned again")
	}
}

// --- Valkyrie's Call ------------------------------------------------------------------

func TestStdB15ValkyriesCallReturnsANontokenNonAngelAsAFlyingAngel(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dftABPush(g, me.ID, "Valkyrie's Call", "Enchantment", stdB15ValkyrieOracle, 0, 0)
	bear := dftABPush(g, me.ID, "Plain Bear", "Creature — Bear", "", 2, 2)
	angel := dftABPush(g, me.ID, "Plain Angel", "Creature — Angel", "", 2, 2)
	theirs := dftABPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	token := dftABPush(g, me.ID, "Token Bear", "Token Creature — Bear", "", 2, 2)
	for _, id := range []uuid.UUID{angel, theirs, token} {
		killCreature(t, g, g.Seats[0].ID, id)
	}
	passPriorityAroundTable(t, g)
	for _, name := range []string{"Plain Angel", "Their Bear", "Token Bear"} {
		if findBattlefieldByName(g, name) != uuid.Nil {
			t.Errorf("%s returned", name)
		}
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	bf := findBattlefieldByName(g, "Plain Bear")
	if bf == uuid.Nil {
		t.Fatal("the Bear did not return")
	}
	card := ringBFCard(t, g, bf)
	if !card.HasSubtype("Angel") || !card.HasSubtype("Bear") {
		t.Errorf("returned types %q, want Angel in addition to Bear", card.TypeLine)
	}
	if !hasEffectiveKeyword(t, g, bf, "flying") {
		t.Error("the returned creature has no flying")
	}
	if got := countersOn(g, bf, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	killCreature(t, g, me.ID, bf)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Plain Bear") != uuid.Nil {
		t.Error("the Bear, now an Angel, returned again")
	}
}

// --- Raid: Gutless Plunderer, Strongbox Raider, Perforating Artist, Midnight Snack ----

func TestStdB15GutlessPlundererKeepsOneOnTopAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := stdB15Library(me, "A", "B", "C", "D")
	// no attack yet: nothing is looked at
	avB16CastCreature(t, g, "Gutless Plunderer", "Creature — Skeleton Pirate", stdB15PlundererOracle, 2, 2)
	if chooseCardsChoiceFor(g, me.ID) != nil || me.Graveyard.Size() != 0 {
		t.Fatal("the raid trigger fired without an attack this turn")
	}
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	avB16CastCreature(t, g, "Gutless Plunderer", "Creature — Skeleton Pirate", stdB15PlundererOracle, 2, 2)
	p := chooseCardsChoiceFor(g, me.ID)
	if p == nil {
		t.Fatal("no look prompt")
	}
	if len(p.ChooseCards) != 3 || p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Fatalf("prompt %v %d..%d, want the top three and 0..1", p.ChooseCards, p.ChooseMin, p.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, ids[1])
	if stdB15Top(me) != ids[1] {
		t.Errorf("the kept card is not on top of the library")
	}
	if me.Graveyard.Size() != 2 {
		t.Errorf("%d cards in the graveyard, want the other two", me.Graveyard.Size())
	}
}

func TestStdB15StrongboxRaiderChoosesOneOfTwoToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := stdB15Library(me, "A", "B", "C")
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	avB16CastCreature(t, g, "Strongbox Raider", "Creature — Orc Pirate", stdB15StrongboxOracle, 5, 2)
	p := chooseCardsChoiceFor(g, me.ID)
	if p == nil {
		t.Fatal("no choice among the exiled cards")
	}
	if len(p.ChooseCards) != 2 {
		t.Fatalf("offered %d cards, want the two exiled", len(p.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, ids[0])
	for i, id := range ids[:2] {
		if z := zoneOf(g, id); z != game.ZoneExile {
			t.Errorf("card %d is in %v, want exile", i, z)
		}
	}
	if stdB15Top(me) != ids[2] {
		t.Error("a third card left the library")
	}
}

func TestStdB15PerforatingArtistPunishesEachOpponentAfterAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dftABPush(g, me.ID, "Perforating Artist", "Creature — Devil", stdB15ArtistOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	pick := latestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent was not asked: %+v", g.PendingChoices)
	}
	life := opp.Life
	blbSTPickOption(t, g, opp.ID, "Lose 3 life")
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 lost", life, opp.Life)
	}
}

func TestStdB15MidnightSnackMakesFoodAfterAnAttackAndDrainsByLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snack := dftABPush(g, me.ID, "Midnight Snack", "Enchantment", stdB15SnackOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(snack, me.ID, 4) })
	floatMana(t, g, me, "{C}{C}{B}")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, snack, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{playerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want 4 lost", life, opp.Life)
	}
	if g.Battlefield.Contains(snack) {
		t.Error("the enchantment was not sacrificed")
	}
	_ = bear
}

func TestStdB15MidnightSnackFoodOnlyAfterAnAttack(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dftABPush(g, me.ID, "Midnight Snack", "Enchantment", stdB15SnackOracle, 0, 0)
		bear := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		if attacked {
			declareAttack(t, g, opp.ID, bear)
			passPriorityAroundTable(t, g)
		}
		advanceToEndStep(t, g)
		passPriorityAroundTable(t, g)
		want := 0
		if attacked {
			want = 1
		}
		if got := countOnBattlefield(g, "Food", me.ID); got != want {
			t.Errorf("attacked=%v: %d Food, want %d", attacked, got, want)
		}
	}
}

// stdB15Top is the id of the top card of p's library (the last element),
// or uuid.Nil when it is empty.
func stdB15Top(p *game.Player) uuid.UUID {
	c, err := p.Library.Top()
	if err != nil {
		return uuid.Nil
	}
	return c.InstanceID
}

// --- Kaito, Cunning Infiltrator -------------------------------------------------

func TestStdB15KaitoLoyaltyAbilitiesAndTheNinjaEmblem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stdB15Library(me, "A", "B", "C")
	kaito := pushWalkerForTest(g, me.ID, "Kaito, Cunning Infiltrator", stdB15KaitoOracle, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, kaito, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the target can be blocked")
	}
	if loyaltyOf(g, kaito) != 4 {
		t.Errorf("loyalty %d after +1, want 4", loyaltyOf(g, kaito))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the loot's draw", hand, me.Hand.Size())
	}
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard after the draw")
	}
	discardFromHand(t, g, me.ID)

	b39NextTurnOf(t, g, 0)
	b16Activate(t, g, me.ID, kaito, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Ninja", me.ID); got != 1 {
		t.Errorf("%d Ninjas after the -2, want 1", got)
	}
	if p, th := effectivePTOf(t, g, findBattlefieldByName(g, "Ninja")); p != 2 || th != 1 {
		t.Errorf("the Ninja is %d/%d, want 2/1", p, th)
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(kaito, game.CounterLoyalty, 9) })
	b39NextTurnOf(t, g, 0)
	b16Activate(t, g, me.ID, kaito, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if len(emblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems %v, want one", emblemsOf(g, me.ID))
	}
	// Any player's spell makes the emblem's owner a Ninja.
	before := countOnBattlefield(g, "Ninja", me.ID)
	castCatalogSpell(t, g, "Uncharted Voyage", "Instant", stdB15VoyageOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Ninja", me.ID); got != before+1 {
		t.Errorf("Ninjas %d -> %d after a spell was cast, want one more", before, got)
	}
}

func TestStdB15KaitoGainsLoyaltyWhenACreatureYouControlConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaito := pushWalkerForTest(g, me.ID, "Kaito, Cunning Infiltrator", stdB15KaitoOracle, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if triggerOrderPrompt(g) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
		passPriorityAroundTable(t, g)
	}
	if got := loyaltyOf(g, kaito); got != 5 {
		t.Errorf("loyalty %d, want 5: one counter for each creature that connected", got)
	}
}

// --- Riverchurn Monument --------------------------------------------------------

func TestStdB15RiverchurnMonumentMillsEachTargetAndExhaustsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monument := dftABPush(g, me.ID, "Riverchurn Monument", "Artifact", stdB15RiverchurnOracle, 0, 0)
	stdB15Library(me, "A", "B", "C", "D", "E", "F")
	stdB15Library(opp, "G", "H", "I", "J", "K", "L")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, monument, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{playerTarget(me.ID), playerTarget(opp.ID)}})
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 2 || opp.Graveyard.Size() != 2 {
		t.Fatalf("graveyards %d and %d, want two milled each", me.Graveyard.Size(), opp.Graveyard.Size())
	}
	g.WithWriteLock(func() { c := battlefieldCardPtr(t, g, monument); c.Tapped = false })
	pushCatalogGraveyardCard(opp, "Extra", "Creature — Bear", "", 1, 1)
	floatMana(t, g, me, "{C}{C}{U}{U}")
	b16Activate(t, g, me.ID, monument, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{playerTarget(me.ID), playerTarget(opp.ID)}})
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 4 || opp.Graveyard.Size() != 6 {
		t.Errorf("graveyards %d and %d, want each to mill as many as they held (4 and 6)", me.Graveyard.Size(), opp.Graveyard.Size())
	}
	g.WithWriteLock(func() { c := battlefieldCardPtr(t, g, monument); c.Tapped = false })
	floatMana(t, g, me, "{C}{C}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, monument, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{playerTarget(me.ID)}}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Fiendish Panda -------------------------------------------------------------

func TestStdB15FiendishPandaGrowsOnLifeGainAndReturnsABoundedNonBear(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	panda := dftABPush(g, me.ID, "Fiendish Panda", "Creature — Bear Demon", stdB15PandaOracle, 3, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(panda, me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, panda, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after gaining life, want 1", got)
	}
	big := me.Graveyard
	push := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		big.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		return id
	}
	good := push("Four Drop", "Creature — Elf", "{2}{G}{G}")
	push("Two Drop", "Creature — Elf", "{1}{G}")
	tooBig := push("Five Drop", "Creature — Giant", "{3}{G}{G}")
	aBear := push("Little Bear", "Creature — Bear", "{1}")
	killCreature(t, g, me.ID, panda)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target for the dies trigger")
	}
	for _, bad := range []uuid.UUID{tooBig, aBear, panda} {
		if slices.Contains(p.PickTargetCards, bad) {
			t.Errorf("%v offered, but the Panda's power is 4 and no Bear may return", bad)
		}
	}
	pickTriggerTarget(t, g, me.ID, good)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(good) {
		t.Error("the Elf did not return")
	}
}

// --- Ravenous Amulet ------------------------------------------------------------

func TestStdB15RavenousAmuletDrawsCountsSoulsAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amulet := dftABPush(g, me.ID, "Ravenous Amulet", "Artifact", stdB15AmuletOracle, 0, 0)
	a := b12Creature(g, me.ID, "Fodder A", "Creature — Bear", 1, 1)
	stdB15Library(me, "A", "B", "C")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a}})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn", hand, me.Hand.Size())
	}
	if got := countersOn(g, amulet, "soul"); got != 1 {
		t.Errorf("%d soul counters, want 1", got)
	}
	if g.Battlefield.Contains(a) {
		t.Error("the creature was not sacrificed")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(amulet, "soul", 2); battlefieldCardPtr(t, g, amulet).Tapped = false })
	floatMana(t, g, me, "{C}{C}{C}{C}")
	life := opp.Life
	b16Activate(t, g, me.ID, amulet, 1, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 lost (one per soul counter)", life, opp.Life)
	}
	if g.Battlefield.Contains(amulet) {
		t.Error("the Amulet was not sacrificed")
	}
}

// --- Coalstoke Gearhulk ---------------------------------------------------------

func TestStdB15CoalstokeGearhulkReanimatesWithFinalityAndExilesAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Their Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	big := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Their Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{G}{G}",
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	avB16CastCreature(t, g, "Coalstoke Gearhulk", "Artifact Creature — Construct", stdB15CoalstokeOracle, 5, 4)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target for the enters trigger")
	}
	if slices.Contains(p.PickTargetCards, big) {
		t.Error("a creature card with mana value 5 was offered")
	}
	pickTriggerTarget(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, small)
	if !ok || card.Controller != me.ID {
		t.Fatalf("the Elf is not on the battlefield under my control: %+v", card)
	}
	if countersOn(g, small, game.CounterFinality) != 1 {
		t.Error("no finality counter")
	}
	for _, kw := range []string{"menace", "deathtouch", "haste"} {
		if !hasEffectiveKeyword(t, g, small, kw) {
			t.Errorf("the Elf lacks %s", kw)
		}
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Exile.Contains(small) {
		t.Error("the Elf was not exiled at the beginning of my next end step")
	}
}

// --- Ketramose, the New Dawn ----------------------------------------------------

func TestStdB15KetramoseCantAttackOrBlockUntilSevenCardsAreInExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ket := dftABPush(g, me.ID, "Ketramose, the New Dawn", "Legendary Creature — God", stdB15KetramoseOracle, 4, 4)
	exile := func(n int) {
		ids := make([]uuid.UUID, n)
		for i := range ids {
			ids[i] = pushCatalogGraveyardCard(me, "Exiled", "Sorcery", "", 0, 0)
		}
		g.WithWriteLock(func() { g.ExileCardsForEffect(ids) })
	}
	exile(6)
	g.RunStateChecksForTest()
	if r := restrictionsOf(t, g, ket); r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Errorf("restrictions %v with six cards in exile, want can't attack or block", r)
	}
	exile(1)
	g.RunStateChecksForTest()
	if r := restrictionsOf(t, g, ket); r&(game.CantAttack|game.CantBlock) != 0 {
		t.Errorf("restrictions %v with seven cards in exile, want none", r)
	}
	for _, kw := range []string{"menace", "lifelink", "indestructible"} {
		if !hasEffectiveKeyword(t, g, ket, kw) {
			t.Errorf("Ketramose lacks %s", kw)
		}
	}
}

func TestStdB15KetramoseDrawsOncePerBatchOnYourTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dftABPush(g, me.ID, "Ketramose, the New Dawn", "Legendary Creature — God", stdB15KetramoseOracle, 4, 4)
	stdB15Library(me, "A", "B", "C", "D")
	a := pushCatalogGraveyardCard(opp, "Their Dead A", "Creature — Bear", "", 2, 2)
	b := pushCatalogGraveyardCard(me, "My Dead B", "Creature — Bear", "", 2, 2)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() {
		g.ExileCardsForEffect([]uuid.UUID{a, b, victim})
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d, want one card and one life for the whole batch", hand, me.Hand.Size(), life, me.Life)
	}
	// An exile on another player's turn does nothing.
	c := pushCatalogGraveyardCard(opp, "Their Dead C", "Creature — Bear", "", 2, 2)
	advanceToNextSeatsTurn(t, g)
	hand, life = me.Hand.Size(), me.Life
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{c}) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("an exile on another turn drew: hand %d -> %d, life %d -> %d", hand, me.Hand.Size(), life, me.Life)
	}
}

// --- Fishing Pole ---------------------------------------------------------------

func TestStdB15FishingPoleBaitsThenMakesAFishWhenTheCreatureUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pole := dftABPush(g, me.ID, "Fishing Pole", "Artifact — Equipment", stdB15FishingPoleOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Angler", "Creature — Human", 2, 2)
	other := dftABPush(g, me.ID, "Fishing Pole", "Artifact — Equipment", stdB15FishingPoleOracle, 0, 0)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, pole, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, bear, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, pole, "bait"); got != 1 {
		t.Fatalf("%d bait counters on the Pole, want 1", got)
	}
	if got := countersOn(g, other, "bait"); got != 0 {
		t.Errorf("%d bait counters on the other Pole, want 0", got)
	}
	if c, _ := battlefieldCard(g, pole); !c.Tapped {
		t.Error("the Pole was not tapped as a cost")
	}
	if c, _ := battlefieldCard(g, bear); !c.Tapped {
		t.Error("the creature was not tapped as a cost")
	}
	// With the Pole tapped the granted ability can't be paid again.
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, bear).Tapped = false })
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the granted ability was activated with Fishing Pole tapped")
	}
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, bear).Tapped = true })
	b39NextTurnOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Fish", me.ID); got != 1 {
		t.Errorf("%d Fish after the creature untapped, want 1", got)
	}
	if got := countersOn(g, pole, "bait"); got != 0 {
		t.Errorf("%d bait counters left, want 0", got)
	}
	// No counter, no Fish.
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, bear).Tapped = true })
	b39NextTurnOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Fish", me.ID); got != 1 {
		t.Errorf("%d Fish with no bait counter, want still 1", got)
	}
}

// --- Cursecloth Wrappings -------------------------------------------------------

func TestStdB15CurseclothWrappingsPumpsZombiesAndGrantsEmbalmForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wrappings := dftABPush(g, me.ID, "Cursecloth Wrappings", "Artifact", stdB15CurseclothOracle, 0, 0)
	zombie := dftABPush(g, me.ID, "Plain Zombie", "Creature — Zombie", "", 2, 2)
	bear := dftABPush(g, me.ID, "Plain Bear", "Creature — Bear", "", 2, 2)
	if p, th := effectivePTOf(t, g, zombie); p != 3 || th != 3 {
		t.Errorf("Zombie %d/%d, want 3/3", p, th)
	}
	if p, th := effectivePTOf(t, g, bear); p != 2 || th != 2 {
		t.Errorf("Bear %d/%d, want 2/2", p, th)
	}
	advanceToMain(t, g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Knight", TypeLine: "Creature — Human Knight", ManaCost: "{1}{W}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	// Without the grant the card has no ability.
	floatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a plain graveyard card was embalmed")
	}
	b16Activate(t, g, me.ID, wrappings, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(dead)})
	passPriorityAroundTable(t, g)
	// Too little mana: the cost is its {1}{W}.
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("embalm was activated for {C} when its cost is {1}{W}")
	}
	floatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("embalm for its mana cost: %v", err)
	}
	if !g.Exile.Contains(dead) {
		t.Fatal("the embalmed card is not in exile")
	}
	passPriorityAroundTable(t, g)
	tok := findBattlefieldByName(g, "Dead Knight")
	if tok == uuid.Nil {
		t.Fatal("no token")
	}
	card := ringBFCard(t, g, tok)
	if !card.HasSubtype("Zombie") || !card.HasSubtype("Knight") || card.ManaCost != "" {
		t.Errorf("token %q cost %q, want a Zombie Knight with no mana cost", card.TypeLine, card.ManaCost)
	}
}

func TestStdB15CurseclothWrappingsGrantEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wrappings := dftABPush(g, me.ID, "Cursecloth Wrappings", "Artifact", stdB15CurseclothOracle, 0, 0)
	advanceToMain(t, g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Knight", TypeLine: "Creature — Human Knight", ManaCost: "{1}{W}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	b16Activate(t, g, me.ID, wrappings, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(dead)})
	passPriorityAroundTable(t, g)
	b39NextTurnOf(t, g, 0)
	floatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the granted embalm outlived the turn")
	}
}
