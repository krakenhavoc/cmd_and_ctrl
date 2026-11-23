package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b14_cards_test.go — Innistrad Remastered creatures and
// enchantments, GAP batch 14 (#5948). Each test plays the real catalog
// entry through the engine.

const (
	gapMausoleumGuardOracle    = "64f3af46-f8ab-4ba4-9dff-7412cefb4eea"
	gapNiblisOfTheUrnOracle    = "3ea7a555-d6fa-4cc9-9348-784e6f9a8cfc"
	gapSlayerOfTheWickedOracle = "d93299d4-3aaf-48fb-8227-26203963c8b9"
	gapSubjugatorAngelOracle   = "88b1e06c-5899-4e83-8204-e2c32c0c6aff"
	gapTriskaidekaOracle       = "d5b131c7-29d4-4745-95e3-70bdf9bf54b3"
	gapBloodMistOracle         = "22188333-69b6-4685-b96b-b8a102ad5531"
	gapBurningVengeanceOracle  = "ef1fb174-9a34-43d0-8caf-9c1938147a12"
	gapSpontMutationOracle     = "ab78fa13-f599-4d08-bd27-d3d8740ae58f"
	gapBoundByMoonsilverOracle = "7a98a1d1-7366-4b38-9047-48269790d328"
	gapFaithUnbrokenOracle     = "a57fe358-baf3-4eb5-9417-8945b41f84de"
	gapLunarchMantleOracle     = "0da104ae-c46c-4106-919f-d975721d42d9"
)

func TestGapB14EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapMausoleumGuardOracle:    "Mausoleum Guard",
		gapNiblisOfTheUrnOracle:    "Niblis of the Urn",
		gapSlayerOfTheWickedOracle: "Slayer of the Wicked",
		gapSubjugatorAngelOracle:   "Subjugator Angel",
		gapTriskaidekaOracle:       "Triskaidekaphobia",
		gapBloodMistOracle:         "Blood Mist",
		gapBurningVengeanceOracle:  "Burning Vengeance",
		gapSpontMutationOracle:     "Spontaneous Mutation",
		gapBoundByMoonsilverOracle: "Bound by Moonsilver",
		gapFaithUnbrokenOracle:     "Faith Unbroken",
		gapLunarchMantleOracle:     "Lunarch Mantle",
	}
	for oracle, name := range want {
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

func TestGapB14MausoleumGuardLeavesTwoFlyingSpiritsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guard := effectstest.B12Push(g, me.ID, "Mausoleum Guard", "Creature — Human Scout", gapMausoleumGuardOracle, 2, 2)
	if got := effectstest.B27CountNamed(g, me.ID, "Spirit"); got != 0 {
		t.Fatalf("%d Spirits before it died", got)
	}
	effectstest.B27Kill(g, guard)
	effectstest.PassPriorityAroundTable(t, g)
	spirits := effectstest.BattlefieldIDsNamed(g, "Spirit")
	if len(spirits) != 2 {
		t.Fatalf("%d Spirit tokens, want 2", len(spirits))
	}
	for _, id := range spirits {
		effectstest.AssertKeywords(t, g, id, "flying")
		if effectstest.EffectivePower(t, g, id) != 1 || effectstest.EffectiveToughness(t, g, id) != 1 {
			t.Error("a Spirit is not 1/1")
		}
	}
}

func TestGapB14NiblisOfTheUrnMayTapATargetWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	niblis := effectstest.PushCatalogPermanent(g, me.ID, "Niblis of the Urn", "Creature — Spirit", gapNiblisOfTheUrnOracle, false)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(niblis, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the target was not tapped")
	}
}

func TestGapB14NiblisOfTheUrnMayDeclineToTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	niblis := effectstest.PushCatalogPermanent(g, me.ID, "Niblis of the Urn", "Creature — Spirit", gapNiblisOfTheUrnOracle, false)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(niblis, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedForTest(t, g, bear) {
		t.Error("the target was tapped after a no")
	}
}

