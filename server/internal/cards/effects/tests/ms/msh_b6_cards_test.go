package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b6_cards_test.go — Marvel Super Heroes, half B, sixth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB6HydeOracle     = "4a7c7ecc-df72-4da0-af35-dd0af58ec02c"
	mshB6MjolnirOracle  = "7f9a8845-d760-44a7-a4c9-8a20dba4e14a"
	mshB6MoleManOracle  = "34994475-b914-44f6-8ca1-c2747b6d5956"
	mshB6MoonGirlOracle = "bfe0818b-ae13-4174-85c8-24b895d28df0"
	mshB6MsMarvelOracle = "ebedff30-ae43-4861-aab2-afe1fc88337e"
	mshB6NamorOracle    = "dff11626-ad4d-443a-b7db-f508790b5fd2"
	mshB6NurseOracle    = "e3559565-0512-4525-addb-bf27903bcd13"
	mshB6OkoyeOracle    = "6d397d26-7e52-47b2-9880-2c10e90cfba4"
	mshB6PounceOracle   = "ead0820e-fc6b-4ab2-9110-048d6d53985d"
	mshB6PetsOracle     = "bbee7921-19c9-4c70-8462-f16e7e9c25db"
	mshB6PhotonOracle   = "a5424e33-ff4e-4c82-8aa9-9d25ced7830a"
	mshB6TriumphOracle  = "0bbceb51-1fb8-44d7-8afc-b4746d292262"
)

func TestMshB6EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB6HydeOracle:     "Mister Hyde, Monster Within",
		mshB6MjolnirOracle:  "Mjölnir, Hammer of Thor",
		mshB6MoleManOracle:  "Mole Man, Moloid Master",
		mshB6MoonGirlOracle: "Moon Girl and Devil Dinosaur",
		mshB6MsMarvelOracle: "Ms. Marvel, Kamala Khan",
		mshB6NamorOracle:    "Namor the Sub-Mariner",
		mshB6NurseOracle:    "Night Nurse, Healer of Heroes",
		mshB6OkoyeOracle:    "Okoye, Dora Milaje Leader",
		mshB6PounceOracle:   "Panther Pounce",
		mshB6PetsOracle:     "Pet Avengers",
		mshB6PhotonOracle:   "Photon Blast Barrage",
		mshB6TriumphOracle:  "Political Triumph",
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

