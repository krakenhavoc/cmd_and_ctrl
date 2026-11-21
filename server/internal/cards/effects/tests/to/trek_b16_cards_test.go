package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b16_cards_test.go — Star Trek, batch 16 (#4973). Each test plays the
// real catalog entry through the engine.

const (
	trekB16GenesisOracle   = "ec763f36-3504-4bfe-8481-1a7d0bff8d8e"
	trekB16HortaOracle     = "e8ac0742-15f5-4cf0-888f-f1c2c97fc144"
	trekB16DiversityOracle = "c7d8ed66-d66e-42fd-9c3e-8c2d8194e009"
	trekB16KhaaanOracle    = "45b1a19b-6a0c-4534-9106-99c2e5fc30d6"
	trekB16KhanOracle      = "b7ea2089-abcd-4e60-8201-93801eae5084"
	trekB16LaAnOracle      = "ed3b80eb-8915-4428-89b8-90f4278f2733"
	trekB16LwaxanaOracle   = "e56472ec-6ae8-48b2-a4ce-0ed154161c06"
	trekB16MunitionsOracle = "0013f005-2b49-48a4-8a74-9bbdadc88c9f"
	trekB16NakedTimeOracle = "2cf79dd0-c9db-403e-969d-6c09e7aaa575"
	trekB16PahWraithOracle = "e47c340d-b20b-4103-955b-3f76424c43f8"
)

func TestTrekB16EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB16GenesisOracle:   "Genesis Effect",
		trekB16HortaOracle:     "Horta",
		trekB16DiversityOracle: "Infinite Diversity",
		trekB16KhaaanOracle:    "Khaaaaaaaaaaaannn!",
		trekB16KhanOracle:      "Khan, Despot by Design",
		trekB16LaAnOracle:      "La'An Noonien-Singh, Security",
		trekB16LwaxanaOracle:   "Lwaxana Troi, Charming Guest",
		trekB16MunitionsOracle: "Munitions Enthusiast",
		trekB16NakedTimeOracle: "The Naked Time",
		trekB16PahWraithOracle: "Pah-Wraith",
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

func trekB16Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	return g, g.Seats[0], g.Seats[1]
}

func trekB16Ref(id uuid.UUID, slot int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot}
}

func trekB16OnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := effectstest.BattlefieldCard(g, id)
	return ok
}

// --- Genesis Effect -------------------------------------------------------

func TestTrekB16GenesisEffectDamagesDestroysAndMakesATappedNewPlanet(t *testing.T) {
	g, me, opp := trekB16Main(t)
	victim := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 6, 6)
	mine := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.FloatMana(t, g, me, "{R}{C}{C}{C}{C}")
	effectstest.CastFromHandForTest(t, g, me, "Genesis Effect", "Instant", "{4}{R}", trekB16GenesisOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{trekB16Ref(victim, 0), trekB16Ref(mine, 1)}})
	effectstest.PassPriorityAroundTable(t, g)
	if trekB16OnBattlefield(g, victim) {
		t.Error("the 6/6 survived 6 damage")
	}
	if trekB16OnBattlefield(g, mine) {
		t.Error("the targeted land was not destroyed")
	}
	var planet *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "New Planet" {
			planet = c
		}
	}
	if planet == nil {
		t.Fatal("no New Planet token")
	}
	if !planet.Tapped || planet.Controller != me.ID || !planet.IsLand() || !planet.HasSubtype("Planet") || len(planet.Colors) != 0 {
		t.Errorf("New Planet = tapped %v controller %v land %v planet %v colors %v", planet.Tapped, planet.Controller == me.ID, planet.IsLand(), planet.HasSubtype("Planet"), planet.Colors)
	}
	mana := game.ManaAbilitiesForCard(*planet)
	if len(mana) != 1 || mana[0].Produced != "{W|U|B|R|G}" || !mana[0].TapCost {
		t.Errorf("New Planet mana abilities = %+v, want one {T}: any colour", mana)
	}
}

func TestTrekB16GenesisEffectNeedsALandYouControl(t *testing.T) {
	g, me, opp := trekB16Main(t)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	effectstest.FloatMana(t, g, me, "{R}{C}{C}{C}{C}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Genesis Effect", TypeLine: "Instant", ManaCost: "{4}{R}", OracleID: trekB16GenesisOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{trekB16Ref(victim, 0), trekB16Ref(theirs, 1)}}); err == nil {
		t.Fatal("Genesis Effect targeted an opponent's land")
	}
}

