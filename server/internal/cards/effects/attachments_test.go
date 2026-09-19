package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_test.go covers the S24 Equipment and Aura batch. The
// engine-side relation is pinned in game/attach_test.go; everything
// here is the catalog half — that equip is an ordinary activated
// ability, that "equipped creature" statics reach the host, and that
// the attachment-conditioned triggers fire.

const (
	bonesplitterOracle   = "452e3f5f-ce17-4682-966b-5cc100210aee"
	skullclampOracle     = "65986c1b-8e51-4604-b685-d82fa7d1263a"
	greavesOracle        = "ca204b66-8d0c-431a-8d34-282f7c2d17da"
	bootsOracle          = "c8b143ad-43ec-4e0d-a440-e348daa31391"
	swordFeastOracle     = "d0901053-6de0-46d0-9ee3-8d40510236c1"
	swordFireOracle      = "2ccdc60a-49a9-44b9-a7af-0ebf18b26785"
	rancorOracle         = "9d2d6479-531c-4ce1-b52b-00e36fa63b64"
	curseOpulenceOracle  = "ba0d3df2-3acf-46d7-8d64-8d67d1579adc"
	equipTypeLine        = "Artifact — Equipment"
	auraTypeLine         = "Enchantment — Aura"
	curseAuraTypeLine    = "Enchantment — Aura Curse"
	testCreatureTypeLine = "Creature — Bear"
)

// advanceToMain walks the turn to a main phase, which is where the
// sorcery-speed equip gate opens (CR 702.6b).
func advanceToMain(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached a main phase")
}

// equipTo activates the equipment's equip ability (always the last
// declared ability on these cards) targeting `creature`, then passes
// priority so it resolves.
func equipTo(t *testing.T, g *game.Game, controller, equipment, creature uuid.UUID) {
	t.Helper()
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracleOfBattlefieldCard(t, g, equipment)})
	idx := len(abilities) - 1
	if idx < 0 {
		t.Fatal("equipment has no activated ability")
	}
	if err := g.ActivateCatalogAbility(controller, equipment, idx, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func oracleOfBattlefieldCard(t *testing.T, g *game.Game, id uuid.UUID) string {
	t.Helper()
	var out string
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.OracleID, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

func attachmentHostOf(t *testing.T, g *game.Game, id uuid.UUID) game.TargetRef {
	t.Helper()
	var out game.TargetRef
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.AttachedTo, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// --- Bonesplitter: the shape every other Equipment varies on ------

func TestEquipAttachesAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})

	if got := effectivePower(t, g, bear); got != 2 {
		t.Fatalf("setup: power %d, want 2", got)
	}
	equipTo(t, g, me.ID, sword, bear)

	if host := attachmentHostOf(t, g, sword); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("AttachedTo = %+v, want card %s", host, bear)
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power %d, want 4", got)
	}
	if got := effectiveToughness(t, g, bear); got != 2 {
		t.Errorf("Bonesplitter is +2/+0, toughness %d, want 2", got)
	}
}

// CR 702.6d — a second activation MOVES the Equipment, and the bonus
// moves with it in the same beat.
func TestReEquipMovesTheBonus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	first := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	second := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ox", TypeLine: testCreatureTypeLine,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})

	equipTo(t, g, me.ID, sword, first)
	equipTo(t, g, me.ID, sword, second)

	if got := effectivePower(t, g, first); got != 2 {
		t.Errorf("first creature kept the bonus: power %d, want 2", got)
	}
	if got := effectivePower(t, g, second); got != 3 {
		t.Errorf("second creature power %d, want 3", got)
	}
}

// CR 702.6b — equip is sorcery-speed.
func TestEquipIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})
	// newCatalogGame lands on Upkeep, which is not a sorcery window.
	err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err != game.ErrSorcerySpeedRequired {
		t.Errorf("equip on upkeep: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// CR 702.6b — "target creature you control".
func TestEquipRejectsACreatureYouDoNotControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("equip should reject a creature you do not control")
	}
}

// --- Skullclamp ---------------------------------------------------

// The whole engine: clamp a 1/1, it becomes a 2/0, the lethal-damage
// SBA kills it, the dies trigger draws two.
func TestSkullclampKillsAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	clamp := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Skullclamp", TypeLine: equipTypeLine,
		OracleID: skullclampOracle, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()

	equipTo(t, g, me.ID, clamp, token)

	if g.Battlefield.Contains(token) {
		t.Fatal("a clamped 1/1 is a 2/0 and dies to the toughness SBA")
	}
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d -> %d, want +2", before, got)
	}
	// CR 704.5n: the Clamp itself stays, unattached, ready to move.
	if !g.Battlefield.Contains(clamp) {
		t.Fatal("Skullclamp should stay on the battlefield")
	}
	if host := attachmentHostOf(t, g, clamp); host.Kind != "" {
		t.Errorf("Skullclamp still attached to the dead creature: %+v", host)
	}
}

// --- Lightning Greaves / Swiftfoot Boots --------------------------

func TestLightningGreavesGrantsHasteAndShroud(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	greaves := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lightning Greaves", TypeLine: equipTypeLine,
		OracleID: greavesOracle, Owner: me.ID, Controller: me.ID,
	})
	equipTo(t, g, me.ID, greaves, bear)

	abilities := effectiveAbilities(t, g, bear)
	for _, want := range []string{"haste", "shroud"} {
		if !containsString(abilities, want) {
			t.Errorf("abilities %v missing %q", abilities, want)
		}
	}
	// Shroud is the half that must be real: NOBODY may target it,
	// its own controller included (CR 702.18a).
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != bear {
				continue
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(me.ID)) {
				t.Error("shroud should stop even the controller targeting it")
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(opp.ID)) {
				t.Error("shroud should stop an opponent targeting it")
			}
		}
	})
}

func TestSwiftfootBootsGrantHexproofAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	boots := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Swiftfoot Boots", TypeLine: equipTypeLine,
		OracleID: bootsOracle, Owner: me.ID, Controller: me.ID,
	})
	equipTo(t, g, me.ID, boots, bear)

	abilities := effectiveAbilities(t, g, bear)
	for _, want := range []string{"hexproof", "haste"} {
		if !containsString(abilities, want) {
			t.Errorf("abilities %v missing %q", abilities, want)
		}
	}
	// Hexproof is the asymmetric one — that asymmetry is the whole
	// reason Boots costs a mana more to equip than Greaves.
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != bear {
				continue
			}
			if !game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(me.ID)) {
				t.Error("hexproof must not stop the controller targeting it")
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(opp.ID)) {
				t.Error("hexproof should stop an opponent targeting it")
			}
		}
	})
}

// --- Rancor: the first Aura ---------------------------------------

func TestRancorAttachesOnResolutionAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	rancor := castCatalogSpell(t, g, "Rancor", auraTypeLine, rancorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(rancor) {
		t.Fatal("Rancor should resolve to the battlefield")
	}
	if host := attachmentHostOf(t, g, rancor); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("Rancor AttachedTo = %+v, want card %s", host, bear)
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("enchanted power %d, want 4", got)
	}
	if got := effectiveToughness(t, g, bear); got != 2 {
		t.Errorf("Rancor is +2/+0, toughness %d, want 2", got)
	}
	if ab := effectiveAbilities(t, g, bear); !containsString(ab, "trample") {
		t.Errorf("abilities %v missing trample", ab)
	}
}

// CR 704.5m plus Rancor's own recursion: the host leaves, the Aura
// goes to the graveyard as a state-based action, and the LTB trigger
// hands it straight back.
func TestRancorReturnsToHandWhenItFallsOff(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	rancor := castCatalogSpell(t, g, "Rancor", auraTypeLine, rancorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	// MoveCardByID rather than DestroyPermanentForEffect: the public
	// mutator runs the state checks on the way out, and the CR 704.5m
	// unattach IS a state check. Nothing else is on the stack to
	// carry a priority pass here.
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		bear,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rancor) {
		t.Fatal("Rancor should have fallen off the battlefield")
	}
	if !me.Hand.Contains(rancor) {
		t.Error("Rancor's own recursion should have put it back in hand")
	}
}

// --- The Swords: attachment-conditioned combat triggers -----------

// dealCombatDamageToPlayer fakes the combat-damage step's event for
// one creature, which is all the Swords' triggers watch. Driving a
// full combat would test the combat engine, not the attachment.
func dealCombatDamageToPlayer(g *game.Game, creature, victim uuid.UUID, amount int) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:   game.EventDealDamage,
			Source: creature,
			Target: victim,
			Amount: amount,
			Combat: true,
		})
	})
}

func TestSwordOfFeastAndFamineDiscardsAndUntapsLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sword of Feast and Famine", TypeLine: equipTypeLine,
		OracleID: swordFeastOracle, Owner: me.ID, Controller: me.ID,
	})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	equipTo(t, g, me.ID, sword, bear)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Fatalf("equipped power %d, want 4", got)
	}
	oppHand := opp.Hand.Size()

	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)

	// #651: the damaged player CHOOSES their discard (CR 701.8a), so
	// the trigger leaves a prompt addressed to them. The untap is not
	// behind it — the lands come back as the trigger resolves.
	if got := discardOwed(g, opp.ID); got != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", got)
	}
	discardFromHand(t, g, opp.ID)
	if got := opp.Hand.Size(); got != oppHand-1 {
		t.Errorf("opponent hand %d -> %d, want -1", oppHand, got)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == land && c.Tapped {
				t.Error("the land should have untapped")
			}
		}
	})
}

func TestSwordOfFireAndIceDamagesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sword of Fire and Ice", TypeLine: equipTypeLine,
		OracleID: swordFireOracle, Owner: me.ID, Controller: me.ID,
	})
	equipTo(t, g, me.ID, sword, bear)
	myHand, oppLife := me.Hand.Size(), opp.Life

	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	// The trigger targets, so it waits for a pick before it reaches
	// the stack. Point it at the player who just took the hit.
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != oppLife-2 {
		t.Errorf("life %d -> %d, want -2", oppLife, opp.Life)
	}
	if got := me.Hand.Size(); got != myHand+1 {
		t.Errorf("hand %d -> %d, want +1", myHand, got)
	}
}

// --- Curse of Opulence: the TargetPlayer branch -------------------

func TestCurseOfOpulenceAttachesToAPlayerAndPays(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	curse := castCatalogSpell(t, g, "Curse of Opulence", curseAuraTypeLine, curseOpulenceOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)

	host := attachmentHostOf(t, g, curse)
	if host.Kind != game.TargetPlayer || host.ID != victim.ID {
		t.Fatalf("Curse AttachedTo = %+v, want player %s", host, victim.ID)
	}

	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.DeclareAttacker(attacker, victim.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	golds := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Gold" && c.Controller == me.ID {
				golds++
			}
		}
	})
	if golds != 1 {
		t.Errorf("Gold tokens for the Curse's controller: %d, want 1", golds)
	}
}
