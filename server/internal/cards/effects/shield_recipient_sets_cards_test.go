package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shield_recipient_sets_cards_test.go — #2045 (ADR 0108 §7, amendment of
// 2026-10-06): every card the recipient set shipped, against the damage
// it must stop, the damage it must not, and a recipient that changes
// after the shield resolved (CR 611.2c: a prevention effect's set is
// read as the damage would be dealt).

const (
	rsForfend          = "59f8f2c4-d30d-411c-9fac-a5a3219bb124"
	rsBlindingFog      = "50555471-a03c-4735-a6ce-6ae96e27d7c0"
	rsDivineLight      = "924bf3bb-9f06-4a2c-a17f-3fe11551bcf2"
	rsSivvisRuse       = "c5edfaae-3ecd-41ad-970f-8cecb09bf1e4"
	rsChameleonBlur    = "6ac8ac31-02ce-48e0-98b9-1c5858e1f593"
	rsCommencement     = "56ea8f6b-a765-44e7-bf0e-4d4a2405684b"
	rsDefendTheHearth  = "1dd64972-5713-4c25-b7a5-a35a56886ef4"
	rsSurgeOfSalvation = "b9a62c6b-5848-44b1-932b-8d5de2560976"
	rsPackLeader       = "3701ed34-a97c-4d7b-a15a-4faec02ef24b"
	rsShieldmage       = "ee988017-fc7e-4d8a-8f5c-0a7e57a8d050"
	rsLoyalUnicorn     = "6cf8191c-acca-4145-86c2-53a138b3fe4a"
)

// rsHit deals 1 from `source` to the player `to` and returns the life
// lost.
func rsHit(t *testing.T, g *game.Game, source uuid.UUID, to *game.Player) int {
	t.Helper()
	before := to.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, to.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	return before - to.Life
}

// Forfend and Blinding Fog: every creature, whoever controls it, a late
// one included; no player. Blinding Fog's hexproof goes to the creatures
// you control as it resolves.
func TestRecipientSetCardsEveryCreature(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{{"Forfend", rsForfend}, {"Blinding Fog", rsBlindingFog}} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rod := sfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
			mine := sfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
			castCatalogSpell(t, g, c.name, "Instant", c.oracle, nil)
			passPriorityAroundTable(t, g)
			late := sfPush(g, opp.ID, "Late", "Creature — Test", 2, []string{"R"})
			pr6Damage(t, g, rod, mine, 2)
			pr6Damage(t, g, rod, late, 2)
			if pr6Marked(g, mine) != 0 || pr6Marked(g, late) != 0 {
				t.Fatalf("marked mine %d, late %d; want 0 and 0", pr6Marked(g, mine), pr6Marked(g, late))
			}
			if lost := rsHit(t, g, rod, me); lost != 1 {
				t.Fatalf("lost %d, want 1: a player is not protected", lost)
			}
			if c.name == "Blinding Fog" {
				if !hasEffectiveKeyword(t, g, mine, "hexproof") {
					t.Error("my creature should have hexproof")
				}
				if hasEffectiveKeyword(t, g, late, "hexproof") {
					t.Error("a creature that entered later, and an opponent's, gains no hexproof (CR 611.2c)")
				}
			}
		})
	}
}

// Divine Light and Sivvi's Ruse: creatures you control, not you, not an
// opponent's creature; control is read as the damage would be dealt.
func TestRecipientSetCardsCreaturesYouControl(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Divine Light", "Sorcery", rsDivineLight},
		{"Sivvi's Ruse", "Instant", rsSivvisRuse},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rod := sfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
			mine := sfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
			theirs := sfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"R"})
			castCatalogSpell(t, g, c.name, c.typeLine, c.oracle, nil)
			passPriorityAroundTable(t, g)
			pr6Damage(t, g, rod, mine, 1)
			pr6Damage(t, g, rod, theirs, 1)
			if pr6Marked(g, mine) != 0 || pr6Marked(g, theirs) != 1 {
				t.Fatalf("marked mine %d (want 0), theirs %d (want 1)", pr6Marked(g, mine), pr6Marked(g, theirs))
			}
			if lost := rsHit(t, g, rod, me); lost != 1 {
				t.Fatalf("lost %d, want 1: you are not a creature you control", lost)
			}
			findBattlefieldCardForTest(g, theirs).Controller = me.ID
			g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
			pr6Damage(t, g, rod, theirs, 1)
			if pr6Marked(g, theirs) != 1 {
				t.Fatalf("marked %d, want still 1: a creature you gained control of is protected", pr6Marked(g, theirs))
			}
		})
	}
}

