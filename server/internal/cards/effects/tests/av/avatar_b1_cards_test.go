package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b1_cards_test.go — Avatar: The Last Airbender, batch 1 (#4459).
// Each test plays the real catalog entry and asserts what its text says.

const (
	avB1AangsJourneyOracle   = "ddf461b9-a205-4dc7-a9c3-c047b1ff709f"
	avB1AscensionOracle      = "c4d2fdf9-637d-4e97-ae33-45f2d27bf8cd"
	avB1AppaOracle           = "c6e28cf0-81c7-4f52-a29b-988cc6fd31ee"
	avB1EnthusiastsOracle    = "7b0d9d79-2e86-4079-8db3-500483c8807e"
	avB1HealerOracle         = "6ef464bd-2ed0-4f99-af4a-ebf276f55384"
	avB1FarmAnimalsOracle    = "e451b5aa-5832-42b3-8d72-f67e126a529f"
	avB1ConfrontationOracle  = "9703140f-36ee-4427-a390-c1dbdb9e4ee9"
	avB1JailerOracle         = "1b799c30-4a43-403d-b822-0555bd513913"
	avB1ProtectorsOracle     = "3bfb3913-808d-444b-97c3-a90af3bc585b"
	avB1EnterAvatarOracle    = "4e552be1-0702-4e18-822f-98611f8a8e70"
	avB1AscensionQuestCounts = "quest"
)

func TestAvatarB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB1AangsJourneyOracle:  "Aang's Journey",
		avB1AscensionOracle:     "Airbender Ascension",
		avB1AppaOracle:          "Appa, Loyal Sky Bison",
		avB1EnthusiastsOracle:   "Avatar Enthusiasts",
		avB1HealerOracle:        "Compassionate Healer",
		avB1FarmAnimalsOracle:   "Curious Farm Animals",
		avB1ConfrontationOracle: "Destined Confrontation",
		avB1JailerOracle:        "Earth Kingdom Jailer",
		avB1ProtectorsOracle:    "Earth Kingdom Protectors",
		avB1EnterAvatarOracle:   "Enter the Avatar State",
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

// --- Aang's Journey -------------------------------------------------------------

func TestAvatarB1AangsJourneyFindsABasicLandAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Plains Card", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Some Shrine", TypeLine: "Legendary Enchantment — Shrine"},
		game.Card{Name: "Nonbasic", TypeLine: "Land — Plains Island"},
	)
	effectstest.AvB1Plains(g, me.ID, 2)
	life := me.Life
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Aang's Journey", TypeLine: "Sorcery — Lesson", ManaCost: "{2}", OracleID: avB1AangsJourneyOracle},
		game.CastSpellParams{Strict: true, AutoTap: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	for _, bad := range [][]uuid.UUID{{ids[1]}, {ids[2]}, {ids[0], ids[1]}} {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, bad); err == nil {
			t.Fatalf("unkicked, the search accepted %v", bad)
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land card should be in hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestAvatarB1AangsJourneyKickedFindsABasicAndAShrine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Plains A", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Plains B", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Shrine A", TypeLine: "Legendary Enchantment — Shrine"},
		game.Card{Name: "Shrine B", TypeLine: "Legendary Enchantment — Shrine"},
	)
	life := me.Life
	// Kicker {2} needs four mana; three cannot pay it.
	effectstest.AvB1Plains(g, me.ID, 3)
	effectstest.AdvanceToMain(t, g)
	short := effectstest.AvB1Hand(me, game.Card{Name: "Aang's Journey", TypeLine: "Sorcery — Lesson", ManaCost: "{2}", OracleID: avB1AangsJourneyOracle})
	if err := g.CastSpell(me.ID, short, game.CastSpellParams{Strict: true, AutoTap: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("a kicked cast was paid with three mana")
	}
	effectstest.AvB1Plains(g, me.ID, 1)
	if err := g.CastSpell(me.ID, short, game.CastSpellParams{Strict: true, AutoTap: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	for _, bad := range [][]uuid.UUID{{ids[0], ids[1]}, {ids[2], ids[3]}, {ids[0], ids[2], ids[3]}} {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, bad); err == nil {
			t.Fatalf("kicked, the search accepted %v", bad)
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1], ids[3])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[3]) {
		t.Error("the basic land card and the Shrine card should be in hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Airbender Ascension ---------------------------------------------------------

func TestAvatarB1AirbenderAscensionAirbendsCountsAndFlickers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	asc := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Airbender Ascension", TypeLine: "Enchantment", OracleID: avB1AscensionOracle}, game.CastSpellParams{})
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertAirbent(t, g, victim, opp.ID)

	// Three counters, then a creature entering under my control makes four;
	// one entering under an opponent's does not.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, avB1AscensionQuestCounts, 2) })
	effectstest.AvB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Elf")
	if n := effectstest.AvB1Card(t, g, asc).Counters[avB1AscensionQuestCounts]; n != 2 {
		t.Fatalf("an opponent's creature added a counter: %d", n)
	}
	effectstest.AvB1ResolveCreature(t, g, me, "My Elf", "Creature — Elf", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AvB1Card(t, g, asc).Counters[avB1AscensionQuestCounts]; n != 3 {
		t.Fatalf("quest counters %d, want 3", n)
	}

	// Three counters at my end step: nothing triggers.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil || g.Stack.Size() > 0 || len(g.StackMeta) > 0 {
		t.Fatal("the end step triggered with three quest counters")
	}
	if !g.Battlefield.Contains(mine) {
		t.Fatal("my creature moved without the fourth counter")
	}
	_ = mine
}

func TestAvatarB1AirbenderAscensionFlickersAtFourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	asc := effectstest.B12Push(g, me.ID, "Airbender Ascension", "Enchantment", avB1AscensionOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(asc, avB1AscensionQuestCounts, 4)
		_ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1)
	})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Fatal("the creature should have left as the old object")
	}
	back := effectstest.AvB1Named(g, me.ID, "My Bear")
	if back == uuid.Nil {
		t.Fatal("the creature never returned to the battlefield")
	}
	if c := effectstest.AvB1Card(t, g, back); c.Owner != me.ID || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("returned as %+v; want a new object under its owner, counters gone", c.Counters)
	}
}

