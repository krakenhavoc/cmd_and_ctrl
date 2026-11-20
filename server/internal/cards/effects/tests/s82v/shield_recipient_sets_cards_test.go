package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
	for _, c := range []struct{ name, oracle string }{{"Forfend", effectstest.RsForfend}, {"Blinding Fog", effectstest.RsBlindingFog}} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rod := effectstest.SfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
			mine := effectstest.SfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
			effectstest.CastCatalogSpell(t, g, c.name, "Instant", c.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			late := effectstest.SfPush(g, opp.ID, "Late", "Creature — Test", 2, []string{"R"})
			effectstest.Pr6Damage(t, g, rod, mine, 2)
			effectstest.Pr6Damage(t, g, rod, late, 2)
			if effectstest.Pr6Marked(g, mine) != 0 || effectstest.Pr6Marked(g, late) != 0 {
				t.Fatalf("marked mine %d, late %d; want 0 and 0", effectstest.Pr6Marked(g, mine), effectstest.Pr6Marked(g, late))
			}
			if lost := rsHit(t, g, rod, me); lost != 1 {
				t.Fatalf("lost %d, want 1: a player is not protected", lost)
			}
			if c.name == "Blinding Fog" {
				if !effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") {
					t.Error("my creature should have hexproof")
				}
				if effectstest.HasEffectiveKeyword(t, g, late, "hexproof") {
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
		{"Divine Light", "Sorcery", effectstest.RsDivineLight},
		{"Sivvi's Ruse", "Instant", effectstest.RsSivvisRuse},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rod := effectstest.SfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
			mine := effectstest.SfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
			theirs := effectstest.SfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"R"})
			effectstest.CastCatalogSpell(t, g, c.name, c.typeLine, c.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.Pr6Damage(t, g, rod, mine, 1)
			effectstest.Pr6Damage(t, g, rod, theirs, 1)
			if effectstest.Pr6Marked(g, mine) != 0 || effectstest.Pr6Marked(g, theirs) != 1 {
				t.Fatalf("marked mine %d (want 0), theirs %d (want 1)", effectstest.Pr6Marked(g, mine), effectstest.Pr6Marked(g, theirs))
			}
			if lost := rsHit(t, g, rod, me); lost != 1 {
				t.Fatalf("lost %d, want 1: you are not a creature you control", lost)
			}
			effectstest.FindBattlefieldCardForTest(g, theirs).Controller = me.ID
			g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
			effectstest.Pr6Damage(t, g, rod, theirs, 1)
			if effectstest.Pr6Marked(g, theirs) != 1 {
				t.Fatalf("marked %d, want still 1: a creature you gained control of is protected", effectstest.Pr6Marked(g, theirs))
			}
		})
	}
}

// Sivvi's Ruse is free only when an opponent controls a Mountain and you
// control a Plains.
func TestRecipientSetSivvisRuseFreeCastCondition(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cond := effects.AnOpponentControlsAAndYouControlA("Mountain", "Plains")
	if cond(g, me.ID) {
		t.Fatal("free with neither land")
	}
	effectstest.SfPush(g, me.ID, "My Mountain", "Basic Land — Mountain", 0, nil)
	effectstest.SfPush(g, me.ID, "Plains", "Basic Land — Plains", 0, nil)
	if cond(g, me.ID) {
		t.Fatal("your own Mountain doesn't count")
	}
	effectstest.SfPush(g, opp.ID, "Mountain", "Basic Land — Mountain", 0, nil)
	if !cond(g, me.ID) {
		t.Fatal("an opponent's Mountain and your Plains make it free")
	}
	effectstest.CastWithAltCost(t, g, "Sivvi's Ruse", "Instant", effectstest.RsSivvisRuse, "free")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.Pr7bSourceShields(g) != 1 {
		t.Fatalf("shields %d, want 1", effectstest.Pr7bSourceShields(g))
	}
}

