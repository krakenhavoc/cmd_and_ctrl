package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_batch3_test.go covers the third S24 attachment batch —
// thirteen Equipment and ten Auras on the surface #374 / #379 / #511
// built.
//
// The plain cards are not retested here. Bonesplitter already pins
// "equip attaches and the 7c static reaches the host", and Loxodon
// Warhammer pins "a grant and a pump move together", so Sword of
// Vengeance and Behemoth Sledge would only re-prove those. What IS
// here is everything this batch does that no earlier card did:
//
//   - a static whose bonus is COUNTED from the board on every
//     recompute rather than fixed (Blackblade, All That Glitters,
//     Ancestral Mask, Blanchwood Armor)
//   - a grant with a CONDITION on the host (Champion's Helm)
//   - an attach ability that is NOT an equip ability, because it has
//     no sorcery-speed gate (Cranial Plating)
//   - a second equip ability with a tighter target clause (Blackblade)
//   - "whenever equipped creature attacks" (Sword of the Animist,
//     Argentum Armor) and a targeted version of it
//   - a damage condition wider than the Swords' (Curiosity, Spirit
//     Link)
//   - ward granted to the HOST rather than printed on the source
//     (Lavaspur Boots)

const (
	swordOfVengeanceOracle  = "e366afb3-c447-4bde-b358-41c8568142d5"
	championsHelmOracle     = "c01aafb8-3da5-4eb1-8731-2a223e747d63"
	blackbladeOracle        = "dca51281-fb21-45b6-beb4-1f13397caee2"
	cranialPlatingOracle    = "75564721-e4e9-463a-b49e-f0c7cd6f53a7"
	swordOfTheAnimistOracle = "d79cbc61-6c15-48ea-bbba-3cffb819ccba"
	argentumArmorOracle     = "d0a1f39b-cda4-4925-83f8-2161f575edfb"
	quietusSpikeOracle      = "0ff0a309-9b4a-4f9e-9ea6-c4d07906ebec"
	goldveinPickOracle      = "c1624d10-8838-4af8-aea1-a96c0fe6fd6b"
	lavaspurBootsOracle     = "63bfb9ca-d9d5-4c17-be39-82eb115bc20c"
	allThatGlittersOracle   = "a4d751e0-41c1-4e90-853d-512f385acd81"
	ancestralMaskOracle     = "db5380ed-ba28-4ea2-abc3-4998e2022903"
	blanchwoodArmorOracle   = "80ea56ad-e741-4a85-b4e8-ce62e7d593d5"
	curiosityOracle         = "223fa044-d387-4884-bf4e-75f1b61c6a46"
	spiritLinkOracle        = "c77ff526-c0a8-45c7-9730-2e306a0d01b8"
	shadowspearOracle       = "8b27326f-e7b8-4a4d-b589-df459246d19a"
)

// seedLegendaryCreature is a 2/2 with the Legendary supertype, which
// is the half Champion's Helm and Blackblade Reforged both read.
func seedLegendaryCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// seedPermanent puts an arbitrary non-creature permanent on the
// battlefield — the things the counted pumps count.
func seedPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
}

// --- Sword of Vengeance: four keywords on one grant ---------------

func TestSwordOfVengeanceGrantsAllFourKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	sword := seedEquipment(g, me.ID, "Sword of Vengeance", swordOfVengeanceOracle)

	equipTo(t, g, me.ID, sword, bear)

	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	ab := effectiveAbilities(t, g, bear)
	for _, want := range []string{"first strike", "vigilance", "trample", "haste"} {
		if !containsString(ab, want) {
			t.Errorf("abilities %v missing %q", ab, want)
		}
	}
}

// --- Champion's Helm: a grant with a condition on the host ---------

// The pump is unconditional and the hexproof is not. One Helm, two
// creatures, two different answers — and the difference is read per
// recompute rather than captured at equip time.
func TestChampionsHelmGrantsHexproofOnlyToALegendaryHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plain := seedBear(g, me.ID)
	legend := seedLegendaryCreature(g, me.ID, "Commander")
	helm := seedEquipment(g, me.ID, "Champion's Helm", championsHelmOracle)

	equipTo(t, g, me.ID, helm, plain)
	if got := effectivePower(t, g, plain); got != 4 {
		t.Errorf("non-legendary power %d, want 4 — the pump is unconditional", got)
	}
	if ab := effectiveAbilities(t, g, plain); containsString(ab, "hexproof") {
		t.Errorf("a non-legendary host got hexproof: %v", ab)
	}

	equipTo(t, g, me.ID, helm, legend)
	if ab := effectiveAbilities(t, g, legend); !containsString(ab, "hexproof") {
		t.Errorf("the legendary host lacks hexproof: %v", ab)
	}
	if ab := effectiveAbilities(t, g, plain); containsString(ab, "hexproof") {
		t.Errorf("hexproof stayed behind on the creature the Helm left: %v", ab)
	}
}

