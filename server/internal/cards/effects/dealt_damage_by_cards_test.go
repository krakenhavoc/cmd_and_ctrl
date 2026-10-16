package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dealt_damage_by_cards_test.go — #1946: the cards built on the turn's
// record of which objects each source dealt damage to. The record itself
// is pinned in game/dealt_damage_by_test.go.

const (
	ddBoneShamanOracle  = "690ba846-964e-44ba-8637-9a1ce62f5931"
	ddRuneswordOracle   = "964c2003-cef0-4ac7-9f66-4e20893c8e50"
	ddKumanoOracle      = "6d776c7b-4ca1-48cd-88c7-dfbbe5a56a0b"
	ddFrostwielderOracl = "f895ec2e-7481-4280-a5c0-38b58c440faf"
	ddPupilsOracle      = "f7f8a186-313e-4a29-bf4e-b6a200dfd1ed"
	ddOracleOracle      = "2c3ed6b9-1a2b-42ad-baa9-ee2a0ad505a7"
	ddBlessingOracle    = "85d81888-8df7-4826-891e-a14fb8bf6549"
)

// ddCard puts a catalog creature on the battlefield for `owner`, ready to
// tap.
func ddCard(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// ddCombatDamage records `n` combat damage from `src` to `victim` the
// way the combat damage step emits it.
func ddCombatDamage(g *game.Game, actor, src, victim uuid.UUID, n int) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: actor, Source: src, Target: victim, Amount: n, Combat: true})
	})
}

func ddFlicker(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var back uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatal(err)
		}
		var err error
		if back, err = g.ReturnFromExileToBattlefieldForEffect(id, uuid.Nil, false); err != nil {
			t.Fatal(err)
		}
	})
	return back
}

func ddCantBeRegenerated(g *game.Game, id uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.PermanentCantBeRegeneratedForEffect(id) })
	return out
}

// --- Kumano, Master Yamabushi ----------------------------------------

