package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch04_test.go — card-level coverage for the card-coverage
// roadmap's batch 04 (#297, `edhrec_rank` 472–577): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// batch01_test.go are reused by name; new ones are b04-prefixed.

const (
	b04TatyovaOracle              = "0715e860-3b3b-4331-9718-207973e94fee"
	b04AccursedMarauderOracle     = "d8ad23a1-0b43-48ea-9fbe-d89b29194509"
	b04ChandrasIgnitionOracle     = "f61680da-606e-4d16-b0a0-361aa5210901"
	b04DiabolicTutorOracle        = "14589b6b-1814-46f9-a364-83cc15dacac2"
	b04BedevilOracle              = "bceecc64-96f1-4e7b-8904-0aef90377764"
	b04TalismanCuriosityOracle    = "8c34b089-aad1-476e-958a-3077bf1bbb51"
	b04FieldOfTheDeadOracle       = "aa959340-c869-4caa-92c7-572bd8d23eef"
	b04SoulWardenOracle           = "f3fad295-1af2-4ecc-8546-b121ad6be27b"
	b04AncientDenOracle           = "02f16726-f2f6-4943-b71a-93f8e26251d3"
	b04AdelineOracle              = "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba"
	b04BorosGarrisonOracle        = "8fa3ac81-3dfe-4565-be99-5554f7597b4b"
	b04TalismanResilienceOracle   = "42b8aa14-10bc-4bd6-88d9-4bb287eadd19"
	b04RiseOfTheDarkRealmsOracle  = "e5223a09-f732-4747-8914-e6546ab0ef4c"
	b04RakdosCarnariumOracle      = "0a023964-2905-4928-9c3e-dc63e6ebd218"
	b04SanguineBondOracle         = "73089a39-a2f6-4aa2-a058-e6551475153d"
	b04SandsteppeCitadelOracle    = "544dbabd-cbfc-40da-a5ba-2fea9cddb453"
	b04AvacynsPilgrimOracle       = "069f6530-e65c-4d52-85f3-e0a2acd148c5"
	b04ExquisiteBloodOracle       = "8f933fae-6c0c-42d7-a817-14760d8285cd"
	b04TerrorOfThePeaksOracle     = "f8e17f4f-080d-4bba-bd05-ca27e94ccecc"
	b04ChordOfCallingOracle       = "6789a170-f2c5-4fc0-8a45-2b2361e67410"
	b04GitaxianProbeOracle        = "1d67f5ff-1fce-45e5-b6a1-416c569351e2"
	b04StripMineOracle            = "d21a89eb-7c5b-459a-acc7-12b20b13bf79"
	b04HedronArchiveOracle        = "32263baa-d3f0-463f-92b3-4e9938476add"
	b04CatharsCrusadeOracle       = "cc65ac73-5bef-4ecb-ad8e-39199084c027"
	b04ElementalBondOracle        = "d9a7e5a6-3e41-4fc6-987a-18fe1b9d67dd"
	b04BeastmasterAscensionOracle = "11b5308d-5bc0-4782-875f-a28be36e665d"
	b04WheelOfFortuneOracle       = "a8abd966-de7b-46a3-8ac7-8747ab35653a"
	b04SelesnyaSanctuaryOracle    = "00ef1c55-dea1-4564-bd57-66de86cba4df"
	b04GravenCairnsOracle         = "5004b84a-33b7-4f6f-b2c2-7086b9087535"
)

// b04WaitForPick passes priority until a pick_target prompt for
// chooser appears — a targeted trigger asks for its target before it
// can go on the stack (the Blood Artist test's loop).
func b04WaitForPick(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && latestPickTarget(g, chooser) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if latestPickTarget(g, chooser) == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
}

// b04Creature pushes a live creature with the given P/T under owner.
func b04Creature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- registration --------------------------------------------------