// --- Blackblade Reforged: a counted pump and a second equip --------

// The bonus is read from the board on every recompute, so a land
// entering after the equip grows the creature with no re-equip.
func TestBlackbladeReforgedCountsLandsAndGrowsWithThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	legend := seedLegendaryCreature(g, me.ID, "Commander")
	blade := seedEquipment(g, me.ID, "Blackblade Reforged", blackbladeOracle)
	for i := 0; i < 3; i++ {
		seedPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	}

	equipTo(t, g, me.ID, blade, legend)
	if got := effectivePower(t, g, legend); got != 5 {
		t.Fatalf("power %d, want 5 (2 base + 3 lands)", got)
	}

	seedPermanent(g, me.ID, "Island", "Basic Land — Island")
	if got := effectivePower(t, g, legend); got != 6 {
		t.Errorf("power %d after a fourth land, want 6 — the count is re-read, not captured", got)
	}
	if got := effectiveToughness(t, g, legend); got != 6 {
		t.Errorf("toughness %d, want 6 — the bonus is +1/+1 per land", got)
	}
}

// An opponent's lands are not "lands you control", and the "you" is
// the EQUIPMENT's controller (CR 109.5).
func TestBlackbladeReforgedIgnoresOpponentsLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	legend := seedLegendaryCreature(g, me.ID, "Commander")
	blade := seedEquipment(g, me.ID, "Blackblade Reforged", blackbladeOracle)
	seedPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	for i := 0; i < 5; i++ {
		seedPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	}

	equipTo(t, g, me.ID, blade, legend)
	if got := effectivePower(t, g, legend); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1 of my lands)", got)
	}
}

// The cheap equip carries a legendary-only target clause, validated
// at announce like any other (CR 601.2c).
func TestBlackbladeReforgedCheapEquipRefusesANonLegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plain := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Blackblade Reforged", blackbladeOracle)

	err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	})
	if err == nil {
		t.Fatal("equip-legendary-creature {3} accepted a non-legendary creature")
	}

	// The {7} ability has no such restriction — it is index 1, and it
	// is what equipTo reaches for.
	equipTo(t, g, me.ID, blade, plain)
	if host := attachmentHostOf(t, g, blade); host.ID != plain {
		t.Errorf("the unrestricted equip did not attach: %+v", host)
	}
}

// --- Cranial Plating: attach without the sorcery-speed gate --------

// The whole card is the missing gate. Declare attackers, then move
// the Plating onto the unblocked creature — a play the equip ability
// on every other Equipment in the catalog cannot make.
func TestCranialPlatingAttachesAtInstantSpeedInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := seedBear(g, me.ID)
	second := seedBear(g, me.ID)
	plating := seedEquipment(g, me.ID, "Cranial Plating", cranialPlatingOracle)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(first, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, plating, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: second}},
	}); err != nil {
		t.Fatalf("the {B}{B} attach is not instant-speed: %v", err)
	}
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, plating); host.Kind != game.TargetCard || host.ID != second {
		t.Fatalf("Plating AttachedTo = %+v, want card %s", host, second)
	}
	// The Plating counts itself: one artifact on the board, +1/+0.
	if got := effectivePower(t, g, second); got != 3 {
		t.Errorf("power %d, want 3 — the Plating is an artifact you control", got)
	}
}

// --- Sword of the Animist: the first attack trigger on an Equipment -

func TestSwordOfTheAnimistTriggersWhenTheEquippedCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	other := seedBear(g, me.ID)
	sword := seedEquipment(g, me.ID, "Sword of the Animist", swordOfTheAnimistOracle)
	equipTo(t, g, me.ID, sword, bear)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if triggersOnStackFrom(g, sword) != 0 || len(g.PendingChoices) != 0 {
		t.Fatalf("an unequipped creature attacking triggered the Sword")
	}

	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if triggersOnStackFrom(g, sword) != 1 {
		t.Errorf("attack triggers from the Sword = %d, want 1", triggersOnStackFrom(g, sword))
	}
}

