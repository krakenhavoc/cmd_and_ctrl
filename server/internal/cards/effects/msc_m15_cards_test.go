package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m15_cards_test.go — Marvel Super Heroes Commander, M–Z batch 15
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	mscM15SpiderManOracle  = "4604c8fb-d38f-42aa-935a-eada20f6aafa"
	mscM15WeaponsOracle    = "760d6cb2-5d8a-495c-9853-f96a1efa5775"
	mscM15SpearOracle      = "e7a0b370-913d-4e96-8fb4-5df168405073"
	mscM15VisionOracle     = "33219792-f0a7-4849-bfca-70ee2d410b9e"
	mscM15TitanOracle      = "f62250c6-0832-46c8-a800-806589504f5f"
	mscM15TriSentinelOracl = "d9cc9fc9-522e-4598-9a09-c95824d77acd"
	mscM15WitnessOracle    = "8ff7c8af-de9a-45f8-a9ba-98209a746891"
	mscM15VictorOracle     = "2ad77eb7-9466-432f-87b4-e80e4b66e143"
	mscM15TChallaOracle    = "ae10eb15-28e7-46d1-9a2e-85763d93febf"
	mscM15MoundOracle      = "6b78e417-44f8-4ab6-9d3c-e71704fc648e"
	mscM15HeroType         = "Legendary Creature — Human Hero"
)

func TestMscM15EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM15SpiderManOracle:  "Spider-Man, New Champion",
		mscM15WeaponsOracle:    "The Immortal Weapons",
		mscM15SpearOracle:      "The Spear of Bashenga",
		mscM15VisionOracle:     "The Vision and Scarlet Witch",
		mscM15TitanOracle:      "Titan of Littjara",
		mscM15TriSentinelOracl: "Tri-Sentinel, Act of Vengeance",
		mscM15WitnessOracle:    "Unlucky Witness",
		mscM15VictorOracle:     "Victor Timely, Wily Tycoon",
		mscM15TChallaOracle:    "T'Challa, the Black Panther",
		mscM15MoundOracle:      "The Great Mound",
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

// mscM15TriggersFrom counts the triggers of `source` that exist: on the
// stack, or waiting on a target pick.
func mscM15TriggersFrom(g *game.Game, source uuid.UUID) int {
	n := triggersOnStackFrom(g, source)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Source == source {
			n++
		}
	}
	return n
}

// --- Spider-Man, New Champion ---------------------------------------------------

func TestMscM15SpiderManDealsThatMuchDamageOnceForATwoCardDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := pushCatalogPermanent(g, me.ID, "Spider-Man, New Champion", "Legendary Creature — Spider Human Hero", mscM15SpiderManOracle, false)
	discarded := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	before := opp.Life

	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 2) })
	answerDiscard(t, g, me.ID, discarded...)
	pickPlayer(t, g, me.ID, opp.ID)
	if got := mscM15TriggersFrom(g, spider); got != 1 {
		t.Fatalf("a two-card discard fires the trigger %d times, want once", got)
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("damage dealt %d, want 2 (one per card discarded)", got)
	}
}

func TestMscM15SpiderManCanHitACreatureForOneCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Spider-Man, New Champion", "Legendary Creature — Spider Human Hero", mscM15SpiderManOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 1)

	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	discardFromHand(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, theirs) {
		t.Error("one damage should have killed the 2/1")
	}
}

func TestMscM15SpiderManIgnoresAnOpponentsDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := pushCatalogPermanent(g, me.ID, "Spider-Man, New Champion", "Legendary Creature — Spider Human Hero", mscM15SpiderManOracle, false)
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(opp.ID, 1) })
	discardFromHand(t, g, opp.ID)
	if got := mscM15TriggersFrom(g, spider); got != 0 {
		t.Errorf("an opponent's discard fired Spider-Man %d times", got)
	}
}

// --- The Immortal Weapons ----------------------------------------------------------

func TestMscM15ImmortalWeaponsReturnsAnInstantOrSorceryWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sorcery := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Old Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	creature := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})

	castCatalogSpell(t, g, "The Immortal Weapons", "Legendary Creature — Human Warrior Hero", mscM15WeaponsOracle, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetCards, creature) || !hasID(p.PickTargetCards, sorcery) {
		t.Fatalf("the pick should offer only the sorcery: %+v", p)
	}
	pickCard(t, g, me.ID, sorcery)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(sorcery) || !me.Hand.Contains(sorcery) {
		t.Error("the sorcery did not return to hand")
	}
	if !me.Graveyard.Contains(creature) {
		t.Error("the creature card left the graveyard")
	}
}

