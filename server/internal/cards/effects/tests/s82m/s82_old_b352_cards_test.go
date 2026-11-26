package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b352_cards_test.go — Khans of Tarkir, batch 352 (#5554). Each test
// plays the real catalog entry and asserts what its text says.

const (
	oldB352Pinions = "c8cdc08a-975b-475f-8e30-85d61f6f3a9b" // Bladed Pinions, a real Equipment to attach
	oldB352Rough   = "10177654-c342-4537-a141-887a3fef13e8"
	oldB352Blaze   = "352844c5-2160-4787-97f1-9cd0fed060eb"
	oldB352Tusker  = "ccad8b03-5ff1-4895-94e0-0c96883cb875"
	oldB352Snake   = "1a668556-da98-4f1e-b1f7-26b7d63357a1"
	oldB352Bell    = "dad07f8c-1f24-4f56-b8a6-7c4b82e27e0f"
	oldB352Warden  = "5342ce88-4afc-431c-814a-054e5fca4b9e"
	oldB352Lich    = "7e4cbfd1-2393-4643-bc05-022d06b30210"
	oldB352EndH    = "7b2e2c4b-b8fd-440c-8a79-de6532022745"
	oldB352Field   = "baf8310f-5baf-459e-95ae-6b43a1d47e24"
	oldB352Corps   = "3b93427c-5b40-4119-96ae-7dce3ca2bbb9"
	oldB352Lens    = "7df50107-25d8-47dc-9a13-27b257cfbf6f"
	oldB352Throne  = "93e40d56-18fd-42e1-a598-636aab79b836"
	oldB352Quiet   = "f87dfde4-8c57-4543-826c-6b3cf9c11b90"
)

func TestOldB352CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB352Rough: "Mardu Roughrider", oldB352Blaze: "Mardu Blazebringer", oldB352Tusker: "Avalanche Tusker", oldB352Snake: "Molting Snakeskin", oldB352Bell: "Singing Bell Strike", oldB352Warden: "Warden of the Eye", oldB352Lich: "Kheru Lich Lord", oldB352EndH: "End Hostilities", oldB352Field: "Suspension Field", oldB352Corps: "Armament Corps", oldB352Lens: "Lens of Clarity", oldB352Throne: "Dragon Throne of Tarkir", oldB352Quiet: "Quiet Contemplation",
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

func oldB352Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func TestOldB352RoughriderStopsTheTargetBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	rider := effectstest.ReqCreature(g, me.ID, "Mardu Roughrider", oldB352Rough, 5, 4, "haste")
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
	other := effectstest.ReqCreature(g, opp.ID, "Other", "", 2, 2)
	effectstest.AttackAndTrigger(t, g, rider)
	effectstest.PickCard(t, g, me.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, wall)&game.CantBlock == 0 {
		t.Error("the target can block")
	}
	if effectstest.RestrictionsOf(t, g, other)&game.CantBlock != 0 {
		t.Error("an untargeted creature can't block")
	}
}

func TestOldB352BlazebringerIsSacrificedAtEndOfCombatWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	id := effectstest.B12Push(g, me.ID, "Mardu Blazebringer", "Creature — Ogre Warrior", oldB352Blaze, 4, 4)
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, id) {
		t.Fatal("the Blazebringer left before the end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, id) {
		t.Error("the Blazebringer survived combat")
	}
}

func TestOldB352TuskerForcesTheTargetToBlockIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	tusker := effectstest.ReqCreature(g, me.ID, "Avalanche Tusker", oldB352Tusker, 6, 4, "haste")
	other := effectstest.ReqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := effectstest.ReqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	effectstest.AttackAndTrigger(t, g, tusker, other)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OnToBlockers(t, g)
	effectstest.ReqRefusal(t, effectstest.ReqBlock(t, g, other, victim))
	if err := effectstest.ReqBlock(t, g, tusker, victim); err != nil {
		t.Fatal(err)
	}
}

