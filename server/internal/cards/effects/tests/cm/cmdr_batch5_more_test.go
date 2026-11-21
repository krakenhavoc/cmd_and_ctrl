package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdrBitterReunionOracle    = "173e86f4-17f3-44f5-8073-78e7e2e12c4a"
	cmdrDragonhawkOracle       = "f7f435bc-1716-4412-9505-a1f2f2d83d08"
	cmdrGoddricOracle          = "35b2b172-5fbc-4e4d-a011-390ee1b6c848"
	cmdrMinionOracle           = "058062c7-071b-4b14-903d-ae8d2876b37f"
	cmdrOrbOracle              = "d9c36ed5-629c-4709-bff5-2f10fcde608a"
	cmdrSarkhanOracle          = "71cb48e2-2328-471c-a04a-7b6461f378f4"
	cmdrScourgeThroneOracle    = "06496a05-7f1a-402c-9594-304cd82bfefd"
	cmdrScourgeValkasOracle    = "0a147009-a57c-4ac1-ab81-d8bd8b737631"
	cmdrSkylineDespotOracle    = "c0ce54a1-14a1-45f3-91d0-c57b581afc63"
	cmdrTyrantsFamiliarOracle  = "63394e24-a456-4cc6-9885-8c62447b050a"
	cmdrChallengerTrollOracle  = "78cb1f05-edad-4430-9095-f3eabd80bf41"
	cmdrDungroveElderOracle    = "32083c9f-04c6-4b6f-9e24-cc8f3cae3e3e"
	cmdrElderGargarothOracle   = "397b9dcf-690a-4a58-8637-bb9baab7cc2f"
	cmdrMonstrousOnslaughtOrcl = "52be3da7-1e55-44c4-88bc-fb578a635ec0"
	cmdrParadiseDruidOracle    = "e119869d-1f66-402b-8251-d351965cf0d3"
	cmdrRamThroughOracle       = "4b1befff-48fa-47a3-832e-bdaf43495c02"
)

func TestCmdrBatch5And6CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdrSoullessOneOracle:      "Soulless One",
		cmdrTendrilsCorrOracle:     "Tendrils of Corruption",
		cmdrBitterReunionOracle:    "Bitter Reunion",
		cmdrBreathWeaponOracle:     "Breath Weapon",
		cmdrDragonhawkOracle:       "Dragonhawk, Fate's Tempest",
		cmdrGoddricOracle:          "Goddric, Cloaked Reveler",
		cmdrMagmaquakeOracle:       "Magmaquake",
		cmdrMinionOracle:           "Minion of the Mighty",
		cmdrOrbOracle:              "Orb of Dragonkind",
		cmdrSarkhanOracle:          "Sarkhan, Dragon Ascendant",
		cmdrScourgeThroneOracle:    "Scourge of the Throne",
		cmdrScourgeValkasOracle:    "Scourge of Valkas",
		cmdrSkylineDespotOracle:    "Skyline Despot",
		cmdrTyrantsFamiliarOracle:  "Tyrant's Familiar",
		cmdrChallengerTrollOracle:  "Challenger Troll",
		cmdrDungroveElderOracle:    "Dungrove Elder",
		cmdrElderGargarothOracle:   "Elder Gargaroth",
		cmdrMonstrousOnslaughtOrcl: "Monstrous Onslaught",
		cmdrParadiseDruidOracle:    "Paradise Druid",
		cmdrRamThroughOracle:       "Ram Through",
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

// --- Bitter Reunion ------------------------------------------------------

func TestBitterReunionDiscardsToDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	discard := effectstest.PushHandCard(g, me)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bitter Reunion", TypeLine: "Enchantment", ManaCost: "{1}{R}", OracleID: cmdrBitterReunionOracle})
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, discard)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand-1+2 {
		t.Errorf("hand = %d, want %d (discard one, draw two)", got, hand+1)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the chosen card was not discarded")
	}
}

func TestBitterReunionDecliningDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.PushHandCard(g, me)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bitter Reunion", TypeLine: "Enchantment", ManaCost: "{1}{R}", OracleID: cmdrBitterReunionOracle})
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("declining changed the hand: %d -> %d", hand, me.Hand.Size())
	}
}

