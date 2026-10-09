package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reconfigure_cards_test.go — CR 702.151 through the real activation
// path (#2639), and the seventeen reconfigure cards.

const (
	lizardBladesOracle       = "f6314e50-a610-4fff-8f60-803d23460ff0"
	rabbitBatteryOracle      = "c739e180-2f14-41ed-8e7e-50b7df985f35"
	cloudsteelKirinOracle    = "a34a71eb-26a4-4e37-8d3c-a0352bea491b"
	bladeOfTheOniOracle      = "6d9ac636-3c6c-40ca-90f3-1a3387836bb4"
	ogreHeadHelmOracle       = "b73c297a-1c39-450f-99c3-c8daa31c5f3f"
	chainflailOracle         = "e77a1809-d21a-4bb6-88a6-35eb13771b1d"
	razorfieldRipperOracle   = "ef842299-8a31-4793-9142-2bcd7a8b2fab"
	lionSashOracle           = "50330e48-db74-4c5f-a0bc-f9a8607e8f31"
	simianSlingOracle        = "e49c4d9a-6413-440f-ba5b-396fea6c03d9"
	tanukiTransplanterOracle = "4d1c7f42-19ef-405e-9edc-50a81b78f97c"
	bronzeplateBoarOracle    = "0a09f825-0484-40b4-8adf-1028385db480"
	acquisitionOctopusOracle = "234ff22f-2ff1-4a73-a7c5-e9c53557c4c6"
	realityChipOracle        = "751f22bc-bcea-4213-a37b-b5f72448a4c4"
	komainuOracle            = "5e4d69de-1876-49a8-8ec6-a891d4a84ccf"
	armguardFamiliarOracle   = "a7955f10-f502-4893-88ad-f36e7ff0af4f"
	leechGauntletOracle      = "1aaddaeb-baab-42b6-a442-d97339da54db"
	webspinnerCuffOracle     = "7826f3f5-c8f2-43f5-a837-ff016da625e6"
)

// seedReconfigure puts a reconfigure Equipment creature on the
// battlefield, unattached and able to attack.
func seedReconfigure(g *game.Game, owner uuid.UUID, name, subtype, oracle string, power, toughness int) uuid.UUID {
	return b12Push(g, owner, name, "Artifact Creature — Equipment "+subtype, oracle, power, toughness)
}

// reconfigureRow is the index of the row whose label ends in `suffix`
// ("(attach)", "(pay {E}{E}{E}, unattach)").
func reconfigureRow(t *testing.T, oracle, suffix string) int {
	t.Helper()
	for i, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle}) {
		if strings.HasSuffix(ab.Label, suffix) {
			return i
		}
	}
	t.Fatalf("%s has no row ending %q", oracle, suffix)
	return -1
}

// reconfigureOnto activates the attach row and lets it resolve.
func reconfigureOnto(t *testing.T, g *game.Game, controller, eq, host uuid.UUID, oracle string) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, eq, reconfigureRow(t, oracle, "(attach)"), game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("reconfigure attach: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// The whole keyword on Lizard Blades: attach to another creature you
// control, stop being a creature, grant double strike; unattach only
// while attached; a creature again afterwards.
func TestReconfigureAttachesAndUnattaches(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	blades := seedReconfigure(g, me.ID, "Lizard Blades", "Lizard", lizardBladesOracle, 1, 1)

	// "Another target creature": not itself.
	if err := g.ActivateCatalogAbility(me.ID, blades, reconfigureRow(t, lizardBladesOracle, "(attach)"), game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blades}},
	}); err == nil {
		t.Fatal("reconfigure targeted the Equipment itself")
	}
	// "Activate only if this permanent is attached to a creature."
	if err := g.ActivateCatalogAbility(me.ID, blades, reconfigureRow(t, lizardBladesOracle, "(unattach)"), game.ActivateAbilityParams{}); err == nil {
		t.Fatal("unattach was activated while unattached")
	}

	reconfigureOnto(t, g, me.ID, blades, bear, lizardBladesOracle)
	if host := attachmentHostOf(t, g, blades); host.ID != bear {
		t.Fatalf("AttachedTo = %+v, want the bear", host)
	}
	if isCreatureNow(g, blades) {
		t.Error("attached Lizard Blades is still a creature (CR 702.151b)")
	}
	if !effectiveAbilitiesContain(t, g, bear, "double strike") {
		t.Error("the equipped creature has no double strike")
	}

	if err := g.ActivateCatalogAbility(me.ID, blades, reconfigureRow(t, lizardBladesOracle, "(unattach)"), game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("unattach: %v", err)
	}
	passPriorityAroundTable(t, g)
	if attachmentHostOf(t, g, blades).Kind != "" {
		t.Error("still attached after unattach")
	}
	if !isCreatureNow(g, blades) {
		t.Error("Lizard Blades is not a creature again")
	}
	if effectiveAbilitiesContain(t, g, bear, "double strike") {
		t.Error("the bear kept double strike")
	}
}