// mshB6Settle passes priority until the stack is empty, keeping every scry
// where it is and re-picking `keep` (or the first legal target) for every
// copy's new-target prompt.
func mshB6Settle(t *testing.T, g *game.Game, who, keep uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if c := effectstest.ScryChoiceFor(g, who); c != nil {
			if err := g.ResolveScry(c.ID, who, nil, c.ScryCards); err != nil {
				t.Fatalf("ResolveScry: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, who); p != nil {
			ref := game.TargetRef{Kind: game.TargetCard, ID: keep}
			if !mshB6Has(p.PickTargetCards, keep) && len(p.PickTargetCards) > 0 {
				ref.ID = p.PickTargetCards[0]
			}
			if err := g.ResolvePickTarget(p.ID, who, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
}

func mshB6Has(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// mshB6Push puts a coloured catalog creature on the battlefield.
func mshB6Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, colors []string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- Mister Hyde, Monster Within ----------------------------------------------

// mshB6UpkeepMode walks to seat 0's next upkeep and answers the mode prompt.
func mshB6UpkeepMode(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	effectstest.B12ToMyNextUpkeep(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		effectstest.PassPriorityAroundTable(t, g)
		c = effectstest.ModePickChoiceFor(g, me.ID)
	}
	if c == nil {
		t.Fatal("the upkeep trigger asked for no mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMshB6HydeFirstModePutsAPlusOneCounterOnHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hyde := effectstest.B12Push(g, me.ID, "Mister Hyde, Monster Within", "Legendary Creature — Human Villain", mshB6HydeOracle, 2, 2)
	mshB6UpkeepMode(t, g, me, 0)
	if got := effectstest.PlusOnes(g, hyde); got != 1 {
		t.Errorf("Hyde has %d +1/+1 counters, want 1", got)
	}
}

func TestMshB6HydeSecondModeRemovesACounterAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mister Hyde, Monster Within", "Legendary Creature — Human Villain", mshB6HydeOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounterBy(me.ID, bear, game.CounterPlusOne, 2); err != nil {
		t.Fatal(err)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	handBefore := me.Hand.Size()
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		effectstest.PassPriorityAroundTable(t, g)
		c = effectstest.ModePickChoiceFor(g, me.ID)
	}
	if c == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("Bear has %d +1/+1 counters, want 1 (one removed)", got)
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want one card drawn", handBefore, got)
	}
}

func TestMshB6HydeSecondModeAsksWhichCounterWhenThereIsAChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mister Hyde, Monster Within", "Legendary Creature — Human Villain", mshB6HydeOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	for _, id := range []uuid.UUID{bear, wolf} {
		if err := g.AddCounterBy(me.ID, id, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	handBefore := me.Hand.Size()
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		effectstest.PassPriorityAroundTable(t, g)
		c = effectstest.ModePickChoiceFor(g, me.ID)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("expected a two-way pick (Bear or Wolf), got %+v", pick)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	// Bear and Wolf are listed in battlefield order; the second one loses it.
	if effectstest.PlusOnes(g, bear)+effectstest.PlusOnes(g, wolf) != 1 {
		t.Errorf("counters left: Bear %d, Wolf %d, want exactly one", effectstest.PlusOnes(g, bear), effectstest.PlusOnes(g, wolf))
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want one card drawn", handBefore, got)
	}
}

func TestMshB6HydeSecondModeDrawsNothingWithoutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mister Hyde, Monster Within", "Legendary Creature — Human Villain", mshB6HydeOracle, 2, 2)
	effectstest.B12ToMyNextUpkeep(t, g)
	handBefore := me.Hand.Size()
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		effectstest.PassPriorityAroundTable(t, g)
		c = effectstest.ModePickChoiceFor(g, me.ID)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d -> %d: no counter was removed, so nothing is drawn", handBefore, got)
	}
}

// --- Mjölnir, Hammer of Thor --------------------------------------------------

func TestMshB6MjolnirDealsFourDamageToATargetCreatureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 4)
	effectstest.B12PlayFromHand(t, g, "Mjölnir, Hammer of Thor", "Legendary Artifact — Equipment", mshB6MjolnirOracle, game.CastSpellParams{})
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: ogre}, true)
	if g.Battlefield.Contains(ogre) {
		t.Error("a 3/4 survived Mjölnir's 4 damage")
	}
}

func TestMshB6MjolnirDoublesAllDamageTheEquippedCreatureDeals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hammer := effectstest.PushCatalogPermanent(g, me.ID, "Mjölnir, Hammer of Thor", "Legendary Artifact — Equipment", mshB6MjolnirOracle, false)
	thor := mshB6Push(g, me.ID, "Thor", "Legendary Creature — God Hero", "", []string{"R"}, 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, hammer, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: thor}},
	}); err != nil {
		t.Fatalf("equip worthy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, thor, other)
	if got := before - opp.Life; got != 6+2 {
		t.Errorf("the opponent lost %d, want 8 (Thor's 3 doubled plus the Bear's 2)", got)
	}
}

func TestMshB6MjolnirEquipWorthyOnlyFitsALegendaryNonVillainThatIsRedOrWhite(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		colors   []string
		ok       bool
	}{
		{"Red Hero", "Legendary Creature — Human Hero", []string{"R"}, true},
		{"White Hero", "Legendary Creature — Human Hero", []string{"W"}, true},
		{"Blue and white Hero", "Legendary Creature — Human Hero", []string{"U", "W"}, true},
		{"Blue Hero", "Legendary Creature — Human Hero", []string{"U"}, false},
		{"Red Villain", "Legendary Creature — Human Villain", []string{"R"}, false},
		{"Red non-legend", "Creature — Human Hero", []string{"R"}, false},
		{"Colorless Hero", "Legendary Creature — Human Hero", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			hammer := effectstest.PushCatalogPermanent(g, me.ID, "Mjölnir, Hammer of Thor", "Legendary Artifact — Equipment", mshB6MjolnirOracle, false)
			target := mshB6Push(g, me.ID, tc.name, tc.typeLine, "", tc.colors, 2, 2)
			effectstest.FloatMana(t, g, me, "{C}")
			err := g.ActivateCatalogAbility(me.ID, hammer, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
			})
			if (err == nil) != tc.ok {
				t.Errorf("equip to %s: err = %v, want ok=%v", tc.name, err, tc.ok)
			}
		})
	}
}

