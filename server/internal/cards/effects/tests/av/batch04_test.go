package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

func TestBatch04CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B04TatyovaOracle:              "Tatyova, Benthic Druid",
		effectstest.B04AccursedMarauderOracle:     "Accursed Marauder",
		effectstest.B04ChandrasIgnitionOracle:     "Chandra's Ignition",
		effectstest.B04DiabolicTutorOracle:        "Diabolic Tutor",
		effectstest.B04BedevilOracle:              "Bedevil",
		effectstest.B04TalismanCuriosityOracle:    "Talisman of Curiosity",
		effectstest.B04FieldOfTheDeadOracle:       "Field of the Dead",
		effectstest.B04SoulWardenOracle:           "Soul Warden",
		effectstest.B04AncientDenOracle:           "Ancient Den",
		effectstest.B04AdelineOracle:              "Adeline, Resplendent Cathar",
		effectstest.B04BorosGarrisonOracle:        "Boros Garrison",
		effectstest.B04TalismanResilienceOracle:   "Talisman of Resilience",
		effectstest.B04RiseOfTheDarkRealmsOracle:  "Rise of the Dark Realms",
		effectstest.B04RakdosCarnariumOracle:      "Rakdos Carnarium",
		effectstest.B04SanguineBondOracle:         "Sanguine Bond",
		effectstest.B04SandsteppeCitadelOracle:    "Sandsteppe Citadel",
		effectstest.B04AvacynsPilgrimOracle:       "Avacyn's Pilgrim",
		effectstest.B04ExquisiteBloodOracle:       "Exquisite Blood",
		effectstest.B04TerrorOfThePeaksOracle:     "Terror of the Peaks",
		effectstest.B04ChordOfCallingOracle:       "Chord of Calling",
		effectstest.B04GitaxianProbeOracle:        "Gitaxian Probe",
		effectstest.B04StripMineOracle:            "Strip Mine",
		effectstest.B04HedronArchiveOracle:        "Hedron Archive",
		effectstest.B04CatharsCrusadeOracle:       "Cathars' Crusade",
		effectstest.B04ElementalBondOracle:        "Elemental Bond",
		effectstest.B04BeastmasterAscensionOracle: "Beastmaster Ascension",
		effectstest.B04WheelOfFortuneOracle:       "Wheel of Fortune",
		effectstest.B04SelesnyaSanctuaryOracle:    "Selesnya Sanctuary",
		effectstest.B04GravenCairnsOracle:         "Graven Cairns",
	}
	if len(want) != 29 {
		t.Fatalf("the batch is 29 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

func TestB04TalismanRowsHaveBothHalves(t *testing.T) {
	for oracle, want := range map[string]string{
		effectstest.B04TalismanCuriosityOracle:  "{G|U}",
		effectstest.B04TalismanResilienceOracle: "{B|G}",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s not registered", oracle)
		}
		if len(spec.ManaAbilities) != 2 || spec.ManaAbilities[0].Produced != "{C}" {
			t.Errorf("%s: want a painless {C} first, got %+v", spec.Name, spec.ManaAbilities)
			continue
		}
		second := spec.ManaAbilities[1]
		if second.Produced != want || second.PainToYou != 1 || second.NarrowToCommanderIdentity {
			t.Errorf("%s: ability 1 must be %s with a damage rider and no identity narrowing", spec.Name, want)
		}
	}
}

// --- lands ---------------------------------------------------------

// #1337: the bounce is a resolution-time choice (ReturnOneYouControl
// / ChoosePermanents), not a target picked when the trigger goes on
// the stack.
func TestB04BounceLandEntersTappedAndBouncesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")

	garrison := effectstest.PlayLandFromHand(t, g, "Boros Garrison", effectstest.B04BorosGarrisonOracle)
	effectstest.Top100AssertEnteredTapped(t, g, garrison, "Boros Garrison")

	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Boros Garrison queued no return pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{plains}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}

	if g.Battlefield.Contains(plains) || !me.Hand.Contains(plains) {
		t.Error("the chosen land should be back in its owner's hand")
	}
	if !g.Battlefield.Contains(garrison) {
		t.Error("the Garrison itself should stay")
	}
}

func TestB04BounceLandTapsForBothColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Rakdos Carnarium", "Land", effectstest.B04RakdosCarnariumOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "R" {
		t.Errorf("pool %v, want [B R]", got)
	}
}

