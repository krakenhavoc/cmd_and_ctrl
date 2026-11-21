package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB90Renegade = "6b5565a2-be5c-46b1-99df-8094e77446cc"
	gapB90Whale    = "3e84d719-bb76-43a4-b61a-fee30c3aac68"
	gapB90Flood    = "677bd655-b240-4862-912c-29209fb295e1"
	gapB90Beetle   = "247a0c13-c611-4c00-bedf-ecf7ea6bb956"
	gapB90Rory     = "16a44586-ea1d-45f1-b9de-da1c11f729df"
	gapB90Sally    = "5c924f9f-84e2-4051-ae18-0c2cc57cf464"
	gapB90Eleventh = "721e8249-ff1b-4c15-964b-4ba2a96beb75"
	gapB90Fifth    = "992ddd9f-e0db-4383-909c-6dacca41d188"
	gapB90First    = "5611b70d-f8d1-4a8f-b9a2-a356264a7c46"
	gapB90Seventh  = "43e1b053-2cb9-4b1d-9e38-100cfa7267b6"
)

var gapB90Names = map[string]string{
	gapB90Renegade: "Renegade Silent",
	gapB90Whale:    "Star Whale",
	gapB90Flood:    "The Flood of Mars",
	gapB90Beetle:   "Time Beetle",
	gapB90Rory:     "Rory Williams",
	gapB90Sally:    "Sally Sparrow",
	gapB90Eleventh: "The Eleventh Doctor",
	gapB90Fifth:    "The Fifth Doctor",
	gapB90First:    "The First Doctor",
	gapB90Seventh:  "The Seventh Doctor",
}

func TestGapB90CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range gapB90Names {
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

func gapB90Clues(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasSubtype("Clue") {
			n++
		}
	}
	return n
}

func TestGapB90StarWhaleGrantsWardToOtherCreaturesAndHasSuspend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(gapB90Whale)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Counters != 6 || spec.SpecialActions[0].Cost != "{1}{U}" {
		t.Errorf("Star Whale's suspend is %+v, want suspend 6 for {1}{U}", spec.SpecialActions)
	}
	whale := effectstest.B12Push(g, me.ID, "Star Whale", "Creature — Alien Whale", gapB90Whale, 8, 8)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.MatWardPrompt(t, g, opp, bear) {
		t.Error("another creature of mine had no ward {2}")
	}
	if effectstest.MatWardPrompt(t, g, opp, whale) {
		t.Error("Star Whale itself has ward")
	}
}

func TestGapB90TimeBeetleTimeTravelsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	beetle := effectstest.B12Push(g, me.ID, "Time Beetle", "Creature — Alien Insect", gapB90Beetle, 1, 1)
	clock := effectstest.B12Permanent(g, me.ID, "Clock", "Artifact")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(clock, game.CounterTime, 1) })
	effectstest.DealCombatDamageToPlayer(g, beetle, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // add one time counter
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AvB1Card(t, g, clock).Counters[game.CounterTime]; n != 2 {
		t.Errorf("the Clock has %d time counters, want 2", n)
	}
}