// Sivvi's Ruse is free only when an opponent controls a Mountain and you
// control a Plains.
func TestRecipientSetSivvisRuseFreeCastCondition(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cond := AnOpponentControlsAAndYouControlA("Mountain", "Plains")
	if cond(g, me.ID) {
		t.Fatal("free with neither land")
	}
	sfPush(g, me.ID, "My Mountain", "Basic Land — Mountain", 0, nil)
	sfPush(g, me.ID, "Plains", "Basic Land — Plains", 0, nil)
	if cond(g, me.ID) {
		t.Fatal("your own Mountain doesn't count")
	}
	sfPush(g, opp.ID, "Mountain", "Basic Land — Mountain", 0, nil)
	if !cond(g, me.ID) {
		t.Fatal("an opponent's Mountain and your Plains make it free")
	}
	castWithAltCost(t, g, "Sivvi's Ruse", "Instant", rsSivvisRuse, "free")
	passPriorityAroundTable(t, g)
	if pr7bSourceShields(g) != 1 {
		t.Fatalf("shields %d, want 1", pr7bSourceShields(g))
	}
}

// Chameleon Blur: damage creatures would deal to players. A noncreature
// source's damage to a player and a creature's damage to a creature are
// dealt.
func TestRecipientSetChameleonBlurStopsCreaturesHittingPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rod := sfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
	bear := sfPush(g, opp.ID, "Bear", "Creature — Bear", 2, []string{"G"})
	mine := sfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
	castCatalogSpell(t, g, "Chameleon Blur", "Instant", rsChameleonBlur, nil)
	passPriorityAroundTable(t, g)
	if lost := rsHit(t, g, bear, me) + rsHit(t, g, bear, opp); lost != 0 {
		t.Fatalf("lost %d, want 0: a creature's damage to any player is prevented", lost)
	}
	if lost := rsHit(t, g, rod, me); lost != 1 {
		t.Fatalf("lost %d, want 1: a noncreature source's damage is dealt", lost)
	}
	pr6Damage(t, g, bear, mine, 1)
	if pr6Marked(g, mine) != 1 {
		t.Fatalf("marked %d, want 1: damage to a creature is dealt", pr6Marked(g, mine))
	}
}

// Commencement of Festivities and Defend the Hearth: combat damage to
// players only. A blocked attacker and its blocker deal theirs; the
// unblocked attacker's is prevented, and noncombat damage to a player
// is dealt.
func TestRecipientSetCardsCombatDamageToPlayers(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Commencement of Festivities", rsCommencement},
		{"Defend the Hearth", rsDefendTheHearth},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			free := sfPush(g, me.ID, "Free", "Creature — Test", 3, []string{"R"})
			held := sfPush(g, me.ID, "Held", "Creature — Test", 2, []string{"R"})
			wall := sfPush(g, opp.ID, "Wall", "Creature — Wall", 1, []string{"W"})
			pr7bAttack(t, g, opp.ID, free, held)
			pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: held})
			sfCast(t, g, me, sfInstant(c.name, c.oracle), game.CastSpellParams{})
			if lost := sfCombat(t, g)[opp.ID]; lost != 0 {
				t.Fatalf("defender lost %d, want 0", lost)
			}
			if pr6Marked(g, wall) != 2 || pr6Marked(g, held) != 1 {
				t.Fatalf("wall %d (want 2), held %d (want 1): combat damage to creatures is dealt", pr6Marked(g, wall), pr6Marked(g, held))
			}
			if lost := rsHit(t, g, free, opp); lost != 1 {
				t.Fatalf("lost %d, want 1: noncombat damage to a player is dealt", lost)
			}
		})
	}
}