func TestBitterReunionSacrificeGivesCreaturesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reunion := effectstest.PushCatalogPermanent(g, me.ID, "Bitter Reunion", "Enchantment", cmdrBitterReunionOracle, false)
	mine := cmdrB5Push(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	theirs := cmdrB5Push(g, opp.ID, "Ogre", "", "Creature — Ogre", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, reunion, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without paying {1}")
	}
	effectstest.FillPool(me, 1)
	effectstest.Pr7Activate(t, g, me.ID, reunion, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(reunion) {
		t.Error("the enchantment was not sacrificed")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "haste") {
		t.Error("your creature did not gain haste")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "haste") {
		t.Error("an opponent's creature gained haste")
	}
}

// --- Skyline Despot ------------------------------------------------------

func TestSkylineDespotMakesYouMonarchAndDragonsEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Skyline Despot", TypeLine: "Creature — Dragon", ManaCost: "{5}{R}{R}", OracleID: cmdrSkylineDespotOracle, Power: 5, Toughness: 5})
	if g.Monarch != me.ID {
		t.Fatal("entering did not make you the monarch")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Dragon"); n != 1 {
		t.Fatalf("%d Dragon tokens after the upkeep, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dragon" && (c.Power != 5 || c.Toughness != 5 || !c.HasSubtype("Dragon")) {
			t.Errorf("the token is %d/%d %s, want a 5/5 Dragon", c.Power, c.Toughness, c.TypeLine)
		}
	}
}

func TestSkylineDespotMakesNoDragonWhenYouAreNotTheMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Skyline Despot", "Creature — Dragon", cmdrSkylineDespotOracle, false)
	g.Monarch = opp.ID
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Dragon"); n != 0 {
		t.Errorf("%d Dragon tokens while an opponent is the monarch, want 0", n)
	}
}

// --- Scourge of Valkas ---------------------------------------------------

func TestScourgeOfValkasDealsDamageEqualToDragonsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdrB5Push(g, me.ID, "Other Dragon", "", "Creature — Dragon", 3, 3)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Scourge of Valkas", TypeLine: "Creature — Dragon", ManaCost: "{2}{R}{R}{R}", OracleID: cmdrScourgeValkasOracle, Power: 4, Toughness: 4})
	before := opp.Life
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d -> %d, want -2 (two Dragons)", before, opp.Life)
	}
}

func TestScourgeOfValkasTriggersForAnotherDragonAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valkas := effectstest.PushCatalogPermanent(g, me.ID, "Scourge of Valkas", "Creature — Dragon", cmdrScourgeValkasOracle, false)
	before := opp.Life
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Baby Dragon", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("another Dragon entering: life %d -> %d, want -2", before, opp.Life)
	}
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if g.Turn.Step == game.StepPrecombatMain {
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Error("a non-Dragon entering triggered the ability")
		}
	}
	basePower := effectstest.LayeredCard(t, g, valkas).CurrentPower()
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.Pr7Activate(t, g, me.ID, valkas, 0, game.ActivateAbilityParams{})
	if got := effectstest.LayeredCard(t, g, valkas).CurrentPower(); got != basePower+1 {
		t.Errorf("{R}: power %d -> %d, want +1", basePower, got)
	}
}

// --- Tyrant's Familiar ---------------------------------------------------

func TestTyrantsFamiliarLieutenantNeedsYourCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fam := cmdrB5Push(g, me.ID, "Tyrant's Familiar", cmdrTyrantsFamiliarOracle, "Creature — Dragon", 5, 5, "flying", "haste")
	if p, tt := cmdrPT(t, g, fam); p != 5 || tt != 5 {
		t.Fatalf("without a commander it is %d/%d, want 5/5", p, tt)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, IsCommander: true, Owner: me.ID, Controller: me.ID})
	g.RunStateChecksForTest()
	if p, tt := cmdrPT(t, g, fam); p != 7 || tt != 7 {
		t.Errorf("with your commander it is %d/%d, want 7/7", p, tt)
	}
}

