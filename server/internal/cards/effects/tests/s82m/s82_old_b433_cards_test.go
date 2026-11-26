package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set AVR, batch 433 (#5570): Avacyn Restored Auras.

const (
	avrGhostlyTouch     = "02d25219-888d-4032-9e9e-8c6f879a9084"
	avrSpiritAway       = "15b618f0-281f-4745-8504-bd70e7b4a267"
	avrPredatorsGambit  = "4a1f94c0-6858-4296-a4a6-27d2201ed7f4"
	avrDualCasting      = "4d22b01b-63dd-4698-a221-0acaefb2000b"
	avrCommandersAuth   = "57bc1106-81a0-4d91-94db-25e3a5f989cb"
	avrDefang           = "5a6c81b8-71f0-468b-85ad-d87e9a712ecf"
	avrGhoulflesh       = "640c0424-b862-438f-a527-ff606151040a"
	avrGuiseOfFire      = "69d4d459-d680-45d9-a3dc-468e4ecce12e"
	avrLightningProwess = "85f40ced-a065-470a-a686-faaca17fa40c"
	avrMaliciousIntent  = "8aac06d0-c353-4af9-a243-feb59b9d0c54"
	avrSpectralPrison   = "b5991a13-5fcd-4d2d-9db2-d3d5af7259f5"
	avrCallToServe      = "bb4be6b2-507c-4d7c-bea4-010670dce3da"
	avrGrounded         = "cca3557c-c911-494e-95e7-cebec3442d63"
)

func avrEffective(g *game.Game, id uuid.UUID) (c game.Characteristic) {
	g.ReadSnapshot(func() {
		if card := effectstest.FindBattlefieldCardForTest(g, id); card != nil {
			c = card.Effective()
		}
	})
	return c
}

func TestAvrB433AurasRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avrGhostlyTouch: "Ghostly Touch", avrSpiritAway: "Spirit Away", avrPredatorsGambit: "Predator's Gambit",
		avrDualCasting: "Dual Casting", avrCommandersAuth: "Commander's Authority", avrDefang: "Defang",
		avrGhoulflesh: "Ghoulflesh", avrGuiseOfFire: "Guise of Fire", avrLightningProwess: "Lightning Prowess",
		avrMaliciousIntent: "Malicious Intent", avrSpectralPrison: "Spectral Prison", avrCallToServe: "Call to Serve",
		avrGrounded: "Grounded",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestAvrB433GhoulfleshShrinksAndAddsBlackAndZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.AuraBear(g, g.Seats[0].ID)
	effectstest.AuraCast(t, g, "Ghoulflesh", avrGhoulflesh, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 1 {
		t.Fatalf("enchanted Bear = %d/%d, want 1/1", p, tough)
	}
	c := avrEffective(g, bear)
	if !slices.Contains(c.Subtypes, "Zombie") {
		t.Errorf("subtypes %v, want a Zombie (in addition to the rest)", c.Subtypes)
	}
	if !slices.Contains(c.Colors, "B") {
		t.Errorf("colors %v, want black", c.Colors)
	}
}

func TestAvrB433PredatorsGambitIntimidateOnlyWhenAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Predator's Gambit", avrPredatorsGambit, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 3 {
		t.Fatalf("enchanted Bear = %d/%d, want 4/3", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "intimidate") {
		t.Error("alone, the creature should have intimidate")
	}
	other := effectstest.AuraBear(g, me.ID)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "intimidate") {
		t.Error("with another creature, intimidate should be gone")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "intimidate") {
		t.Error("intimidate should come back once it is alone again")
	}
}

func TestAvrB433GuiseOfFirePumpsAndForcesAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.AuraBear(g, g.Seats[0].ID)
	effectstest.AuraCast(t, g, "Guise of Fire", avrGuiseOfFire, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 1 {
		t.Fatalf("enchanted Bear = %d/%d, want 3/1", p, tough)
	}
	if n := len(avrEffective(g, bear).AttackRequirements); n != 1 {
		t.Errorf("attack requirements = %d, want 1 (attacks each combat if able)", n)
	}
}

func TestAvrB433CallToServeIsAnAngelWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Call to Serve", avrCallToServe, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 4 {
		t.Fatalf("enchanted Bear = %d/%d, want 3/4", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("enchanted creature should fly")
	}
	if !slices.Contains(avrEffective(g, bear).Subtypes, "Angel") {
		t.Errorf("subtypes %v, want Angel", avrEffective(g, bear).Subtypes)
	}
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Call to Serve", effectstest.AuraTypeLine, avrCallToServe,
		[]game.TargetRef{{Kind: game.TargetCard, ID: black}}); err == nil {
		t.Error("Call to Serve enchants a nonblack creature only")
	}
}

func TestAvrB433GroundedRemovesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake", Keywords: []string{"flying"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if !effectstest.HasKeywordOnBattlefield(t, g, flier, "flying") {
		t.Fatal("setup: the Drake should fly")
	}
	effectstest.AuraCast(t, g, "Grounded", avrGrounded, flier)
	if effectstest.HasKeywordOnBattlefield(t, g, flier, "flying") {
		t.Error("the enchanted creature should have lost flying")
	}
}

func TestAvrB433SpiritAwayStealsAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Spirit Away", avrSpiritAway, theirs)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Errorf("controller %s, want the Aura's controller", got)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 4 || tough != 4 {
		t.Errorf("enchanted Bear = %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "flying") {
		t.Error("enchanted creature should fly")
	}
}

func TestAvrB433DefangPreventsItsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	other := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Defang", avrDefang, bear)
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(bear, opp.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	if opp.Life != life {
		t.Errorf("life = %d, want %d: the enchanted creature's damage is prevented", opp.Life, life)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(other, opp.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	if opp.Life != life-3 {
		t.Errorf("life = %d, want %d: another creature's damage is not prevented", opp.Life, life-3)
	}
}

func TestAvrB433LightningProwessGivesHasteAndAPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Lightning Prowess", avrLightningProwess, bear)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("enchanted creature should have haste")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
}

func TestAvrB433MaliciousIntentStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Malicious Intent", avrMaliciousIntent, bear)
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, theirs)&game.CantBlock == 0 {
		t.Error("the targeted creature should be unable to block this turn")
	}
	if effectstest.RestrictionsOf(t, g, bear)&game.CantBlock != 0 {
		t.Error("the enchanted creature itself is not restricted")
	}
}

func TestAvrB433CommandersAuthorityMakesAHumanEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Commander's Authority", avrCommandersAuth, bear)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Human"); n < 1 {
		t.Errorf("Human tokens = %d, want one at the enchanted creature's controller's upkeep", n)
	}
}

func TestAvrB433SpectralPrisonHoldsTheCreatureAndFallsWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.AuraBear(g, opp.ID)
	aura := effectstest.AuraCast(t, g, "Spectral Prison", avrSpectralPrison, theirs)
	if err := g.TapCard(theirs, true); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if c := effectstest.FindBattlefieldCardForTest(g, theirs); c == nil || !c.Tapped {
		t.Fatal("the enchanted creature untapped in its controller's untap step")
	}
	// A spell that targets it: the Aura is sacrificed.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("Spectral Prison should be sacrificed when the creature is targeted by a spell")
	}
	_ = me
}

func TestAvrB433DualCastingCopiesASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Dual Casting", avrDualCasting, bear)
	life := opp.Life
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.AcPlayerRef(opp.ID))
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.SettleThroughCopyRetarget(t, g, me.ID, opp.ID)
	if opp.Life != life-6 {
		t.Errorf("life = %d, want %d (the Bolt and its copy)", opp.Life, life-6)
	}
}

func TestAvrB433GhostlyTouchTapsOrUntapsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Ghostly Touch", avrGhostlyTouch, bear)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).SummonedThisTurn = false })
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("\"tap it\" left the target untapped")
	}
}
