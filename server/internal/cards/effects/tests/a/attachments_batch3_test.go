package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Sword of Vengeance: four keywords on one grant ---------------

func TestSwordOfVengeanceGrantsAllFourKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Sword of Vengeance", effectstest.SwordOfVengeanceOracle)

	effectstest.EquipTo(t, g, me.ID, sword, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	ab := effectstest.EffectiveAbilities(t, g, bear)
	for _, want := range []string{"first strike", "vigilance", "trample", "haste"} {
		if !effectstest.ContainsString(ab, want) {
			t.Errorf("abilities %v missing %q", ab, want)
		}
	}
}

// --- Champion's Helm: a grant with a condition on the host ---------

// The pump is unconditional and the hexproof is not. One Helm, two
// creatures, two different answers — and the difference is read per
// recompute rather than captured at equip time.
func TestChampionsHelmGrantsHexproofOnlyToALegendaryHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plain := effectstest.SeedBear(g, me.ID)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	helm := effectstest.SeedEquipment(g, me.ID, "Champion's Helm", effectstest.ChampionsHelmOracle)

	effectstest.EquipTo(t, g, me.ID, helm, plain)
	if got := effectstest.EffectivePower(t, g, plain); got != 4 {
		t.Errorf("non-legendary power %d, want 4 — the pump is unconditional", got)
	}
	if ab := effectstest.EffectiveAbilities(t, g, plain); effectstest.ContainsString(ab, "hexproof") {
		t.Errorf("a non-legendary host got hexproof: %v", ab)
	}

	effectstest.EquipTo(t, g, me.ID, helm, legend)
	if ab := effectstest.EffectiveAbilities(t, g, legend); !effectstest.ContainsString(ab, "hexproof") {
		t.Errorf("the legendary host lacks hexproof: %v", ab)
	}
	if ab := effectstest.EffectiveAbilities(t, g, plain); effectstest.ContainsString(ab, "hexproof") {
		t.Errorf("hexproof stayed behind on the creature the Helm left: %v", ab)
	}
}

// --- Blackblade Reforged: a counted pump and a second equip --------

// The bonus is read from the board on every recompute, so a land
// entering after the equip grows the creature with no re-equip.
func TestBlackbladeReforgedCountsLandsAndGrowsWithThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	blade := effectstest.SeedEquipment(g, me.ID, "Blackblade Reforged", effectstest.BlackbladeOracle)
	for i := 0; i < 3; i++ {
		effectstest.SeedPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	}

	effectstest.EquipTo(t, g, me.ID, blade, legend)
	if got := effectstest.EffectivePower(t, g, legend); got != 5 {
		t.Fatalf("power %d, want 5 (2 base + 3 lands)", got)
	}

	effectstest.SeedPermanent(g, me.ID, "Island", "Basic Land — Island")
	if got := effectstest.EffectivePower(t, g, legend); got != 6 {
		t.Errorf("power %d after a fourth land, want 6 — the count is re-read, not captured", got)
	}
	if got := effectstest.EffectiveToughness(t, g, legend); got != 6 {
		t.Errorf("toughness %d, want 6 — the bonus is +1/+1 per land", got)
	}
}

// An opponent's lands are not "lands you control", and the "you" is
// the EQUIPMENT's controller (CR 109.5).
func TestBlackbladeReforgedIgnoresOpponentsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	blade := effectstest.SeedEquipment(g, me.ID, "Blackblade Reforged", effectstest.BlackbladeOracle)
	effectstest.SeedPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	for i := 0; i < 5; i++ {
		effectstest.SeedPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	}

	effectstest.EquipTo(t, g, me.ID, blade, legend)
	if got := effectstest.EffectivePower(t, g, legend); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1 of my lands)", got)
	}
}

