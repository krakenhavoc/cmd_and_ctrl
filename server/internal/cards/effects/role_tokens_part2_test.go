package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// role_tokens_part2_test.go — #1945 part 2: the Sorcerer, Young Hero,
// Virtuous and Chef Roles (a trigger or a count the ENCHANTED CREATURE
// has, not the Role) and the cards that make them.

const (
	livingLecternOracle      = "061e7fb2-2c04-4845-b60a-0e2512a2dec1"
	splashySpellcasterOracle = "21486218-70ec-4c60-9844-8f7c80912333"
	spellbookVendorOracle    = "51dad7d3-f91d-4bd8-aaed-235da42a448b"
	unassumingSageOracle     = "1cf4e16b-2273-41c7-9bcd-0cdc78e12198"
	cutInOracle              = "9b205900-ec39-455e-900b-615ece08a07f"
	embereth                 = "2c232056-9dd6-4a78-be08-3bff19daa88d"
	merryBardsOracle         = "699fa8e6-84f5-4010-a8fe-7b2e4f280ef9"
	protectiveParentsOracle  = "d9717522-358b-41ac-9183-8a6b3981f921"
	returnTriumphantOracle   = "a4814615-ffa2-4692-86bc-3a1b7fc2acc8"
	elliverOracle            = "97f9b82f-b3cf-44ff-9abc-2d92d7cbaa27"
	charmedClothierOracle    = "ad0e6b12-a6bd-4860-9eac-68b2492d9567"
	redtoothOracle           = "f6cb9c5a-601e-4580-9fab-903e9b00e35e"
	syrArmontOracle          = "53fc09e1-1e5d-4f22-b623-2c5c770183ab"
)

// giveRole attaches a Role of `kind` to `host`, controlled by
// `controller`, through the same primitive the cards use.
func giveRole(t *testing.T, g *game.Game, kind RoleKind, host, controller uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: controller}
		err = CreateRoleToken{Role: kind, Host: host, Controller: controller}.Apply(NewContext(g, item))
	})
	if err != nil {
		t.Fatalf("CreateRoleToken %s: %v", kind, err)
	}
	g.RunStateChecksForTest()
}

func handCount(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Hand.Size() })
	return n
}

// --- the Roles ---------------------------------------------------

// The Sorcerer Role's scry is the CREATURE's trigger: the stack item
// names the bear as its source, not the Role.
func TestSorcererRoleGivesTheCreatureAScryOnAttackTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	giveRole(t, g, RoleSorcerer, bear, me.ID)

	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tough)
	}
	declareAttack(t, g, opp.ID, bear)

	if triggersOnStackFrom(g, bear) != 1 {
		t.Fatalf("triggers sourced from the bear = %d, want 1", triggersOnStackFrom(g, bear))
	}
	role := findBattlefieldByName(g, "Sorcerer Role")
	if role == uuid.Nil || triggersOnStackFrom(g, role) != 0 {
		t.Fatalf("the Role (%v) must not be the trigger's source", role)
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry prompt for the enchanted creature's controller")
	}
}

// A creature that is not enchanted has no such trigger, and a newer
// Role takes the grant away with the older Role (CR 704.5z).
func TestSorcererRoleTriggerEndsWhenTheRoleIsReplaced(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	giveRole(t, g, RoleSorcerer, bear, me.ID)
	giveRole(t, g, RoleMonster, bear, me.ID)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Sorcerer Role") != uuid.Nil {
		t.Fatal("the older Sorcerer Role is still on the battlefield")
	}
	declareAttack(t, g, opp.ID, bear)
	if triggersOnStackFrom(g, bear) != 0 {
		t.Error("the bear still triggers a scry after losing its Sorcerer Role")
	}
}

func TestYoungHeroRolePutsACounterOnAnAttackerWithToughnessThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	giveRole(t, g, RoleYoungHero, bear, me.ID)

	declareAttack(t, g, opp.ID, bear)
	if triggersOnStackFrom(g, bear) != 1 {
		t.Fatalf("Young Hero trigger on the stack = %d, want 1", triggersOnStackFrom(g, bear))
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the bear = %d, want 1", n)
	}
}

// The intervening if (CR 603.4): toughness 4 never triggers at all.
func TestYoungHeroRoleDoesNothingOnAToughnessFourAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := pushVanillaCreature(g, me.ID, "Ox", 2, 4)
	giveRole(t, g, RoleYoungHero, wall, me.ID)

	declareAttack(t, g, opp.ID, wall)
	if triggersOnStackFrom(g, wall) != 0 {
		t.Fatal("a toughness-4 attacker triggered Young Hero")
	}
	if n := counterCount(g, wall, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want 0", n)
	}
}

// Second check on resolution: a pump in response that lifts the
// toughness past 3 leaves the counter off.
func TestYoungHeroRoleRechecksToughnessOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	giveRole(t, g, RoleYoungHero, bear, me.ID)

	declareAttack(t, g, opp.ID, bear)
	if triggersOnStackFrom(g, bear) != 1 {
		t.Fatal("no trigger to respond to")
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2)
	})
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("counters = %d, want the 2 from the pump and no third", n)
	}
}