func TestOldB352SnakeskinPumpsAndGrantsRegeneration(t *testing.T) {
	g, me, _ := oldB352Table(t)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Molting Snakeskin", oldB352Snake, host)
	if p := effectstest.EffectivePower(t, g, host); p != 4 {
		t.Fatalf("enchanted power %d, want 4", p)
	}
	effectstest.FloatForTest(g, me, "BBB")
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, host)
	if !effectstest.BattlefieldHas(g, host) {
		t.Error("the regeneration shield did not save the creature")
	}
}

func TestOldB352SingingBellStrikeTapsLocksAndGrantsAnUntap(t *testing.T) {
	g, me, opp := oldB352Table(t)
	host := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Singing Bell Strike", oldB352Bell, host)
	if !effectstest.PermanentTapped(g, host) {
		t.Fatal("the enchanted creature did not tap as the Aura entered")
	}
	_ = me
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if !effectstest.PermanentTapped(g, host) {
		t.Fatal("the enchanted creature untapped in its controller's untap step")
	}
	effectstest.FloatForTest(g, opp, "WWWWWW")
	if err := g.ActivateCatalogAbility(opp.ID, host, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{6}: untap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PermanentTapped(g, host) {
		t.Error("the {6} ability did not untap the creature")
	}
}

func TestOldB352WardenReturnsOnlyANoncreatureNonlandCard(t *testing.T) {
	g, me, _ := oldB352Table(t)
	spell := effectstest.PushGraveyardPermanent(me, "A Sorcery", "Sorcery", "{1}")
	rock := effectstest.PushGraveyardPermanent(me, "A Rock", "Artifact", "{1}")
	bear := effectstest.PushGraveyardPermanent(me, "A Bear", "Creature — Bear", "{1}")
	forest := effectstest.PushGraveyardPermanent(me, "A Forest", "Basic Land — Forest", "")
	effectstest.CastAndResolveCreature(t, g, "Warden of the Eye", "Creature — Djinn Wizard", oldB352Warden)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, id := range []uuid.UUID{bear, forest} {
		if effectstest.HasID(p.PickTargetCards, id) {
			t.Errorf("a creature or land card %s was offered", id)
		}
	}
	for _, id := range []uuid.UUID{spell, rock} {
		if !effectstest.HasID(p.PickTargetCards, id) {
			t.Errorf("a noncreature, nonland card %s was not offered", id)
		}
	}
	effectstest.PickCard(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == spell {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the sorcery card did not return to hand")
	}
}

func TestOldB352KheruLichLordReturnsARandomCreatureWithKeywordsAndExilesIt(t *testing.T) {
	g, me, _ := oldB352Table(t)
	effectstest.PoolPush(g, me.ID, "Kheru Lich Lord", oldB352Lich, 4, 4)
	only := effectstest.AdGraveyardCreature(me, "Dead Bear")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.FloatForTest(g, me, "BBB")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, only) {
		t.Fatal("the creature card did not return")
	}
	for _, kw := range []string{"flying", "trample", "haste"} {
		if !effectstest.AvB22Has(t, g, only, kw) {
			t.Errorf("the returned creature lacks %s", kw)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, only) {
		t.Error("the returned creature was not exiled at the end step")
	}
	if z := g.FindCardZoneForEffect(only); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the returned creature is in %v, want exile", z)
	}
}

func TestOldB352KheruLichLordDecliningLeavesTheGraveyardAlone(t *testing.T) {
	g, me, _ := oldB352Table(t)
	effectstest.PoolPush(g, me.ID, "Kheru Lich Lord", oldB352Lich, 4, 4)
	only := effectstest.AdGraveyardCreature(me, "Dead Bear")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, only) {
		t.Error("the creature returned without the {2}{B}")
	}
}