func TestMscM15ImmortalWeaponsPumpsAndGivesMenaceOnANoncreatureCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	weapons := pushCatalogPermanent(g, me.ID, "The Immortal Weapons", "Legendary Creature — Human Warrior Hero", mscM15WeaponsOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Doom Blade", "Instant", "", nil)
	mscM13PassUntilPickTarget(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("Bear power %d, want 4", got)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "menace") {
		t.Error("the Bear did not gain menace")
	}
	if hasKeywordOnBattlefield(t, g, weapons, "menace") {
		t.Error("the Weapons themselves gained menace")
	}
	powerUpPassTo(t, g, 1)
	if got := effectivePower(t, g, bear); got != 2 || hasKeywordOnBattlefield(t, g, bear, "menace") {
		t.Errorf("the boost should end with the turn: power %d", got)
	}
}

func TestMscM15ImmortalWeaponsIgnoresACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	weapons := pushCatalogPermanent(g, me.ID, "The Immortal Weapons", "Legendary Creature — Human Warrior Hero", mscM15WeaponsOracle, false)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	if got := mscM15TriggersFrom(g, weapons); got != 0 {
		t.Errorf("a creature spell fired the Weapons %d times", got)
	}
}

// --- The Vision and Scarlet Witch -------------------------------------------------

func TestMscM15VisionAddsRedAndGrowsOnEachSpellYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vision := pushCatalogPermanent(g, me.ID, "The Vision and Scarlet Witch", "Legendary Artifact Creature — Mutant Hero", mscM15VisionOracle, false)

	castCatalogSpell(t, g, "Doom Blade", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, vision, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("mana pool %d, want 1 ({R})", got)
	}
	castCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, vision, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d after a creature spell, want 2", got)
	}
}

func TestMscM15VisionIgnoresAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	vision := pushCatalogPermanent(g, g.Seats[0].ID, "The Vision and Scarlet Witch", "Legendary Artifact Creature — Mutant Hero", mscM15VisionOracle, false)
	powerUpPassTo(t, g, 1)
	castCatalogSpell(t, g, "Doom Blade", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, vision, game.CounterPlusOne); got != 0 || opp.ManaPool.Total() != 0 {
		t.Errorf("an opponent's spell: counters %d, their pool %d", got, opp.ManaPool.Total())
	}
}

// --- T'Challa, the Black Panther ---------------------------------------------------

func TestMscM15TChallaMakesATappedVibraniumWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "T'Challa, the Black Panther", "Legendary Creature — Human Noble Hero", mscM15TChallaOracle, nil)
	passPriorityAroundTable(t, g)
	if n, tapped := countTokensNamed(g, me.ID, "Vibranium"); n != 1 || tapped != 1 {
		t.Fatalf("after entering: %d Vibranium, %d tapped, want 1 and 1", n, tapped)
	}
}

func TestMscM15TChallaMakesATappedVibraniumWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tchalla := pushCatalogPermanent(g, me.ID, "T'Challa, the Black Panther", "Legendary Creature — Human Noble Hero", mscM15TChallaOracle, false)
	declareAttack(t, g, opp.ID, tchalla)
	passPriorityAroundTable(t, g)
	if n, tapped := countTokensNamed(g, me.ID, "Vibranium"); n != 1 || tapped != 1 {
		t.Errorf("after attacking: %d Vibranium, %d tapped, want 1 and 1", n, tapped)
	}
}

func TestMscM15TChallaGrowsOnlyForABigArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tchalla := pushCatalogPermanent(g, me.ID, "T'Challa, the Black Panther", "Legendary Creature — Human Noble Hero", mscM15TChallaOracle, false)
	cast := func(name, typeLine, cost string) {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		passPriorityAroundTable(t, g)
	}
	cast("Cheap Trinket", "Artifact", "{3}")
	cast("Big Creature", "Creature — Bear", "{5}")
	if got := counterCount(g, tchalla, game.CounterPlusOne); got != 0 {
		t.Fatalf("a cheap artifact or a nonartifact grew T'Challa: %d counters", got)
	}
	cast("Big Trinket", "Artifact", "{4}")
	if got := counterCount(g, tchalla, game.CounterPlusOne); got != 2 {
		t.Errorf("a four-mana artifact gave %d counters, want 2", got)
	}
}

