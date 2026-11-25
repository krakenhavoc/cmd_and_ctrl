package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b292_cards_test.go — Shadows over Innistrad, S82 older sets
// batch 292 (#5543). Each test plays the real catalog entries through the
// engine.

const (
	soiB292PyreHound    = "0a6f45e9-5c7a-4e2c-bf8a-fc5fe0ae0793"
	soiB292Mentor       = "17cbeec4-b180-48e9-9156-42271f918a9d"
	soiB292Brain        = "321dbd10-1d48-49fc-ba6a-1df241a53338"
	soiB292Dreadknight  = "73e2d8aa-3a23-4458-8cae-e9209a102197"
	soiB292Captain      = "2d8eba16-cfca-4a38-bde0-6ee36c67231a"
	soiB292Cult         = "4ea5409e-9b39-434c-bdb8-8c798128f199"
	soiB292Partisan     = "dfeb239b-8904-4ef0-9a91-edc901b0c49a"
	soiB292Emissary     = "f21bbacb-0436-46fa-b361-582b766d3964"
	soiB292Geist        = "5c975dbb-aabb-43fe-877c-c3950019a584"
	soiB292PaleRider    = "6cc47c8e-9154-4517-940d-d1a7f35e5677"
	soiB292StartledWake = "99d42ae8-8cee-449b-812b-98a81a593077"
)

func TestSoiB292EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		soiB292PyreHound: "Pyre Hound", soiB292Mentor: "Stromkirk Mentor", soiB292Brain: "Brain in a Jar",
		soiB292Dreadknight: "Markov Dreadknight", soiB292Captain: "Hanweir Militia Captain",
		soiB292Captain + "#1": "Westvale Cult Leader", soiB292Cult: "Cult of the Waxing Moon",
		soiB292Partisan: "Silverfur Partisan", soiB292Emissary: "Emissary of the Sleepless",
		soiB292Geist: "Uninvited Geist", soiB292Geist + "#1": "Unimpeded Trespasser",
		soiB292PaleRider: "Pale Rider of Trostad", soiB292StartledWake: "Startled Awake",
		soiB292StartledWake + "#1": "Persistent Nightmare",
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

func soiB292PickAndSettle(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	for i := 0; i < 4; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, chooser) == nil && effectstest.StackFullyEmpty(g) {
			return
		}
	}
}

// soiB292Row is a transforming card's catalog row.
func soiB292Row(oracle, front, frontType, frontCost, frontP, frontT, back, backType, backP, backT string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: frontType + " // " + backType,
		OracleID: uuid.MustParse(oracle),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: frontType, ManaCost: frontCost, Power: frontP, Toughness: frontT},
			{Name: back, TypeLine: backType, Power: backP, Toughness: backT},
		},
	}
}

func soiB292Put(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	c.SummonedThisTurn = false
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

// --- Pyre Hound ------------------------------------------------------------

func TestSoiB292PyreHoundGrowsOnlyForYourInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hound := effectstest.B12Push(g, me.ID, "Pyre Hound", "Creature — Elemental Dog", soiB292PyreHound, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, hound, "trample") {
		t.Error("Pyre Hound has no trample")
	}
	effectstest.CastCatalogSpell(t, g, "Shock Test", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Ponder Test", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, hound, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d counters after an instant and a sorcery, want 2", n)
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, hound, game.CounterPlusOne); n != 2 {
		t.Errorf("an artifact spell added a counter: %d", n)
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: uuid.New()}) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, hound, game.CounterPlusOne); n != 2 {
		t.Errorf("an opponent's cast added a counter: %d", n)
	}
}

// --- Stromkirk Mentor ------------------------------------------------------

func TestSoiB292MentorCountersAnotherVampireYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vamp := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Vampire", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Vampire", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := effectstest.CastHoldCreature(t, g, "Stromkirk Mentor", "Creature — Vampire Soldier", soiB292Mentor, 4, 2)
	soiB292PickAndSettle(t, g, me.ID, vamp)
	if n := effectstest.CountersOn(g, vamp, game.CounterPlusOne); n != 1 {
		t.Errorf("the other Vampire has %d counters, want 1", n)
	}
	for name, other := range map[string]uuid.UUID{"the Mentor": id, "an opposing Vampire": theirs, "a Bear": bear} {
		if n := effectstest.CountersOn(g, other, game.CounterPlusOne); n != 0 {
			t.Errorf("%s got %d counters", name, n)
		}
	}
}

// --- Markov Dreadknight ----------------------------------------------------

func TestSoiB292DreadknightDiscardsToGrowByTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	knight := effectstest.B12Push(g, me.ID, "Markov Dreadknight", "Creature — Vampire Knight", soiB292Dreadknight, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, knight, "flying") {
		t.Error("Markov Dreadknight has no flying")
	}
	me.Hand.Cards = nil
	effectstest.FloatForTest(g, me, "BBB")
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no card to discard")
	}
	fodder := effectstest.B13HandCard(me, "Fodder", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder) || !me.Graveyard.Contains(fodder) {
		t.Error("the card was not discarded")
	}
	if n := effectstest.CountersOn(g, knight, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2", n)
	}
}

// --- Pale Rider of Trostad -------------------------------------------------

func TestSoiB292PaleRiderDiscardsAndHasSkulk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastHoldCreature(t, g, "Pale Rider of Trostad", "Creature — Spirit", soiB292PaleRider, 3, 3)
	before := me.Hand.Size()
	if !effectstest.HasEffectiveKeyword(t, g, id, "skulk") {
		t.Error("Pale Rider has no skulk")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("the controller was not asked to discard")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before-1 {
		t.Errorf("hand %d, want %d after the discard", got, before-1)
	}
}

// --- Emissary of the Sleepless ---------------------------------------------

func TestSoiB292EmissaryMakesASpiritOnlyIfACreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Emissary of the Sleepless", "Creature — Spirit", soiB292Emissary, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("%d Spirits with no death this turn", n)
	}
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastHoldCreature(t, g, "Emissary of the Sleepless", "Creature — Spirit", soiB292Emissary, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Spirit"); n != 1 {
		t.Errorf("%d Spirits after a creature died, want 1", n)
	}
}

// --- Silverfur Partisan ----------------------------------------------------

func TestSoiB292PartisanMakesAWolfWhenYourWolfIsTargetedBySpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Silverfur Partisan", "Creature — Wolf Warrior", soiB292Partisan, 2, 2)
	wolf := effectstest.B12Creature(g, me.ID, "Pup", "Creature — Wolf", 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != before {
		t.Fatalf("a spell targeting a Bear made a Wolf (%d)", n)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: wolf}})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != before+1 {
		t.Errorf("%d Wolf tokens after an instant targeted a Wolf, want %d", n, before+1)
	}
}

// --- Cult of the Waxing Moon -----------------------------------------------

func TestSoiB292CultMakesAWolfWhenYourPermanentTransformsIntoANonHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Cult of the Waxing Moon", "Creature — Human Shaman", soiB292Cult, 5, 4)
	// A Human that turns into a Werewolf makes a Wolf; one that turns into
	// another Human does not.
	howler := soiB292Put(t, g, soiB292Row("11111111-1111-4111-8111-111111111111",
		"Test Villager", "Creature — Human Werewolf", "{1}{G}", "2", "2",
		"Test Howler", "Creature — Werewolf", "3", "3"), me)
	human := soiB292Put(t, g, soiB292Row("22222222-2222-4222-8222-222222222222",
		"Test Captain", "Creature — Human Soldier", "{1}{W}", "2", "2",
		"Test Cleric", "Creature — Human Cleric", "3", "3"), me)
	effectstest.IwTransform(t, g, human)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != 0 {
		t.Fatalf("a Human back face made %d Wolves", n)
	}
	effectstest.IwTransform(t, g, howler)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != 1 {
		t.Errorf("%d Wolves after a Werewolf appeared, want 1", n)
	}
	// Turning it back to its Human face is not "into a non-Human".
	effectstest.IwTransform(t, g, howler)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != 1 {
		t.Errorf("turning back to a Human made a Wolf: %d", n)
	}
}

// --- Hanweir Militia Captain // Westvale Cult Leader ------------------------

func soiB292CaptainRow() cards.Card {
	return soiB292Row(soiB292Captain, "Hanweir Militia Captain", "Creature — Human Soldier", "{1}{W}", "2", "2",
		"Westvale Cult Leader", "Creature — Human Cleric", "*", "*")
}