// Surge of Salvation: black and/or red sources, to creatures you control.
// A green source's damage and damage to you are dealt; you and your
// permanents gain hexproof.
func TestRecipientSetSurgeOfSalvation(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := sfPush(g, opp.ID, "Red", "Creature — Test", 2, []string{"R"})
	gold := sfPush(g, opp.ID, "Gold", "Creature — Test", 2, []string{"B", "G"})
	green := sfPush(g, opp.ID, "Green", "Creature — Test", 2, []string{"G"})
	mine := sfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
	theirs := sfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"W"})
	castCatalogSpell(t, g, "Surge of Salvation", "Instant", rsSurgeOfSalvation, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "hexproof") || hasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("my permanents gain hexproof, an opponent's do not")
	}
	pr6Damage(t, g, red, mine, 1)
	pr6Damage(t, g, gold, mine, 1)
	if pr6Marked(g, mine) != 0 {
		t.Fatalf("marked %d, want 0: red and black-green sources are stopped", pr6Marked(g, mine))
	}
	pr6Damage(t, g, green, mine, 1)
	pr6Damage(t, g, red, theirs, 1)
	if pr6Marked(g, mine) != 1 || pr6Marked(g, theirs) != 1 {
		t.Fatalf("mine %d (want 1, the green source's), theirs %d (want 1)", pr6Marked(g, mine), pr6Marked(g, theirs))
	}
	if lost := rsHit(t, g, red, me); lost != 1 {
		t.Fatalf("lost %d, want 1: damage to you is not prevented", lost)
	}
}

// Pack Leader: other Dogs get +1/+1; when it attacks, combat damage to
// Dogs you control is prevented — not to you, not to your other
// creatures, and not noncombat damage.
func TestRecipientSetPackLeaderShieldsDogsInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leader := apaPush(g, me.ID, me.ID, game.Card{Name: "Pack Leader", TypeLine: "Creature — Dog", OracleID: rsPackLeader,
		Power: 2, Toughness: 2, Colors: []string{"W"}})
	dog := sfPush(g, me.ID, "Dog", "Creature — Dog", 1, []string{"W"})
	cat := sfPush(g, me.ID, "Cat", "Creature — Cat", 1, []string{"W"})
	if p := findBattlefieldCardForTest(g, dog).CurrentPower(); p != 2 {
		t.Fatalf("Dog power %d, want 2 with the anthem", p)
	}
	if p := findBattlefieldCardForTest(g, leader).CurrentPower(); p != 2 {
		t.Fatalf("Pack Leader power %d, want 2: the anthem is for other Dogs", p)
	}
	b1 := sfPush(g, opp.ID, "B1", "Creature — Test", 3, []string{"R"})
	b2 := sfPush(g, opp.ID, "B2", "Creature — Test", 3, []string{"R"})
	b3 := sfPush(g, opp.ID, "B3", "Creature — Test", 3, []string{"R"})
	pr7bAttack(t, g, opp.ID, leader, dog, cat)
	if pr7bSourceShields(g) != 1 {
		t.Fatalf("shields %d, want 1 from the attack trigger", pr7bSourceShields(g))
	}
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: leader, b2: dog, b3: cat})
	sfCombat(t, g)
	if pr6Marked(g, leader) != 0 || pr6Marked(g, dog) != 0 {
		t.Fatalf("leader %d, dog %d: combat damage to Dogs you control is prevented", pr6Marked(g, leader), pr6Marked(g, dog))
	}
	if pr6Marked(g, cat) != 3 {
		t.Fatalf("cat %d, want 3: the Cat is not a Dog", pr6Marked(g, cat))
	}
	pr6Damage(t, g, b1, dog, 1)
	if pr6Marked(g, dog) != 1 {
		t.Fatalf("dog %d, want 1: noncombat damage is dealt", pr6Marked(g, dog))
	}
}

// Ethersworn Shieldmage: damage to artifact creatures, whoever controls
// them, is prevented; a plain creature's is not.
func TestRecipientSetEtherswornShieldmageShieldsArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rod := sfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
	golem := sfPush(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, nil)
	bear := sfPush(g, me.ID, "Bear", "Creature — Bear", 2, []string{"G"})
	mage := sfCast(t, g, me, game.Card{Name: "Ethersworn Shieldmage", TypeLine: "Artifact Creature — Vedalken Wizard",
		OracleID: rsShieldmage, Power: 2, Toughness: 2, Colors: []string{"W", "U"}}, game.CastSpellParams{})
	if pr7bSourceShields(g) != 1 {
		t.Fatalf("shields %d, want 1 from the enters trigger", pr7bSourceShields(g))
	}
	pr6Damage(t, g, rod, golem, 1)
	pr6Damage(t, g, rod, mage, 1)
	pr6Damage(t, g, rod, bear, 1)
	if pr6Marked(g, golem) != 0 || pr6Marked(g, mage) != 0 || pr6Marked(g, bear) != 1 {
		t.Fatalf("golem %d, mage %d (want 0 and 0), bear %d (want 1)", pr6Marked(g, golem), pr6Marked(g, mage), pr6Marked(g, bear))
	}
}