func TestMshB6MjolnirDiscardedFromHandDealsTwoToEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 2)
	mine := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Small", "Creature — Bear", 2, 2)
	hammer := effectstest.PushCatalogHandCard(me, "Mjölnir, Hammer of Thor", "Legendary Artifact — Equipment", mshB6MjolnirOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hammer, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("discard ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, hammer) != game.ZoneGraveyard {
		t.Errorf("the card is in %q, want the graveyard", effectstest.B12ZoneOf(g, hammer))
	}
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(theirs) {
		t.Error("a creature with toughness 2 survived 2 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("a 3/3 did not survive 2 damage")
	}
}

// --- Mole Man, Moloid Master --------------------------------------------------

func TestMshB6MoleManPlaysLandsFromYourGraveyardAndMakesAMoloid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Mole Man, Moloid Master", "Legendary Creature — Human Villain", mshB6MoleManOracle, false)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not enter")
	}
	moloid := effectstest.MshB1Token(t, g, me.ID, "Minion")
	if moloid.Name != "Moloid" || moloid.Power != 1 || moloid.Toughness != 1 || !moloid.HasColor("G") {
		t.Errorf("token is %+v, want a 1/1 green Minion named Moloid", moloid)
	}
}

func TestMshB6MoleManNeededToPlayLandsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Error("a land was played from the graveyard with no Mole Man")
	}
}

func TestMshB6MoloidMaySelfMillWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Mole Man, Moloid Master", "Legendary Creature — Human Villain", mshB6MoleManOracle, false)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	moloid := effectstest.MshB1Token(t, g, me.ID, "Minion").InstanceID
	effectstest.SeedLibrary(me, "Top", "Second")

	graveBefore := me.Graveyard.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: moloid, Actor: me.ID, Target: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != graveBefore+1 {
		t.Errorf("graveyard %d -> %d, want one card milled", graveBefore, got)
	}

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: moloid, Actor: me.ID, Target: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != graveBefore+1 {
		t.Errorf("graveyard %d after declining, want it unchanged at %d", got, graveBefore+1)
	}
}

// --- Moon Girl and Devil Dinosaur ---------------------------------------------

func TestMshB6MoonGirlBecomesSixSixWithTrampleOnTheSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	girl := effectstest.B12Push(g, me.ID, "Moon Girl and Devil Dinosaur", "Legendary Creature — Human Dinosaur Hero", mshB6MoonGirlOracle, 2, 2)

	effectstest.HobA5Draw(t, g, me, nil)
	if p, tough := effectstest.EffectivePTOf(t, g, girl); p != 2 || tough != 2 {
		t.Fatalf("after one draw: %d/%d, want 2/2", p, tough)
	}
	effectstest.HobA5Draw(t, g, me, nil)
	if p, tough := effectstest.EffectivePTOf(t, g, girl); p != 6 || tough != 6 {
		t.Errorf("after the second draw: %d/%d, want 6/6", p, tough)
	}
	if !effectstest.HasKeywordOf(g, girl, "trample") {
		t.Error("no trample after the second draw")
	}
	// Her turn comes round: until end of turn is over.
	effectstest.PowerUpPassTo(t, g, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, girl); p != 2 || tough != 2 {
		t.Errorf("a turn later: %d/%d, want 2/2", p, tough)
	}
}