func TestTyrantsFamiliarAttackDealsSevenToADefendersCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fam := cmdrB5Push(g, me.ID, "Tyrant's Familiar", cmdrTyrantsFamiliarOracle, "Creature — Dragon", 5, 5, "flying", "haste")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, IsCommander: true, Owner: me.ID, Controller: me.ID})
	victim := cmdrB5Push(g, opp.ID, "Wall", "", "Creature — Wall", 0, 9)
	bystander := cmdrB5Push(g, g.Seats[2].ID, "Bystander", "", "Creature — Bear", 2, 2)
	g.RunStateChecksForTest()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fam, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 7 {
		t.Errorf("the defender's creature took %d, want 7", got)
	}
	if effectstest.DamageMarkedOn(g, bystander) != 0 {
		t.Error("a creature of a player who is not defending was hit")
	}
}

func TestTyrantsFamiliarAttackWithoutACommanderDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fam := cmdrB5Push(g, me.ID, "Tyrant's Familiar", cmdrTyrantsFamiliarOracle, "Creature — Dragon", 5, 5, "flying", "haste")
	victim := cmdrB5Push(g, opp.ID, "Wall", "", "Creature — Wall", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fam, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 0 {
		t.Errorf("the creature took %d with no commander, want 0", got)
	}
}

// --- Scourge of the Throne -----------------------------------------------

func TestScourgeOfTheThroneDethronesAndGetsAnExtraCombatAgainstTheLeader(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scourge := cmdrB5Push(g, me.ID, "Scourge of the Throne", cmdrScourgeThroneOracle, "Creature — Dragon", 5, 5, "flying")
	other := cmdrB5Push(g, me.ID, "Attacker", "", "Creature — Bear", 2, 2)
	opp.Life = 45
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{scourge, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, scourge); got != 1 {
		t.Errorf("dethrone: %d +1/+1 counters, want 1", got)
	}
	effectstest.B16Tap(g, other)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after combat: %+v, want an additional combat", next)
	}
}

func TestScourgeOfTheThroneDoesNothingAgainstAPlayerBehindOnLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scourge := cmdrB5Push(g, me.ID, "Scourge of the Throne", cmdrScourgeThroneOracle, "Creature — Dragon", 5, 5, "flying")
	g.Seats[2].Life = 50
	opp.Life = 30
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(scourge, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, scourge); got != 0 {
		t.Errorf("%d counters attacking a player who is not leading, want 0", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step == game.StepBeginCombat {
		t.Error("an additional combat was granted against a player who is not leading")
	}
}

// --- Challenger Troll ----------------------------------------------------

func TestChallengerTrollLimitsBigCreaturesToOneBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdrB5Push(g, me.ID, "Challenger Troll", cmdrChallengerTrollOracle, "Creature — Troll", 6, 5)
	big := cmdrB5Push(g, me.ID, "Big", "", "Creature — Bear", 4, 4)
	small := cmdrB5Push(g, me.ID, "Small", "", "Creature — Bear", 3, 3)
	b1 := cmdrB5Push(g, opp.ID, "Blocker A", "", "Creature — Bear", 2, 2)
	b2 := cmdrB5Push(g, opp.ID, "Blocker B", "", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{big, small} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(b1, big); err != nil {
		t.Fatalf("first blocker on the 4-power creature: %v", err)
	}
	if err := g.DeclareBlocker(b2, big); err == nil {
		t.Error("a second creature blocked a creature with power 4 or greater")
	}
	if err := g.DeclareBlocker(b2, small); err != nil {
		t.Fatalf("blocker on the power-3 creature: %v", err)
	}
}

// --- Dungrove Elder / Paradise Druid ---------------------------------------

func TestDungroveElderCountsForestsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elder := cmdrB5Push(g, me.ID, "Dungrove Elder", cmdrDungroveElderOracle, "Creature — Treefolk", 0, 0, "hexproof")
	if p, tt := cmdrPT(t, g, elder); p != 0 || tt != 0 {
		t.Fatalf("with no Forest it is %d/%d, want 0/0", p, tt)
	}
	cmdrB5Push(g, me.ID, "Forest", "", "Basic Land — Forest", 0, 0)
	cmdrB5Push(g, me.ID, "Stomping Ground", "", "Land — Mountain Forest", 0, 0)
	cmdrB5Push(g, opp.ID, "Forest", "", "Basic Land — Forest", 0, 0)
	cmdrB5Push(g, me.ID, "Mountain", "", "Basic Land — Mountain", 0, 0)
	g.RunStateChecksForTest()
	if p, tt := cmdrPT(t, g, elder); p != 2 || tt != 2 {
		t.Errorf("with two Forests of its controller's it is %d/%d, want 2/2", p, tt)
	}
}