// --- Argentum Armor: a TARGETED attack trigger --------------------

func TestArgentumArmorDestroysTheChosenPermanentOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	victim := seedPermanent(g, opp.ID, "Signet", "Artifact")
	armor := seedEquipment(g, me.ID, "Argentum Armor", argentumArmorOracle)
	equipTo(t, g, me.ID, armor, bear)

	if got := effectivePower(t, g, bear); got != 8 {
		t.Fatalf("power %d, want 8", got)
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick_target prompt from Argentum Armor")
	}
	if !hasID(prompt.PickTargetCards, victim) {
		t.Fatalf("legal set %v missing the opponent's artifact", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("Argentum Armor did not destroy the chosen permanent")
	}
}

// --- Quietus Spike: life loss, halved and rounded up --------------

func TestQuietusSpikeHalvesLifeRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	spike := seedEquipment(g, me.ID, "Quietus Spike", quietusSpikeOracle)
	equipTo(t, g, me.ID, spike, bear)

	if ab := effectiveAbilities(t, g, bear); !containsString(ab, "deathtouch") {
		t.Errorf("abilities %v missing deathtouch", ab)
	}

	// An odd life total is the case "rounded up" exists for.
	opp.Life = 21
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)

	if opp.Life != 10 {
		t.Errorf("life %d, want 10 — 21 loses 11, rounded up", opp.Life)
	}
}

// --- Goldvein Pick: a Treasure per connection ---------------------

func TestGoldveinPickMakesATreasureOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	pick := seedEquipment(g, me.ID, "Goldvein Pick", goldveinPickOracle)
	equipTo(t, g, me.ID, pick, bear)

	before := countTreasuresControlledBy(g, me.ID)
	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	passPriorityAroundTable(t, g)

	if got := countTreasuresControlledBy(g, me.ID) - before; got != 1 {
		t.Errorf("made %d Treasures, want 1", got)
	}
}

func countTreasuresControlledBy(g *game.Game, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == "Treasure" {
				n++
			}
		}
	})
	return n
}

// --- Lavaspur Boots: ward granted to the HOST ---------------------

// The ward trigger watches the EQUIPPED CREATURE, not the Equipment,
// and it behaves like every other ward: the creature is a LEGAL
// target, the payment is offered to the spell's controller, and a
// decline counters the spell. A targeting-gate implementation would
// pass none of that.
func TestLavaspurBootsWardsTheEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	boots := seedEquipment(g, me.ID, "Lavaspur Boots", lavaspurBootsOracle)
	equipTo(t, g, me.ID, boots, bear)

	if ab := effectiveAbilities(t, g, bear); !containsString(ab, "haste") {
		t.Errorf("abilities %v missing haste", ab)
	}

	blade := castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's removal spell")
	}
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the equipped creature's controller must not be asked to pay")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("a declined ward must counter the spell — the bear should live")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
}

// Your own spell on your own equipped creature does not trigger it —
// the difference between ward and shroud, measured against the HOST's
// controller.
func TestLavaspurBootsDoesNotWardAgainstItsOwnController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	boots := seedEquipment(g, me.ID, "Lavaspur Boots", lavaspurBootsOracle)
	equipTo(t, g, me.ID, boots, bear)

	castAtWardedCreature(t, g, me, bear)
	passPriorityAroundTable(t, g)

	if hasPayUnlessFor(g, me.ID) {
		t.Error("ward triggered on its own controller's spell")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the unwarded Doom Blade should have killed the bear")
	}
}

// --- All That Glitters: a counted Aura that counts itself ---------

func TestAllThatGlittersCountsItselfAndYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)
	seedPermanent(g, me.ID, "Signet", "Artifact")
	seedPermanent(g, me.ID, "Ghostly Prison", "Enchantment")
	seedPermanent(g, opp.ID, "Their Signet", "Artifact")

	castCatalogSpell(t, g, "All That Glitters", auraTypeLine, allThatGlittersOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	// 1 artifact + 1 enchantment + the Aura itself = +3/+3.
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5 — the Aura counts itself and ignores the opponent's", got)
	}
	if got := effectiveToughness(t, g, bear); got != 5 {
		t.Errorf("toughness %d, want 5", got)
	}
}

// --- Ancestral Mask: "other", and everyone's ----------------------

