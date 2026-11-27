package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 631 (Apocalypse: Auras): Squee's Embrace, Foul Presence, Shield
// of Duty and Reason, Manacles of Decay, Coalition Flag, Soul Link, Bloodfire
// Infusion, Yavimaya's Embrace.

const (
	oldB631Squee      = "095f8dac-15b8-4c28-ae33-6dc71152bcc7"
	oldB631Foul       = "16f3a09f-e29b-4fb5-95ed-d76ad550c2ad"
	oldB631Shield     = "1c64ab87-1da1-4b07-a3f2-73631ae3d501"
	oldB631Manacles   = "69095c25-6ae6-42f5-9c3d-1d8f86f03fae"
	oldB631Flag       = "b45a0a2f-39fd-47b9-bc9d-3d4321b377d3"
	oldB631SoulLink   = "bb81aa50-49db-4743-85fc-6b4d9ed29bc6"
	oldB631Bloodfire  = "d5bd7c0b-c412-47ed-afb8-1a9ba04be77c"
	oldB631Yavimaya   = "d5cb0370-a6f0-40c0-8b8f-553fca23778d"
	oldB631BoltOracle = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestOldB631EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB631Squee: "Squee's Embrace", oldB631Foul: "Foul Presence", oldB631Shield: "Shield of Duty and Reason",
		oldB631Manacles: "Manacles of Decay", oldB631Flag: "Coalition Flag", oldB631SoulLink: "Soul Link",
		oldB631Bloodfire: "Bloodfire Infusion", oldB631Yavimaya: "Yavimaya's Embrace",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldB631Size(t *testing.T, g *game.Game, id uuid.UUID, p, tough int) {
	t.Helper()
	if gp, gt := effectstest.PtOf(t, g, id); gp != p || gt != tough {
		t.Errorf("creature is %d/%d, want %d/%d", gp, gt, p, tough)
	}
}

// --- Squee's Embrace ---------------------------------------------------------

func TestOldB631SqueesEmbracePumpsAndReturnsTheCardToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.Enchant(t, g, "Squee's Embrace", oldB631Squee, bear)
	oldB631Size(t, g, bear, 4, 4)
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the dead creature did not return to its owner's hand")
	}
	if me.Hand.Contains(bear) {
		t.Error("the card went to the Aura's controller's hand")
	}
}

// --- Foul Presence -----------------------------------------------------------

func TestOldB631FoulPresenceShrinksTheHostAndGivesItAShrinkingTapAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.AuraBear(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.Enchant(t, g, "Foul Presence", oldB631Foul, host)
	oldB631Size(t, g, host, 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(victim)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB631Size(t, g, victim, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(victim)}); err == nil {
		t.Error("the tapped host activated again")
	}
}

func TestOldB631FoulPresenceCanKillItsOwnHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 2)
	effectstest.Enchant(t, g, "Foul Presence", oldB631Foul, host)
	oldB631Size(t, g, host, 0, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(host)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the host survived its own -1/-1 at toughness 1")
	}
}

// --- Shield of Duty and Reason -------------------------------------------------

func TestOldB631ShieldGrantsProtectionFromGreenAndBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.AuraBear(g, me.ID)
	effectstest.Enchant(t, g, "Shield of Duty and Reason", oldB631Shield, host)
	for _, kw := range []string{"protection from green", "protection from blue"} {
		if !effectstest.HasEffectiveKeyword(t, g, host, kw) {
			t.Errorf("the host lacks %s", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, host, "protection from red") {
		t.Error("the host has protection from red")
	}
}

// --- Manacles of Decay -----------------------------------------------------------

func TestOldB631ManaclesStopsAttackingAndItsAbilitiesShrinkOrStopBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 3, 3)
	aura := effectstest.Enchant(t, g, "Manacles of Decay", oldB631Manacles, host)
	card := effectstest.LayeredCard(t, g, host)
	if !game.Restricted(&card, game.CantAttack) {
		t.Fatal("the enchanted creature can attack")
	}
	if game.Restricted(&card, game.CantBlock) {
		t.Fatal("the enchanted creature can't block before the ability")
	}
	effectstest.AdvanceToMain(t, g)
	activate := func(idx int, mana string) error {
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		effectstest.FloatForTest(g, me, mana)
		err := g.ActivateCatalogAbility(me.ID, aura, idx, game.ActivateAbilityParams{Strict: true})
		if err == nil {
			effectstest.PassPriorityAroundTable(t, g)
		}
		return err
	}
	if err := activate(0, "R"); err == nil {
		t.Error("the -1/-1 ability took red mana")
	}
	if err := activate(0, "B"); err != nil {
		t.Fatalf("{B}: %v", err)
	}
	oldB631Size(t, g, host, 2, 2)
	if err := activate(0, "B"); err != nil {
		t.Fatalf("{B} again: %v", err)
	}
	oldB631Size(t, g, host, 1, 1)
	if err := activate(1, "R"); err != nil {
		t.Fatalf("{R}: %v", err)
	}
	card = effectstest.LayeredCard(t, g, host)
	if !game.Restricted(&card, game.CantBlock) {
		t.Error("the enchanted creature can still block this turn")
	}
}