// The cheap equip carries a legendary-only target clause, validated
// at announce like any other (CR 601.2c).
func TestBlackbladeReforgedCheapEquipRefusesANonLegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plain := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Blackblade Reforged", effectstest.BlackbladeOracle)

	err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	})
	if err == nil {
		t.Fatal("equip-legendary-creature {3} accepted a non-legendary creature")
	}

	// The {7} ability has no such restriction — it is index 1, and it
	// is what equipTo reaches for.
	effectstest.EquipTo(t, g, me.ID, blade, plain)
	if host := effectstest.AttachmentHostOf(t, g, blade); host.ID != plain {
		t.Errorf("the unrestricted equip did not attach: %+v", host)
	}
}

// --- Cranial Plating: attach without the sorcery-speed gate --------

// The whole card is the missing gate. Declare attackers, then move
// the Plating onto the unblocked creature — a play the equip ability
// on every other Equipment in the catalog cannot make.
func TestCranialPlatingAttachesAtInstantSpeedInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.SeedBear(g, me.ID)
	second := effectstest.SeedBear(g, me.ID)
	plating := effectstest.SeedEquipment(g, me.ID, "Cranial Plating", effectstest.CranialPlatingOracle)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(first, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, plating, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: second}},
	}); err != nil {
		t.Fatalf("the {B}{B} attach is not instant-speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, plating); host.Kind != game.TargetCard || host.ID != second {
		t.Fatalf("Plating AttachedTo = %+v, want card %s", host, second)
	}
	// The Plating counts itself: one artifact on the board, +1/+0.
	if got := effectstest.EffectivePower(t, g, second); got != 3 {
		t.Errorf("power %d, want 3 — the Plating is an artifact you control", got)
	}
}

// --- Sword of the Animist: the first attack trigger on an Equipment -

func TestSwordOfTheAnimistTriggersWhenTheEquippedCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Sword of the Animist", effectstest.SwordOfTheAnimistOracle)
	effectstest.EquipTo(t, g, me.ID, sword, bear)

	// Both attack in ONE declaration (CR 508.1), which is what the
	// lock-in harvests from (#859): the equipped creature is the
	// Sword's subject and the other one is not, so the whole
	// declaration is worth exactly one trigger.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{other, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	if effectstest.TriggersOnStackFrom(g, sword) != 1 {
		t.Errorf("attack triggers from the Sword = %d, want 1 (only the equipped creature)", effectstest.TriggersOnStackFrom(g, sword))
	}
}

// --- Argentum Armor: a TARGETED attack trigger --------------------

func TestArgentumArmorDestroysTheChosenPermanentOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	victim := effectstest.SeedPermanent(g, opp.ID, "Signet", "Artifact")
	armor := effectstest.SeedEquipment(g, me.ID, "Argentum Armor", effectstest.ArgentumArmorOracle)
	effectstest.EquipTo(t, g, me.ID, armor, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 8 {
		t.Fatalf("power %d, want 8", got)
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick_target prompt from Argentum Armor")
	}
	if !effectstest.HasID(prompt.PickTargetCards, victim) {
		t.Fatalf("legal set %v missing the opponent's artifact", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("Argentum Armor did not destroy the chosen permanent")
	}
}

// --- Quietus Spike: life loss, halved and rounded up --------------

func TestQuietusSpikeHalvesLifeRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	spike := effectstest.SeedEquipment(g, me.ID, "Quietus Spike", effectstest.QuietusSpikeOracle)
	effectstest.EquipTo(t, g, me.ID, spike, bear)

	if ab := effectstest.EffectiveAbilities(t, g, bear); !effectstest.ContainsString(ab, "deathtouch") {
		t.Errorf("abilities %v missing deathtouch", ab)
	}

	// An odd life total is the case "rounded up" exists for.
	opp.Life = 21
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != 10 {
		t.Errorf("life %d, want 10 — 21 loses 11, rounded up", opp.Life)
	}
}

// --- Goldvein Pick: a Treasure per connection ---------------------

func TestGoldveinPickMakesATreasureOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	pick := effectstest.SeedEquipment(g, me.ID, "Goldvein Pick", effectstest.GoldveinPickOracle)
	effectstest.EquipTo(t, g, me.ID, pick, bear)

	before := effectstest.CountTreasuresControlledBy(g, me.ID)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountTreasuresControlledBy(g, me.ID) - before; got != 1 {
		t.Errorf("made %d Treasures, want 1", got)
	}
}