func TestGapB14SlayerOfTheWickedDestroysAVampireWerewolfOrZombie(t *testing.T) {
	for _, tl := range []string{"Creature — Vampire", "Creature — Werewolf", "Creature — Zombie Warrior"} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		victim := effectstest.B12Creature(g, opp.ID, "Victim", tl, 2, 2)
		effectstest.CastHoldCreature(t, g, "Slayer of the Wicked", "Creature — Human Soldier", gapSlayerOfTheWickedOracle, 3, 2)
		effectstest.PickCard(t, g, g.Seats[0].ID, victim)
		effectstest.AnswerLatestTriggerPrompt(t, g, g.Seats[0].ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) {
			t.Errorf("%s survived", tl)
		}
	}
}

func TestGapB14SlayerOfTheWickedLeavesOtherCreaturesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Slayer of the Wicked", "Creature — Human Soldier", gapSlayerOfTheWickedOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a Bear was destroyed")
	}
}

func TestGapB14SlayerOfTheWickedMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Zombie", 2, 2)
	effectstest.CastHoldCreature(t, g, "Slayer of the Wicked", "Creature — Human Soldier", gapSlayerOfTheWickedOracle, 3, 2)
	effectstest.PickCard(t, g, g.Seats[0].ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, g.Seats[0].ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the Zombie was destroyed after a no")
	}
}

func TestGapB14SubjugatorAngelTapsOnlyYourOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Subjugator Angel", "Creature — Angel", gapSubjugatorAngelOracle, 4, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, theirs) || !effectstest.TappedForTest(t, g, other) {
		t.Error("an opponent's creature stayed untapped")
	}
	if effectstest.TappedForTest(t, g, mine) {
		t.Error("my creature was tapped")
	}
}

func gapTriskaidekaTable(t *testing.T, lives ...int) *game.Game {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Triskaidekaphobia", "Enchantment", gapTriskaidekaOracle, 0, 0)
	g.WithWriteLock(func() {
		for i, l := range lives {
			g.Seats[i].Life = l
		}
	})
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	return g
}

func TestGapB14TriskaidekaphobiaGainModeKillsThirteensThenGains(t *testing.T) {
	g := gapTriskaidekaTable(t, 20, 13, 14, 12)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, g.Seats[0].ID), g.Seats[0].ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Seats[1].Eliminated {
		t.Error("the player at exactly 13 did not lose")
	}
	for _, i := range []int{0, 2, 3} {
		if g.Seats[i].Eliminated {
			t.Errorf("seat %d lost", i)
		}
	}
	for i, want := range map[int]int{0: 21, 2: 15, 3: 13} {
		if g.Seats[i].Life != want {
			t.Errorf("seat %d life %d, want %d", i, g.Seats[i].Life, want)
		}
	}
}

func TestGapB14TriskaidekaphobiaLoseModeDropsEveryoneByOne(t *testing.T) {
	g := gapTriskaidekaTable(t, 14, 13, 20, 12)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, g.Seats[0].ID), g.Seats[0].ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Seats[1].Eliminated {
		t.Error("the player at exactly 13 did not lose")
	}
	for i, want := range map[int]int{0: 13, 2: 19, 3: 11} {
		if g.Seats[i].Life != want {
			t.Errorf("seat %d life %d, want %d", i, g.Seats[i].Life, want)
		}
	}
}

func TestGapB14TriskaidekaphobiaWithNobodyAtThirteenOnlyChangesLife(t *testing.T) {
	g := gapTriskaidekaTable(t, 20, 12, 14, 30)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, g.Seats[0].ID), g.Seats[0].ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for i, want := range []int{21, 13, 15, 31} {
		if g.Seats[i].Eliminated || g.Seats[i].Life != want {
			t.Errorf("seat %d: eliminated %v life %d, want %d", i, g.Seats[i].Eliminated, g.Seats[i].Life, want)
		}
	}
}

func TestGapB14TriskaidekaphobiaTwoLastPlayersAtThirteenBothLose(t *testing.T) {
	g := gapTriskaidekaTable(t, 13, 13, 20, 20)
	g.WithWriteLock(func() {
		g.Seats[2].Eliminated = true
		g.Seats[3].Eliminated = true
	})
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, g.Seats[0].ID), g.Seats[0].ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	res := g.Result()
	if res == nil || res.Kind != game.OutcomeDraw {
		t.Errorf("outcome %+v, want a draw: both players lost at once", res)
	}
}

func TestGapB14BloodMistGivesDoubleStrikeAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Blood Mist", "Enchantment", gapBloodMistOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, bear, "double strike")
}

