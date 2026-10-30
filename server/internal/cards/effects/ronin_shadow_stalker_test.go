package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ronin_shadow_stalker_test.go — #4600, Ronin, Shadow Stalker (CR 106.6,
// CR 702.6, CR 119.4): "Pay 2 life: Add two mana of any one color. Spend
// this mana only to cast Equipment spells or activate equip abilities.
// Activate only once each turn." and the sorcery-speed -4/-4. The engine
// half (the tag's matcher and the pool solver) is
// game/equip_mana_test.go; the enumerator's is legal/equip_mana_test.go.

const (
	roninOracle = "8ceb1acf-5ff3-4f1a-af04-9f07ad454526"
	roninType   = "Legendary Creature — Human Rogue Hero"
)

func roninTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, _ := spendTable(t)
	return g, me, pushCatalogPermanent(g, me.ID, "Ronin, Shadow Stalker", roninType, roninOracle, false)
}

// payLifeForRoninMana fires the mana ability for {R}{R} and checks what it
// left in the pool.
func payLifeForRoninMana(t *testing.T, g *game.Game, me *game.Player, ronin uuid.UUID) {
	t.Helper()
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, ronin, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("Ronin's mana ability: %v", err)
	}
	if me.Life != before-2 {
		t.Errorf("life = %d, want %d (paid 2)", me.Life, before-2)
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %v, want two mana", me.ManaPool)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "R" || len(tok.Restrictions) != 1 {
			t.Fatalf("token %+v, want red mana carrying the one restriction", tok)
		}
	}
}

func TestRoninIsFull(t *testing.T) {
	spec, ok := Lookup(roninOracle)
	if !ok {
		t.Fatal("Ronin, Shadow Stalker is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v, want Full and none", spec.Completeness, spec.Caveats)
	}
}

// The mana pays for an Equipment spell, both cost symbols of it.
func TestRoninManaPaysAnEquipmentSpell(t *testing.T) {
	g, me, ronin := roninTable(t)
	payLifeForRoninMana(t, g, me, ronin)
	blade := handSpell(me, "Test Blade", "Artifact — Equipment", "{1}{R}")
	if err := g.CastSpell(me.ID, blade, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an Equipment spell with Ronin's mana: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want it spent", me.ManaPool)
	}
}

// The mana pays for an equip ability: Shadowspear's equip {2}.
func TestRoninManaPaysAnEquipAbility(t *testing.T) {
	g, me, ronin := roninTable(t)
	spear := pushCatalogPermanent(g, me.ID, "Shadowspear", equipTypeLine, shadowspearOracle, false)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	payLifeForRoninMana(t, g, me, ronin)
	if err := g.ActivateCatalogAbility(me.ID, spear, 1, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("equip with Ronin's mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, spear); host.ID != bear {
		t.Errorf("Shadowspear is attached to %v, want the bear", host)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want it spent", me.ManaPool)
	}
}

// It pays nothing else: not a creature spell, not a non-Equipment
// artifact spell, and not the other activated ability of an Equipment
// (Shadowspear's {1}), which is an activation whose source is an
// Equipment but is not an equip ability.
func TestRoninManaPaysNothingButEquipmentSpellsAndEquipAbilities(t *testing.T) {
	g, me, ronin := roninTable(t)
	spear := pushCatalogPermanent(g, me.ID, "Shadowspear", equipTypeLine, shadowspearOracle, false)
	payLifeForRoninMana(t, g, me, ronin)

	bear := handSpell(me, "Test Bear", "Creature — Bear", "{1}{R}")
	refusedForMana(t, g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}), "a creature spell")
	rock := handSpell(me, "Test Rock", "Artifact", "{1}{R}")
	refusedForMana(t, g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true}), "a non-Equipment artifact spell")
	bolt := handSpell(me, "Test Bolt", "Instant", "{1}{R}")
	refusedForMana(t, g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true}), "an instant")
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, spear, 0, game.ActivateAbilityParams{Strict: true}),
		"Shadowspear's non-equip ability")
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool = %v, want both mana still floating after the refusals", me.ManaPool)
	}
}