func TestParadiseDruidHasHexproofOnlyWhileUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.PushCatalogPermanent(g, me.ID, "Paradise Druid", "Creature — Elf Druid", cmdrParadiseDruidOracle, false)
	if !effectstest.EffectiveAbilitiesContain(t, g, druid, "hexproof") {
		t.Fatal("an untapped Druid lacks hexproof")
	}
	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 {
		t.Errorf("pool = %v, want one mana", got)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, druid, "hexproof") {
		t.Error("a tapped Druid still has hexproof")
	}
}

// --- Elder Gargaroth -----------------------------------------------------

func TestElderGargarothOffersThreeModesOnAttack(t *testing.T) {
	for mode, check := range map[int]func(t *testing.T, g *game.Game, me *game.Player, before [3]int){
		0: func(t *testing.T, g *game.Game, me *game.Player, _ [3]int) {
			if n := effectstest.CountNamed(g, "Beast"); n != 1 {
				t.Errorf("%d Beast tokens, want 1", n)
			}
		},
		1: func(t *testing.T, g *game.Game, me *game.Player, before [3]int) {
			if me.Life != before[0]+3 {
				t.Errorf("life %d -> %d, want +3", before[0], me.Life)
			}
		},
		2: func(t *testing.T, g *game.Game, me *game.Player, before [3]int) {
			if me.Hand.Size() != before[1]+1 {
				t.Errorf("hand %d -> %d, want +1", before[1], me.Hand.Size())
			}
		},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(me, "A", "B")
		garg := cmdrB5Push(g, me.ID, "Elder Gargaroth", cmdrElderGargarothOracle, "Creature — Beast", 6, 6, "reach", "vigilance", "trample")
		before := [3]int{me.Life, me.Hand.Size(), 0}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(garg, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		p := effectstest.ModePickChoiceFor(g, me.ID)
		if p == nil {
			t.Fatalf("mode %d: no mode prompt for the attack trigger (pending: %+v)", mode, g.PendingChoices)
		}
		effectstest.AnswerMode(t, g, p, me.ID, mode)
		effectstest.PassPriorityAroundTable(t, g)
		check(t, g, me, before)
	}
}

func TestElderGargarothTriggersOnBlockToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garg := cmdrB5Push(g, opp.ID, "Elder Gargaroth", cmdrElderGargarothOracle, "Creature — Beast", 6, 6, "reach", "vigilance", "trample")
	atk := cmdrB5Push(g, me.ID, "Attacker", "", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(garg, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if effectstest.ModePickChoiceFor(g, opp.ID) == nil {
		t.Fatalf("blocking did not trigger Elder Gargaroth (pending: %+v)", g.PendingChoices)
	}
}

// --- Monstrous Onslaught -------------------------------------------------

func TestMonstrousOnslaughtDividesTheGreatestPowerAmongTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdrB5Push(g, me.ID, "Ogre", "", "Creature — Ogre", 5, 5)
	cmdrB5Push(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	a := cmdrB5Push(g, opp.ID, "Target A", "", "Creature — Bear", 3, 3)
	b := cmdrB5Push(g, opp.ID, "Target B", "", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Monstrous Onslaught", TypeLine: "Sorcery", OracleID: cmdrMonstrousOnslaughtOrcl, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 2}})
	if err == nil {
		t.Fatal("a division that does not add up to the greatest power (5) was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 3, b: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("Target A took 3 and should die")
	}
	if !g.Battlefield.Contains(b) || effectstest.DamageMarkedOn(g, b) != 2 {
		t.Errorf("Target B should have 2 damage, has %d", effectstest.DamageMarkedOn(g, b))
	}
}

// --- Ram Through ---------------------------------------------------------

func castRamThrough(t *testing.T, g *game.Game, mine, theirs uuid.UUID) error {
	t.Helper()
	return effectstest.CastCatalogSpellErr(t, g, "Ram Through", "Instant", cmdrRamThroughOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
}

func TestRamThroughIsOneSidedAndTrampleSendsTheExcessToTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := cmdrB5Push(g, me.ID, "Plain Ogre", "", "Creature — Ogre", 5, 5)
	tramp := cmdrB5Push(g, me.ID, "Trampler", "", "Creature — Rhino", 5, 5, "trample")
	t1 := cmdrB5Push(g, opp.ID, "Victim One", "", "Creature — Bear", 2, 2)
	t2 := cmdrB5Push(g, opp.ID, "Victim Two", "", "Creature — Bear", 2, 3)
	life := opp.Life
	if err := castRamThrough(t, g, plain, t1); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(t1) {
		t.Error("the victim should die")
	}
	if opp.Life != life {
		t.Errorf("a creature without trample dealt excess: life %d -> %d", life, opp.Life)
	}
	if effectstest.DamageMarkedOn(g, plain) != 0 {
		t.Error("the creature you control was dealt damage back")
	}
	if err := castRamThrough(t, g, tramp, t2); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(t2) {
		t.Error("the second victim should die")
	}
	if opp.Life != life-2 {
		t.Errorf("trample excess: life %d -> %d, want -2 (5 power, 3 toughness)", life, opp.Life)
	}
}

func TestRamThroughRefusesYourCreatureAsTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := cmdrB5Push(g, me.ID, "A", "", "Creature — Ogre", 5, 5)
	b := cmdrB5Push(g, me.ID, "B", "", "Creature — Ogre", 5, 5)
	if err := castRamThrough(t, g, a, b); err == nil {
		t.Error("a creature you control was accepted as the target you don't control")
	}
}

// --- Dragonhawk, Fate's Tempest --------------------------------------------

func TestDragonhawkExilesXCardsAndBurnsForEachStillExiledAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	played := effectstest.SeedLibraryTop(me, "Free Bear", "Creature — Bear", "")
	kept := effectstest.SeedLibraryTop(me, "Kept Card", "Sorcery", "{9}")
	cmdrB5Push(g, me.ID, "Big One", "", "Creature — Giant", 4, 4)
	cmdrB5Push(g, me.ID, "Small One", "", "Creature — Bear", 3, 3)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Dragonhawk, Fate's Tempest", TypeLine: "Legendary Creature — Bird Dragon",
		ManaCost: "{3}{R}{R}", OracleID: cmdrDragonhawkOracle, Power: 5, Toughness: 5})
	// X = Dragonhawk (5) + Big One (4): the top two cards, in order.
	if !g.Exile.Contains(kept) || !g.Exile.Contains(played) {
		t.Fatalf("X=2 exiled kept=%v played=%v, want both", g.Exile.Contains(kept), g.Exile.Contains(played))
	}
	if err := g.CastSpell(me.ID, played, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing an exiled card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i]
		if i != 0 {
			want -= 2 // one card is still exiled
		}
		if p.Life != want {
			t.Errorf("seat %d life = %d, want %d", i, p.Life, want)
		}
	}
}