func TestBatch04CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b04TatyovaOracle:              "Tatyova, Benthic Druid",
		b04AccursedMarauderOracle:     "Accursed Marauder",
		b04ChandrasIgnitionOracle:     "Chandra's Ignition",
		b04DiabolicTutorOracle:        "Diabolic Tutor",
		b04BedevilOracle:              "Bedevil",
		b04TalismanCuriosityOracle:    "Talisman of Curiosity",
		b04FieldOfTheDeadOracle:       "Field of the Dead",
		b04SoulWardenOracle:           "Soul Warden",
		b04AncientDenOracle:           "Ancient Den",
		b04AdelineOracle:              "Adeline, Resplendent Cathar",
		b04BorosGarrisonOracle:        "Boros Garrison",
		b04TalismanResilienceOracle:   "Talisman of Resilience",
		b04RiseOfTheDarkRealmsOracle:  "Rise of the Dark Realms",
		b04RakdosCarnariumOracle:      "Rakdos Carnarium",
		b04SanguineBondOracle:         "Sanguine Bond",
		b04SandsteppeCitadelOracle:    "Sandsteppe Citadel",
		b04AvacynsPilgrimOracle:       "Avacyn's Pilgrim",
		b04ExquisiteBloodOracle:       "Exquisite Blood",
		b04TerrorOfThePeaksOracle:     "Terror of the Peaks",
		b04ChordOfCallingOracle:       "Chord of Calling",
		b04GitaxianProbeOracle:        "Gitaxian Probe",
		b04StripMineOracle:            "Strip Mine",
		b04HedronArchiveOracle:        "Hedron Archive",
		b04CatharsCrusadeOracle:       "Cathars' Crusade",
		b04ElementalBondOracle:        "Elemental Bond",
		b04BeastmasterAscensionOracle: "Beastmaster Ascension",
		b04WheelOfFortuneOracle:       "Wheel of Fortune",
		b04SelesnyaSanctuaryOracle:    "Selesnya Sanctuary",
		b04GravenCairnsOracle:         "Graven Cairns",
	}
	if len(want) != 29 {
		t.Fatalf("the batch is 29 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
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
		b04TalismanCuriosityOracle:  "{G|U}",
		b04TalismanResilienceOracle: "{B|G}",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s not registered", oracle)
		}
		if len(spec.ManaAbilities) != 2 || spec.ManaAbilities[0].Produced != "{C}" {
			t.Errorf("%s: want a painless {C} first, got %+v", spec.Name, spec.ManaAbilities)
			continue
		}
		second := spec.ManaAbilities[1]
		if second.Produced != want || second.Rider == nil || second.NarrowToCommanderIdentity {
			t.Errorf("%s: ability 1 must be %s with a damage rider and no identity narrowing", spec.Name, want)
		}
	}
}

// --- lands ---------------------------------------------------------

// #1337: the bounce is a resolution-time choice (ReturnOneYouControl
// / ChoosePermanents), not a target picked when the trigger goes on
// the stack.
func TestB04BounceLandEntersTappedAndBouncesALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")

	garrison := playLandFromHand(t, g, "Boros Garrison", b04BorosGarrisonOracle)
	top100AssertEnteredTapped(t, g, garrison, "Boros Garrison")

	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Rakdos Carnarium", "Land", b04RakdosCarnariumOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "R" {
		t.Errorf("pool %v, want [B R]", got)
	}
}

func TestB04SandsteppeCitadelEntersTappedAndOffersThreeColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	played := playLandFromHand(t, g, "Sandsteppe Citadel", b04SandsteppeCitadelOracle)
	top100AssertEnteredTapped(t, g, played, "Sandsteppe Citadel")

	ready := seedPermanentWithOracle(g, me.ID, "Sandsteppe Citadel", "Land", b04SandsteppeCitadelOracle)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 3 {
		t.Fatalf("a tri-land must ask among exactly three colours, got %+v", pick)
	}
}

func TestB04AncientDenAndAvacynsPilgrimTapForWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	den := seedPermanentWithOracle(g, me.ID, "Ancient Den", "Artifact Land", b04AncientDenOracle)
	pilgrim := pushCatalogPermanent(g, me.ID, "Avacyn's Pilgrim", "Creature — Human Monk", b04AvacynsPilgrimOracle, false)
	for _, id := range []uuid.UUID{den, pilgrim} {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "W" || got[1] != "W" {
		t.Errorf("pool %v, want [W W]", got)
	}
}

func TestB04HedronArchiveTapsForTwoAndCashesIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := seedPermanentWithOracle(g, me.ID, "Hedron Archive", "Artifact", b04HedronArchiveOracle)
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool %v, want [C C]", got)
	}

	fresh := seedPermanentWithOracle(g, me.ID, "Hedron Archive", "Artifact", b04HedronArchiveOracle)
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, fresh, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fresh) {
		t.Error("the Archive is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestB04StripMineDestroysALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCatalogPermanent(g, me.ID, "Strip Mine", "Land", b04StripMineOracle, false)
	forest := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := seedCreature(g, "Bear", opp.ID)

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
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !opp.Graveyard.Contains(forest) {
		t.Error("the targeted land should be destroyed")
	}
}

func TestB04FieldOfTheDeadMakesAZombieAtSevenNames(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, name := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest", "Command Tower"} {
		seedLandOnBattlefield(g, me.ID, name, "Land")
	}
	// The Field is the seventh name, and its own entry counts.
	field := playLandFromHand(t, g, "Field of the Dead", b04FieldOfTheDeadOracle)
	top100AssertEnteredTapped(t, g, field, "Field of the Dead")
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Zombie") != 1 {
		t.Fatal("seven differently named lands should make one Zombie")
	}
	// An eighth land, one more Zombie.
	playLandFromHand(t, g, "Wastes", "")
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Zombie") != 2 {
		t.Error("every later land drop should make another Zombie")
	}
}

