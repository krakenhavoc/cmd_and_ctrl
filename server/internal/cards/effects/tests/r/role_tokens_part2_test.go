package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- the Roles ---------------------------------------------------

// The Sorcerer Role's scry is the CREATURE's trigger: the stack item
// names the bear as its source, not the Role.
func TestSorcererRoleGivesTheCreatureAScryOnAttackTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleSorcerer, bear, me.ID)

	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tough)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)

	if effectstest.TriggersOnStackFrom(g, bear) != 1 {
		t.Fatalf("triggers sourced from the bear = %d, want 1", effectstest.TriggersOnStackFrom(g, bear))
	}
	role := effectstest.FindBattlefieldByName(g, "Sorcerer Role")
	if role == uuid.Nil || effectstest.TriggersOnStackFrom(g, role) != 0 {
		t.Fatalf("the Role (%v) must not be the trigger's source", role)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry prompt for the enchanted creature's controller")
	}
}

// A creature that is not enchanted has no such trigger, and a newer
// Role takes the grant away with the older Role (CR 704.5z).
func TestSorcererRoleTriggerEndsWhenTheRoleIsReplaced(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleSorcerer, bear, me.ID)
	effectstest.GiveRole(t, g, effects.RoleMonster, bear, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Sorcerer Role") != uuid.Nil {
		t.Fatal("the older Sorcerer Role is still on the battlefield")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.TriggersOnStackFrom(g, bear) != 0 {
		t.Error("the bear still triggers a scry after losing its Sorcerer Role")
	}
}

func TestYoungHeroRolePutsACounterOnAnAttackerWithToughnessThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleYoungHero, bear, me.ID)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.TriggersOnStackFrom(g, bear) != 1 {
		t.Fatalf("Young Hero trigger on the stack = %d, want 1", effectstest.TriggersOnStackFrom(g, bear))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the bear = %d, want 1", n)
	}
}

// The intervening if (CR 603.4): toughness 4 never triggers at all.
func TestYoungHeroRoleDoesNothingOnAToughnessFourAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushVanillaCreature(g, me.ID, "Ox", 2, 4)
	effectstest.GiveRole(t, g, effects.RoleYoungHero, wall, me.ID)

	effectstest.DeclareAttack(t, g, opp.ID, wall)
	if effectstest.TriggersOnStackFrom(g, wall) != 0 {
		t.Fatal("a toughness-4 attacker triggered Young Hero")
	}
	if n := effectstest.CounterCount(g, wall, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want 0", n)
	}
}

// Second check on resolution: a pump in response that lifts the
// toughness past 3 leaves the counter off.
func TestYoungHeroRoleRechecksToughnessOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleYoungHero, bear, me.ID)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.TriggersOnStackFrom(g, bear) != 1 {
		t.Fatal("no trigger to respond to")
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("counters = %d, want the 2 from the pump and no third", n)
	}
}

func TestVirtuousRoleCountsTheEnchantmentsItsControllerControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleVirtuous, bear, me.ID)
	// The Role is itself an enchantment you control: +1/+1.
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("bear with the Role alone is %d/%d, want 3/3", p, tough)
	}
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mine", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID,
	})
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("bear with one more of mine and one of theirs has power %d, want 4", p)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(mine) })
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("after the enchantment left, power %d, want 3", p)
	}
}

func TestChefRoleMakesAFoodOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleChef, bear, me.ID)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountOnBattlefield(g, "Food", me.ID) != 1 {
		t.Errorf("Food tokens = %d, want 1", effectstest.CountOnBattlefield(g, "Food", me.ID))
	}
}

// --- Living Lectern ----------------------------------------------

func TestLivingLecternDrawsAndPutsASorcererRoleOnAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lectern := effectstest.PushCatalogPermanent(g, me.ID, "Living Lectern", "Artifact Creature — Construct", livingLecternOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.B06AddMana(me, "C")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand := effectstest.HandCount(g, me)

	if err := g.ActivateCatalogAbility(me.ID, lectern, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(lectern)}); err == nil {
		t.Fatal("the Lectern named itself as the other creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, lectern, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.HandCount(g, me) != hand+1 {
		t.Errorf("hand %d, want %d", effectstest.HandCount(g, me), hand+1)
	}
	if roles := effectstest.RolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Sorcerer Role" {
		t.Errorf("roles on the bear = %+v, want a Sorcerer Role", roles)
	}
	if effectstest.FindBattlefieldByName(g, "Living Lectern") != uuid.Nil {
		t.Error("the Lectern was not sacrificed")
	}
}

func TestLivingLecternWithNoTargetStillDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lectern := effectstest.PushCatalogPermanent(g, me.ID, "Living Lectern", "Artifact Creature — Construct", livingLecternOracle, false)
	effectstest.B06AddMana(me, "C")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand := effectstest.HandCount(g, me)

	if err := g.ActivateCatalogAbility(me.ID, lectern, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandCount(g, me) != hand+1 {
		t.Errorf("hand %d, want %d", effectstest.HandCount(g, me), hand+1)
	}
}

// --- Embereth Veteran --------------------------------------------

func TestEmberethVeteranPutsAYoungHeroRoleOnAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vet := effectstest.PushCatalogPermanent(g, me.ID, "Embereth Veteran", "Creature — Human Knight", embereth, false)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.B06AddMana(me, "C")

	if err := g.ActivateCatalogAbility(me.ID, vet, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	roles := effectstest.RolesOn(g, theirs)
	if len(roles) != 1 || roles[0].Name != "Young Hero Role" || roles[0].Controller != me.ID {
		t.Fatalf("roles on their bear = %+v, want my Young Hero Role", roles)
	}
	if effectstest.FindBattlefieldByName(g, "Embereth Veteran") != uuid.Nil {
		t.Error("the Veteran was not sacrificed")
	}
}

// --- Splashy Spellcaster -----------------------------------------

func TestSplashySpellcasterRolesAnotherCreatureWhenYouCastAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Splashy Spellcaster", "Creature — Elemental Wizard", splashySpellcasterOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if roles := effectstest.RolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Sorcerer Role" {
		t.Errorf("roles on the bear = %+v, want a Sorcerer Role", roles)
	}
}

// --- Unassuming Sage ---------------------------------------------

func TestUnassumingSagePayingTwoAttachesASorcererRoleToItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sage := effectstest.CastCatalogSpell(t, g, "Unassuming Sage", "Creature — Human Peasant Wizard", unassumingSageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment prompt for the Sage")
	}
	effectstest.B06AddMana(me, "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if roles := effectstest.RolesOn(g, sage); len(roles) != 1 || roles[0].Name != "Sorcerer Role" {
		t.Errorf("roles on the Sage = %+v, want a Sorcerer Role", roles)
	}
}

func TestUnassumingSageDecliningLeavesItBare(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sage := effectstest.CastCatalogSpell(t, g, "Unassuming Sage", "Creature — Human Peasant Wizard", unassumingSageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if roles := effectstest.RolesOn(g, sage); len(roles) != 0 {
		t.Errorf("roles on the Sage = %+v, want none", roles)
	}
}

// --- Spellbook Vendor and Merry Bards: "you may pay. When you do" ---

func TestSpellbookVendorPaysThenTargetsAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Spellbook Vendor", "Creature — Human Peasant", spellbookVendorOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment prompt at the beginning of combat")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for before the payment was made")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if roles := effectstest.RolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Sorcerer Role" {
		t.Errorf("roles on the bear = %+v, want a Sorcerer Role", roles)
	}
}

func TestMerryBardsDecliningAsksForNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Merry Bards", "Creature — Human Bard", merryBardsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt appeared after declining")
	}
	if roles := effectstest.RolesOn(g, bear); len(roles) != 0 {
		t.Errorf("roles = %+v, want none", roles)
	}
}

func TestMerryBardsPayingPutsAYoungHeroRoleOnTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Merry Bards", "Creature — Human Bard", merryBardsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if roles := effectstest.RolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Young Hero Role" {
		t.Errorf("roles on the bear = %+v, want a Young Hero Role", roles)
	}
}

// --- Protective Parents ------------------------------------------

func TestProtectiveParentsDyingRolesAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	parents := effectstest.PushCatalogPermanent(g, me.ID, "Protective Parents", "Creature — Human Peasant", protectiveParentsOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(parents) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if roles := effectstest.RolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Young Hero Role" {
		t.Errorf("roles on the bear = %+v, want a Young Hero Role", roles)
	}
}