// --- The Great Mound ---------------------------------------------------------------

func TestMscM15GreatMoundMakesATappedVibraniumForThreeAndDrawsForSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mound := pushCatalogPermanent(g, me.ID, "The Great Mound", "Land", mscM15MoundOracle, false)

	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, mound, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{3},{T}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, tapped := countTokensNamed(g, me.ID, "Vibranium"); n != 1 || tapped != 1 {
		t.Errorf("%d Vibranium, %d tapped, want 1 and 1", n, tapped)
	}
	if err := g.ActivateCatalogAbility(me.ID, mound, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the tapped land activated the draw ability")
	}

	powerUpPassTo(t, g, 0)
	floatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, mound, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("five mana paid for a {6} ability")
	}
	floatForTest(g, me, "C")
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, mound, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{6},{T}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- The Spear of Bashenga ---------------------------------------------------------

func TestMscM15SpearCrownsYouOnlyWhenThereIsNoMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "The Spear of Bashenga", "Legendary Artifact — Equipment", mscM15SpearOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch %v, want the caster", g.Monarch)
	}

	g2 := newCatalogGame(t)
	g2.WithWriteLock(func() { _ = g2.SetMonarchForEffect(g2.Seats[1].ID) })
	castCatalogSpell(t, g2, "The Spear of Bashenga", "Legendary Artifact — Equipment", mscM15SpearOracle, nil)
	passPriorityAroundTable(t, g2)
	if g2.Monarch != g2.Seats[1].ID {
		t.Errorf("monarch %v, want the opponent to keep the crown", g2.Monarch)
	}
	_ = opp
}

func TestMscM15SpearGivesPlusTwoPlusTwoAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	spear := pushCatalogPermanent(g, me.ID, "The Spear of Bashenga", "Legendary Artifact — Equipment", mscM15SpearOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	equipTo(t, g, me.ID, spear, bear)
	if card := mshB2PT(t, g, bear); card.Power != 4 || card.Toughness != 4 {
		t.Errorf("equipped Bear = %d/%d, want 4/4", card.Power, card.Toughness)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Error("the equipped creature lacks vigilance")
	}
}

func TestMscM15SpearDestroysATappedNonlandPermanentOfTheMonarchItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	spear := pushCatalogPermanent(g, me.ID, "The Spear of Bashenga", "Legendary Artifact — Equipment", mscM15SpearOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tappedRock := b12Creature(g, opp.ID, "Tapped Rock", "Artifact", 0, 0)
	untappedRock := b12Creature(g, opp.ID, "Untapped Rock", "Artifact", 0, 0)
	tappedLand := b12Creature(g, opp.ID, "Tapped Land", "Land", 0, 0)
	thirdsRock := b12Creature(g, third.ID, "Third Rock", "Artifact", 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case tappedRock, tappedLand, thirdsRock:
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		_ = g.SetMonarchForEffect(opp.ID)
	})
	equipTo(t, g, me.ID, spear, bear)

	declareAttack(t, g, opp.ID, bear)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the Spear's trigger")
	}
	if len(p.PickTargetCards) != 1 || p.PickTargetCards[0] != tappedRock {
		t.Fatalf("offered %v, want only the monarch's tapped nonland permanent", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, tappedRock)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, tappedRock) {
		t.Error("the tapped permanent survived")
	}
	for name, id := range map[string]uuid.UUID{"untapped": untappedRock, "land": tappedLand, "third player's": thirdsRock} {
		if !onBattlefield(g, id) {
			t.Errorf("the %s permanent was destroyed", name)
		}
	}
}

func TestMscM15SpearDoesNothingWhenTheDefenderIsNotTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	spear := pushCatalogPermanent(g, me.ID, "The Spear of Bashenga", "Legendary Artifact — Equipment", mscM15SpearOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Creature(g, opp.ID, "Tapped Rock", "Artifact", 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		_ = g.SetMonarchForEffect(third.ID)
	})
	equipTo(t, g, me.ID, spear, bear)
	declareAttack(t, g, opp.ID, bear)
	if got := mscM15TriggersFrom(g, spear); got != 0 {
		t.Errorf("attacking a non-monarch fired the Spear %d times", got)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, rock) {
		t.Error("a permanent was destroyed though the defender was not the monarch")
	}
}