func TestAncestralMaskExcludesItselfAndCountsEveryPlayersEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)

	mask := castCatalogSpell(t, g, "Ancestral Mask", auraTypeLine, ancestralMaskOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mask) {
		t.Fatal("Ancestral Mask should resolve to the battlefield")
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Fatalf("power %d, want 2 — \"each OTHER enchantment\" excludes the Mask", got)
	}

	seedPermanent(g, me.ID, "Mine", "Enchantment")
	seedPermanent(g, opp.ID, "Theirs", "Enchantment")
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("power %d, want 6 — two other enchantments at +2/+2 each, controller irrelevant", got)
	}
}

// --- Blanchwood Armor: the land TYPE, not the name ----------------

func TestBlanchwoodArmorCountsForestsByLandType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)
	seedPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	// A dual with the Forest type counts; a plain Island does not.
	seedPermanent(g, me.ID, "Stomping Ground", "Land — Mountain Forest")
	seedPermanent(g, me.ID, "Island", "Basic Land — Island")

	castCatalogSpell(t, g, "Blanchwood Armor", auraTypeLine, blanchwoodArmorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 — two Forests by type, the Island excluded", got)
	}
}

// --- Curiosity: the wider damage condition ------------------------

// Not combat damage, and an OPPONENT rather than any player.
func TestCuriosityDrawsOnNoncombatDamageToAnOpponentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)

	castCatalogSpell(t, g, "Curiosity", auraTypeLine, curiosityOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	// Damage to my own face draws nothing.
	emitNoncombatDamage(g, bear, me.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("Curiosity asked about damage dealt to its own controller")
	}

	before := me.Hand.Size()
	emitNoncombatDamage(g, bear, opp.ID, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// The "may" is a real question, and No means no card.
func TestCuriositysMayIsAnswerable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)

	castCatalogSpell(t, g, "Curiosity", auraTypeLine, curiosityOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	before := me.Hand.Size()
	emitNoncombatDamage(g, bear, opp.ID, 1)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != before {
		t.Errorf("declining the draw still drew a card")
	}
}

func emitNoncombatDamage(g *game.Game, source, victim uuid.UUID, amount int) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:   game.EventDealDamage,
			Source: source,
			Target: victim,
			Amount: amount,
		})
	})
}

// --- Spirit Link: damage to ANYTHING, life to the Aura's owner ----

func TestSpiritLinkGainsLifeOnDamageDealtToACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := seedBear(g, opp.ID)
	mine := seedBear(g, me.ID)

	castCatalogSpell(t, g, "Spirit Link", auraTypeLine, spiritLinkOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)

	before := me.Life
	// Damage to a CREATURE — the Swords' condition would miss this.
	emitNoncombatDamage(g, mine, theirs, 3)
	passPriorityAroundTable(t, g)

	if got := me.Life - before; got != 3 {
		t.Errorf("gained %d life, want 3", got)
	}
}

// --- Shadowspear: a turn-scoped removal of two keywords -----------

// The ability strips the two keywords that blank removal, from
// OPPONENTS' permanents only, and gives them back at end of turn.
func TestShadowspearStripsHexproofAndIndestructibleFromOpponentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)

	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Hexproof Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		Keywords: []string{"hexproof", "indestructible"},
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Hexproof Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"hexproof", "indestructible"},
	})
	spear := seedEquipment(g, me.ID, "Shadowspear", shadowspearOracle)

	if ab := effectiveAbilities(t, g, theirs); !containsString(ab, "hexproof") {
		t.Fatalf("setup: the opponent's creature should start with hexproof, got %v", ab)
	}

	if err := g.ActivateCatalogAbility(me.ID, spear, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Shadowspear: %v", err)
	}
	passPriorityAroundTable(t, g)

	if ab := effectiveAbilities(t, g, theirs); containsString(ab, "hexproof") || containsString(ab, "indestructible") {
		t.Errorf("the opponent's permanent kept %v", ab)
	}
	if ab := effectiveAbilities(t, g, mine); !containsString(ab, "hexproof") || !containsString(ab, "indestructible") {
		t.Errorf("my own permanent lost keywords it should keep: %v", ab)
	}

	advanceToNextSeatsTurn(t, g)
	if ab := effectiveAbilities(t, g, theirs); !containsString(ab, "hexproof") {
		t.Errorf("the removal outlived the turn: %v", ab)
	}
}