func TestOldB352EndHostilitiesDestroysCreaturesAndWhatIsAttachedToThem(t *testing.T) {
	g, me, opp := oldB352Table(t)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	sword := effectstest.SeedEquipment(g, me.ID, "Bladed Pinions", oldB352Pinions)
	effectstest.EquipTo(t, g, me.ID, sword, mine)
	loose := effectstest.SeedEquipment(g, opp.ID, "Loose Pinions", oldB352Pinions)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "End Hostilities", "Sorcery", oldB352EndH, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, mine, sword} {
		if effectstest.BattlefieldHas(g, id) {
			t.Errorf("%s survived", id)
		}
	}
	for _, id := range []uuid.UUID{loose, rock} {
		if !effectstest.BattlefieldHas(g, id) {
			t.Errorf("%s was destroyed though it is not attached to a creature", id)
		}
	}
}

func TestOldB352SuspensionFieldExilesATougherCreatureUntilItLeaves(t *testing.T) {
	g, me, opp := oldB352Table(t)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	field := effectstest.CastCatalogSpell(t, g, "Suspension Field", "Enchantment", oldB352Field, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, small) || !effectstest.HasID(p.PickTargetCards, big) {
		t.Fatalf("legal targets %v, want only the toughness 4 creature", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, big) {
		t.Fatal("the target was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(field) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Big", opp.ID); n != 1 {
		t.Errorf("after the enchantment left, %d of the exiled creature under its owner's control, want 1", n)
	}
}

func TestOldB352ArmamentCorpsDistributesTwoCounters(t *testing.T) {
	g, me, _ := oldB352Table(t)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	corps := effectstest.CastCatalogSpell(t, g, "Armament Corps", "Creature — Human Soldier", oldB352Corps, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the enters trigger did not ask for targets")
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{corps: 1, other: 1}, corps, other)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{corps, other} {
		if n := effectstest.E2Card(t, g, id).Counters[game.CounterPlusOne]; n != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", id, n)
		}
	}
}

func TestOldB352LensOfClarityDeclaresBothLooks(t *testing.T) {
	spec, ok := effects.Lookup(oldB352Lens)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.LibraryTopVisible != game.LibraryTopOwner || !spec.LooksAtFaceDownCreatures {
		t.Errorf("library top %v, face-down look %v", spec.LibraryTopVisible, spec.LooksAtFaceDownCreatures)
	}
}

func TestOldB352DragonThroneGivesDefenderAndTheRallyAbility(t *testing.T) {
	g, me, _ := oldB352Table(t)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 3, 3)
	pal := effectstest.B12Creature(g, me.ID, "Pal", "Creature — Bear", 1, 1)
	throne := effectstest.PushCatalogPermanent(g, me.ID, "Dragon Throne of Tarkir", "Legendary Artifact — Equipment", oldB352Throne, false)
	effectstest.EquipTo(t, g, me.ID, throne, host)
	if !effectstest.AvB22Has(t, g, host, "defender") {
		t.Fatal("the equipped creature lacks defender")
	}
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("rally: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, pal); p != 4 {
		t.Errorf("the other creature has power %d, want 1+3", p)
	}
	if !effectstest.AvB22Has(t, g, pal, "trample") {
		t.Error("the other creature lacks trample")
	}
	if p := effectstest.EffectivePower(t, g, host); p != 3 || effectstest.AvB22Has(t, g, host, "trample") {
		t.Errorf("the equipped creature changed: power %d", p)
	}
}

func TestOldB352QuietContemplationTapsAndFreezesForOne(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g, me, opp := oldB352Table(t)
		effectstest.PushCatalogPermanent(g, me.ID, "Quiet Contemplation", "Enchantment", oldB352Quiet, false)
		victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
		effectstest.PickCard(t, g, me.ID, victim)
		effectstest.FloatForTest(g, me, "C")
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.PermanentTapped(g, victim); got != pay {
			t.Errorf("pay=%v: the creature is tapped = %v", pay, got)
		}
	}
}