// --- Tri-Sentinel, Act of Vengeance ------------------------------------------------

func TestMscM15TriSentinelDealsThreeToOneCreatureOfEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := make([]uuid.UUID, 0, 3)
	for _, opp := range g.Seats[1:] {
		small = append(small, b12Creature(g, opp.ID, "Spare "+opp.Name, "Creature — Bear", 2, 2))
		b12Creature(g, opp.ID, "Big "+opp.Name, "Creature — Bear", 2, 5)
	}
	castCatalogSpell(t, g, "Tri-Sentinel, Act of Vengeance", "Legendary Artifact Creature — Robot Villain", mscM15TriSentinelOracl, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if latestPickTarget(g, me.ID) == nil {
			t.Fatalf("clause %d: no target prompt", i)
		}
		pickCard(t, g, me.ID, small[i])
	}
	passPriorityAroundTable(t, g)
	for i, id := range small {
		if onBattlefield(g, id) {
			t.Errorf("opponent %d's chosen creature survived 3 damage", i+1)
		}
	}
	for _, opp := range g.Seats[1:] {
		if n := battlefieldNamed(g, opp.ID, "Big "+opp.Name); n != 1 {
			t.Errorf("%s's unchosen creature is gone", opp.Name)
		}
	}
}

func TestMscM15TriSentinelCanBeUnearthedForSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Tri-Sentinel, Act of Vengeance", OracleID: mscM15TriSentinelOracl,
		TypeLine: "Legendary Artifact Creature — Robot Villain", ManaCost: "{7}", Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	floatForTest(g, me, "CCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("unearth with six mana")
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth {7}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, id) {
		t.Fatal("the unearthed Tri-Sentinel is not on the battlefield")
	}
	if !hasKeywordOnBattlefield(t, g, id, "haste") {
		t.Error("the unearthed Tri-Sentinel lacks haste")
	}
}

// --- Unlucky Witness ---------------------------------------------------------------

func TestMscM15UnluckyWitnessExilesTwoAndLetsYouPlayOnlyOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	witness := pushCatalogPermanent(g, me.ID, "Unlucky Witness", "Creature — Human Citizen", mscM15WitnessOracle, false)
	top := []uuid.UUID{uuid.New(), uuid.New()}
	for i, id := range top {
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Exiled Land " + string(rune('A'+i)), TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	killCreature(t, g, me.ID, witness)
	passPriorityAroundTable(t, g)
	if !inExile(g, top[0]) || !inExile(g, top[1]) {
		t.Fatalf("the top two cards were not exiled")
	}
	if !b12ExiledPlayableBy(t, g, top[0], me.ID) || !b12ExiledPlayableBy(t, g, top[1], me.ID) {
		t.Fatal("both exiled cards should be playable at first")
	}
	if err := g.CastSpell(me.ID, top[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the first exiled land: %v", err)
	}
	if b12ExiledPlayableBy(t, g, top[1], me.ID) {
		t.Error("the second exiled card is still playable after one was played")
	}
}

func TestMscM15UnluckyWitnessWindowEndsAtYourNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	witness := pushCatalogPermanent(g, me.ID, "Unlucky Witness", "Creature — Human Citizen", mscM15WitnessOracle, false)
	seedLibrary(me, "Exiled A", "Exiled B")
	killCreature(t, g, me.ID, witness)
	passPriorityAroundTable(t, g)
	exiled := make([]uuid.UUID, 0, 2)
	for _, c := range g.Exile.Cards {
		if c.Owner == me.ID {
			exiled = append(exiled, c.InstanceID)
		}
	}
	if len(exiled) != 2 {
		t.Fatalf("%d cards exiled, want 2", len(exiled))
	}
	advanceToEndStep(t, g)
	if b12ExiledPlayableBy(t, g, exiled[0], me.ID) {
		t.Error("the exiled card is still playable after this turn's end step began")
	}
}

// --- Victor Timely, Wily Tycoon ----------------------------------------------------

