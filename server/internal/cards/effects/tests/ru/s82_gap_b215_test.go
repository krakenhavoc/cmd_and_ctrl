package testsru

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB215JonOracle      = "c5c19d6e-16d5-4d24-be2d-c63646bdf00c"
	gapB215RilsaOracle    = "57f6a8c2-9d84-4640-9b92-8b8611e05773"
	gapB215BaneOracle     = "ca507eea-1b04-4488-813e-ca970640e227"
	gapB215AstarionOracle = "03048844-32bb-45a2-a754-73f75eaa4566"
	gapB215FelisaOracle   = "39144f72-789e-457a-a7e7-862f90c34e15"
	gapB215PriestOracle   = "d8070ee7-dbd1-4d38-9715-ba14e103e379"
	gapB215MintharaOracle = "fad0c174-4c4a-4662-bd1f-224886e826e1"
	gapB215NaliaOracle    = "1bf6248c-658c-4540-8bee-81af79d349be"
	gapB215DragonbornOrcl = "934094d6-c897-4ae2-89cd-08d106600d3c"
	gapB215BronzeOracle   = "8ce74097-97b1-493d-9585-bd1336f95927"
	gapB215MammothOracle  = "7a6db509-7ef4-483c-b8a0-62a4f3e16ef9"
)

func TestGapB215EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB215JonOracle:      "Jon Irenicus, Shattered One",
		gapB215RilsaOracle:    "Rilsa Rael, Kingpin",
		gapB215BaneOracle:     "Bane, Lord of Darkness",
		gapB215AstarionOracle: "Astarion, the Decadent",
		gapB215FelisaOracle:   "Felisa, Fang of Silverquill",
		gapB215PriestOracle:   "High Priest of Penance",
		gapB215MintharaOracle: "Minthara, Merciless Soul",
		gapB215NaliaOracle:    "Nalia de'Arnise",
		gapB215DragonbornOrcl: "Ambitious Dragonborn",
		gapB215BronzeOracle:   "Ancient Bronze Dragon",
		gapB215MammothOracle:  "Battle Mammoth",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB215SetLife(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, n-p.Life); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
}

func gapB215Experience(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			if err := g.AddPlayerCounterByForEffect(p.ID, p.ID, game.CounterExperience, 1); err != nil {
				t.Fatalf("experience counter: %v", err)
			}
		}
	})
}

// --- High Priest of Penance ---------------------------------------------

func TestGapB215HighPriestDestroysANonlandPermanentWhenDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	priest := effectstest.B12Push(g, me.ID, "High Priest of Penance", "Creature — Human Cleric", gapB215PriestOracle, 1, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Pr6Damage(t, g, victim, priest, 1)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, victim)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, victim); z != game.ZoneGraveyard {
		t.Errorf("their creature is in %s, want the graveyard", z)
	}
}

func TestGapB215HighPriestMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	priest := effectstest.B12Push(g, me.ID, "High Priest of Penance", "Creature — Human Cleric", gapB215PriestOracle, 1, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Pr6Damage(t, g, victim, priest, 1)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, victim)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, victim); z != game.ZoneBattlefield {
		t.Errorf("their creature is in %s although the controller declined", z)
	}
}

// --- Ambitious Dragonborn ------------------------------------------------

func TestGapB215DragonbornEntersWithTheGreatestPowerAsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 3, 3)
	effectstest.B17GraveyardCard(me, "Dead Giant", "Creature — Giant", "{4}{R}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].Name == "Dead Giant" {
				me.Graveyard.Cards[i].Power = 6
			}
		}
	})
	id := effectstest.B13PlayAs(t, g, 0, "Ambitious Dragonborn", "Creature — Dragon Barbarian", gapB215DragonbornOrcl)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 6 {
		t.Errorf("%d +1/+1 counters, want 6 (the Giant card in the graveyard)", n)
	}
}

func TestGapB215DragonbornCountsOnlyCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 4, 4)
	effectstest.B12Creature(g, opp.ID, "Their Titan", "Creature — Giant", 9, 9)
	id := effectstest.B13PlayAs(t, g, 0, "Ambitious Dragonborn", "Creature — Dragon Barbarian", gapB215DragonbornOrcl)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 4 {
		t.Errorf("%d +1/+1 counters, want 4", n)
	}
}

// --- Ancient Bronze Dragon -----------------------------------------------