// Chameleon Blur: damage creatures would deal to players. A noncreature
// source's damage to a player and a creature's damage to a creature are
// dealt.
func TestRecipientSetChameleonBlurStopsCreaturesHittingPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rod := effectstest.SfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
	bear := effectstest.SfPush(g, opp.ID, "Bear", "Creature — Bear", 2, []string{"G"})
	mine := effectstest.SfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
	effectstest.CastCatalogSpell(t, g, "Chameleon Blur", "Instant", effectstest.RsChameleonBlur, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if lost := rsHit(t, g, bear, me) + rsHit(t, g, bear, opp); lost != 0 {
		t.Fatalf("lost %d, want 0: a creature's damage to any player is prevented", lost)
	}
	if lost := rsHit(t, g, rod, me); lost != 1 {
		t.Fatalf("lost %d, want 1: a noncreature source's damage is dealt", lost)
	}
	effectstest.Pr6Damage(t, g, bear, mine, 1)
	if effectstest.Pr6Marked(g, mine) != 1 {
		t.Fatalf("marked %d, want 1: damage to a creature is dealt", effectstest.Pr6Marked(g, mine))
	}
}

// Commencement of Festivities and Defend the Hearth: combat damage to
// players only. A blocked attacker and its blocker deal theirs; the
// unblocked attacker's is prevented, and noncombat damage to a player
// is dealt.
func TestRecipientSetCardsCombatDamageToPlayers(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Commencement of Festivities", effectstest.RsCommencement},
		{"Defend the Hearth", effectstest.RsDefendTheHearth},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			free := effectstest.SfPush(g, me.ID, "Free", "Creature — Test", 3, []string{"R"})
			held := effectstest.SfPush(g, me.ID, "Held", "Creature — Test", 2, []string{"R"})
			wall := effectstest.SfPush(g, opp.ID, "Wall", "Creature — Wall", 1, []string{"W"})
			effectstest.Pr7bAttack(t, g, opp.ID, free, held)
			effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: held})
			sfCast(t, g, me, sfInstant(c.name, c.oracle), game.CastSpellParams{})
			if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 0 {
				t.Fatalf("defender lost %d, want 0", lost)
			}
			if effectstest.Pr6Marked(g, wall) != 2 || effectstest.Pr6Marked(g, held) != 1 {
				t.Fatalf("wall %d (want 2), held %d (want 1): combat damage to creatures is dealt", effectstest.Pr6Marked(g, wall), effectstest.Pr6Marked(g, held))
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := effectstest.SfPush(g, opp.ID, "Red", "Creature — Test", 2, []string{"R"})
	gold := effectstest.SfPush(g, opp.ID, "Gold", "Creature — Test", 2, []string{"B", "G"})
	green := effectstest.SfPush(g, opp.ID, "Green", "Creature — Test", 2, []string{"G"})
	mine := effectstest.SfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
	theirs := effectstest.SfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"W"})
	effectstest.CastCatalogSpell(t, g, "Surge of Salvation", "Instant", effectstest.RsSurgeOfSalvation, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") || effectstest.HasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("my permanents gain hexproof, an opponent's do not")
	}
	effectstest.Pr6Damage(t, g, red, mine, 1)
	effectstest.Pr6Damage(t, g, gold, mine, 1)
	if effectstest.Pr6Marked(g, mine) != 0 {
		t.Fatalf("marked %d, want 0: red and black-green sources are stopped", effectstest.Pr6Marked(g, mine))
	}
	effectstest.Pr6Damage(t, g, green, mine, 1)
	effectstest.Pr6Damage(t, g, red, theirs, 1)
	if effectstest.Pr6Marked(g, mine) != 1 || effectstest.Pr6Marked(g, theirs) != 1 {
		t.Fatalf("mine %d (want 1, the green source's), theirs %d (want 1)", effectstest.Pr6Marked(g, mine), effectstest.Pr6Marked(g, theirs))
	}
	if lost := rsHit(t, g, red, me); lost != 1 {
		t.Fatalf("lost %d, want 1: damage to you is not prevented", lost)
	}
}