func TestMshB6MoonGirlDrawsOnceEachTurnForAnArtifactYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Moon Girl and Devil Dinosaur", "Legendary Creature — Human Dinosaur Hero", mshB6MoonGirlOracle, 2, 2)

	// An opponent's artifact draws nothing.
	before := me.Hand.Size()
	effectstest.B12Permanent(g, opp.ID, "Their Trinket", "Artifact")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: effectstest.FindBattlefieldByName(g, "Their Trinket"), Actor: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Fatalf("an opponent's artifact drew %d cards", got-before)
	}

	effectstest.B12PlayFromHand(t, g, "Trinket", "Artifact", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	afterFirst := me.Hand.Size()
	if afterFirst != before+1 {
		t.Fatalf("hand %d -> %d after the first artifact (the Trinket left the hand, a card was drawn)", before, afterFirst)
	}
	effectstest.B12PlayFromHand(t, g, "Gadget", "Artifact", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != afterFirst {
		t.Errorf("hand %d -> %d: the second artifact this turn drew again", afterFirst, got)
	}
}

// --- Ms. Marvel, Kamala Khan --------------------------------------------------

func TestMshB6MsMarvelHasReachVigilanceAndNoMaximumHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marvel := effectstest.B12Push(g, me.ID, "Ms. Marvel, Kamala Khan", "Legendary Creature — Mutant Inhuman Hero", mshB6MsMarvelOracle, 1, 4)
	if !effectstest.HasKeywordOf(g, marvel, "reach") || !effectstest.HasKeywordOf(g, marvel, "vigilance") {
		t.Error("Ms. Marvel lacks reach or vigilance")
	}
	var limit int
	g.WithWriteLock(func() { limit = g.EffectiveMaxHandSizeLocked(me) })
	if limit != game.NoMaxHandSize {
		t.Errorf("maximum hand size %d, want none", limit)
	}
}

func TestMshB6MsMarvelDrawsAndHerBasePowerBecomesYourHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	marvel := effectstest.B12Push(g, me.ID, "Ms. Marvel, Kamala Khan", "Legendary Creature — Mutant Inhuman Hero", mshB6MsMarvelOracle, 1, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, marvel); p != 1 {
		t.Fatalf("before the trigger her power is %d, want 1", p)
	}
	handBefore := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Djeru's Resolve", "Instant", effectstest.Pr7aDjerusResolve, []game.TargetRef{effectstest.Pr7aCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want one drawn (the spell itself left the hand)", handBefore, got)
	}
	// Her base power is the hand size, and it follows the hand.
	if p, tough := effectstest.EffectivePTOf(t, g, marvel); p != me.Hand.Size() || tough != 4 {
		t.Errorf("she is %d/%d with %d cards in hand, want power = hand size and toughness 4", p, tough, me.Hand.Size())
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if p, _ := effectstest.EffectivePTOf(t, g, marvel); p != me.Hand.Size() {
		t.Errorf("after another card was drawn her power is %d, hand %d", p, me.Hand.Size())
	}
	// Until end of turn only.
	effectstest.PowerUpPassTo(t, g, 0)
	if p, _ := effectstest.EffectivePTOf(t, g, marvel); p != 1 {
		t.Errorf("a turn later her power is %d, want 1", p)
	}
}

func TestMshB6MsMarvelIgnoresASpellTargetingAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ms. Marvel, Kamala Khan", "Legendary Creature — Mutant Inhuman Hero", mshB6MsMarvelOracle, 1, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Djeru's Resolve", "Instant", effectstest.Pr7aDjerusResolve, []game.TargetRef{effectstest.Pr7aCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d -> %d: the spell was cast and nothing was drawn", handBefore, got)
	}
}