// --- Coalition Flag ---------------------------------------------------------------

func TestOldB631CoalitionFlagMakesTheCreatureAFlagbearerAndForcesOpponentsToTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flagged := effectstest.B12Creature(g, me.ID, "Flagged", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Coalition Flag", oldB631Flag, flagged)
	c := effectstest.LayeredCard(t, g, flagged)
	if !c.HasSubtype("Flagbearer") || c.HasSubtype("Bear") {
		t.Errorf("type line %q, want only Flagbearer", c.TypeLine)
	}
	if o := effectstest.LayeredCard(t, g, other); o.HasSubtype("Flagbearer") {
		t.Error("another creature became a Flagbearer")
	}
	// An opponent's spell must choose the Flagbearer if it can.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if err := effectstest.CastCatalogSpellErr(t, g, "Lightning Bolt", "Instant", oldB631BoltOracle, effectstest.CardRefs(other)); err == nil {
		t.Error("an opponent's spell targeted a non-Flagbearer while a Flagbearer was legal")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Lightning Bolt", "Instant", oldB631BoltOracle, effectstest.CardRefs(flagged)); err != nil {
		t.Errorf("an opponent's spell could not target the Flagbearer: %v", err)
	}
	_ = opp
}

func TestOldB631CoalitionFlagEnchantsOnlyYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Coalition Flag", effectstest.AuraTypeLine, oldB631Flag, effectstest.CardRefs(theirs))
	if err == nil {
		t.Error("Coalition Flag enchanted an opponent's creature")
	}
}

// --- Soul Link --------------------------------------------------------------------

func TestOldB631SoulLinkGainsLifeForDamageDealtAndDamageTaken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Their Host", "Creature — Bear", 2, 5)
	other := effectstest.B12Creature(g, me.ID, "My Other", "Creature — Bear", 2, 5)
	effectstest.Enchant(t, g, "Soul Link", oldB631SoulLink, host)
	life, oppLife := me.Life, opp.Life
	effectstest.EmitNoncombatDamage(g, host, other, 3) // the host deals damage
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d after the host dealt 3, want %d", me.Life, life+3)
	}
	effectstest.EmitNoncombatDamage(g, other, host, 2) // the host is dealt damage
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d after the host was dealt 2, want %d", me.Life, life+5)
	}
	effectstest.EmitNoncombatDamage(g, other, other, 4) // neither: no gain
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 || opp.Life != oppLife {
		t.Errorf("life %d / %d, want %d / %d after unrelated damage", me.Life, opp.Life, life+5, oppLife)
	}
}

// --- Bloodfire Infusion ------------------------------------------------------------

func TestOldB631BloodfireInfusionSacrificesTheHostAndDamagesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 3)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	aura := effectstest.Enchant(t, g, "Bloodfire Infusion", oldB631Bloodfire, host)
	effectstest.AdvanceToMain(t, g)
	// Only the enchanted creature may be sacrificed.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("sacrificed a creature the Aura does not enchant")
	}
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(host) {
		t.Fatal("a refused activation sacrificed something")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{host}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(theirs) {
		t.Error("a creature survived 3 damage it should not have")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 5/5 died to 3 damage")
	}
}

func TestOldB631BloodfireInfusionNeedsRedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 3)
	aura := effectstest.Enchant(t, g, "Bloodfire Infusion", oldB631Bloodfire, host)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{host}}); err == nil {
		t.Error("activated without red mana")
	}
	if !g.Battlefield.Contains(host) {
		t.Error("the creature was sacrificed by a refused activation")
	}
}

// --- Yavimaya's Embrace -------------------------------------------------------------

func TestOldB631YavimayasEmbraceStealsPumpsAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Yavimaya's Embrace", oldB631Yavimaya, bear)
	if c := effectstest.B12Card(t, g, bear); c.Controller != me.ID {
		t.Errorf("the creature is controlled by %v, want the Aura's controller", c.Controller)
	}
	oldB631Size(t, g, bear, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
	effectstest.KillCreature(t, g, me.ID, aura)
	if c := effectstest.B12Card(t, g, bear); c.Controller != opp.ID {
		t.Errorf("control did not revert when the Aura left: %v", c.Controller)
	}
	oldB631Size(t, g, bear, 2, 2)
}