func TestB04SandsteppeCitadelEntersTappedAndOffersThreeColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	played := effectstest.PlayLandFromHand(t, g, "Sandsteppe Citadel", effectstest.B04SandsteppeCitadelOracle)
	effectstest.Top100AssertEnteredTapped(t, g, played, "Sandsteppe Citadel")

	ready := effectstest.SeedPermanentWithOracle(g, me.ID, "Sandsteppe Citadel", "Land", effectstest.B04SandsteppeCitadelOracle)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 3 {
		t.Fatalf("a tri-land must ask among exactly three colours, got %+v", pick)
	}
}

func TestB04AncientDenAndAvacynsPilgrimTapForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	den := effectstest.SeedPermanentWithOracle(g, me.ID, "Ancient Den", "Artifact Land", effectstest.B04AncientDenOracle)
	pilgrim := effectstest.PushCatalogPermanent(g, me.ID, "Avacyn's Pilgrim", "Creature — Human Monk", effectstest.B04AvacynsPilgrimOracle, false)
	for _, id := range []uuid.UUID{den, pilgrim} {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "W" || got[1] != "W" {
		t.Errorf("pool %v, want [W W]", got)
	}
}

func TestB04HedronArchiveTapsForTwoAndCashesIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.SeedPermanentWithOracle(g, me.ID, "Hedron Archive", "Artifact", effectstest.B04HedronArchiveOracle)
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool %v, want [C C]", got)
	}

	fresh := effectstest.SeedPermanentWithOracle(g, me.ID, "Hedron Archive", "Artifact", effectstest.B04HedronArchiveOracle)
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, fresh, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fresh) {
		t.Error("the Archive is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestB04StripMineDestroysALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Strip Mine", "Land", effectstest.B04StripMineOracle, false)
	forest := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)

	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a creature was accepted for 'target land'")
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: forest}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the Mine is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !opp.Graveyard.Contains(forest) {
		t.Error("the targeted land should be destroyed")
	}
}

func TestB04FieldOfTheDeadMakesAZombieAtSevenNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, name := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest", "Command Tower"} {
		effectstest.SeedLandOnBattlefield(g, me.ID, name, "Land")
	}
	// The Field is the seventh name, and its own entry counts.
	field := effectstest.PlayLandFromHand(t, g, "Field of the Dead", effectstest.B04FieldOfTheDeadOracle)
	effectstest.Top100AssertEnteredTapped(t, g, field, "Field of the Dead")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Zombie") != 1 {
		t.Fatal("seven differently named lands should make one Zombie")
	}
	// An eighth land, one more Zombie.
	effectstest.PlayLandFromHand(t, g, "Wastes", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Zombie") != 2 {
		t.Error("every later land drop should make another Zombie")
	}
}

func TestB04FieldOfTheDeadCountsNamesNotLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 8; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.PlayLandFromHand(t, g, "Field of the Dead", effectstest.B04FieldOfTheDeadOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Zombie") != 0 {
		t.Error("eight Forests are two names, not seven")
	}
}

// --- spells --------------------------------------------------------

func TestB04DiabolicTutorSearchesAnyCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.StapleLibraryCard(me, "Needle", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Diabolic Tutor", "Sorcery", effectstest.B04DiabolicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the tutor should ask which card to take")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("the chosen card did not reach the hand")
	}
}

func TestB04BedevilDestroysAnArtifactAndRefusesAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	aura := effectstest.SeedPermanentFor(g, opp.ID, "Glory", "Enchantment")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bedevil", TypeLine: "Instant",
		OracleID: effectstest.B04BedevilOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}},
	}); err == nil {
		t.Fatal("an enchantment was accepted for 'artifact, creature, or planeswalker'")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestB04ChandrasIgnitionDealsPowerToEverythingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B04Creature(g, me.ID, "Big", 4, 4)
	mine := effectstest.B04Creature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.B04Creature(g, opp.ID, "Theirs", 3, 3)
	before := effectstest.LifeOfOpponents(g)
	meBefore := me.Life

	effectstest.CastCatalogSpell(t, g, "Chandra's Ignition", "Sorcery", effectstest.B04ChandrasIgnitionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(big) {
		t.Error("the source creature does not damage itself")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("every OTHER creature takes 4")
	}
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-4 {
			t.Errorf("opponent %d: %d -> %d, want -4", i+1, b, got)
		}
	}
	if me.Life != meBefore {
		t.Error("the caster is not an opponent")
	}
}

func TestB04RiseOfTheDarkRealmsTakesEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushGraveyardCardForTest(me, "My Dead")
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Dead")
	sorcery := effectstest.Batch01GraveyardCard(opp, "Their Spell", "Sorcery")

	effectstest.CastCatalogSpell(t, g, "Rise of the Dark Realms", "Sorcery", effectstest.B04RiseOfTheDarkRealmsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Errorf("%s did not return", id)
			continue
		}
		if card.Controller != me.ID {
			t.Errorf("%s returned under %s, want the caster", id, card.Controller)
		}
	}
	if !opp.Graveyard.Contains(sorcery) {
		t.Error("a noncreature card must stay in the graveyard")
	}
}