func TestDragonhawkExilesNothingWithoutAPowerFourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.SeedLibraryTop(me, "Top", "Sorcery", "{1}")
	d := cmdrB5Push(g, me.ID, "Dragonhawk, Fate's Tempest", cmdrDragonhawkOracle, "Legendary Creature — Bird Dragon", 3, 3)
	g.WithWriteLock(func() {
		_ = d
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(d, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("X = 0 exiled a card")
	}
}

// --- Goddric, Cloaked Reveler ----------------------------------------------

func TestGoddricBecomesADragonOnlyWhileCelebrating(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gid := cmdrEnterGoddric(t, g, me.ID)
	g.RunStateChecksForTest()
	c := effectstest.LayeredCard(t, g, gid)
	if c.HasSubtype("Dragon") || !c.HasSubtype("Human") || c.CurrentPower() != 3 {
		t.Fatalf("with one permanent entered Goddric is %s %d/%d, want a 3/3 Human Noble", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.CelebrationTokens(t, g, me.ID, 1)
	g.RunStateChecksForTest()
	c = effectstest.LayeredCard(t, g, gid)
	if !c.HasSubtype("Dragon") || c.HasSubtype("Human") || c.HasSubtype("Noble") {
		t.Errorf("celebrating, his creature types are not just Dragon: %+v", c.Effective())
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("celebrating he is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if !c.IsLegendary() {
		t.Error("he lost the legendary supertype")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, gid, "flying") || !effectstest.EffectiveAbilitiesContain(t, g, gid, "haste") {
		t.Errorf("celebrating he should have flying and haste: %v", effectstest.EffectiveAbilities(t, g, gid))
	}
}

func TestGoddricPumpsDragonsYouControlWhileCelebrating(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gid := cmdrEnterGoddric(t, g, me.ID)
	dragon := cmdrB5Push(g, me.ID, "Drake", "", "Creature — Dragon", 3, 3)
	bear := cmdrB5Push(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, gid, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Dragon pump was available before celebrating")
	}
	effectstest.CelebrationTokens(t, g, me.ID, 1)
	g.RunStateChecksForTest()
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.Pr7Activate(t, g, me.ID, gid, 0, game.ActivateAbilityParams{Strict: true})
	if p, _ := cmdrPT(t, g, dragon); p != 4 {
		t.Errorf("the Dragon's power = %d, want 4", p)
	}
	if p, _ := cmdrPT(t, g, bear); p != 2 {
		t.Errorf("the Bear's power = %d, want 2", p)
	}
	if p, _ := cmdrPT(t, g, gid); p != 5 {
		t.Errorf("Goddric (a Dragon) power = %d, want 5", p)
	}
}

// --- Minion of the Mighty --------------------------------------------------

func TestMinionOfTheMightyPutsADragonOntoTheBattlefieldAttackingWithPackTactics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	minion := cmdrB5Push(g, me.ID, "Minion of the Mighty", cmdrMinionOracle, "Creature — Kobold", 0, 1, "menace")
	big := cmdrB5Push(g, me.ID, "Big", "", "Creature — Giant", 6, 6)
	dragon := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: dragon, Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	notDragon := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: notDragon, Name: "Dragon Whelp Of Name Only", TypeLine: "Creature — Lizard", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{minion, big} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no choose-cards prompt for the Dragon (pending: %+v)", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != dragon {
		t.Fatalf("candidates = %v, want only the Dragon card", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dragon)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID); p != nil {
		if err := g.ResolveOptionPick(p.ID, me.ID, 0); err != nil {
			t.Fatalf("choose what it attacks: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardFor(g, dragon)
	if c == nil || !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("the Dragon should be on the battlefield tapped and attacking: %+v", c)
	}
}

func TestMinionOfTheMightyDoesNothingBelowSixPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	minion := cmdrB5Push(g, me.ID, "Minion of the Mighty", cmdrMinionOracle, "Creature — Kobold", 0, 1, "menace")
	small := cmdrB5Push(g, me.ID, "Small", "", "Creature — Bear", 5, 5)
	dragon := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: dragon, Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{minion, small} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || g.Battlefield.Contains(dragon) {
		t.Error("pack tactics triggered with total power 5")
	}
}

// --- Orb of Dragonkind -----------------------------------------------------

func TestOrbOfDragonkindManaPaysOnlyForDragons(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	orb := effectstest.PushCatalogPermanent(g, me.ID, "Orb of Dragonkind", "Artifact", cmdrOrbOracle, false)
	if err := g.ActivateManaAbility(me.ID, orb, 0, game.ManaAbilityParams{Colors: []string{"R", "U"}}); err == nil {
		t.Fatal("tapped without paying {1}")
	}
	effectstest.FillPool(me, 1)
	if err := g.ActivateManaAbility(me.ID, orb, 0, game.ManaAbilityParams{Colors: []string{"R", "U"}}); err != nil {
		t.Fatalf("tap the Orb: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 {
		t.Fatalf("pool = %v, want two mana", got)
	}
	cost, _ := game.ParseCost("{R}{U}")
	dragon := game.Card{Name: "Drake", TypeLine: "Creature — Dragon", ManaCost: "{R}{U}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{R}{U}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(dragon)) {
		t.Error("refused a Dragon spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(dragon)) {
		t.Error("refused an ability of a Dragon")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a non-Dragon spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("paid for a non-Dragon's ability")
	}
}

func TestOrbOfDragonkindSacrificeFindsADragonCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	orb := effectstest.PushCatalogPermanent(g, me.ID, "Orb of Dragonkind", "Artifact", cmdrOrbOracle, false)
	for i := 0; i < 5; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Creature — Bear", "")
	}
	dragon := effectstest.SeedLibraryTop(me, "Real Dragon", "Creature — Dragon", "{4}{R}{R}")
	effectstest.SeedLibraryTop(me, "Dragon's Name Only", "Sorcery", "{1}")
	hand := me.Hand.Size()
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.Pr7Activate(t, g, me.ID, orb, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(orb) {
		t.Error("the Orb was not sacrificed")
	}
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		for _, c := range pick.ChooseCards {
			if c != dragon {
				t.Errorf("offered %s, which is not a Dragon card", c)
			}
		}
		effectstest.AnswerChooseCards(t, g, me.ID, dragon)
		effectstest.PassPriorityAroundTable(t, g)
	} else {
		t.Fatal("no prompt to take a Dragon card")
	}
	if !me.Hand.Contains(dragon) || me.Hand.Size() != hand+1 {
		t.Errorf("the Dragon card did not reach the hand (hand %d -> %d)", hand, me.Hand.Size())
	}
}

// --- Sarkhan, Dragon Ascendant -----------------------------------------------

func TestSarkhanBeholdingADragonInHandRevealsItAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: dragon, Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Sarkhan, Dragon Ascendant", TypeLine: "Legendary Creature — Human Druid",
		ManaCost: "{1}{R}", OracleID: cmdrSarkhanOracle, Power: 2, Toughness: 2})
	effectstest.AnswerChooseCards(t, g, me.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
	if !me.Hand.Contains(dragon) {
		t.Error("beholding moved the Dragon card out of the hand")
	}
}

func TestSarkhanBeholdingADragonYouControlWorksAndDecliningMakesNoTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cmdrB5Push(g, me.ID, "Drake", "", "Creature — Dragon", 3, 3)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Sarkhan, Dragon Ascendant", TypeLine: "Legendary Creature — Human Druid",
		ManaCost: "{1}{R}", OracleID: cmdrSarkhanOracle, Power: 2, Toughness: 2})
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Treasure"); n != 0 {
		t.Errorf("declined the behold, got %d Treasure tokens", n)
	}
}

func TestSarkhanWithNothingToBeholdAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Sarkhan, Dragon Ascendant", TypeLine: "Legendary Creature — Human Druid",
		ManaCost: "{1}{R}", OracleID: cmdrSarkhanOracle, Power: 2, Toughness: 2})
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a behold prompt was raised with no Dragon to behold")
	}
	if n := effectstest.CountNamed(g, "Treasure"); n != 0 {
		t.Errorf("%d Treasure tokens with nothing beheld", n)
	}
}

func TestSarkhanGrowsAndBecomesAFlyingDragonWhenADragonEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sarkhan := effectstest.PushCatalogPermanent(g, me.ID, "Sarkhan, Dragon Ascendant", "Legendary Creature — Human Druid", cmdrSarkhanOracle, false)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if got := effectstest.PlusOnes(g, sarkhan); got != 0 {
		t.Fatalf("a Bear entering gave %d counters", got)
	}
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Drake", TypeLine: "Creature — Dragon", Power: 3, Toughness: 3})
	if got := effectstest.PlusOnes(g, sarkhan); got != 1 {
		t.Errorf("a Dragon entering gave %d counters, want 1", got)
	}
	c := effectstest.LayeredCard(t, g, sarkhan)
	if !c.HasSubtype("Dragon") || !c.HasSubtype("Human") || !c.HasSubtype("Druid") {
		t.Errorf("Sarkhan is %q, want a Human Druid Dragon", c.TypeLine)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, sarkhan, "flying") {
		t.Error("Sarkhan lacks flying")
	}
}

// cmdrEnterGoddric puts a 3/3 Goddric onto the battlefield through the
// entry pipeline, so his entry counts toward celebration.
func cmdrEnterGoddric(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		g.PlayerByIDForEffect(owner).Hand.PushTop(game.Card{
			InstanceID: id, Name: "Goddric, Cloaked Reveler", TypeLine: "Legendary Creature — Human Noble", OracleID: cmdrGoddricOracle,
			Power: 3, Toughness: 3, Owner: owner, Controller: owner,
		})
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: owner})
	})
	if err != nil || entered == uuid.Nil {
		t.Fatalf("put Goddric onto the battlefield: %v", err)
	}
	return entered
}