func TestGapB215BronzeDragonPutsTheRollOnUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Ancient Bronze Dragon", "Creature — Elder Dragon", gapB215BronzeOracle, 7, 7)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	} else {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	na, nb := effectstest.CounterCount(g, a, game.CounterPlusOne), effectstest.CounterCount(g, b, game.CounterPlusOne)
	if na < 1 || na > 20 || na != nb {
		t.Errorf("counters %d and %d, want the same d20 result (1..20) on both", na, nb)
	}
	if n := effectstest.CounterCount(g, dragon, game.CounterPlusOne); n != 0 {
		t.Errorf("the Dragon itself got %d counters although it was not targeted", n)
	}
}

// --- Battle Mammoth ------------------------------------------------------

func TestGapB215MammothDrawsWhenAnOpponentTargetsYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Battle Mammoth", "Creature — Elephant", gapB215MammothOracle, 6, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}

func TestGapB215MammothIgnoresYourOwnSpellsAndPlayerTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Battle Mammoth", "Creature — Elephant", gapB215MammothOracle, 6, 5)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Error("a spell that targets a player (not a permanent) triggered the Mammoth")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, me.Hand.Size())
	}
}

func TestGapB215MammothHasForetell(t *testing.T) {
	spec, _ := effects.Lookup(gapB215MammothOracle)
	if len(spec.SpecialActions) != 1 {
		t.Errorf("%d special actions, want the foretell action", len(spec.SpecialActions))
	}
}

// --- Minthara ------------------------------------------------------------

func TestGapB215MintharaGivesCreaturesPlusOneForEachExperienceCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	min := effectstest.B12Push(g, me.ID, "Minthara, Merciless Soul", "Legendary Creature — Elf Cleric", gapB215MintharaOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, mine); p != 2 {
		t.Fatalf("with no counters my Bear is %d power, want 2", p)
	}
	gapB215Experience(t, g, me, 3)
	if p := effectstest.EffectivePower(t, g, mine); p != 5 {
		t.Errorf("my Bear is %d power, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, min); p != 5 {
		t.Errorf("Minthara is %d power, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("their Bear is %d power, want 2", p)
	}
}

func TestGapB215MintharaGivesAnExperienceCounterAtEndStepIfAPermanentLeft(t *testing.T) {
	for _, left := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Minthara, Merciless Soul", "Legendary Creature — Elf Cleric", gapB215MintharaOracle, 2, 2)
		if left {
			effectstest.B18Kill(t, g, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if left {
			want = 1
		}
		if got := effects.ExperienceCounters(g, me.ID); got != want {
			t.Errorf("left=%v: %d experience counters, want %d", left, got, want)
		}
	}
}

func TestGapB215MintharaWardChargesOneManaPerExperienceCounter(t *testing.T) {
	for _, exp := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		minth := effectstest.B12Push(g, me.ID, "Minthara, Merciless Soul", "Legendary Creature — Elf Cleric", gapB215MintharaOracle, 2, 2)
		gapB215Experience(t, g, me, exp)
		spell := effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, "",
			[]game.TargetRef{{Kind: game.TargetCard, ID: minth}})
		effectstest.PassPriorityAroundTable(t, g)
		asked := effectstest.HasPayUnlessFor(g, opp.ID)
		if exp == 0 && asked {
			t.Errorf("with no experience counters the ward asked for a payment")
		}
		if exp == 2 {
			if !asked {
				t.Fatalf("with two experience counters the ward did not ask for {2}")
			}
			c := effectstest.AnswerPayUnless(t, g, opp.ID, false)
			if !strings.Contains(c.PayCost, "2") {
				t.Errorf("the ward asked for %q, want {2}", c.PayCost)
			}
			if !effectstest.P7Countered(g, spell) {
				t.Errorf("the spell was not countered after declining")
			}
		}
	}
}

// --- Felisa --------------------------------------------------------------

func TestGapB215FelisaMakesOneTappedInklingPerCounterOnADyingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Felisa, Fang of Silverquill", "Legendary Creature — Vampire Wizard", gapB215FelisaOracle, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	_ = g.AddCounter(bear, game.CounterPlusOne, 2)
	_ = g.AddCounter(bear, "charge", 1)
	effectstest.B18Kill(t, g, bear)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Inkling" {
			n++
			if !c.Tapped || c.Power != 2 || c.Toughness != 1 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Errorf("Inkling %d/%d tapped=%v, want a tapped 2/1 flyer", c.Power, c.Toughness, c.Tapped)
			}
		}
	}
	if n != 3 {
		t.Errorf("%d Inklings, want 3 (two +1/+1 counters and a charge counter)", n)
	}
}