// Sorcery speed (CR 702.151a): not in combat.
func TestReconfigureIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	boar := seedReconfigure(g, me.ID, "Bronzeplate Boar", "Boar", bronzeplateBoarOracle, 3, 2)
	declareAttack(t, g, opp.ID, bear)
	if err := g.ActivateCatalogAbility(me.ID, boar, reconfigureRow(t, bronzeplateBoarOracle, "(attach)"), game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("reconfigure was activated during combat")
	}
}

// The stat sticks: each pumps and grants as printed.
func TestReconfigureStatEquipment(t *testing.T) {
	cases := []struct {
		name, subtype, oracle string
		p, tough              int
		wantP, wantT          int
		keyword               string
	}{
		{"Rabbit Battery", "Rabbit", rabbitBatteryOracle, 1, 1, 3, 3, "haste"},
		{"Bronzeplate Boar", "Boar", bronzeplateBoarOracle, 3, 2, 5, 4, "trample"},
		{"Leech Gauntlet", "Leech", leechGauntletOracle, 2, 2, 2, 2, "lifelink"},
		{"Webspinner Cuff", "Spider", webspinnerCuffOracle, 1, 4, 3, 6, "reach"},
		{"Armguard Familiar", "Beast", armguardFamiliarOracle, 2, 1, 4, 3, ""},
		{"Simian Sling", "Monkey", simianSlingOracle, 1, 1, 3, 3, ""},
		{"Komainu Battle Armor", "Dog", komainuOracle, 2, 2, 4, 4, "menace"},
		{"Ogre-Head Helm", "Ogre", ogreHeadHelmOracle, 2, 2, 4, 4, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			bear := seedBear(g, me.ID)
			eq := seedReconfigure(g, me.ID, tc.name, tc.subtype, tc.oracle, tc.p, tc.tough)
			reconfigureOnto(t, g, me.ID, eq, bear, tc.oracle)
			if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != tc.wantP || tg != tc.wantT {
				t.Errorf("bear is %d/%d, want %d/%d", p, tg, tc.wantP, tc.wantT)
			}
			if tc.keyword != "" && !effectiveAbilitiesContain(t, g, bear, tc.keyword) {
				t.Errorf("the bear has no %s", tc.keyword)
			}
			if isCreatureNow(g, eq) {
				t.Error("the attached Equipment is still a creature")
			}
		})
	}
}

// Blade of the Oni: base 5/5, menace, and a black Demon in addition.
func TestBladeOfTheOniMakesABlackDemon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	blade := seedReconfigure(g, me.ID, "Blade of the Oni", "Demon", bladeOfTheOniOracle, 3, 1)
	reconfigureOnto(t, g, me.ID, blade, bear, bladeOfTheOniOracle)

	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("bear is %d/%d, want 5/5", p, tg)
	}
	if !effectiveAbilitiesContain(t, g, bear, "menace") {
		t.Error("no menace")
	}
	g.ReadSnapshot(func() {
		c, _ := battlefieldCard(g, bear)
		if !c.HasColor("B") || !c.HasSubtype("Demon") || !c.HasSubtype("Bear") {
			t.Errorf("colors %v subtypes %v, want black and Demon in addition to Bear", c.EffectiveColors(), c.Effective().Subtypes)
		}
		b, _ := battlefieldCard(g, blade)
		if b.HasSubtype("Demon") {
			t.Error("the attached Blade kept its creature type (CR 205.3d)")
		}
	})
}