func TestVirtuousRoleCountsTheEnchantmentsItsControllerControls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	giveRole(t, g, RoleVirtuous, bear, me.ID)
	// The Role is itself an enchantment you control: +1/+1.
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("bear with the Role alone is %d/%d, want 3/3", p, tough)
	}
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mine", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID,
	})
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("bear with one more of mine and one of theirs has power %d, want 4", p)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(mine) })
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("after the enchantment left, power %d, want 3", p)
	}
}

func TestChefRoleMakesAFoodOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	giveRole(t, g, RoleChef, bear, me.ID)

	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if countOnBattlefield(g, "Food", me.ID) != 1 {
		t.Errorf("Food tokens = %d, want 1", countOnBattlefield(g, "Food", me.ID))
	}
}

// --- Living Lectern ----------------------------------------------

func TestLivingLecternDrawsAndPutsASorcererRoleOnAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lectern := pushCatalogPermanent(g, me.ID, "Living Lectern", "Artifact Creature — Construct", livingLecternOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b06AddMana(me, "C")
	advanceTo(t, g, game.StepPrecombatMain)
	hand := handCount(g, me)

	if err := g.ActivateCatalogAbility(me.ID, lectern, 0, game.ActivateAbilityParams{Targets: cardRefs(lectern)}); err == nil {
		t.Fatal("the Lectern named itself as the other creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, lectern, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if handCount(g, me) != hand+1 {
		t.Errorf("hand %d, want %d", handCount(g, me), hand+1)
	}
	if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Sorcerer Role" {
		t.Errorf("roles on the bear = %+v, want a Sorcerer Role", roles)
	}
	if findBattlefieldByName(g, "Living Lectern") != uuid.Nil {
		t.Error("the Lectern was not sacrificed")
	}
}

func TestLivingLecternWithNoTargetStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lectern := pushCatalogPermanent(g, me.ID, "Living Lectern", "Artifact Creature — Construct", livingLecternOracle, false)
	b06AddMana(me, "C")
	advanceTo(t, g, game.StepPrecombatMain)
	hand := handCount(g, me)

	if err := g.ActivateCatalogAbility(me.ID, lectern, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if handCount(g, me) != hand+1 {
		t.Errorf("hand %d, want %d", handCount(g, me), hand+1)
	}
}

// --- Embereth Veteran --------------------------------------------

func TestEmberethVeteranPutsAYoungHeroRoleOnAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vet := pushCatalogPermanent(g, me.ID, "Embereth Veteran", "Creature — Human Knight", embereth, false)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b06AddMana(me, "C")

	if err := g.ActivateCatalogAbility(me.ID, vet, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	roles := rolesOn(g, theirs)
	if len(roles) != 1 || roles[0].Name != "Young Hero Role" || roles[0].Controller != me.ID {
		t.Fatalf("roles on their bear = %+v, want my Young Hero Role", roles)
	}
	if findBattlefieldByName(g, "Embereth Veteran") != uuid.Nil {
		t.Error("the Veteran was not sacrificed")
	}
}

// --- Splashy Spellcaster -----------------------------------------

func TestSplashySpellcasterRolesAnotherCreatureWhenYouCastAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Splashy Spellcaster", "Creature — Elemental Wizard", splashySpellcasterOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)

	if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Sorcerer Role" {
		t.Errorf("roles on the bear = %+v, want a Sorcerer Role", roles)
	}
}

// --- Unassuming Sage ---------------------------------------------

func TestUnassumingSagePayingTwoAttachesASorcererRoleToItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := castCatalogSpell(t, g, "Unassuming Sage", "Creature — Human Peasant Wizard", unassumingSageOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment prompt for the Sage")
	}
	b06AddMana(me, "C", "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if roles := rolesOn(g, sage); len(roles) != 1 || roles[0].Name != "Sorcerer Role" {
		t.Errorf("roles on the Sage = %+v, want a Sorcerer Role", roles)
	}
}

func TestUnassumingSageDecliningLeavesItBare(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := castCatalogSpell(t, g, "Unassuming Sage", "Creature — Human Peasant Wizard", unassumingSageOracle, nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if roles := rolesOn(g, sage); len(roles) != 0 {
		t.Errorf("roles on the Sage = %+v, want none", roles)
	}
}

// --- Spellbook Vendor and Merry Bards: "you may pay. When you do" ---

func TestSpellbookVendorPaysThenTargetsAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Spellbook Vendor", "Creature — Human Peasant", spellbookVendorOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment prompt at the beginning of combat")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for before the payment was made")
	}
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Sorcerer Role" {
		t.Errorf("roles on the bear = %+v, want a Sorcerer Role", roles)
	}
}

func TestMerryBardsDecliningAsksForNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Merry Bards", "Creature — Human Bard", merryBardsOracle, nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt appeared after declining")
	}
	if roles := rolesOn(g, bear); len(roles) != 0 {
		t.Errorf("roles = %+v, want none", roles)
	}
}