func TestGapB215FelisaIgnoresCreaturesWithoutCountersTokensAndOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Felisa, Fang of Silverquill", "Legendary Creature — Vampire Wizard", gapB215FelisaOracle, 3, 2)
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	_ = g.AddCounter(theirs, game.CounterPlusOne, 2)
	effectstest.B18Kill(t, g, plain)
	effectstest.B18Kill(t, g, theirs)
	if n := effectstest.B16CountNamed(g, "Inkling"); n != 0 {
		t.Errorf("%d Inklings, want none", n)
	}
}

// --- Astarion ------------------------------------------------------------

func gapB215ChooseMode(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestGapB215AstarionFriendsGainsLifeEqualToLifeGainedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Astarion, the Decadent", "Legendary Creature — Vampire Elf Rogue", gapB215AstarionOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5) })
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	gapB215ChooseMode(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d -> %d, want +5", life, me.Life)
	}
}

func TestGapB215AstarionFeedMakesAnOpponentLoseWhatTheyLostThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Astarion, the Decadent", "Legendary Creature — Vampire Elf Rogue", gapB215AstarionOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -4) })
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	gapB215ChooseMode(t, g, me, 0)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4", life, opp.Life)
	}
}

// --- Nalia de'Arnise -----------------------------------------------------

func TestGapB215NaliaOffersTheFourPartyTypesFromTheTopOfTheLibrary(t *testing.T) {
	spec, _ := effects.Lookup(gapB215NaliaOracle)
	seen := map[string]bool{}
	for _, p := range spec.CastPermissions {
		if p.Zone != game.ZoneLibrary || !p.TopOfLibraryOnly || !p.CastOnly {
			t.Errorf("permission %+v is not a cast from the top of the library", p)
		}
		for _, ty := range p.Filter.CreatureTypesAny {
			seen[ty] = true
		}
	}
	for _, ty := range []string{"Cleric", "Rogue", "Warrior", "Wizard"} {
		if !seen[ty] {
			t.Errorf("no permission for %s spells", ty)
		}
	}
	if spec.LibraryTopVisible != game.LibraryTopOwner {
		t.Error("the top card is not visible to its owner")
	}
}

func TestGapB215NaliaCastsAPartyTypeSpellFromTheTopButNotOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Nalia de'Arnise", "Legendary Creature — Human Rogue", gapB215NaliaOracle, 3, 3)
	bear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a Bear was cast from the top of the library")
	}
	wizard := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: wizard, Name: "Top Wizard", TypeLine: "Creature — Human Wizard", ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, wizard, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("casting a Wizard from the top: %v", err)
	}
}

func TestGapB215NaliaBuffsTheTeamOnlyWithAFullParty(t *testing.T) {
	for _, full := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Nalia de'Arnise", "Legendary Creature — Human Rogue", gapB215NaliaOracle, 3, 3)
		cleric := effectstest.B12Creature(g, me.ID, "Cleric", "Creature — Human Cleric", 1, 1)
		effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 1, 1)
		if full {
			effectstest.B12Creature(g, me.ID, "Wizard", "Creature — Human Wizard", 1, 1)
		}
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		n := effectstest.CounterCount(g, cleric, game.CounterPlusOne)
		if full && (n != 1 || !effectstest.HasEffectiveKeyword(t, g, cleric, "deathtouch")) {
			t.Errorf("full party: %d counters, deathtouch=%v", n, effectstest.HasEffectiveKeyword(t, g, cleric, "deathtouch"))
		}
		if !full && (n != 0 || effectstest.HasEffectiveKeyword(t, g, cleric, "deathtouch")) {
			t.Errorf("three types: %d counters, deathtouch=%v, want none", n, effectstest.HasEffectiveKeyword(t, g, cleric, "deathtouch"))
		}
	}
}

// --- Rilsa Rael ----------------------------------------------------------

func TestGapB215RilsaTakesTheInitiativeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B13PlayAs(t, g, 0, "Rilsa Rael, Kingpin", "Legendary Creature — Human Rogue", gapB215RilsaOracle)
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		if !effects.YouHaveTheInitiative(g, me.ID) {
			t.Error("the controller did not take the initiative")
		}
	})
}