func TestB04FieldOfTheDeadCountsNamesNotLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 8; i++ {
		seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	}
	playLandFromHand(t, g, "Field of the Dead", b04FieldOfTheDeadOracle)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Zombie") != 0 {
		t.Error("eight Forests are two names, not seven")
	}
}

// --- spells --------------------------------------------------------

func TestB04DiabolicTutorSearchesAnyCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := stapleLibraryCard(me, "Needle", "Sorcery")
	castCatalogSpell(t, g, "Diabolic Tutor", "Sorcery", b04DiabolicTutorOracle, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("the tutor should ask which card to take")
	}
	answerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("the chosen card did not reach the hand")
	}
}

func TestB04BedevilDestroysAnArtifactAndRefusesAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	aura := seedPermanentFor(g, opp.ID, "Glory", "Enchantment")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bedevil", TypeLine: "Instant",
		OracleID: b04BedevilOracle, Owner: me.ID, Controller: me.ID})
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
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestB04ChandrasIgnitionDealsPowerToEverythingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b04Creature(g, me.ID, "Big", 4, 4)
	mine := b04Creature(g, me.ID, "Mine", 2, 2)
	theirs := b04Creature(g, opp.ID, "Theirs", 3, 3)
	before := lifeOfOpponents(g)
	meBefore := me.Life

	castCatalogSpell(t, g, "Chandra's Ignition", "Sorcery", b04ChandrasIgnitionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushGraveyardCardForTest(me, "My Dead")
	theirs := pushGraveyardCardForTest(opp, "Their Dead")
	sorcery := batch01GraveyardCard(opp, "Their Spell", "Sorcery")

	castCatalogSpell(t, g, "Rise of the Dark Realms", "Sorcery", b04RiseOfTheDarkRealmsOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		card, ok := battlefieldCard(g, id)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	titan := pushLibraryCardForTest(me, game.Card{Name: "Titan", TypeLine: "Creature — Giant", ManaCost: "{4}{G}{G}"})

	castXSpell(t, g, "Chord of Calling", "Instant", b04ChordOfCallingOracle, "{X}{G}{G}{G}", 2, nil)
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, bear); !ok {
		t.Error("the only creature with mana value ≤ 2 should have been put onto the battlefield")
	}
	if _, wrong := battlefieldCard(g, titan); wrong {
		t.Error("a mana value 6 creature is not a legal find for X=2")
	}
}

func TestB04GitaxianProbeLetsOnlyTheCasterSeeTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Gitaxian Probe", "Sorcery", b04GitaxianProbeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Wheel of Fortune", "Sorcery", b04WheelOfFortuneOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Tatyova, Benthic Druid", "Legendary Creature — Merfolk Druid", b04TatyovaOracle, false)
	life := me.Life
	playLandFromHand(t, g, "Forest", "")
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("landfall: life %d→%d, hand %d→%d; want +1 and +1", life, me.Life, hand, me.Hand.Size())
	}
}

func TestB04AccursedMarauderEdictsNontokenCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := seedCreature(g, "Bear", opp1.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp2.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)

	marauder := castCatalogSpell(t, g, "Accursed Marauder", "Creature — Zombie Warrior", b04AccursedMarauderOracle, nil)
	passPriorityAroundTable(t, g)

	if c := sacrificeChoiceFor(g, opp1.ID); c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != bear {
		t.Errorf("opponent 1 should be offered exactly their Bear, got %+v", c)
	}
	if sacrificeChoiceFor(g, opp2.ID) != nil {
		t.Error("a player whose only creature is a token has nothing to sacrifice")
	}
	if c := sacrificeChoiceFor(g, me.ID); c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != marauder {
		t.Errorf("the caster's only nontoken creature is the Marauder itself, got %+v", c)
	}
}

func TestB04SoulWardenGainsOnAnyOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	castCatalogSpell(t, g, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Error("the Warden's own entry is not 'another creature'")
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
	})
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1 (an opponent's creature counts, an artifact does not)", before, me.Life)
	}
}