// The ping that kills exiles; a creature it pinged that dies later is
// exiled too; one it never damaged dies normally.
func TestKumanoExilesWhatItDamagedThisTurn(t *testing.T) {
	g, me, opp := flipMain(t)
	kumano := ddCard(g, me.ID, "Kumano, Master Yamabushi", "Legendary Creature — Human Shaman", ddKumanoOracle, 4, 4)
	small := p1Creature(g, opp.ID, 1, 1)
	big := p1Creature(g, opp.ID, 3, 3)
	untouched := p1Creature(g, opp.ID, 3, 3)
	for _, target := range []uuid.UUID{small, big} {
		if err := g.ActivateCatalogAbility(me.ID, kumano, 0, game.ActivateAbilityParams{Targets: pr6Card(target)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	g.RunStateChecksForTest()
	p1WantZone(t, g, small, game.ZoneExile, "the 1/1 Kumano's ping killed")
	p1WantZone(t, g, big, game.ZoneBattlefield, "the 3/3 Kumano pinged")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 3/3 destroyed later this turn")
	p1Destroy(t, g, untouched)
	p1WantZone(t, g, untouched, game.ZoneGraveyard, "a creature Kumano never damaged")
}

// Combat damage counts; damage prevented to 0 does not; a creature that
// was damaged and then flickered is a new object.
func TestKumanoReadsCombatDamagePreventionAndObjects(t *testing.T) {
	g, me, opp := flipMain(t)
	kumano := ddCard(g, me.ID, "Kumano, Master Yamabushi", "Legendary Creature — Human Shaman", ddKumanoOracle, 4, 4)
	blocker := p1Creature(g, opp.ID, 2, 6)
	shielded := p1Creature(g, opp.ID, 2, 6)
	blinked := p1Creature(g, opp.ID, 2, 6)

	ddCombatDamage(g, me.ID, kumano, blocker, 4)
	p1ShieldCreature(g, shielded)
	damage(t, g, kumano, shielded, 4)
	damage(t, g, kumano, blinked, 1)
	blinked = ddFlicker(t, g, blinked)

	p1Destroy(t, g, blocker)
	p1WantZone(t, g, blocker, game.ZoneExile, "the blocker Kumano dealt combat damage to")
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature whose damage was prevented to 0")
	p1Destroy(t, g, blinked)
	p1WantZone(t, g, blinked, game.ZoneGraveyard, "a creature flickered after Kumano damaged it")
}

// The replacement is Kumano's static: once Kumano has left, nothing is
// exiled, and next turn last turn's damage counts for nothing.
func TestKumanoNeedsToBeThereAndForgetsAtTheTurn(t *testing.T) {
	g, me, opp := flipMain(t)
	kumano := ddCard(g, me.ID, "Kumano, Master Yamabushi", "Legendary Creature — Human Shaman", ddKumanoOracle, 4, 4)
	gone := p1Creature(g, opp.ID, 2, 6)
	later := p1Creature(g, opp.ID, 2, 6)
	damage(t, g, kumano, later, 1)
	damage(t, g, kumano, gone, 1)

	// Next turn the record is empty.
	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	p1Destroy(t, g, later)
	p1WantZone(t, g, later, game.ZoneGraveyard, "a creature Kumano damaged last turn")

	damage(t, g, kumano, gone, 1)
	p1Destroy(t, g, kumano)
	p1Destroy(t, g, gone)
	p1WantZone(t, g, gone, game.ZoneGraveyard, "a creature that dies after Kumano has left")
}

// --- Frostwielder, Kumano's Pupils, Incendiary Oracle -----------------

func TestFrostwielderExilesTheCreatureItPings(t *testing.T) {
	g, me, opp := flipMain(t)
	fw := ddCard(g, me.ID, "Frostwielder", "Creature — Human Shaman", ddFrostwielderOracl, 1, 2)
	bear := p1Creature(g, opp.ID, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, fw, 0, game.ActivateAbilityParams{Targets: pr6Card(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	p1WantZone(t, g, bear, game.ZoneExile, "the 1/1 Frostwielder pinged")
}

func TestKumanosPupilsExilesWhatItFought(t *testing.T) {
	g, me, opp := flipMain(t)
	pupils := ddCard(g, me.ID, "Kumano's Pupils", "Creature — Human Shaman", ddPupilsOracle, 3, 3)
	blocker := p1Creature(g, opp.ID, 2, 5)
	ddCombatDamage(g, me.ID, pupils, blocker, 3)
	p1Destroy(t, g, blocker)
	p1WantZone(t, g, blocker, game.ZoneExile, "the creature Kumano's Pupils dealt combat damage to")
}

func TestIncendiaryOraclePumpsAndExiles(t *testing.T) {
	g, me, opp := flipMain(t)
	oracle := ddCard(g, me.ID, "Incendiary Oracle", "Creature — Human Shaman", ddOracleOracle, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := apaLive(g, oracle).Effective().Power; p != 3 {
		t.Errorf("power after {1}{R} = %d, want 3", p)
	}
	blocker := p1Creature(g, opp.ID, 2, 5)
	ddCombatDamage(g, me.ID, oracle, blocker, 3)
	p1Destroy(t, g, blocker)
	p1WantZone(t, g, blocker, game.ZoneExile, "the creature Incendiary Oracle damaged")
}

// --- Kumano's Blessing -----------------------------------------------

// "Enchanted creature" is whatever the Aura is on as the other creature
// would die: damage dealt before the Aura arrived counts, and damage
// dealt by a creature the Aura has left does not.
func TestKumanosBlessingFollowsTheEnchantedCreature(t *testing.T) {
	g, me, opp := flipMain(t)
	host := p1Creature(g, me.ID, 3, 3)
	other := p1Creature(g, me.ID, 3, 3)
	first := p1Creature(g, opp.ID, 2, 6)
	second := p1Creature(g, opp.ID, 2, 6)
	damage(t, g, host, first, 1)
	damage(t, g, host, second, 1)
	aura := ddCard(g, me.ID, "Kumano's Blessing", "Enchantment — Aura", ddBlessingOracle, 0, 0)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	p1Destroy(t, g, first)
	p1WantZone(t, g, first, game.ZoneExile, "a creature the enchanted creature damaged before the Aura arrived")

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: other}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	p1Destroy(t, g, second)
	p1WantZone(t, g, second, game.ZoneGraveyard, "a creature damaged by the creature the Aura has left")
}

// --- Bone Shaman -----------------------------------------------------

// The gained ability covers the creatures Bone Shaman damaged earlier
// this turn, refuses the shield without spending it, and leaves every
// other creature alone.
func TestBoneShamanStopsRegenerationOfWhatItDamaged(t *testing.T) {
	g, me, opp := flipMain(t)
	shaman := ddCard(g, me.ID, "Bone Shaman", "Creature — Giant Shaman", ddBoneShamanOracle, 3, 3)
	troll := p1Creature(g, opp.ID, 2, 6)
	other := p1Creature(g, opp.ID, 2, 6)
	ddCombatDamage(g, me.ID, shaman, troll, 3)
	p1Regenerate(t, g, troll)
	p1Regenerate(t, g, other)
	if ddCantBeRegenerated(g, troll) {
		t.Fatal("before {B} the creature can be regenerated")
	}
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !ddCantBeRegenerated(g, troll) {
		t.Fatal("after {B} the creature Bone Shaman damaged can still be regenerated")
	}
	if ddCantBeRegenerated(g, other) {
		t.Error("a creature Bone Shaman never damaged can't be regenerated")
	}
	p1Destroy(t, g, troll)
	p1WantZone(t, g, troll, game.ZoneGraveyard, "the creature Bone Shaman damaged")
	p1Destroy(t, g, other)
	p1WantZone(t, g, other, game.ZoneBattlefield, "the creature it never damaged, which regenerates")
}

// It is an ability Bone Shaman has: losing all abilities, or leaving the
// battlefield, ends it.
func TestBoneShamanAbilityEndsWithTheAbility(t *testing.T) {
	g, me, opp := flipMain(t)
	shaman := ddCard(g, me.ID, "Bone Shaman", "Creature — Giant Shaman", ddBoneShamanOracle, 3, 3)
	troll := p1Creature(g, opp.ID, 2, 6)
	ddCombatDamage(g, me.ID, shaman, troll, 3)
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !ddCantBeRegenerated(g, troll) {
		t.Fatal("the gained ability is not in force")
	}
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(shaman),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "Humble")
	})
	if ddCantBeRegenerated(g, troll) {
		t.Error("a Bone Shaman that lost all abilities still stops regeneration")
	}

	g2, me2, opp2 := flipMain(t)
	shaman2 := ddCard(g2, me2.ID, "Bone Shaman", "Creature — Giant Shaman", ddBoneShamanOracle, 3, 3)
	troll2 := p1Creature(g2, opp2.ID, 2, 6)
	ddCombatDamage(g2, me2.ID, shaman2, troll2, 3)
	if err := g2.ActivateCatalogAbility(me2.ID, shaman2, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g2)
	p1Destroy(t, g2, shaman2)
	if ddCantBeRegenerated(g2, troll2) {
		t.Error("the ability outlived Bone Shaman")
	}
}

// --- Runesword -------------------------------------------------------

func runeswordBoard(t *testing.T) (g *game.Game, me *game.Player, sword, attacker, opp uuid.UUID) {
	t.Helper()
	g, mine, theirs := flipMain(t)
	sword = ddCard(g, mine.ID, "Runesword", "Artifact", ddRuneswordOracle, 0, 0)
	attacker = p1Creature(g, mine.ID, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = theirs.ID
			}
		}
	})
	return g, mine, sword, attacker, theirs.ID
}