// --- Namor the Sub-Mariner ----------------------------------------------------

func TestMshB6NamorsPowerIsTheMerfolkYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	namor := effectstest.B12Push(g, me.ID, "Namor the Sub-Mariner", "Legendary Creature — Mutant Merfolk Villain", mshB6NamorOracle, 0, 4)
	if p, tough := effectstest.EffectivePTOf(t, g, namor); p != 1 || tough != 4 {
		t.Fatalf("alone he is %d/%d, want 1/4 (he is a Merfolk himself)", p, tough)
	}
	effectstest.B12Creature(g, me.ID, "Merfolk Scout", "Creature — Merfolk Scout", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, namor); p != 2 {
		t.Errorf("with another Merfolk he is %dp, want 2 (not the Bear, not the opponent's Merfolk)", p)
	}
	if !effectstest.HasKeywordOf(g, namor, "flying") {
		t.Error("Namor has no flying")
	}
}

func TestMshB6NamorMakesAMerfolkForEachBlueSymbolOfANoncreatureSpell(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
		pool                 []string
		want                 int
	}{
		{"Test Sorcery", "Sorcery", "{1}{U}{U}", []string{"U", "U", "U"}, 2},
		{"Test Hybrid", "Instant", "{U/R}{U/R}{R}", []string{"U", "U", "R"}, 2},
		{"Test Blue Instant", "Instant", "{U}", []string{"U"}, 1},
		{"Test Red Instant", "Instant", "{R}", []string{"R"}, 0},
		{"Test Blue Creature", "Creature — Bear", "{U}", []string{"U"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			effectstest.B12Push(g, me.ID, "Namor the Sub-Mariner", "Legendary Creature — Mutant Merfolk Villain", mshB6NamorOracle, 0, 4)
			effectstest.AssistCardCastX(t, g, tc.name, tc.typeLine, "", tc.cost, tc.pool, 0, nil)
			mshB6Settle(t, g, me.ID, uuid.Nil)
			if got := effectstest.MshB1Tokens(g, me.ID, "Merfolk"); got != tc.want {
				t.Errorf("%d Merfolk tokens, want %d", got, tc.want)
			}
			if tc.want > 0 {
				m := effectstest.MshB1Token(t, g, me.ID, "Merfolk")
				if m.Power != 1 || m.Toughness != 1 || !m.HasColor("U") {
					t.Errorf("token is %+v, want 1/1 blue", m)
				}
			}
		})
	}
}

// --- Night Nurse, Healer of Heroes --------------------------------------------

func TestMshB6NightNurseReturnsAPermanentCardPutIntoYourGraveyardThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	old := effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	bear := effectstest.B12Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Spent Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, spell); err != nil {
		t.Fatal(err)
	}

	effectstest.CastPowerUpCreature(t, g, me, "Night Nurse, Healer of Heroes", "Legendary Creature — Human Doctor Hero", "{1}{W}", "{C}{W}", mshB6NurseOracle, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MshB2PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.B12ZoneOf(g, bear) != game.ZoneHand {
		t.Errorf("the fresh Bear is in %q, want the hand", effectstest.B12ZoneOf(g, bear))
	}
	if effectstest.B12ZoneOf(g, old) != game.ZoneGraveyard || effectstest.B12ZoneOf(g, spell) != game.ZoneGraveyard {
		t.Error("a card that was not a permanent put there this turn moved")
	}
}

func TestMshB6NightNurseHasFlashAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nurse := effectstest.B12Push(g, me.ID, "Night Nurse, Healer of Heroes", "Legendary Creature — Human Doctor Hero", mshB6NurseOracle, 2, 1)
	if !effectstest.HasKeywordOf(g, nurse, "flash") || !effectstest.HasKeywordOf(g, nurse, "lifelink") {
		t.Error("Night Nurse lacks flash or lifelink")
	}
}