// Pack Leader: other Dogs get +1/+1; when it attacks, combat damage to
// Dogs you control is prevented — not to you, not to your other
// creatures, and not noncombat damage.
func TestRecipientSetPackLeaderShieldsDogsInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leader := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Pack Leader", TypeLine: "Creature — Dog", OracleID: effectstest.RsPackLeader,
		Power: 2, Toughness: 2, Colors: []string{"W"}})
	dog := effectstest.SfPush(g, me.ID, "Dog", "Creature — Dog", 1, []string{"W"})
	cat := effectstest.SfPush(g, me.ID, "Cat", "Creature — Cat", 1, []string{"W"})
	if p := effectstest.FindBattlefieldCardForTest(g, dog).CurrentPower(); p != 2 {
		t.Fatalf("Dog power %d, want 2 with the anthem", p)
	}
	if p := effectstest.FindBattlefieldCardForTest(g, leader).CurrentPower(); p != 2 {
		t.Fatalf("Pack Leader power %d, want 2: the anthem is for other Dogs", p)
	}
	b1 := effectstest.SfPush(g, opp.ID, "B1", "Creature — Test", 3, []string{"R"})
	b2 := effectstest.SfPush(g, opp.ID, "B2", "Creature — Test", 3, []string{"R"})
	b3 := effectstest.SfPush(g, opp.ID, "B3", "Creature — Test", 3, []string{"R"})
	effectstest.Pr7bAttack(t, g, opp.ID, leader, dog, cat)
	if effectstest.Pr7bSourceShields(g) != 1 {
		t.Fatalf("shields %d, want 1 from the attack trigger", effectstest.Pr7bSourceShields(g))
	}
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: leader, b2: dog, b3: cat})
	effectstest.SfCombat(t, g)
	if effectstest.Pr6Marked(g, leader) != 0 || effectstest.Pr6Marked(g, dog) != 0 {
		t.Fatalf("leader %d, dog %d: combat damage to Dogs you control is prevented", effectstest.Pr6Marked(g, leader), effectstest.Pr6Marked(g, dog))
	}
	if effectstest.Pr6Marked(g, cat) != 3 {
		t.Fatalf("cat %d, want 3: the Cat is not a Dog", effectstest.Pr6Marked(g, cat))
	}
	effectstest.Pr6Damage(t, g, b1, dog, 1)
	if effectstest.Pr6Marked(g, dog) != 1 {
		t.Fatalf("dog %d, want 1: noncombat damage is dealt", effectstest.Pr6Marked(g, dog))
	}
}

// Ethersworn Shieldmage: damage to artifact creatures, whoever controls
// them, is prevented; a plain creature's is not.
func TestRecipientSetEtherswornShieldmageShieldsArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rod := effectstest.SfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
	golem := effectstest.SfPush(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, nil)
	bear := effectstest.SfPush(g, me.ID, "Bear", "Creature — Bear", 2, []string{"G"})
	mage := sfCast(t, g, me, game.Card{Name: "Ethersworn Shieldmage", TypeLine: "Artifact Creature — Vedalken Wizard",
		OracleID: effectstest.RsShieldmage, Power: 2, Toughness: 2, Colors: []string{"W", "U"}}, game.CastSpellParams{})
	if effectstest.Pr7bSourceShields(g) != 1 {
		t.Fatalf("shields %d, want 1 from the enters trigger", effectstest.Pr7bSourceShields(g))
	}
	effectstest.Pr6Damage(t, g, rod, golem, 1)
	effectstest.Pr6Damage(t, g, rod, mage, 1)
	effectstest.Pr6Damage(t, g, rod, bear, 1)
	if effectstest.Pr6Marked(g, golem) != 0 || effectstest.Pr6Marked(g, mage) != 0 || effectstest.Pr6Marked(g, bear) != 1 {
		t.Fatalf("golem %d, mage %d (want 0 and 0), bear %d (want 1)", effectstest.Pr6Marked(g, golem), effectstest.Pr6Marked(g, mage), effectstest.Pr6Marked(g, bear))
	}
}