func TestMerryBardsPayingPutsAYoungHeroRoleOnTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Merry Bards", "Creature — Human Bard", merryBardsOracle, nil)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Young Hero Role" {
		t.Errorf("roles on the bear = %+v, want a Young Hero Role", roles)
	}
}

// --- Protective Parents ------------------------------------------

func TestProtectiveParentsDyingRolesAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	parents := pushCatalogPermanent(g, me.ID, "Protective Parents", "Creature — Human Peasant", protectiveParentsOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(parents) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
	}
	if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Young Hero Role" {
		t.Errorf("roles on the bear = %+v, want a Young Hero Role", roles)
	}
}

// --- Return Triumphant -------------------------------------------

func TestReturnTriumphantReturnsTheCreatureWithAYoungHeroRole(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := pushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")

	castCatalogSpell(t, g, "Return Triumphant", "Sorcery", returnTriumphantOracle, cardRefs(card))
	passPriorityAroundTable(t, g)

	if findBattlefieldByName(g, "Fallen Bear") == uuid.Nil {
		t.Fatal("the creature did not return")
	}
	if roles := rolesOn(g, card); len(roles) != 1 || roles[0].Name != "Young Hero Role" {
		t.Errorf("roles on the returned creature = %+v, want a Young Hero Role", roles)
	}
}

func TestReturnTriumphantRefusesAFourDropTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := pushGraveyardPermanent(me, "Big Bear", "Creature — Bear", "{3}{G}")
	if err := castCatalogSpellErr(t, g, "Return Triumphant", "Sorcery", returnTriumphantOracle, cardRefs(card)); err == nil {
		t.Fatal("a mana value 4 card was a legal target")
	}
}

// --- Cut In ------------------------------------------------------

func TestCutInDamagesOneCreatureAndRolesAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 4)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	castCatalogSpell(t, g, "Cut In", "Sorcery", cutInOracle, cardRefs(victim, mine))
	passPriorityAroundTable(t, g)

	if findBattlefieldByName(g, "Their Bear") != uuid.Nil {
		t.Error("4 damage did not destroy a 2/4")
	}
	if roles := rolesOn(g, mine); len(roles) != 1 || roles[0].Name != "Young Hero Role" {
		t.Errorf("roles on my bear = %+v, want a Young Hero Role", roles)
	}
}

func TestCutInWithNoRoleTargetStillDamages(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 4)

	castCatalogSpell(t, g, "Cut In", "Sorcery", cutInOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Their Bear") != uuid.Nil {
		t.Error("4 damage did not destroy a 2/4")
	}
}

func TestCutInCannotRoleAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := pushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 2)
	if err := castCatalogSpellErr(t, g, "Cut In", "Sorcery", cutInOracle, cardRefs(a, b)); err == nil {
		t.Fatal("the Role clause accepted a creature I don't control")
	}
}

// --- Ellivere, Charmed Clothier, Redtooth Genealogist, Syr Armont ---

func TestElliverePutsAVirtuousRoleOnAnotherCreatureAndDrawsOnEnchantedHits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	castCatalogSpell(t, g, "Ellivere of the Wild Court", "Legendary Creature — Human Knight", elliverOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the enter trigger")
	}
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Virtuous Role" {
		t.Fatalf("roles on the bear = %+v, want a Virtuous Role", roles)
	}
	hand := handCount(g, me)
	attackWith(t, g, opp.ID, bear, plain)
	passPriorityAroundTable(t, g)
	// Only the enchanted bear draws: one card, not two.
	if got := handCount(g, me); got != hand+1 {
		t.Errorf("hand grew by %d, want 1 (only the enchanted creature connects for a card)", got-hand)
	}
}

func TestCharmedClothierAndRedtoothGenealogistGiveARoyalRole(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Charmed Clothier", "Creature — Faerie Advisor", charmedClothierOracle},
		{"Redtooth Genealogist", "Creature — Elf Advisor", redtoothOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			self := castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			if latestPickTarget(g, me.ID) == nil {
				t.Fatal("no target prompt")
			}
			pickTriggerTarget(t, g, me.ID, bear)
			passPriorityAroundTable(t, g)
			if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Royal Role" {
				t.Errorf("roles on the bear = %+v, want a Royal Role", roles)
			}
			if roles := rolesOn(g, self); len(roles) != 0 {
				t.Error("the card put a Role on itself")
			}
		})
	}
}

func TestSyrArmontMakesEnchantedCreaturesYouControlBigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	giveRole(t, g, RoleMonster, theirs, opp.ID)
	castCatalogSpell(t, g, "Syr Armont, the Redeemer", "Legendary Creature — Human Knight", syrArmontOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	// 2/2 +1/+1 (Monster) +1/+1 (Armont).
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tough)
	}
	if p := effectivePower(t, g, plain); p != 2 {
		t.Errorf("unenchanted creature has power %d, want 2", p)
	}
	// Their enchanted creature gets the Monster Role's own +1/+1 only.
	if p := effectivePower(t, g, theirs); p != 3 {
		t.Errorf("opponent's enchanted creature has power %d, want 3", p)
	}
}