func TestRuneswordPumpsAndRidesTheDamage(t *testing.T) {
	g, me, sword, attacker, opp := runeswordBoard(t)
	before := p1Creature(g, opp, 2, 6)
	ddCombatDamage(g, me.ID, attacker, before, 1)

	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Targets: pr6Card(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := apaLive(g, attacker).Effective().Power; p != 4 {
		t.Errorf("attacker power = %d, want 4 (+2/+0)", p)
	}

	troll := p1Creature(g, opp, 2, 6)
	bystander := p1Creature(g, opp, 2, 6)
	ddCombatDamage(g, me.ID, attacker, troll, 4)
	p1Regenerate(t, g, troll)
	if !ddCantBeRegenerated(g, troll) {
		t.Error("a creature it damages after Runesword resolved can be regenerated")
	}
	if ddCantBeRegenerated(g, before) {
		t.Error("\"if the creature deals damage\" reaches damage dealt before Runesword resolved")
	}
	if ddCantBeRegenerated(g, bystander) {
		t.Error("a creature it never damaged can't be regenerated")
	}
	p1Destroy(t, g, troll)
	p1WantZone(t, g, troll, game.ZoneExile, "the regenerating creature it damaged")
	p1Destroy(t, g, before)
	p1WantZone(t, g, before, game.ZoneExile, "a creature it damaged before Runesword resolved")
	p1Destroy(t, g, bystander)
	p1WantZone(t, g, bystander, game.ZoneGraveyard, "a creature it never damaged")
}

// "When that creature leaves the battlefield this turn, sacrifice this
// artifact" — and the riders outlast the creature.
func TestRuneswordIsSacrificedWhenTheCreatureLeaves(t *testing.T) {
	g, me, sword, attacker, opp := runeswordBoard(t)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Targets: pr6Card(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	victim := p1Creature(g, opp, 2, 6)
	ddCombatDamage(g, me.ID, attacker, victim, 4)
	p1WantZone(t, g, sword, game.ZoneBattlefield, "Runesword while its creature is in play")

	p1Destroy(t, g, attacker)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, sword, game.ZoneGraveyard, "Runesword after its creature left")
	p1Destroy(t, g, victim)
	p1WantZone(t, g, victim, game.ZoneExile, "a creature the departed attacker damaged")
}