// --- Lavaspur Boots: ward granted to the HOST ---------------------

// The ward trigger watches the EQUIPPED CREATURE, not the Equipment,
// and it behaves like every other ward: the creature is a LEGAL
// target, the payment is offered to the spell's controller, and a
// decline counters the spell. A targeting-gate implementation would
// pass none of that.
func TestLavaspurBootsWardsTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	boots := effectstest.SeedEquipment(g, me.ID, "Lavaspur Boots", effectstest.LavaspurBootsOracle)
	effectstest.EquipTo(t, g, me.ID, boots, bear)

	if ab := effectstest.EffectiveAbilities(t, g, bear); !effectstest.ContainsString(ab, "haste") {
		t.Errorf("abilities %v missing haste", ab)
	}

	blade := effectstest.CastAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's removal spell")
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the equipped creature's controller must not be asked to pay")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	boots := effectstest.SeedEquipment(g, me.ID, "Lavaspur Boots", effectstest.LavaspurBootsOracle)
	effectstest.EquipTo(t, g, me.ID, boots, bear)

	effectstest.CastAtWardedCreature(t, g, me, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("ward triggered on its own controller's spell")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the unwarded Doom Blade should have killed the bear")
	}
}

// --- All That Glitters: a counted Aura that counts itself ---------

func TestAllThatGlittersCountsItselfAndYourArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.SeedPermanent(g, me.ID, "Signet", "Artifact")
	effectstest.SeedPermanent(g, me.ID, "Ghostly Prison", "Enchantment")
	effectstest.SeedPermanent(g, opp.ID, "Their Signet", "Artifact")

	effectstest.CastCatalogSpell(t, g, "All That Glitters", effectstest.AuraTypeLine, effectstest.AllThatGlittersOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	// 1 artifact + 1 enchantment + the Aura itself = +3/+3.
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5 — the Aura counts itself and ignores the opponent's", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 5 {
		t.Errorf("toughness %d, want 5", got)
	}
}

// --- Ancestral Mask: "other", and everyone's ----------------------

func TestAncestralMaskExcludesItselfAndCountsEveryPlayersEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)

	mask := effectstest.CastCatalogSpell(t, g, "Ancestral Mask", effectstest.AuraTypeLine, effectstest.AncestralMaskOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mask) {
		t.Fatal("Ancestral Mask should resolve to the battlefield")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Fatalf("power %d, want 2 — \"each OTHER enchantment\" excludes the Mask", got)
	}

	effectstest.SeedPermanent(g, me.ID, "Mine", "Enchantment")
	effectstest.SeedPermanent(g, opp.ID, "Theirs", "Enchantment")
	if got := effectstest.EffectivePower(t, g, bear); got != 6 {
		t.Errorf("power %d, want 6 — two other enchantments at +2/+2 each, controller irrelevant", got)
	}
}

// --- Blanchwood Armor: the land TYPE, not the name ----------------

func TestBlanchwoodArmorCountsForestsByLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.SeedPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	// A dual with the Forest type counts; a plain Island does not.
	effectstest.SeedPermanent(g, me.ID, "Stomping Ground", "Land — Mountain Forest")
	effectstest.SeedPermanent(g, me.ID, "Island", "Basic Land — Island")

	effectstest.CastCatalogSpell(t, g, "Blanchwood Armor", effectstest.AuraTypeLine, effectstest.BlanchwoodArmorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 — two Forests by type, the Island excluded", got)
	}
}

// --- Curiosity: the wider damage condition ------------------------