func TestB04AdelineAttacksWithATappedHumanPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	adeline := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Adeline, Resplendent Cathar",
		TypeLine: "Legendary Creature — Human Knight", OracleID: b04AdelineOracle,
		Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if p := effectivePower(t, g, adeline); p != 2 {
		t.Fatalf("Adeline with two creatures = %d power, want 2", p)
	}
	before := lifeOfOpponents(g)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{adeline, bear} {
		if err := g.DeclareAttacker(id, victim.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if n := countBattlefieldNamed(g, me.ID, "Human"); n != 3 {
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
	if p := effectivePower(t, g, adeline); p != 5 {
		t.Errorf("Adeline with five creatures = %d power, want 5", p)
	}
	if card, _ := battlefieldCard(g, adeline); card.Tapped {
		t.Error("vigilance: Adeline attacks untapped")
	}

	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Sanguine Bond", "Enchantment", b04SanguineBondOracle, false)
	oppBefore := opp.Life

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != oppBefore-3 {
		t.Errorf("target lost %d, want 3", oppBefore-opp.Life)
	}
}

func TestB04ExquisiteBloodGainsOnAnyOpponentLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Exquisite Blood", "Enchantment", b04ExquisiteBloodOracle, false)
	before := me.Life

	// Damage: emits only EventDealDamage.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Fatalf("after a Bolt to an opponent: %d → %d, want +3", before, me.Life)
	}
	// Life loss: emits EventChangeLife.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	passPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("after a 2-life drain: %d → %d, want +5", before, me.Life)
	}
	// My own loss is not an opponent's.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	passPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("after my own loss: %d → %d, want +4 (no gain)", before, me.Life)
	}
}

func TestB04TerrorOfThePeaksShootsForTheNewcomersPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Terror of the Peaks", "Creature — Dragon", b04TerrorOfThePeaksOracle, false)
	oppBefore := opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Beast"), 1) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != oppBefore-3 {
		t.Errorf("a 3/3 entering should deal 3: %d → %d", oppBefore, opp.Life)
	}
	// An opponent's creature is not "you control".
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("3/3 green Beast"), 1) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's creature must not trigger the Dragon")
	}
}

func TestB04ElementalBondDrawsForPowerThreeOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Elemental Bond", "Enchantment", b04ElementalBondOracle, false)
	before := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Beast"), 1)
		_ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1)
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1 (the 3/3 counts, the 1/1 does not)", got)
	}
}

func TestB04BeastmasterAscensionSwitchesOnAtSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beastmaster Ascension", TypeLine: "Enchantment",
		OracleID: b04BeastmasterAscensionOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{"quest": 6},
	})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Fatalf("six counters: bear = %d power, want 2 (not yet)", p)
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, victim.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	lockInAttacks(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("seven counters: bear = %d/%d, want 7/7", p, tough)
	}
}

func TestB04CatharsCrusadeGrowsTheWholeTeam(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Cathars' Crusade", "Enchantment", b04CatharsCrusadeOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)

	// Counters live on the card and fold into CurrentPower, not into
	// the layer engine's effective power.
	if card, _ := battlefieldCard(g, bear); card.Counters["+1/+1"] != 1 || card.CurrentPower() != 3 {
		t.Errorf("the Bear should have one +1/+1 counter and power 3: %v / %d", card.Counters, card.CurrentPower())
	}
	goblin := findBattlefieldByName(g, "Goblin")
	if card, _ := battlefieldCard(g, goblin); card.Counters["+1/+1"] != 1 || card.CurrentPower() != 2 {
		t.Errorf("the entering Goblin gets its own counter: %v / %d", card.Counters, card.CurrentPower())
	}
}

// --- Graven Cairns -------------------------------------------------

func TestB04GravenCairnsFiltersOneHybridIntoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cairns := seedPermanentWithOracle(g, me.ID, "Graven Cairns", "Land", b04GravenCairnsOracle)

	// The wrong colour floating: refused, Cairns untouched.
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateManaAbility(me.ID, cairns, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{B/R} was paid with {G}")
	}
	if card, _ := battlefieldCard(g, cairns); card.Tapped {
		t.Fatal("a refused activation must not tap the Cairns")
	}

	// {R} pays the hybrid; the pool then asks twice from {B, R}.
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateManaAbility(me.ID, cairns, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the {R} should have been spent on the cost, pool = %v", batch01PoolColors(me))
	}
	picks := 0
	for _, c := range g.PendingChoices {
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
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want [B B]", got)
	}
	if card, _ := battlefieldCard(g, cairns); !card.Tapped {
		t.Error("the Cairns must be tapped after a paid activation")
	}
}