// The land left in response: the creature is still dealt damage and the
// token is still made (CR 608.2b).
func TestTrekB16GenesisEffectStillResolvesWithOneTargetGone(t *testing.T) {
	g, me, opp := trekB16Main(t)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.FloatMana(t, g, me, "{R}{C}{C}{C}{C}")
	effectstest.CastFromHandForTest(t, g, me, "Genesis Effect", "Instant", "{4}{R}", trekB16GenesisOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{trekB16Ref(victim, 0), trekB16Ref(mine, 1)}})
	effectstest.B27Kill(g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if trekB16OnBattlefield(g, victim) {
		t.Error("the creature was not dealt damage")
	}
	if effectstest.CountBattlefieldByName(g, "New Planet") != 1 {
		t.Error("no New Planet token")
	}
}

// --- Horta ------------------------------------------------------------------

func TestTrekB16HortaHasTrampleAndBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	horta := effectstest.B12Push(g, g.Seats[0].ID, "Horta", "Creature — Elemental", trekB16HortaOracle, 7, 6)
	if !effectstest.HasEffectiveKeyword(t, g, horta, "trample") {
		t.Error("Horta lacks trample")
	}
	g2 := effectstest.NewCatalogGame(t)
	me := g2.Seats[g2.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Ketria Triome", "Land — Forest Island Mountain"),
	)
	id, _ := effectstest.CycleFromHand(t, g2, "Horta", "Creature — Elemental", trekB16HortaOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("Horta was not discarded to cycle")
	}
	effectstest.PassPriorityAroundTable(t, g2)
	c := effectstest.SearchChoiceFor(g2, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search offers %v, want only the basic Mountain", c)
	}
	if err := g2.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic did not reach the hand")
	}
}

func TestTrekB16HortaCannotCycleWithoutTwoMana(t *testing.T) {
	g, me, _ := trekB16Main(t)
	id := effectstest.PushCatalogHandCard(me, "Horta", "Creature — Elemental", trekB16HortaOracle)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Horta landcycled for one mana")
	}
}

// --- Infinite Diversity -----------------------------------------------------

func trekB16CastDiversity(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.FloatMana(t, g, me, "{R}{R}{C}{C}{C}{C}{C}{C}{C}")
	return effectstest.CastFromHandForTest(t, g, me, "Infinite Diversity", "Sorcery", "{7}{R}{R}", trekB16DiversityOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
}

// X counts each creature type once among your non-Borg creatures: Human,
// Soldier and Elf make 3. A Borg and a changeling (every type, Borg
// included) add nothing, and an opponent's creature is not yours.
func TestTrekB16InfiniteDiversityCopiesTheCardXTimes(t *testing.T) {
	g, me, opp := trekB16Main(t)
	effectstest.B12Creature(g, me.ID, "Officer", "Creature — Human Soldier", 1, 1)
	effectstest.B12Creature(g, me.ID, "Ranger", "Creature — Human Elf", 1, 1)
	effectstest.B12Creature(g, me.ID, "Drone", "Creature — Borg Wizard", 1, 1)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{"changeling"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)

	spell := trekB16CastDiversity(t, g, me, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the targeted card is not exiled")
	}
	if !effectstest.InExile(g, spell) || me.Graveyard.Contains(spell) {
		t.Error("Infinite Diversity did not exile itself")
	}
	copies := effectstest.CardCopies(g)
	if len(copies) != 3 {
		t.Fatalf("%d copies, want 3 (Human, Soldier, Elf)", len(copies))
	}
	before := opp.Life
	for _, cp := range copies[:2] {
		if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
			t.Fatalf("free cast of a copy: %v", err)
		}
	}
	effectstest.DeclineResolutionCasts(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("life %d, want %d: two Bolt copies cast", opp.Life, before-6)
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left", n)
	}
}

func TestTrekB16InfiniteDiversityWithNoTypesCopiesNothing(t *testing.T) {
	g, me, _ := trekB16Main(t)
	effectstest.B12Creature(g, me.ID, "Drone", "Creature — Borg", 1, 1)
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	spell := trekB16CastDiversity(t, g, me, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bolt) || !effectstest.InExile(g, spell) {
		t.Error("the card and the spell are both exiled")
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies with X = 0", n)
	}
}