func TestGapB215RilsaGrantsDeathtouchAndAfterADungeonMore(t *testing.T) {
	for _, completed := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Rilsa Rael, Kingpin", "Legendary Creature — Human Rogue", gapB215RilsaOracle, 2, 5)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if completed {
			me.CompletedDungeons = append(me.CompletedDungeons, "lost-mine-of-phandelver")
		}
		effectstest.DeclareAttack(t, g, opp.ID, bear)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.PickCard(t, g, me.ID, bear)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
			t.Errorf("completed=%v: the attacker lacks deathtouch", completed)
		}
		first, menace := effectstest.HasEffectiveKeyword(t, g, bear, "first strike"), effectstest.HasEffectiveKeyword(t, g, bear, "menace")
		power := effectstest.EffectivePower(t, g, bear)
		if completed && (!first || !menace || power != 7) {
			t.Errorf("after a dungeon: first strike=%v menace=%v power=%d, want true true 7", first, menace, power)
		}
		if !completed && (first || menace || power != 2) {
			t.Errorf("no dungeon: first strike=%v menace=%v power=%d, want false false 2", first, menace, power)
		}
	}
}

// --- Bane ----------------------------------------------------------------

func TestGapB215BaneIsIndestructibleAtHalfStartingLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bane := effectstest.B12Push(g, me.ID, "Bane, Lord of Darkness", "Legendary Creature — God", gapB215BaneOracle, 5, 2)
	gapB215SetLife(t, g, me, 21)
	if effectstest.HasEffectiveKeyword(t, g, bane, "indestructible") {
		t.Error("Bane is indestructible at 21 life")
	}
	gapB215SetLife(t, g, me, 20)
	if !effectstest.HasEffectiveKeyword(t, g, bane, "indestructible") {
		t.Error("Bane is not indestructible at 20 life")
	}
}

func gapB215BaneDies(t *testing.T, g *game.Game, me, opp *game.Player, toughness int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Bane, Lord of Darkness", "Legendary Creature — God", gapB215BaneOracle, 5, 2)
	victim := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, toughness)
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return victim
}

func TestGapB215BaneDrawsWhenTheOpponentSaysYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	gapB215BaneDies(t, g, me, opp, 3)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}

func TestGapB215BaneLetsYouPutACreatureWithEqualOrLesserToughnessWhenTheyDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B13HandCard(me, "Small Elf", "Creature — Elf", "")
	big := effectstest.B13HandCard(me, "Big Ogre", "Creature — Ogre", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			switch me.Hand.Cards[i].InstanceID {
			case small:
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 1, 3
			case big:
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 5, 4
			}
		}
	})
	gapB215BaneDies(t, g, me, opp, 3)
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("the controller was not offered a creature card from their hand")
	}
	for _, id := range pick.ChooseCards {
		if id == big {
			t.Error("a creature with greater toughness was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{small}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, small); z != game.ZoneBattlefield {
		t.Errorf("the small creature is in %s, want the battlefield", z)
	}
}

// --- Jon Irenicus --------------------------------------------------------

func TestGapB215JonGivesAnOpponentACreatureWithCountersTappedAndGoaded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Jon Irenicus, Shattered One", "Legendary Creature — Elf Wizard", gapB215JonOracle, 3, 3)
	gift := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: gift, Slot: 1}}); err != nil {
			t.Fatalf("creature clause: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, gift)
	if !ok {
		t.Fatal("the creature left the battlefield")
	}
	if c.Controller != opp.ID {
		t.Errorf("controller %v, want the opponent", c.Controller)
	}
	if n := effectstest.CounterCount(g, gift, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
	if !c.Tapped {
		t.Error("the creature is not tapped")
	}
	if !c.IsGoadedBy(me.ID) {
		t.Error("the creature is not goaded by Jon's controller")
	}
	if !g.CantBeSacrificedForEffect(gift) {
		t.Error("the creature can be sacrificed")
	}
}

func TestGapB215JonDrawsWhenACreatureYouOwnButDontControlAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Jon Irenicus, Shattered One", "Legendary Creature — Elf Wizard", gapB215JonOracle, 3, 3)
	stolen := effectstest.B12Creature(g, opp.ID, "Stolen Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stolen {
				g.Battlefield.Cards[i].Owner = me.ID
			}
		}
	})
	own := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, g.Seats[2].ID, own)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a creature I own and control drew a card (%d -> %d)", hand, me.Hand.Size())
	}
}