// --- Okoye, Dora Milaje Leader ------------------------------------------------

func TestMshB6OkoyeMakesTwoSoldiersWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	effectstest.CastPowerUpCreature(t, g, me, "Okoye, Dora Milaje Leader", "Legendary Creature — Human Warrior Hero", "{3}{W}", "{C}{C}{C}{W}", mshB6OkoyeOracle, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MshB1Tokens(g, me.ID, "Soldier"); got != 2 {
		t.Errorf("%d Soldier tokens, want 2", got)
	}
}

func TestMshB6OkoyeGivesFirstStrikeToAttackingTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Okoye, Dora Milaje Leader", "Legendary Creature — Human Warrior Hero", mshB6OkoyeOracle, 3, 2)
	token := effectstest.B12Creature(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	idle := effectstest.B12Creature(g, me.ID, "Idle Soldier", "Token Creature — Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Soldier", "Token Creature — Soldier", 1, 1)

	if effectstest.HasKeywordOf(g, token, "first strike") {
		t.Fatal("a token that is not attacking has first strike")
	}
	effectstest.DeclareAttack(t, g, opp.ID, token, bear)
	if !effectstest.HasKeywordOf(g, token, "first strike") {
		t.Error("an attacking token has no first strike")
	}
	if effectstest.HasKeywordOf(g, bear, "first strike") {
		t.Error("an attacking nontoken has first strike")
	}
	if effectstest.HasKeywordOf(g, idle, "first strike") {
		t.Error("a token that is not attacking has first strike")
	}
	if effectstest.HasKeywordOf(g, theirs, "first strike") {
		t.Error("an opponent's token has first strike")
	}
}

// --- Panther Pounce -----------------------------------------------------------

func TestMshB6PantherPounceInvestigatesPumpsGrantsFlyingAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.TapCard(bear, true); err != nil {
		t.Fatalf("tap: %v", err)
	}

	effectstest.CastCatalogSpell(t, g, "Panther Pounce", "Instant", mshB6PounceOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: bear, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)

	clues := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && c.IsToken() && c.HasSubtype("Clue") {
			clues++
		}
	}
	if clues != 1 {
		t.Errorf("the targeted player made %d Clues, want 1", clues)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("Bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "flying") {
		t.Error("Bear has no flying")
	}
	if effectstest.B13Tapped(t, g, bear) {
		t.Error("Bear is still tapped")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() {
			t.Error("the caster made a token when another player investigated")
		}
	}
}

func TestMshB6PantherPounceCanTargetYourselfForTheClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Panther Pounce", "Instant", mshB6PounceOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		{Kind: game.TargetCard, ID: bear, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MshB1Tokens(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Pet Avengers -------------------------------------------------------------

func TestMshB6PetAvengersPowerUpPutsACounterAndMakesAHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	pets := effectstest.CastPowerUpCreature(t, g, me, "Pet Avengers", "Creature — Dragon Cat Dog Bird Frog Hero", "{3}{G}", "{C}{C}{C}{G}", mshB6PetsOracle, 4)
	if !effectstest.HasKeywordOf(g, pets, "reach") {
		t.Error("Pet Avengers has no reach")
	}
	if got := effectstest.ChargedPowerUp(t, g, me.ID, pets); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3} ({6}{G} less {3}{G})", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, pets, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, pets); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	hero := effectstest.MshB1Token(t, g, me.ID, "Hero")
	if hero.Power != 3 || hero.Toughness != 2 || !hero.HasColor("W") || !effectstest.HasKeywordOf(g, hero.InstanceID, "vigilance") {
		t.Errorf("token is %+v, want a 3/2 white Hero with vigilance", hero)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, pets, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a second power-up activation was accepted")
	}
}

// --- Photon Blast Barrage -----------------------------------------------------