func TestTrekB16InfiniteDiversityTargetsOnlyYourInstantsAndSorceries(t *testing.T) {
	g, me, opp := trekB16Main(t)
	bear := effectstest.GyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	theirs := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	effectstest.FloatMana(t, g, me, "{R}{R}{C}{C}{C}{C}{C}{C}{C}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Infinite Diversity", TypeLine: "Sorcery", ManaCost: "{7}{R}{R}", OracleID: trekB16DiversityOracle, Owner: me.ID, Controller: me.ID})
	})
	for name, target := range map[string]uuid.UUID{"a creature card": bear, "an opponent's card": theirs} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err == nil {
			t.Errorf("Infinite Diversity targeted %s", name)
		}
	}
}

// --- Khaaaaaaaaaaaannn! ----------------------------------------------------

func TestTrekB16KhaaanDealsTwiceXAndExilesIfItDies(t *testing.T) {
	g, me, opp := trekB16Main(t)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 4, 4)
	effectstest.FloatMana(t, g, me, "{R}{C}{C}")
	effectstest.CastFromHandForTest(t, g, me, "Khaaaaaaaaaaaannn!", "Instant", "{X}{R}", trekB16KhaaanOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if trekB16OnBattlefield(g, victim) {
		t.Fatal("a 4/4 survived twice X = 4 damage")
	}
	if !effectstest.InExile(g, victim) || opp.Graveyard.Contains(victim) {
		t.Error("the creature died instead of being exiled")
	}
}

func TestTrekB16KhaaanTwiceOneDoesNotKillAThreeToughness(t *testing.T) {
	g, me, opp := trekB16Main(t)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.FloatMana(t, g, me, "{R}{C}")
	effectstest.CastFromHandForTest(t, g, me, "Khaaaaaaaaaaaannn!", "Instant", "{X}{R}", trekB16KhaaanOracle,
		game.CastSpellParams{Strict: true, XValue: 1, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok || c.DamageMarked != 2 {
		t.Fatalf("on battlefield %v, damage %d; want 2 damage marked", ok, c.DamageMarked)
	}
	// "If that creature would die this turn": a later death this turn is
	// an exile too.
	effectstest.B27Kill(g, victim)
	g.RunStateChecksForTest()
	if !effectstest.InExile(g, victim) {
		t.Error("the creature died later this turn and was not exiled")
	}
}

func TestTrekB16KhaaanCannotTargetAPlayer(t *testing.T) {
	g, me, opp := trekB16Main(t)
	effectstest.FloatMana(t, g, me, "{R}{C}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Khaaaaaaaaaaaannn!", TypeLine: "Instant", ManaCost: "{X}{R}", OracleID: trekB16KhaaanOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 1, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Fatal("Khaaaaaaaaaaaannn! targeted a player")
	}
}

// --- Khan, Despot by Design -------------------------------------------------

func TestTrekB16KhanPingsEachOpponentAndTheirCreatures(t *testing.T) {
	g, me, opp := trekB16Main(t)
	mine := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	tough := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	khan := effectstest.CastCreatureCard(t, g, "Khan, Despot by Design", "Legendary Creature — Human Rogue", trekB16KhanOracle, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !effectstest.HasEffectiveKeyword(t, g, khan, "trample") {
		t.Error("Khan lacks trample")
	}
	for _, s := range g.Seats {
		want := lives[s.ID] - 1
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("seat life %d, want %d", s.Life, want)
		}
	}
	if !trekB16OnBattlefield(g, mine) {
		t.Error("Khan damaged his controller's creature")
	}
	if trekB16OnBattlefield(g, theirs) {
		t.Error("the opponent's 1/1 survived")
	}
	if c, ok := effectstest.BattlefieldCard(g, tough); !ok || c.DamageMarked != 1 {
		t.Errorf("the opponent's 2/2 has damage %d, want 1", c.DamageMarked)
	}
}

// --- La'An Noonien-Singh, Security -------------------------------------------

func trekB16LaAnAttack(t *testing.T, g *game.Game, me *game.Player, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
}

func TestTrekB16LaAnStopsABlockerWithEightPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	laan := effectstest.B12Push(g, me.ID, "La'An Noonien-Singh, Security", "Legendary Creature — Human Soldier", trekB16LaAnOracle, 4, 3)
	effectstest.B12Creature(g, me.ID, "Home Guard", "Creature — Human", 4, 4) // stays home and still counts
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 5)
	trekB16LaAnAttack(t, g, me, laan)
	effectstest.AnswerPickTarget(t, g, blocker)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, blocker)
	if !game.Restricted(&c, game.CantBlock) {
		t.Error("the targeted creature can still block")
	}
}

func TestTrekB16LaAnNeedsEightTotalPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	laan := effectstest.B12Push(g, me.ID, "La'An Noonien-Singh, Security", "Legendary Creature — Human Soldier", trekB16LaAnOracle, 4, 3)
	effectstest.B12Creature(g, me.ID, "Ensign", "Creature — Human", 3, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 5)
	trekB16LaAnAttack(t, g, me, laan)
	if effectstest.PendingOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("La'An triggered with total power 7")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, blocker); game.Restricted(&c, game.CantBlock) {
		t.Error("the creature can't block without the trigger")
	}
}

// --- Lwaxana Troi, Charming Guest -------------------------------------------

func trekB16TwoNamedLands(g *game.Game, me *game.Player) {
	effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
}

// Lwaxana's chosen creature is Pah-Wraith: its attack trigger triggers an
// additional time while she stays tapped, so two lands' worth of bonus is
// applied twice. A creature she did not choose is not doubled.
func TestTrekB16LwaxanaDoublesTheChosenCreaturesTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trekB16TwoNamedLands(g, me)
	lwaxana := effectstest.B12Push(g, me.ID, "Lwaxana Troi, Charming Guest", "Legendary Creature — Betazoid Advisor", trekB16LwaxanaOracle, 2, 2)
	wraith := effectstest.B12Push(g, me.ID, "Pah-Wraith", "Creature — Wraith", trekB16PahWraithOracle, 4, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.RtActivate(t, g, me.ID, lwaxana, 0, effectstest.LtCardTarget(wraith), "C", "C")
	if !effectstest.RtTapped(t, g, lwaxana) {
		t.Fatal("Lwaxana did not tap")
	}
	trekB16LaAnAttack(t, g, me, wraith)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, wraith); p != 8 {
		t.Errorf("Pah-Wraith power %d, want 8 (4 + 2 + 2: the trigger doubled)", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, lwaxana); p != 6 {
		t.Errorf("Lwaxana power %d, want 6", p)
	}
}

func TestTrekB16LwaxanaStopsDoublingOnceSheUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trekB16TwoNamedLands(g, me)
	lwaxana := effectstest.B12Push(g, me.ID, "Lwaxana Troi, Charming Guest", "Legendary Creature — Betazoid Advisor", trekB16LwaxanaOracle, 2, 2)
	wraith := effectstest.B12Push(g, me.ID, "Pah-Wraith", "Creature — Wraith", trekB16PahWraithOracle, 4, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.RtActivate(t, g, me.ID, lwaxana, 0, effectstest.LtCardTarget(wraith), "C", "C")
	effectstest.RtUntap(g, lwaxana)
	trekB16LaAnAttack(t, g, me, wraith)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, wraith); p != 6 {
		t.Errorf("Pah-Wraith power %d, want 6: Lwaxana untapped, so one trigger", p)
	}
}

func TestTrekB16LwaxanaTargetsAnotherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lwaxana := effectstest.B12Push(g, me.ID, "Lwaxana Troi, Charming Guest", "Legendary Creature — Betazoid Advisor", trekB16LwaxanaOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	for name, target := range map[string]uuid.UUID{"herself": lwaxana, "an opponent's creature": theirs} {
		effectstest.B06AddMana(me, "C", "C")
		if err := g.ActivateCatalogAbility(me.ID, lwaxana, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.LtCardTarget(target)}); err == nil {
			t.Errorf("Lwaxana targeted %s", name)
		}
	}
}

func TestTrekB16LwaxanaMayStayTapped(t *testing.T) {
	spec, _ := effects.Lookup(trekB16LwaxanaOracle)
	if len(spec.UntapOptOuts) != 1 {
		t.Fatalf("Lwaxana declares %d untap opt-outs, want 1", len(spec.UntapOptOuts))
	}
}

// --- Munitions Enthusiast ---------------------------------------------------

func TestTrekB16MunitionsEnthusiastPingsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Munitions Enthusiast", "Creature — Klingon Warrior", trekB16MunitionsOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Munitions Enthusiast lacks trample")
	}
	life := opp.Life
	effectstest.B27Kill(g, id)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