func mscM15VictorWithGraveyard(t *testing.T) (g *game.Game, me *game.Player, cheapArtifact, dearArtifact, sorcery, creature uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	push := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		return id
	}
	cheapArtifact = push("Cheap Trinket", "Artifact", "{4}")
	dearArtifact = push("Dear Trinket", "Artifact", "{5}")
	sorcery = push("Old Sorcery", "Sorcery", "{2}")
	creature = push("Old Bear", "Creature — Bear", "{1}")
	castCatalogSpell(t, g, "Victor Timely, Wily Tycoon", "Legendary Creature — Human Artificer Villain", mscM15VictorOracle, nil)
	passPriorityAroundTable(t, g)
	return g, me, cheapArtifact, dearArtifact, sorcery, creature
}

func TestMscM15VictorOffersOnlyCheapArtifactInstantOrSorcery(t *testing.T) {
	g, me, cheap, dear, sorcery, creature := mscM15VictorWithGraveyard(t)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range []uuid.UUID{cheap, sorcery} {
		if !hasID(p.PickTargetCards, id) {
			t.Errorf("a legal card was not offered: %v", id)
		}
	}
	for _, id := range []uuid.UUID{dear, creature} {
		if hasID(p.PickTargetCards, id) {
			t.Errorf("an illegal card was offered: %v", id)
		}
	}
}

func TestMscM15VictorCastsTheArtifactForFree(t *testing.T) {
	g, me, cheap, _, _, _ := mscM15VictorWithGraveyard(t)
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, cheap)
	if err := castThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, cheap) {
		t.Error("the artifact did not resolve onto the battlefield")
	}
}

func TestMscM15VictorExilesTheSpellIfItWouldGoToTheGraveyard(t *testing.T) {
	g, me, _, _, sorcery, _ := mscM15VictorWithGraveyard(t)
	pickCard(t, g, me.ID, sorcery)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	if err := castThroughResolutionCast(t, g, me.ID, sorcery, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(sorcery) || !inExile(g, sorcery) {
		t.Error("the sorcery should be exiled, not back in the graveyard")
	}
}

func TestMscM15VictorDecliningLeavesTheCardInTheGraveyard(t *testing.T) {
	g, me, cheap, _, _, _ := mscM15VictorWithGraveyard(t)
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(cheap) {
		t.Error("the declined card left the graveyard")
	}
}

// --- Titan of Littjara -------------------------------------------------------------

func TestMscM15TitanSharesTheChosenTypeAndDrawsPerSharingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)

	castCatalogSpell(t, g, "Titan of Littjara", "Creature — Illusion", mscM15TitanOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Bear")
	passPriorityAroundTable(t, g)

	titan := findOnBattlefieldByOracle(g, mscM15TitanOracle)
	card, _ := g.LookupCardForEffect(titan)
	if !card.HasSubtype("Bear") {
		t.Errorf("the Titan did not become a Bear: %v", card.Effective().Subtypes)
	}
	before := me.Hand.Size()
	answerMayChoice(t, g, me.ID, true)
	if got := me.Hand.Size() - before; got != 2 {
		t.Fatalf("drew %d, want 2 (two Bears share a type)", got)
	}
	discardFromHand(t, g, me.ID)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("net hand change %d, want +1 after the discard", got)
	}
}

func TestMscM15TitanDecliningDrawsAndDiscardsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Titan of Littjara", "Creature — Illusion", mscM15TitanOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Bear")
	passPriorityAroundTable(t, g)
	before, gy := me.Hand.Size(), me.Graveyard.Size()
	answerMayChoice(t, g, me.ID, false)
	if me.Hand.Size() != before || me.Graveyard.Size() != gy {
		t.Error("declining drew or discarded")
	}
}

func TestMscM15TitanAsksNothingWithNoSharingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	castCatalogSpell(t, g, "Titan of Littjara", "Creature — Illusion", mscM15TitanOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Bear")
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("asked to draw with no creature sharing a type")
	}
}

func TestMscM15TitanDrawsAgainWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	titan := pushCatalogPermanent(g, me.ID, "Titan of Littjara", "Creature — Illusion", mscM15TitanOracle, false)
	b12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == titan {
				g.Battlefield.Cards[i].NamedTribe = "Bear"
			}
		}
	})
	declareAttack(t, g, opp.ID, titan)
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()
	answerMayChoice(t, g, me.ID, true)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d on the attack, want 1", got)
	}
	discardFromHand(t, g, me.ID)
}