// Chainflail Centipede pumps whichever creature attacked: itself, or
// its host.
func TestChainflailCentipedePumpsTheAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	centipede := seedReconfigure(g, me.ID, "Chainflail Centipede", "Insect", chainflailOracle, 2, 2)
	declareAttack(t, g, opp.ID, centipede)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, centipede); got != 4 {
		t.Errorf("the attacking Centipede has power %d, want 4", got)
	}

	g2 := newCatalogGame(t)
	me, opp = g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	bear := seedBear(g2, me.ID)
	centipede = seedReconfigure(g2, me.ID, "Chainflail Centipede", "Insect", chainflailOracle, 2, 2)
	reconfigureOnto(t, g2, me.ID, centipede, bear, chainflailOracle)
	declareAttack(t, g2, opp.ID, bear)
	passPriorityAroundTable(t, g2)
	if got := effectivePower(t, g2, bear); got != 4 {
		t.Errorf("the equipped bear has power %d, want 4", got)
	}
}

// Razorfield Ripper: reconfigure paid with energy, and the attack
// trigger counts the energy it just gave.
func TestRazorfieldRipperPaysEnergyAndPumpsByIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	ripper := seedReconfigure(g, me.ID, "Razorfield Ripper", "Rhino", razorfieldRipperOracle, 3, 3)

	if err := g.ActivateCatalogAbility(me.ID, ripper, reconfigureRow(t, razorfieldRipperOracle, "(pay {E}{E}{E}, attach)"), game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("the energy row was paid with no energy")
	}
	if err := g.SetEnergy(me.ID, 3); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ripper, reconfigureRow(t, razorfieldRipperOracle, "(pay {E}{E}{E}, attach)"), game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("energy reconfigure: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Energy != 0 {
		t.Fatalf("energy %d after paying {E}{E}{E}, want 0", me.Energy)
	}
	if host := attachmentHostOf(t, g, ripper); host.ID != bear {
		t.Fatalf("not attached: %+v", host)
	}

	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Energy != 1 {
		t.Errorf("energy %d, want 1", me.Energy)
	}
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("bear is %d/%d, want 3/3 (+1/+1 for one energy)", p, tg)
	}
}

// Acquisition Octopus draws for itself or for its host.
func TestAcquisitionOctopusDrawsForEitherCreature(t *testing.T) {
	for _, attached := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		bear := seedBear(g, me.ID)
		octopus := seedReconfigure(g, me.ID, "Acquisition Octopus", "Octopus", acquisitionOctopusOracle, 2, 2)
		hitter := octopus
		if attached {
			reconfigureOnto(t, g, me.ID, octopus, bear, acquisitionOctopusOracle)
			hitter = bear
		}
		before := me.Hand.Size()
		dealCombatDamageToPlayer(g, hitter, opp.ID, 2)
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != before+1 {
			t.Errorf("attached=%v: hand %d -> %d, want +1", attached, before, got)
		}
	}
}

// Ogre-Head Helm: yes sacrifices it, discards the hand, draws three.
func TestOgreHeadHelmSacrificesForANewHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	helm := seedReconfigure(g, me.ID, "Ogre-Head Helm", "Ogre", ogreHeadHelmOracle, 2, 2)
	reconfigureOnto(t, g, me.ID, helm, bear, ogreHeadHelmOracle)
	hand := me.Hand.Size()
	if hand == 0 {
		t.Fatal("setup: empty hand")
	}

	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, helm); ok {
		t.Error("the Helm was not sacrificed")
	}
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 (the old hand of %d discarded)", got, hand)
	}
}

// Simian Sling: the blocked creature deals 1 damage to the defender.
func TestSimianSlingPingsTheDefendingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	sling := seedReconfigure(g, me.ID, "Simian Sling", "Monkey", simianSlingOracle, 1, 1)
	reconfigureOnto(t, g, me.ID, sling, bear, simianSlingOracle)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)

	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatal(err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("defender life %d, want %d", opp.Life, life-1)
	}
}