// --- The Naked Time -----------------------------------------------------------

func TestTrekB16TheNakedTimeChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Snowy Peak", "Land — Mountain"),
	)
	for i := 0; i < 6; i++ { // draws for the turns the Saga waits
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Instant"})
	}
	saga := effectstest.CastCatalogSpell(t, g, "The Naked Time", "Enchantment — Saga", trekB16NakedTimeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// I: only the basic Mountain, onto the battlefield tapped.
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("chapter I search offers %v, want only the basic Mountain", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if card, ok := effectstest.BattlefieldCard(g, ids[0]); !ok || !card.Tapped {
		t.Fatalf("the Mountain is on the battlefield %v, tapped %v", ok, card.Tapped)
	}

	// II: a creature that was already there is not touched; one that
	// enters after the chapter gets a counter and haste.
	early := effectstest.B12Creature(g, me.ID, "Early Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoreCountersOn(g, saga) != 2 {
		t.Fatalf("lore counters %d, want 2", effectstest.LoreCountersOn(g, saga))
	}
	late := effectstest.CastCreatureCard(t, g, "Late Bear", "Creature — Bear", "", 2, 2)
	if n := effectstest.PlusOneCounters(g, late); n != 1 {
		t.Errorf("the entering creature has %d counters, want 1", n)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, late, "haste") {
		t.Error("the entering creature did not gain haste")
	}
	if n := effectstest.PlusOneCounters(g, early); n != 0 {
		t.Errorf("a creature already there has %d counters", n)
	}

	// III: sacrifice a land of your choice; then the Saga goes.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no sacrifice prompt at chapter III")
	}
	if len(ch.ChooseCards) != 1 || ch.ChooseCards[0] != ids[0] {
		t.Fatalf("sacrifice candidates %v, want the one land", ch.ChooseCards)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if trekB16OnBattlefield(g, ids[0]) {
		t.Error("the land was not sacrificed")
	}
	if trekB16OnBattlefield(g, saga) {
		t.Error("the Saga was not sacrificed after III")
	}
}

// --- Pah-Wraith -----------------------------------------------------------------

func TestTrekB16PahWraithPumpsByDifferentlyNamedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB16TwoNamedLands(g, me) // two Mountains and a Forest: two names
	effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	wraith := effectstest.B12Push(g, me.ID, "Pah-Wraith", "Creature — Wraith", trekB16PahWraithOracle, 4, 4)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Human", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, kw := range []string{"first strike", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, wraith, kw) {
			t.Errorf("Pah-Wraith lacks %s", kw)
		}
	}
	trekB16LaAnAttack(t, g, me, wraith)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{wraith: {6, 4}, ally: {3, 1}, theirs: {2, 2}} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%v is %d/%d, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
}

// Haste: Pah-Wraith attacks the turn it arrives.
func TestTrekB16PahWraithAttacksTheTurnItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	wraith := effectstest.CastCreatureCard(t, g, "Pah-Wraith", "Creature — Wraith", trekB16PahWraithOracle, 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wraith, g.Seats[1].ID); err != nil {
		t.Fatalf("a hasty Pah-Wraith could not attack: %v", err)
	}
}

// A dies trigger looks back (CR 603.10a): the chosen creature dying while
// Lwaxana is tapped still triggers twice.
func TestTrekB16LwaxanaDoublesTheChosenCreaturesDiesTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lwaxana := effectstest.B12Push(g, me.ID, "Lwaxana Troi, Charming Guest", "Legendary Creature — Betazoid Advisor", trekB16LwaxanaOracle, 2, 2)
	munitions := effectstest.B12Push(g, me.ID, "Munitions Enthusiast", "Creature — Klingon Warrior", trekB16MunitionsOracle, 3, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.RtActivate(t, g, me.ID, lwaxana, 0, effectstest.LtCardTarget(munitions), "C", "C")
	life := opp.Life
	effectstest.B27Kill(g, munitions)
	g.RunStateChecksForTest()
	asked := 0
	for i := 0; i < 4 && effectstest.PendingOfKind(g, game.PendingChoicePickTarget) != nil; i++ {
		effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
		asked++
	}
	effectstest.PassPriorityAroundTable(t, g)
	if asked != 2 {
		t.Errorf("%d target prompts, want 2: each instance picks its own target", asked)
	}
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d: the dies trigger triggers twice", opp.Life, life-2)
	}
}