// Not combat damage, and an OPPONENT rather than any player.
func TestCuriosityDrawsOnNoncombatDamageToAnOpponentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)

	effectstest.CastCatalogSpell(t, g, "Curiosity", effectstest.AuraTypeLine, effectstest.CuriosityOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	// Damage to my own face draws nothing.
	effectstest.EmitNoncombatDamage(g, bear, me.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("Curiosity asked about damage dealt to its own controller")
	}

	before := me.Hand.Size()
	effectstest.EmitNoncombatDamage(g, bear, opp.ID, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// The "may" is a real question, and No means no card.
func TestCuriositysMayIsAnswerable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)

	effectstest.CastCatalogSpell(t, g, "Curiosity", effectstest.AuraTypeLine, effectstest.CuriosityOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	before := me.Hand.Size()
	effectstest.EmitNoncombatDamage(g, bear, opp.ID, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != before {
		t.Errorf("declining the draw still drew a card")
	}
}

// --- Spirit Link: damage to ANYTHING, life to the Aura's owner ----

func TestSpiritLinkGainsLifeOnDamageDealtToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.SeedBear(g, opp.ID)
	mine := effectstest.SeedBear(g, me.ID)

	effectstest.CastCatalogSpell(t, g, "Spirit Link", effectstest.AuraTypeLine, effectstest.SpiritLinkOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)

	before := me.Life
	// Damage to a CREATURE — the Swords' condition would miss this.
	effectstest.EmitNoncombatDamage(g, mine, theirs, 3)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Life - before; got != 3 {
		t.Errorf("gained %d life, want 3", got)
	}
}

// --- Shadowspear: a turn-scoped removal of two keywords -----------

// The ability strips the two keywords that blank removal, from
// OPPONENTS' permanents only, and gives them back at end of turn.
func TestShadowspearStripsHexproofAndIndestructibleFromOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)

	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Hexproof Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		Keywords: []string{"hexproof", "indestructible"},
	})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Hexproof Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"hexproof", "indestructible"},
	})
	spear := effectstest.SeedEquipment(g, me.ID, "Shadowspear", effectstest.ShadowspearOracle)

	if ab := effectstest.EffectiveAbilities(t, g, theirs); !effectstest.ContainsString(ab, "hexproof") {
		t.Fatalf("setup: the opponent's creature should start with hexproof, got %v", ab)
	}

	if err := g.ActivateCatalogAbility(me.ID, spear, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Shadowspear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if ab := effectstest.EffectiveAbilities(t, g, theirs); effectstest.ContainsString(ab, "hexproof") || effectstest.ContainsString(ab, "indestructible") {
		t.Errorf("the opponent's permanent kept %v", ab)
	}
	if ab := effectstest.EffectiveAbilities(t, g, mine); !effectstest.ContainsString(ab, "hexproof") || !effectstest.ContainsString(ab, "indestructible") {
		t.Errorf("my own permanent lost keywords it should keep: %v", ab)
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	if ab := effectstest.EffectiveAbilities(t, g, theirs); !effectstest.ContainsString(ab, "hexproof") {
		t.Errorf("the removal outlived the turn: %v", ab)
	}
}

// --- Angelic Gift: an Aura with an enters trigger -----------------

// The Aura attaches at resolution (ADR 0036 decision 5) AND fires an
// enters trigger, and the two do not interfere: the Gift ends up on
// the creature, the creature ends up with flying, and the card
// replaces itself.
func TestAngelicGiftAttachesGrantsFlyingAndCantrips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)

	before := me.Hand.Size()
	gift := effectstest.CastCatalogSpell(t, g, "Angelic Gift", effectstest.AuraTypeLine, effectstest.AngelicGiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(gift) {
		t.Fatal("Angelic Gift should resolve to the battlefield")
	}
	if host := effectstest.AttachmentHostOf(t, g, gift); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("Angelic Gift AttachedTo = %+v, want card %s", host, bear)
	}
	if ab := effectstest.EffectiveAbilities(t, g, bear); !effectstest.ContainsString(ab, "flying") {
		t.Errorf("abilities %v missing flying", ab)
	}
	// `before` is measured before castCatalogSpell seeds the Gift, so
	// the seed (+1) and the cast (-1) cancel and what is left is the
	// trigger's draw: the Aura replaces itself.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d — the enters trigger did not draw", got, before+1)
	}
}