// --- Return Triumphant -------------------------------------------

func TestReturnTriumphantReturnsTheCreatureWithAYoungHeroRole(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.PushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")

	effectstest.CastCatalogSpell(t, g, "Return Triumphant", "Sorcery", returnTriumphantOracle, effectstest.CardRefs(card))
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.FindBattlefieldByName(g, "Fallen Bear") == uuid.Nil {
		t.Fatal("the creature did not return")
	}
	if roles := effectstest.RolesOn(g, card); len(roles) != 1 || roles[0].Name != "Young Hero Role" {
		t.Errorf("roles on the returned creature = %+v, want a Young Hero Role", roles)
	}
}

func TestReturnTriumphantRefusesAFourDropTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.PushGraveyardPermanent(me, "Big Bear", "Creature — Bear", "{3}{G}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Return Triumphant", "Sorcery", returnTriumphantOracle, effectstest.CardRefs(card)); err == nil {
		t.Fatal("a mana value 4 card was a legal target")
	}
}

// --- Cut In ------------------------------------------------------

func TestCutInDamagesOneCreatureAndRolesAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 4)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Cut In", "Sorcery", cutInOracle, effectstest.CardRefs(victim, mine))
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.FindBattlefieldByName(g, "Their Bear") != uuid.Nil {
		t.Error("4 damage did not destroy a 2/4")
	}
	if roles := effectstest.RolesOn(g, mine); len(roles) != 1 || roles[0].Name != "Young Hero Role" {
		t.Errorf("roles on my bear = %+v, want a Young Hero Role", roles)
	}
}

func TestCutInWithNoRoleTargetStillDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 4)

	effectstest.CastCatalogSpell(t, g, "Cut In", "Sorcery", cutInOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Their Bear") != uuid.Nil {
		t.Error("4 damage did not destroy a 2/4")
	}
}

func TestCutInCannotRoleAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Cut In", "Sorcery", cutInOracle, effectstest.CardRefs(a, b)); err == nil {
		t.Fatal("the Role clause accepted a creature I don't control")
	}
}

// --- Ellivere, Charmed Clothier, Redtooth Genealogist, Syr Armont ---

func TestElliverePutsAVirtuousRoleOnAnotherCreatureAndDrawsOnEnchantedHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ellivere of the Wild Court", "Legendary Creature — Human Knight", elliverOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the enter trigger")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if roles := effectstest.RolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Virtuous Role" {
		t.Fatalf("roles on the bear = %+v, want a Virtuous Role", roles)
	}
	hand := effectstest.HandCount(g, me)
	effectstest.AttackWith(t, g, opp.ID, bear, plain)
	effectstest.PassPriorityAroundTable(t, g)
	// Only the enchanted bear draws: one card, not two.
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand grew by %d, want 1 (only the enchanted creature connects for a card)", got-hand)
	}
}

func TestCharmedClothierAndRedtoothGenealogistGiveARoyalRole(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Charmed Clothier", "Creature — Faerie Advisor", charmedClothierOracle},
		{"Redtooth Genealogist", "Creature — Elf Advisor", redtoothOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
			self := effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.LatestPickTarget(g, me.ID) == nil {
				t.Fatal("no target prompt")
			}
			effectstest.PickTriggerTarget(t, g, me.ID, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if roles := effectstest.RolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Royal Role" {
				t.Errorf("roles on the bear = %+v, want a Royal Role", roles)
			}
			if roles := effectstest.RolesOn(g, self); len(roles) != 0 {
				t.Error("the card put a Role on itself")
			}
		})
	}
}

func TestSyrArmontMakesEnchantedCreaturesYouControlBigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleMonster, theirs, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Syr Armont, the Redeemer", "Legendary Creature — Human Knight", syrArmontOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	// 2/2 +1/+1 (Monster) +1/+1 (Armont).
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, plain); p != 2 {
		t.Errorf("unenchanted creature has power %d, want 2", p)
	}
	// Their enchanted creature gets the Monster Role's own +1/+1 only.
	if p := effectstest.EffectivePower(t, g, theirs); p != 3 {
		t.Errorf("opponent's enchanted creature has power %d, want 3", p)
	}
}