func TestSoiB292CaptainTransformsAtYourUpkeepWithFourCreatures(t *testing.T) {
	g, _ := effectstest.IwTable(t)
	next := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := soiB292Put(t, g, soiB292CaptainRow(), next)
	for i := 0; i < 2; i++ {
		effectstest.B12Creature(g, next.ID, "Extra", "Creature — Bear", 1, 1)
	}
	effectstest.IwNextUpkeep(t, g, nil)
	if f := effectstest.IwFace(g, id); f != 0 {
		t.Fatalf("with three creatures the Captain is on face %d, want 0", f)
	}
	effectstest.B12Creature(g, next.ID, "Fourth", "Creature — Bear", 1, 1)
	// Walk a full round to the same player's next upkeep.
	for i := 0; i < len(g.Seats); i++ {
		effectstest.IwNextUpkeep(t, g, nil)
	}
	if f := effectstest.IwFace(g, id); f != 1 {
		t.Fatalf("with four creatures the Captain is on face %d, want 1", f)
	}
	if p := effectstest.EffectivePower(t, g, id); p != 4 {
		t.Errorf("Westvale Cult Leader's power %d, want the number of creatures you control (4)", p)
	}
}

func TestSoiB292CultLeaderMakesAClericAtYourEndStep(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := soiB292Put(t, g, soiB292CaptainRow(), me)
	effectstest.IwTransform(t, g, id)
	for i := 0; i < 40 && g.Turn.Step != game.StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Human Cleric"); n != 1 {
		t.Errorf("%d Human Cleric tokens at the end step, want 1", n)
	}
}

// --- Uninvited Geist // Unimpeded Trespasser -------------------------------

func TestSoiB292GeistTransformsOnCombatDamageToAPlayer(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := soiB292Put(t, g, soiB292Row(soiB292Geist, "Uninvited Geist", "Creature — Spirit", "{2}{U}", "2", "2",
		"Unimpeded Trespasser", "Creature — Spirit", "3", "3"), me)
	if !effectstest.HasEffectiveKeyword(t, g, id, "skulk") {
		t.Error("Uninvited Geist has no skulk")
	}
	effectstest.EmitNoncombatDamage(g, id, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, id); f != 0 {
		t.Fatalf("noncombat damage turned it over (face %d)", f)
	}
	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, id); f != 1 {
		t.Fatalf("combat damage left it on face %d, want 1", f)
	}
}

// --- Startled Awake // Persistent Nightmare --------------------------------

func soiB292AwakeRow() cards.Card {
	return soiB292Row(soiB292StartledWake, "Startled Awake", "Sorcery", "{2}{U}{U}", "", "",
		"Persistent Nightmare", "Creature — Nightmare", "1", "1")
}

func TestSoiB292StartledAwakeMillsThirteenAndReturnsAsNightmare(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.Msc18Library(opp, 20)
	libBefore := opp.Library.Size()
	card := effectstest.ImportToHand(soiB292AwakeRow(), me)
	effectstest.FloatForTest(g, me, "UUUU")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := opp.Library.Size(); n != libBefore-13 {
		t.Fatalf("opponent's library %d, want %d after milling thirteen", n, libBefore-13)
	}
	if !me.Graveyard.Contains(card) {
		t.Fatal("Startled Awake is not in the graveyard")
	}
	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the mana")
	}
	effectstest.FloatForTest(g, me, "UUCCC")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, card)
	if !ok || c.ActiveFace != 1 || c.Name != "Persistent Nightmare" {
		t.Fatalf("did not return as Persistent Nightmare: %+v", c)
	}
	if !effectstest.HasEffectiveKeyword(t, g, card, "skulk") {
		t.Error("Persistent Nightmare has no skulk")
	}
	effectstest.DealCombatDamageToPlayer(g, card, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("Persistent Nightmare was not returned to its owner's hand")
	}
}

// --- Brain in a Jar --------------------------------------------------------

func TestSoiB292BrainChargesAndFreeCastsAnExactManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	brain := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Brain in a Jar", TypeLine: "Artifact", ManaCost: "{2}", OracleID: soiB292Brain})
	two := uuid.New()
	one := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: two, Name: "Two Spell", TypeLine: "Sorcery", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: one, Name: "One Spell", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, brain, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("charge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, brain, game.CounterCharge); n != 1 {
		t.Fatalf("%d charge counters, want 1", n)
	}
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, one)
	for _, id := range offer.DiscardOptions {
		if id == two {
			t.Error("a spell with mana value 2 was offered at one counter")
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, one, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
}

func TestSoiB292BrainScriesForTheCountersRemoved(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	brain := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Brain in a Jar", TypeLine: "Artifact", ManaCost: "{2}", OracleID: soiB292Brain,
		Counters: map[string]int{game.CounterCharge: 2}})
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, brain, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the three mana")
	}
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, brain, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{brain}, CounterCounts: []int{2}}); err != nil {
		t.Fatalf("scry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, brain, game.CounterCharge); n != 0 {
		t.Errorf("%d charge counters left, want 0", n)
	}
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 2 {
		t.Errorf("no scry of two: %+v", c)
	}
}