// The intervening "if" is checked again on resolution: a counter removed in
// response stops the flicker.
func TestAvatarB1AirbenderAscensionRechecksOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	asc := effectstest.B12Push(g, me.ID, "Airbender Ascension", "Enchantment", avB1AscensionOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, avB1AscensionQuestCounts, 4) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PickCard(t, g, me.ID, mine)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, avB1AscensionQuestCounts, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Fatal("the flicker ran though the enchantment had three counters on resolution")
	}
}

// --- Appa, Loyal Sky Bison --------------------------------------------------------

func TestAvatarB1AppaEntersGrantsFlyingOrAirbends(t *testing.T) {
	for _, mode := range []int{0, 1} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		appa := effectstest.AvB1ResolveCreature(t, g, me, "Appa, Loyal Sky Bison", "Legendary Creature — Bison Ally", avB1AppaOracle)
		if !effectstest.EffectiveAbilitiesContain(t, g, appa, "flying") {
			t.Fatal("Appa has printed flying")
		}
		c := effectstest.ModePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("mode %d: no mode prompt on entry", mode)
		}
		effectstest.AnswerMode(t, g, c, me.ID, mode)
		if mode == 1 {
			// "Another": Appa is no target for his own airbend.
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
				if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: appa}}); err == nil {
					t.Fatal("Appa airbent himself")
				}
			}
		}
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		switch mode {
		case 0:
			if !effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
				t.Error("the Bear should have flying until end of turn")
			}
		case 1:
			effectstest.AssertAirbent(t, g, bear, me.ID)
		}
	}
}

func TestAvatarB1AppaTriggersOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	appa := effectstest.B12Push(g, me.ID, "Appa, Loyal Sky Bison", "Legendary Creature — Bison Ally", avB1AppaOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(appa, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt on attack")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the Bear should have flying until end of turn")
	}
}

// --- Avatar Enthusiasts -----------------------------------------------------------

func TestAvatarB1EnthusiastsGrowOnAnotherAllyOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fan := effectstest.B12Push(g, me.ID, "Avatar Enthusiasts", "Creature — Human Peasant Ally", avB1EnthusiastsOracle, 2, 2)
	effectstest.AvB1ResolveCreature(t, g, me, "Plain Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AvB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Human Ally")
	if n := effectstest.AvB1Card(t, g, fan).Counters[game.CounterPlusOne]; n != 0 {
		t.Fatalf("a non-Ally or an opponent's Ally added a counter: %d", n)
	}
	effectstest.AvB1ResolveCreature(t, g, me, "My Ally", "Creature — Human Ally", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AvB1Card(t, g, fan).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("+1/+1 counters %d, want 1", n)
	}
}

// --- Compassionate Healer --------------------------------------------------------

func TestAvatarB1CompassionateHealerGainsAndScriesWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	healer := effectstest.B12Push(g, me.ID, "Compassionate Healer", "Creature — Human Cleric Ally", avB1HealerOracle, 2, 2)
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(healer, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if sc := effectstest.ScryChoiceFor(g, me.ID); sc == nil {
		t.Fatal("attacking tapped the Healer: a scry 1 is owed")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

// --- Curious Farm Animals --------------------------------------------------------

func TestAvatarB1CuriousFarmAnimalsDestroyAndGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	animals := effectstest.B12Push(g, me.ID, "Curious Farm Animals", "Creature — Boar Elk Bird Ox", avB1FarmAnimalsOracle, 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.AvB1Plains(g, me.ID, 1)
	if err := g.ActivateCatalogAbility(me.ID, animals, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(rock), Strict: true, AutoTap: true,
	}); err == nil {
		t.Fatal("one mana paid {2}")
	}
	effectstest.AvB1Plains(g, me.ID, 1)
	if err := g.ActivateCatalogAbility(me.ID, animals, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear), Strict: true, AutoTap: true,
	}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, animals, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(rock), Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(animals) || g.Battlefield.Contains(rock) {
		t.Error("the Animals are sacrificed and the artifact destroyed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (the sacrifice is a death)", me.Life, life+3)
	}
}