func TestGapB14BurningVengeanceHitsWhenYouCastFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Burning Vengeance", "Enchantment", gapBurningVengeanceOracle, 0, 0)
	id := effectstest.SeedGraveyardCard(t, g, "Army of the Damned", "Sorcery", effectstest.ArmyOfTheDamnedOracle)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-2)
	}
}

func TestGapB14BurningVengeanceIgnoresASpellCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Burning Vengeance", "Enchantment", gapBurningVengeanceOracle, 0, 0)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Army of the Damned", "Sorcery", effectstest.ArmyOfTheDamnedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent at %d, want %d: the spell was cast from hand", opp.Life, life)
	}
}

func TestGapB14SpontaneousMutationShrinksByYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	yard := me.Graveyard.Size()
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogGraveyardCard(me, "Filler", "Sorcery", "", 0, 0)
	}
	effectstest.Enchant(t, g, "Spontaneous Mutation", gapSpontMutationOracle, bear)
	want := 5 - (yard + 3)
	if got := effectstest.EffectivePower(t, g, bear); got != want {
		t.Errorf("power %d, want %d", got, want)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 5 {
		t.Errorf("toughness %d, want 5 (-X/-0)", got)
	}
	effectstest.B27Kill(g, effectstest.B12Creature(g, me.ID, "Doomed", "Creature — Bear", 1, 1))
	g.RunStateChecksForTest()
	if got := effectstest.EffectivePower(t, g, bear); got != want-1 {
		t.Errorf("power %d after another card, want %d", got, want-1)
	}
}

func TestGapB14BoundByMoonsilverStopsAttackBlockAndTransform(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Bound by Moonsilver", gapBoundByMoonsilverOracle, bear)
	effectstest.AssertRestrictions(t, g, bear, game.CantAttack|game.CantBlock)
	if !effectstest.LayeredCard(t, g, bear).Effective().CantTransform {
		t.Error("the enchanted creature can still transform")
	}
}

func TestGapB14BoundByMoonsilverMovesForASacrificedPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.B12Creature(g, opp.ID, "First", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second", "Creature — Bear", 2, 2)
	fodder := effectstest.B12Permanent(g, me.ID, "Fodder", "Artifact")
	aura := effectstest.Enchant(t, g, "Bound by Moonsilver", gapBoundByMoonsilverOracle, first)
	effectstest.AdvanceToMain(t, g)
	activate := func(sac, target uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{
			Targets: effectstest.CardRefs(target), SacrificeIDs: []uuid.UUID{sac}, Strict: true,
		})
	}
	if err := activate(aura, second); err == nil {
		t.Fatal("the Aura sacrificed itself: it says another permanent")
	}
	if err := activate(fodder, second); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed permanent is still out")
	}
	effectstest.AssertRestrictions(t, g, second, game.CantAttack|game.CantBlock)
	effectstest.AssertRestrictions(t, g, first, 0)
}

func TestGapB14FaithUnbrokenExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Faith Unbroken", gapFaithUnbrokenOracle, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the opponent's creature was not exiled")
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 4 {
		t.Errorf("enchanted creature power %d, want 4", got)
	}
	effectstest.B27Kill(g, aura)
	g.RunStateChecksForTest()
	if effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID) != 1 {
		t.Error("the exiled creature did not return when the Aura left")
	}
}

func TestGapB14FaithUnbrokenOnlyEnchantsYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Faith Unbroken", "Enchantment — Aura", gapFaithUnbrokenOracle, effectstest.CardRefs(theirs))
	if err == nil {
		t.Error("Faith Unbroken enchanted a creature its caster doesn't control")
	}
}

func TestGapB14LunarchMantlePumpsAndGrantsAFlyingSacrificeAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	fodder := effectstest.B12Permanent(g, me.ID, "Fodder", "Artifact")
	effectstest.Enchant(t, g, "Lunarch Mantle", gapLunarchMantleOracle, bear)
	if effectstest.EffectivePower(t, g, bear) != 4 || effectstest.EffectiveToughness(t, g, bear) != 4 {
		t.Errorf("bear is %d/%d, want 4/4", effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear))
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, bear, "flying")
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed permanent is still out")
	}
}