func TestB04ChordOfCallingFindsACreatureWithinX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	titan := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Titan", TypeLine: "Creature — Giant", ManaCost: "{4}{G}{G}"})

	effectstest.CastXSpell(t, g, "Chord of Calling", "Instant", effectstest.B04ChordOfCallingOracle, "{X}{G}{G}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the only creature with mana value ≤ 2 should have been put onto the battlefield")
	}
	if _, wrong := effectstest.BattlefieldCard(g, titan); wrong {
		t.Error("a mana value 6 creature is not a legal find for X=2")
	}
}

func TestB04GitaxianProbeLetsOnlyTheCasterSeeTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Gitaxian Probe", "Sorcery", effectstest.B04GitaxianProbeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if opp.Hand.Size() == 0 {
		t.Fatal("the opponent needs a hand to look at")
	}
	for _, c := range opp.Hand.Cards {
		if !c.KnownBy[me.ID] {
			t.Fatal("the caster should know every card in the target's hand")
		}
		if c.KnownBy[third.ID] {
			t.Fatal("'look at' must not reveal the hand to the table")
		}
	}
}

func TestB04WheelOfFortuneRefillsEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Wheel of Fortune", "Sorcery", effectstest.B04WheelOfFortuneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != 7 {
			t.Errorf("seat %d holds %d cards, want 7", i, p.Hand.Size())
		}
		if p.Graveyard.Size() == 0 {
			t.Errorf("seat %d discarded nothing", i)
		}
	}
}

// --- creatures and enchantments ------------------------------------

func TestB04TatyovaGainsAndDrawsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Tatyova, Benthic Druid", "Legendary Creature — Merfolk Druid", effectstest.B04TatyovaOracle, false)
	life := me.Life
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("landfall: life %d→%d, hand %d→%d; want +1 and +1", life, me.Life, hand, me.Hand.Size())
	}
}

func TestB04AccursedMarauderEdictsNontokenCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp1, opp2 := g.Seats[1], g.Seats[2]
	bear := effectstest.SeedCreature(g, "Bear", opp1.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp2.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)

	marauder := effectstest.CastCatalogSpell(t, g, "Accursed Marauder", "Creature — Zombie Warrior", effectstest.B04AccursedMarauderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// Every seat has at most one nontoken creature, so every pick is
	// forced and collected unasked (ADR 0175, #3664).
	if effectstest.SacrificeChoiceFor(g, opp2.ID) != nil {
		t.Error("a player whose only creature is a token has nothing to sacrifice")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("opponent 1's only nontoken creature, their Bear, is sacrificed")
	}
	if effectstest.FindBattlefieldByName(g, "Goblin") == uuid.Nil {
		t.Error("a token is not a nontoken creature")
	}
	if g.Battlefield.Contains(marauder) {
		t.Error("the caster's only nontoken creature is the Marauder itself")
	}
}

func TestB04SoulWardenGainsOnAnyOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Soul Warden", "Creature — Human Cleric", effectstest.B04SoulWardenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Error("the Warden's own entry is not 'another creature'")
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1 (an opponent's creature counts, an artifact does not)", before, me.Life)
	}
}

func TestB04AdelineAttacksWithATappedHumanPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	adeline := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Adeline, Resplendent Cathar",
		TypeLine: "Legendary Creature — Human Knight", OracleID: effectstest.B04AdelineOracle,
		Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, adeline); p != 2 {
		t.Fatalf("Adeline with two creatures = %d power, want 2", p)
	}
	before := effectstest.LifeOfOpponents(g)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{adeline, bear} {
		if err := g.DeclareAttacker(id, victim.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human"); n != 3 {
		t.Fatalf("two attackers made %d Humans, want 3 — one per opponent, once per attack", n)
	}
	attacking := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human" && c.Controller == me.ID {
			if !c.Tapped || c.AttackingTarget == uuid.Nil {
				t.Error("each Human must enter tapped and attacking")
			}
			attacking[c.AttackingTarget]++
		}
	}
	if len(attacking) != 3 {
		t.Errorf("the Humans should attack three different opponents, got %v", attacking)
	}
	if p := effectstest.EffectivePower(t, g, adeline); p != 5 {
		t.Errorf("Adeline with five creatures = %d power, want 5", p)
	}
	if card, _ := effectstest.BattlefieldCard(g, adeline); card.Tapped {
		t.Error("vigilance: Adeline attacks untapped")
	}

	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	// The victim takes Adeline (5) + the Bear (2) + one Human (1); the
	// other two opponents take one Human each.
	if got := victim.Life; got != before[0]-8 {
		t.Errorf("victim %d -> %d, want -8", before[0], got)
	}
	for i := 1; i < 3; i++ {
		if got := g.Seats[i+1].Life; got != before[i]-1 {
			t.Errorf("opponent %d: %d -> %d, want -1 from its Human", i+1, before[i], got)
		}
	}
}