// The auto-tapper's top-up reads the floating restricted mana through the
// same spend context: it finishes an equip with Ronin's mana and leaves
// that mana alone for a creature spell and for the Equipment's other
// ability, which it then cannot pay.
func TestRoninManaIsToppedUpByAutoTapOnlyForEquipmentThings(t *testing.T) {
	g, me, ronin := roninTable(t)
	spear := pushCatalogPermanent(g, me.ID, "Shadowspear", equipTypeLine, shadowspearOracle, false)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	payLifeForRoninMana(t, g, me, ronin)

	creature := handSpell(me, "Test Bear", "Creature — Bear", "{1}{R}")
	refusedForMana(t, g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true, AutoTap: true}), "a creature spell")
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, spear, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}),
		"Shadowspear's non-equip ability")
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %v, want both mana still floating after the refusals", me.ManaPool)
	}
	if err := g.ActivateCatalogAbility(me.ID, spear, 1, game.ActivateAbilityParams{
		Strict: true, AutoTap: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("equip with auto-tap over Ronin's mana: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want it spent", me.ManaPool)
	}
}

// "Activate only once each turn": the second activation is refused and
// costs no life.
func TestRoninManaIsOncePerTurn(t *testing.T) {
	g, me, ronin := roninTable(t)
	payLifeForRoninMana(t, g, me, ronin)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, ronin, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("a second activation in the same turn was accepted")
	}
	if me.Life != life || me.ManaPool.Total() != 2 {
		t.Errorf("a refused activation changed life to %d or the pool to %v", me.Life, me.ManaPool)
	}
}

// The -4/-4 needs an Equipment attached to Ronin to sacrifice.
func TestRoninMinusFourNeedsAnAttachedEquipment(t *testing.T) {
	g, me, ronin := roninTable(t)
	opp := g.Seats[1]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	loose := pushCatalogPermanent(g, me.ID, "Shadowspear", equipTypeLine, shadowspearOracle, false)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}

	if err := g.ActivateCatalogAbility(me.ID, ronin, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated with no Equipment to sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, ronin, 0, game.ActivateAbilityParams{
		Targets: target, SacrificeIDs: []uuid.UUID{loose},
	}); err == nil {
		t.Fatal("sacrificed an Equipment that is not attached to Ronin")
	}
	if !g.Battlefield.Contains(loose) || !g.Battlefield.Contains(victim) {
		t.Error("a refused activation cost something")
	}
}

// Equipped, Ronin sacrifices the Equipment and the 4/4 dies; the cost
// taps him.
func TestRoninMinusFourKillsAFourToughnessCreature(t *testing.T) {
	g, me, ronin := roninTable(t)
	opp := g.Seats[1]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	spear := pushCatalogPermanent(g, me.ID, "Shadowspear", equipTypeLine, shadowspearOracle, false)
	// The equip itself is paid with Ronin's mana: the two halves together.
	payLifeForRoninMana(t, g, me, ronin)
	if err := g.ActivateCatalogAbility(me.ID, spear, 1, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ronin}},
	}); err != nil {
		t.Fatalf("equip Ronin: %v", err)
	}
	passPriorityAroundTable(t, g)

	if err := g.ActivateCatalogAbility(me.ID, ronin, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		SacrificeIDs: []uuid.UUID{spear},
	}); err != nil {
		t.Fatalf("Ronin's -4/-4: %v", err)
	}
	if g.Battlefield.Contains(spear) {
		t.Error("the Equipment was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 4/4 survived -4/-4")
	}
	if !g.Battlefield.Contains(ronin) {
		t.Error("Ronin left the battlefield")
	}
}

// "Activate only as a sorcery."
func TestRoninMinusFourIsSorcerySpeed(t *testing.T) {
	g, me, ronin := roninTable(t)
	opp := g.Seats[1]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	spear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shadowspear", TypeLine: equipTypeLine,
		OracleID: shadowspearOracle, Owner: me.ID, Controller: me.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: ronin},
	})
	advanceTo(t, g, game.StepBeginCombat)
	err := g.ActivateCatalogAbility(me.ID, ronin, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		SacrificeIDs: []uuid.UUID{spear},
	})
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("at the beginning of combat: %v, want ErrSorcerySpeedRequired", err)
	}
}