func TestGapB90SallySparrowInvestigatesOnceEachTurnWhenAnotherCreatureLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sally := effectstest.B12Push(g, me.ID, "Sally Sparrow", "Legendary Creature — Human Detective", gapB90Sally, 2, 3)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	spec, _ := effects.Lookup(gapB90Sally)
	if len(spec.CastTimings) != 1 || !spec.CastTimings[0].Filter.CreatureOnly {
		t.Errorf("Sally's cast timing is %+v, want creature spells as though they had flash", spec.CastTimings)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB90Clues(g, me.ID); got != 1 {
		t.Fatalf("clues %d after another creature left, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(b) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB90Clues(g, me.ID); got != 1 {
		t.Errorf("clues %d after a second departure this turn, want still 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(sally) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB90Clues(g, me.ID); got != 1 {
		t.Errorf("Sally leaving made a Clue: %d", got)
	}
}

func TestGapB90FifthDoctorPumpsAndUntapsCreaturesThatDidNotAttackOrEnter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	doctor := effectstest.B12Push(g, me.ID, "The Fifth Doctor", "Legendary Creature — Time Lord Doctor", gapB90Fifth, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(idle, me.ID) })
	effectstest.StdB42Attack(t, g, opp.ID, attacker)
	// A creature that enters this turn.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Newcomer", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}, 1)
	})
	var newcomer uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Newcomer" {
			newcomer = c.InstanceID
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, idle, game.CounterPlusOne); got != 1 {
		t.Errorf("the idle creature has %d counters, want 1", got)
	}
	if effectstest.AvB1Card(t, g, idle).Tapped {
		t.Error("the idle creature was not untapped")
	}
	if got := effectstest.CountersOn(g, doctor, game.CounterPlusOne); got != 1 {
		t.Errorf("the Doctor has %d counters, want 1 (it did not attack)", got)
	}
	if got := effectstest.CountersOn(g, attacker, game.CounterPlusOne); got != 0 {
		t.Errorf("the creature that attacked got %d counters", got)
	}
	if got := effectstest.CountersOn(g, newcomer, game.CounterPlusOne); got != 0 {
		t.Errorf("the creature that entered got %d counters", got)
	}
}

func TestGapB90RenegadeSilentGoadsGrowsAndPhasesOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	silent := effectstest.B12Push(g, me.ID, "Renegade Silent", "Creature — Alien Horror", gapB90Renegade, 3, 3)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB1Card(t, g, foe).Goaded() {
		t.Error("the target is not goaded")
	}
	if !g.PhasedOut.Contains(silent) {
		t.Fatal("the Silent did not phase out")
	}
	for _, c := range g.PhasedOut.Cards {
		if c.InstanceID == silent && c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("the Silent has %v counters, want 1", c.Counters)
		}
	}
}

func TestGapB90FloodOfMarsMakesACreatureACopyAndALandAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	flood := effectstest.B12Push(g, me.ID, "The Flood of Mars", "Creature — Alien Zombie Horror", gapB90Flood, 3, 3)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	effectstest.StdB42Attack(t, g, opp.ID, flood)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, foe, "flood"); got != 1 {
		t.Errorf("the target has %d flood counters, want 1", got)
	}
	if c := effectstest.LayeredCard(t, g, foe); c.Name != "The Flood of Mars" || c.CurrentPower() != 3 {
		t.Errorf("the target is %q %d/%d, want a copy of The Flood of Mars", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
}

func TestGapB90FloodOfMarsMakesALandAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	flood := effectstest.B12Push(g, me.ID, "The Flood of Mars", "Creature — Alien Zombie Horror", gapB90Flood, 3, 3)
	land := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	effectstest.StdB42Attack(t, g, opp.ID, flood)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, land)
	if !c.HasSubtype("Island") || !c.HasSubtype("Mountain") {
		t.Errorf("the land is %q, want an Island in addition to its other types", c.TypeLine)
	}
	if got := effectstest.CountersOn(g, land, "flood"); got != 1 {
		t.Errorf("the land has %d flood counters, want 1", got)
	}
}

func TestGapB90EleventhDoctorExilesACardWithTimeCountersEqualToItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	doctor := effectstest.B12Push(g, me.ID, "The Eleventh Doctor", "Legendary Creature — Time Lord Doctor", gapB90Eleventh, 3, 2)
	spell := game.Card{InstanceID: uuid.New(), Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{2}{U}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(spell)
	effectstest.DealCombatDamageToPlayer(g, doctor, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, spell.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(spell.InstanceID)
	if !ok || !g.Exile.Contains(spell.InstanceID) {
		t.Fatal("the card was not exiled")
	}
	if c.Counters[game.CounterTime] != 3 || !game.CardHasSuspend(c) {
		t.Errorf("the exiled card has %d time counters (suspend %v), want 3 and suspend", c.Counters[game.CounterTime], game.CardHasSuspend(c))
	}
}

func TestGapB90EleventhDoctorMakesASmallCreatureUnblockableAndRefusesABigOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	doctor := effectstest.B12Push(g, me.ID, "The Eleventh Doctor", "Legendary Creature — Time Lord Doctor", gapB90Eleventh, 3, 2)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, doctor, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(big)}); err == nil {
		t.Fatal("targeted a creature with power 4")
	}
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, doctor, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small)})
	if effectstest.RestrictionsOf(t, g, small)&game.CantBeBlocked == 0 {
		t.Error("the target can be blocked")
	}
}