func TestMshB6PhotonBlastBarrageCopiesItXTimes(t *testing.T) {
	for _, tc := range []struct {
		x         int
		wantDead  bool
		toughness int
	}{
		{0, false, 2},
		{1, true, 2},
		{2, false, 4},
		{3, true, 4},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, tc.toughness)
		pool := []string{"R", "R"}
		for i := 0; i < tc.x; i++ {
			pool = append(pool, "R")
		}
		effectstest.AssistCardCastX(t, g, "Photon Blast Barrage", "Sorcery", mshB6PhotonOracle, "{X}{R}{R}", pool, tc.x,
			[]game.TargetRef{effectstest.Pr7aCard(victim)})
		mshB6Settle(t, g, me.ID, victim)
		if dead := !g.Battlefield.Contains(victim); dead != tc.wantDead {
			t.Errorf("X=%d vs toughness %d: dead = %v, want %v", tc.x, tc.toughness, dead, tc.wantDead)
		}
	}
}

func TestMshB6PhotonBlastBarrageCopiesMayChooseNewTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	first := effectstest.B12Creature(g, opp.ID, "First", "Creature — Ogre", 3, 1)
	second := effectstest.B12Creature(g, opp.ID, "Second", "Creature — Ogre", 3, 1)
	effectstest.AssistCardCastX(t, g, "Photon Blast Barrage", "Sorcery", mshB6PhotonOracle, "{X}{R}{R}", []string{"R", "R", "R"}, 1,
		[]game.TargetRef{effectstest.Pr7aCard(first)})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the copy was offered no new targets")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, effectstest.Pr7aCard(second)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	mshB6Settle(t, g, me.ID, second)
	if g.Battlefield.Contains(first) || g.Battlefield.Contains(second) {
		t.Error("the spell and its copy should have killed both targets")
	}
}

// --- Political Triumph --------------------------------------------------------

func TestMshB6PoliticalTriumphScriesAndCountsEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	plan := effectstest.PushCatalogPermanent(g, me.ID, "Political Triumph", "Enchantment — Plan", mshB6TriumphOracle, false)
	effectstest.SeedLibrary(me, "Top", "Second")

	effectstest.B12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("no scry 1 prompt: %+v", c)
	}
	mshB6Settle(t, g, me.ID, uuid.Nil)
	if got := effectstest.CounterCount(g, plan, "plan"); got != 1 {
		t.Errorf("%d plan counters, want 1", got)
	}

	// An opponent's creature, and a noncreature permanent of mine, count nothing.
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs, Actor: opp.ID})
	})
	effectstest.B12PlayFromHand(t, g, "Trinket", "Artifact", "", game.CastSpellParams{})
	mshB6Settle(t, g, me.ID, uuid.Nil)
	if got := effectstest.CounterCount(g, plan, "plan"); got != 1 {
		t.Errorf("%d plan counters after a foreign creature and an artifact, want 1", got)
	}
}

func TestMshB6PoliticalTriumphFourthCounterSacrificesDrawsAndGrowsEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plan := effectstest.PushCatalogPermanent(g, me.ID, "Political Triumph", "Enchantment — Plan", mshB6TriumphOracle, false)
	old := effectstest.B12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounterBy(me.ID, plan, "plan", 3); err != nil {
		t.Fatal(err)
	}
	effectstest.SeedLibrary(me, "Top", "Second", "Third")
	handBefore := me.Hand.Size()

	newbie := effectstest.B12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	mshB6Settle(t, g, me.ID, uuid.Nil)

	if g.Battlefield.Contains(plan) {
		t.Error("Political Triumph was not sacrificed on its fourth plan counter")
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want one card drawn (the Bear came and left the hand)", handBefore, got)
	}
	for name, id := range map[string]uuid.UUID{"the old Bear": old, "the new Bear": newbie} {
		if got := effectstest.PlusOnes(g, id); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, got)
		}
	}
}