// Loyal Unicorn: with your commander, the beginning of combat shields
// your creatures from combat damage and gives the others vigilance;
// without it, nothing.
func TestRecipientSetLoyalUnicornLieutenant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Loyal Unicorn", TypeLine: "Creature — Unicorn",
		OracleID: effectstest.RsLoyalUnicorn, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.Pr7bSourceShields(g) != 0 {
		t.Fatal("no shield without your commander")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	unicorn := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Loyal Unicorn", TypeLine: "Creature — Unicorn",
		OracleID: effectstest.RsLoyalUnicorn, Power: 3, Toughness: 4, Owner: me2.ID, Controller: me2.ID})
	commander := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "My Commander",
		TypeLine: "Legendary Creature — Bear", Power: 2, Toughness: 5, Owner: me2.ID, Controller: me2.ID, IsCommander: true})
	rod := effectstest.SfPush(g2, opp2.ID, "Rod", "Artifact", 0, nil)
	blocker := effectstest.SfPush(g2, opp2.ID, "Blocker", "Creature — Test", 3, []string{"R"})
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.Pr7bSourceShields(g2) != 1 {
		t.Fatalf("shields %d, want 1", effectstest.Pr7bSourceShields(g2))
	}
	if !effectstest.HasEffectiveKeyword(t, g2, commander, "vigilance") {
		t.Error("the other creature you control gains vigilance")
	}
	effectstest.Pr7bAttack(t, g2, opp2.ID, commander)
	effectstest.Pr7bBlock(t, g2, map[uuid.UUID]uuid.UUID{blocker: commander})
	effectstest.SfCombat(t, g2)
	if effectstest.Pr6Marked(g2, commander) != 0 || effectstest.Pr6Marked(g2, blocker) != 2 {
		t.Fatalf("commander %d (want 0), blocker %d (want 2)", effectstest.Pr6Marked(g2, commander), effectstest.Pr6Marked(g2, blocker))
	}
	effectstest.Pr6Damage(t, g2, rod, unicorn, 1)
	if effectstest.Pr6Marked(g2, unicorn) != 1 {
		t.Fatalf("unicorn %d, want 1: noncombat damage is dealt", effectstest.Pr6Marked(g2, unicorn))
	}
}

// crystalFragmentsCard is the double-faced card as the deck importer
// builds it, front face up.
func crystalFragmentsCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   effects.CrystalFragmentsOracleID,
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frags := effectstest.PushBattlefieldCardWithTimestamp(g, crystalFragmentsCard(me.ID))
	mine := effectstest.SfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"W"})
	theirs := effectstest.SfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"R"})
	rod := effectstest.SfPush(g, opp.ID, "Rod", "Artifact", 0, nil)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("W", "W", "W", "W", "W", "W", "W")...) })
	if err := g.ActivateCatalogAbility(me.ID, frags, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var saga *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.OracleID == effects.CrystalFragmentsOracleID {
			saga = c
		}
	}
	if saga == nil || saga.Name != "Summon: Alexander" || !saga.IsCreature() {
		t.Fatalf("the Equipment should come back as Summon: Alexander, a creature; got %+v", saga)
	}
	if saga.Counters[game.CounterLore] != 1 {
		t.Fatalf("lore %d, want 1 (CR 714.3a)", saga.Counters[game.CounterLore])
	}
	if !effectstest.HasEffectiveKeyword(t, g, saga.InstanceID, "flying") {
		t.Error("Summon: Alexander has flying")
	}
	effectstest.Pr6Damage(t, g, rod, mine, 1)
	effectstest.Pr6Damage(t, g, rod, theirs, 1)
	if effectstest.Pr6Marked(g, mine) != 0 || effectstest.Pr6Marked(g, theirs) != 1 {
		t.Fatalf("mine %d (want 0), theirs %d (want 1)", effectstest.Pr6Marked(g, mine), effectstest.Pr6Marked(g, theirs))
	}
	if lost := rsHit(t, g, rod, me); lost != 1 {
		t.Fatalf("lost %d, want 1: you are not protected", lost)
	}
}