func TestGapB90RoryExilesItselfOnCastWithThreeTimeCountersAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.AvB1Hand(me, game.Card{Name: "Rory Williams", TypeLine: "Legendary Creature — Human Soldier", OracleID: gapB90Rory, Power: 3, Toughness: 3})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(id)
	if !ok || !g.Exile.Contains(id) {
		t.Fatal("Rory was not exiled")
	}
	if c.Counters[game.CounterTime] != 3 || !game.CardHasSuspend(c) {
		t.Errorf("Rory has %d time counters (suspend %v), want 3 and suspend", c.Counters[game.CounterTime], game.CardHasSuspend(c))
	}
	if g.Battlefield.Contains(id) {
		t.Error("Rory resolved onto the battlefield")
	}
	if gapB90Clues(g, me.ID) != 1 {
		t.Errorf("clues %d, want 1", gapB90Clues(g, me.ID))
	}
}

func TestGapB90FirstDoctorFetchesTARDISAndGrowsOnCascadeSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tardis := game.Card{InstanceID: uuid.New(), Name: "TARDIS", TypeLine: "Legendary Artifact — Vehicle", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Library.PushTop(tardis) })
	doctor := effectstest.Msc17Play(t, g, me, game.Card{Name: "The First Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", OracleID: gapB90First, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
	effectstest.AnswerSearchByID(t, g, me.ID, tardis.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(tardis.InstanceID) {
		t.Fatal("TARDIS was not put into the hand")
	}
	// A cascade spell: put a counter on the target.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := effectstest.AvB1Hand(me, game.Card{Name: "Bloodbraid Elf", TypeLine: "Creature — Elf Berserker", OracleID: "3f0c9466-5ab9-4205-a84f-b4b27b5a678e", Power: 3, Toughness: 2})
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d counters, want 1", got)
	}
	_ = doctor
}

func TestGapB90SeventhDoctorWrongGuessOffersTheFreeCastAndARightOneInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	doctor := effectstest.B12Push(g, me.ID, "The Seventh Doctor", "Legendary Creature — Time Lord Doctor", gapB90Seventh, 3, 6)
	spell := game.Card{InstanceID: uuid.New(), Name: "Big Creature", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(spell)
	effectstest.StdB42Attack(t, g, opp.ID, doctor)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, spell.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	// Mana value 5 is greater than 0 artifacts; the defender guesses "not greater".
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, spell.InstanceID)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB90Clues(g, me.ID) != 1 {
		t.Errorf("clues %d after declining the free cast, want 1", gapB90Clues(g, me.ID))
	}
}

func TestGapB90SeventhDoctorRightGuessInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	doctor := effectstest.B12Push(g, me.ID, "The Seventh Doctor", "Legendary Creature — Time Lord Doctor", gapB90Seventh, 3, 6)
	spell := game.Card{InstanceID: uuid.New(), Name: "Big Creature", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(spell)
	effectstest.StdB42Attack(t, g, opp.ID, doctor)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, spell.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB90Clues(g, me.ID) != 1 {
		t.Errorf("clues %d after a right guess, want 1", gapB90Clues(g, me.ID))
	}
	if !me.Hand.Contains(spell.InstanceID) {
		t.Error("the card left the hand")
	}
}