// Loyal Unicorn: with your commander, the beginning of combat shields
// your creatures from combat damage and gives the others vigilance;
// without it, nothing.
func TestRecipientSetLoyalUnicornLieutenant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Loyal Unicorn", TypeLine: "Creature — Unicorn",
		OracleID: rsLoyalUnicorn, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if pr7bSourceShields(g) != 0 {
		t.Fatal("no shield without your commander")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	unicorn := pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Loyal Unicorn", TypeLine: "Creature — Unicorn",
		OracleID: rsLoyalUnicorn, Power: 3, Toughness: 4, Owner: me2.ID, Controller: me2.ID})
	commander := pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "My Commander",
		TypeLine: "Legendary Creature — Bear", Power: 2, Toughness: 5, Owner: me2.ID, Controller: me2.ID, IsCommander: true})
	rod := sfPush(g2, opp2.ID, "Rod", "Artifact", 0, nil)
	blocker := sfPush(g2, opp2.ID, "Blocker", "Creature — Test", 3, []string{"R"})
	advanceTo(t, g2, game.StepBeginCombat)
	passPriorityAroundTable(t, g2)
	if pr7bSourceShields(g2) != 1 {
		t.Fatalf("shields %d, want 1", pr7bSourceShields(g2))
	}
	if !hasEffectiveKeyword(t, g2, commander, "vigilance") {
		t.Error("the other creature you control gains vigilance")
	}
	pr7bAttack(t, g2, opp2.ID, commander)
	pr7bBlock(t, g2, map[uuid.UUID]uuid.UUID{blocker: commander})
	sfCombat(t, g2)
	if pr6Marked(g2, commander) != 0 || pr6Marked(g2, blocker) != 2 {
		t.Fatalf("commander %d (want 0), blocker %d (want 2)", pr6Marked(g2, commander), pr6Marked(g2, blocker))
	}
	pr6Damage(t, g2, rod, unicorn, 1)
	if pr6Marked(g2, unicorn) != 1 {
		t.Fatalf("unicorn %d, want 1: noncombat damage is dealt", pr6Marked(g2, unicorn))
	}
}

// crystalFragmentsCard is the double-faced card as the deck importer
// builds it, front face up.
func crystalFragmentsCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   crystalFragmentsOracleID,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Crystal Fragments", TypeLine: "Artifact — Equipment", ManaCost: "{W}", Colors: []string{"W"}},
			{Name: "Summon: Alexander", TypeLine: "Enchantment Creature — Saga Construct", Colors: []string{"W"},
				Power: 4, Toughness: 3},
		},
	}
	c.SetFace(0)
	return c
}

// Crystal Fragments: the flip returns Summon: Alexander with its first
// lore counter, and chapter I shields the creatures you control (not
// you).
func TestRecipientSetCrystalFragmentsFlipsIntoASagaThatShieldsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frags := pushBattlefieldCardWithTimestamp(g, crystalFragmentsCard(me.ID))
	mine := sfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
	theirs := sfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"R"})
	rod := sfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("W", "W", "W", "W", "W", "W", "W")...) })
	if err := g.ActivateCatalogAbility(me.ID, frags, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	var saga *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.OracleID == crystalFragmentsOracleID {
			saga = c
		}
	}
	if saga == nil || saga.Name != "Summon: Alexander" || !saga.IsCreature() {
		t.Fatalf("the Equipment should come back as Summon: Alexander, a creature; got %+v", saga)
	}
	if saga.Counters[game.CounterLore] != 1 {
		t.Fatalf("lore %d, want 1 (CR 714.3a)", saga.Counters[game.CounterLore])
	}
	if !hasEffectiveKeyword(t, g, saga.InstanceID, "flying") {
		t.Error("Summon: Alexander has flying")
	}
	pr6Damage(t, g, rod, mine, 1)
	pr6Damage(t, g, rod, theirs, 1)
	if pr6Marked(g, mine) != 0 || pr6Marked(g, theirs) != 1 {
		t.Fatalf("mine %d (want 0), theirs %d (want 1)", pr6Marked(g, mine), pr6Marked(g, theirs))
	}
	if lost := rsHit(t, g, rod, me); lost != 1 {
		t.Fatalf("lost %d, want 1: you are not protected", lost)
	}
}