func TestB04SanguineBondDrainsTheTargetOnLifegain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sanguine Bond", "Enchantment", effectstest.B04SanguineBondOracle, false)
	oppBefore := opp.Life

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != oppBefore-3 {
		t.Errorf("target lost %d, want 3", oppBefore-opp.Life)
	}
}

func TestB04ExquisiteBloodGainsOnAnyOpponentLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Exquisite Blood", "Enchantment", effectstest.B04ExquisiteBloodOracle, false)
	before := me.Life

	// Damage: emits only EventDealDamage.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Fatalf("after a Bolt to an opponent: %d → %d, want +3", before, me.Life)
	}
	// Life loss: emits EventChangeLife.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("after a 2-life drain: %d → %d, want +5", before, me.Life)
	}
	// My own loss is not an opponent's.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("after my own loss: %d → %d, want +4 (no gain)", before, me.Life)
	}
}

func TestB04TerrorOfThePeaksShootsForTheNewcomersPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	oppBefore := opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != oppBefore-3 {
		t.Errorf("a 3/3 entering should deal 3: %d → %d", oppBefore, opp.Life)
	}
	// An opponent's creature is not "you control".
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's creature must not trigger the Dragon")
	}
}

func TestB04ElementalBondDrawsForPowerThreeOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Elemental Bond", "Enchantment", effectstest.B04ElementalBondOracle, false)
	before := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1)
		_ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1 (the 3/3 counts, the 1/1 does not)", got)
	}
}

func TestB04BeastmasterAscensionSwitchesOnAtSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beastmaster Ascension", TypeLine: "Enchantment",
		OracleID: effectstest.B04BeastmasterAscensionOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{"quest": 6},
	})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Fatalf("six counters: bear = %d power, want 2 (not yet)", p)
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, victim.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("seven counters: bear = %d/%d, want 7/7", p, tough)
	}
}

func TestB04CatharsCrusadeGrowsTheWholeTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Cathars' Crusade", "Enchantment", effectstest.B04CatharsCrusadeOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)

	// Counters live on the card and fold into CurrentPower, not into
	// the layer engine's effective power.
	if card, _ := effectstest.BattlefieldCard(g, bear); card.Counters["+1/+1"] != 1 || card.CurrentPower() != 3 {
		t.Errorf("the Bear should have one +1/+1 counter and power 3: %v / %d", card.Counters, card.CurrentPower())
	}
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	if card, _ := effectstest.BattlefieldCard(g, goblin); card.Counters["+1/+1"] != 1 || card.CurrentPower() != 2 {
		t.Errorf("the entering Goblin gets its own counter: %v / %d", card.Counters, card.CurrentPower())
	}
}

// --- Graven Cairns -------------------------------------------------

func TestB04GravenCairnsFiltersOneHybridIntoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cairns := effectstest.SeedPermanentWithOracle(g, me.ID, "Graven Cairns", "Land", effectstest.B04GravenCairnsOracle)

	// The wrong colour floating: refused, Cairns untouched.
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateManaAbility(me.ID, cairns, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{B/R} was paid with {G}")
	}
	if card, _ := effectstest.BattlefieldCard(g, cairns); card.Tapped {
		t.Fatal("a refused activation must not tap the Cairns")
	}

	// {R} pays the hybrid; the pool then asks twice from {B, R}.
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateManaAbility(me.ID, cairns, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("the {R} should have been spent on the cost, pool = %v", effectstest.Batch01PoolColors(me))
	}
	// #3939: one choice, asked once per pick.
	picks := 0
	for c := effectstest.RiderLatestManaPick(g, me.ID); c != nil && picks < 8; c = effectstest.RiderLatestManaPick(g, me.ID) {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			picks++
			if len(c.ColorOptions) != 2 {
				t.Errorf("each slot must offer exactly B and R, got %v", c.ColorOptions)
			}
			if err := g.ResolveManaChoice(c.ID, me.ID, "B"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if picks != 2 {
		t.Fatalf("want two colour picks, got %d", picks)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want [B B]", got)
	}
	if card, _ := effectstest.BattlefieldCard(g, cairns); !card.Tapped {
		t.Error("the Cairns must be tapped after a paid activation")
	}
}