func TestAvatarB1CuriousFarmAnimalsActivatesWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	animals := effectstest.B12Push(g, me.ID, "Curious Farm Animals", "Creature — Boar Elk Bird Ox", avB1FarmAnimalsOracle, 1, 1)
	life := me.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.AvB1Plains(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, animals, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("up to one: zero targets is legal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Destined Confrontation ------------------------------------------------------

func TestAvatarB1DestinedConfrontationKeepsTotalPowerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBig := effectstest.B12Creature(g, me.ID, "My Big", "Creature — Wurm", 5, 5)
	myTwo := effectstest.B12Creature(g, me.ID, "My Two", "Creature — Bear", 2, 2)
	myOther := effectstest.B12Creature(g, me.ID, "My Other Two", "Creature — Bear", 2, 2)
	theirFour := effectstest.B12Creature(g, opp.ID, "Their Four", "Creature — Beast", 4, 4)
	theirOne := effectstest.B12Creature(g, opp.ID, "Their One", "Creature — Elf", 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Destined Confrontation", "Sorcery", avB1ConfrontationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mine := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if mine == nil || effectstest.ChooseCardsChoiceFor(g, opp.ID) == nil {
		t.Fatal("every player with creatures is asked")
	}
	if err := g.ResolveChooseCards(mine.ID, me.ID, []uuid.UUID{myBig}); err == nil {
		t.Fatal("a set with total power 5 was accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, myTwo, myOther)
	effectstest.AnswerChooseCards(t, g, opp.ID, theirFour)
	effectstest.PassPriorityAroundTable(t, g)
	for _, keep := range []uuid.UUID{myTwo, myOther, theirFour, rock} {
		if !g.Battlefield.Contains(keep) {
			t.Errorf("%s should be kept", keep)
		}
	}
	if g.Battlefield.Contains(myBig) || g.Battlefield.Contains(theirOne) {
		t.Error("everything not chosen is sacrificed")
	}
}

// --- Earth Kingdom Jailer --------------------------------------------------------

func TestAvatarB1EarthKingdomJailerExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.PushTypedCard(g, opp.ID, "Their Three", "Creature — Beast", "{2}{G}")
	small := effectstest.PushTypedCard(g, opp.ID, "Their Two", "Artifact", "{2}")
	mineBig := effectstest.PushTypedCard(g, me.ID, "My Three", "Enchantment", "{2}{W}")
	land := effectstest.PushTypedCard(g, opp.ID, "Their Land", "Land", "")

	jailer := effectstest.AvB1ResolveCreature(t, g, me, "Earth Kingdom Jailer", "Creature — Human Soldier Ally", avB1JailerOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the entry trigger")
	}
	for _, bad := range []uuid.UUID{small, mineBig, land} {
		if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(bad)); err == nil {
			t.Fatalf("%s is not a legal target", bad)
		}
	}
	effectstest.PickCard(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Exile.Contains(big) {
		t.Fatal("the target should be in exile")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(jailer) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AvB1Named(g, opp.ID, "Their Three") == uuid.Nil {
		t.Fatal("the exiled card should return when the Jailer leaves")
	}
}

// --- Earth Kingdom Protectors ----------------------------------------------------

func TestAvatarB1EarthKingdomProtectorsShieldAnotherAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Earth Kingdom Protectors", "Creature — Human Soldier Ally", avB1ProtectorsOracle, 1, 1)
	ally := effectstest.B12Creature(g, me.ID, "My Ally", "Creature — Human Ally", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ally", "Creature — Human Ally", 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, guard, "vigilance") {
		t.Error("printed vigilance is missing")
	}
	effectstest.AdvanceToMain(t, g)
	for _, bad := range []uuid.UUID{guard, bear, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bad), Strict: true}); err == nil {
			t.Fatalf("%s is not another Ally you control", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ally), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(guard) {
		t.Error("the Protectors are sacrificed as the cost")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, ally, "indestructible") {
		t.Error("the Ally should be indestructible until end of turn")
	}
}

// --- Enter the Avatar State ------------------------------------------------------

func TestAvatarB1EnterTheAvatarState(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AvB1Plains(g, me.ID, 1)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.AvB1Hand(me, game.Card{Name: "Enter the Avatar State", TypeLine: "Instant — Lesson", ManaCost: "{W}", OracleID: avB1EnterAvatarOracle})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(theirs), Strict: true, AutoTap: true}); err == nil {
		t.Fatal("an opponent's creature is not a creature you control")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(mine), Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB1Card(t, g, mine).HasSubtype("Avatar") {
		t.Error("the creature should be an Avatar until end of turn")
	}
	if !effectstest.AvB1Card(t, g, mine).HasSubtype("Bear") {
		t.Error("the Avatar type is in addition to its other types")
	}
	for _, kw := range []string{"flying", "first strike", "lifelink", "hexproof"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, mine, kw) {
			t.Errorf("missing %s", kw)
		}
	}
}