// Tanuki Transplanter adds {G} equal to the attacker's power.
func TestTanukiTransplanterAddsGreenEqualToPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tanuki := seedReconfigure(g, me.ID, "Tanuki Transplanter", "Dog", tanukiTransplanterOracle, 2, 4)
	declareAttack(t, g, opp.ID, tanuki)
	passPriorityAroundTable(t, g)
	if got := mkColors(me)["G"]; got != 2 {
		t.Errorf("pool has %d {G}, want 2", got)
	}
}

// Komainu Battle Armor goads every creature of the player it hit.
func TestKomainuBattleArmorGoadsTheDamagedPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	armor := seedReconfigure(g, me.ID, "Komainu Battle Armor", "Dog", komainuOracle, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, other.ID, "Bystander", "Creature — Bear", 2, 2)
	dealCombatDamageToPlayer(g, armor, opp.ID, 2)
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		c, _ := battlefieldCard(g, theirs)
		if !c.IsGoaded() {
			t.Error("the damaged player's creature is not goaded")
		}
		b, _ := battlefieldCard(g, bystander)
		if b.IsGoaded() {
			t.Error("another player's creature was goaded")
		}
	})
}

// Lion Sash: exile a permanent card for a counter; the counters pump
// the host.
func TestLionSashGrowsAndPumpsItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sash := seedReconfigure(g, me.ID, "Lion Sash", "Cat", lionSashOracle, 1, 1)
	creature := pushGraveyardPermanent(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	instant := pushGraveyardPermanent(opp, "Old Bolt", "Instant", "{R}")

	for _, id := range []uuid.UUID{creature, instant} {
		if err := g.ActivateCatalogAbility(me.ID, sash, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if c, _ := battlefieldCard(g, sash); c.Counters["+1/+1"] != 1 {
		t.Fatalf("Sash has %d +1/+1 counters, want 1 (the instant is not a permanent card)", c.Counters["+1/+1"])
	}

	bear := seedBear(g, me.ID)
	reconfigureOnto(t, g, me.ID, sash, bear, lionSashOracle)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("bear is %d/%d, want 3/3", p, tg)
	}
}

// Cloudsteel Kirin: the equipped creature's controller can't lose.
func TestCloudsteelKirinProtectsTheEquippedCreaturesController(t *testing.T) {
	g := newTwoSeatCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	kirin := seedReconfigure(g, me.ID, "Cloudsteel Kirin", "Kirin", cloudsteelKirinOracle, 3, 2)

	g.WithWriteLock(func() {
		if len(g.CantLoseCausesForEffect(me)) != 0 {
			t.Error("an unattached Kirin protects its controller")
		}
	})
	reconfigureOnto(t, g, me.ID, kirin, bear, cloudsteelKirinOracle)
	if !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("no flying")
	}
	me.Life = -3
	checkState(t, g)
	if me.Eliminated {
		t.Fatal("lost the game under an attached Kirin")
	}
}

// The Reality Chip opens the top of the library only while attached.
func TestTheRealityChipPlaysFromTheTopWhileAttached(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	chip := seedReconfigure(g, me.ID, "The Reality Chip", "Jellyfish", realityChipOracle, 0, 4)
	var top uuid.UUID
	g.WithWriteLock(func() {
		c := game.NewCard("On Top", me.ID)
		c.TypeLine = "Instant"
		me.Library.PushTop(c)
		top = c.InstanceID
	})
	granted := func() bool {
		out := false
		g.ReadSnapshot(func() {
			if card, ok := g.LookupCardForEffect(top); ok {
				out = g.CastPermissionForLocked(me.ID, card, game.ZoneLibrary).Granted()
			}
		})
		return out
	}
	if got := g.LibraryTopVisibilityForEffect(me.ID); got != game.LibraryTopOwner {
		t.Errorf("look at the top any time: visibility %v", got)
	}
	if granted() {
		t.Error("the top is playable while the Chip is unattached")
	}
	reconfigureOnto(t, g, me.ID, chip, bear, realityChipOracle)
	if !granted() {
		t.Error("the top is not playable while the Chip is attached")
	}
}
