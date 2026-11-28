package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b686_test.go — Mercadian Masques Auras, batch 686 (#5619). Every card is
// the real catalog entry played through the engine.

const (
	n686Maggot     = "ab4887e6-f71b-462d-9243-9ebe54da98f4"
	n686Claim      = "ac285479-46c2-49f3-a7ff-31affe23c21f"
	n686Stamina    = "b77217b8-22d1-4dec-a5b4-18fccfcb70e9"
	n686Muzzle     = "bdf497b5-166f-46a7-8e18-ae0b8f97768c"
	n686Charisma   = "d2abfcd1-c00c-43fa-a9ca-a9577b3c8a60"
	n686Hunger     = "d37a361e-83e8-4f0a-bd09-3135fb918471"
	n686Buoyancy   = "f4e4060d-bfff-4991-9ae5-8f848304cd1e"
	n686ChoManno   = "fe9f2a67-b8af-48fd-83cd-d15168d2e64f"
	n686AllPool    = "{2}{G}{G}{B}{B}{B}{U}{U}{U}{R}{W}{W}"
	n686AuraType   = "Enchantment — Aura"
	n686CreatureTL = "Creature — Test"
)

func TestS82OldB686EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		n686Maggot: "Maggot Therapy", n686Claim: "Pretender's Claim", n686Stamina: "Stamina",
		n686Muzzle: "Muzzle", n686Charisma: "Charisma", n686Hunger: "Unnatural Hunger",
		n686Buoyancy: "Buoyancy", n686ChoManno: "Cho-Manno's Blessing",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// n686Cast casts an Aura from the active player's hand onto host under strict rules, paid from a
// full pool, and lets it resolve. It returns the Aura.
func n686Cast(t *testing.T, g *game.Game, name, cost, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, n686AllPool)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: n686AuraType, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestS82OldB686MaggotTherapyShrinksAndKills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.SeedBear(g, opp.ID)
	n686Cast(t, g, "Maggot Therapy", "{2}{B}", n686Maggot, bear)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 enchanted with +2/-2 survived")
	}
	g2 := effectstest.NewCatalogGame(t)
	big := effectstest.PushTestPermanent(g2, g2.Seats[1].ID, game.Card{Name: "Ox", TypeLine: n686CreatureTL, Power: 2, Toughness: 4})
	n686Cast(t, g2, "Maggot Therapy", "{2}{B}", n686Maggot, big)
	if p, th := effectstest.PtOf(t, g2, big); p != 4 || th != 2 {
		t.Errorf("enchanted 2/4 is %d/%d, want 4/2", p, th)
	}
}

func TestS82OldB686BuoyancyGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.SeedBear(g, g.Seats[1].ID)
	n686Cast(t, g, "Buoyancy", "{1}{U}", n686Buoyancy, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature has no flying")
	}
}

func TestS82OldB686MuzzlePreventsDamageDealtByTheEnchantedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	biter := effectstest.SeedBear(g, opp.ID)
	other := effectstest.SeedBear(g, opp.ID)
	n686Cast(t, g, "Muzzle", "{1}{W}", n686Muzzle, biter)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(biter, me.ID, 2) })
	if me.Life != life {
		t.Errorf("the enchanted creature's damage went through: %d, want %d", me.Life, life)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(other, me.ID, 2) })
	if me.Life != life-2 {
		t.Errorf("another creature's damage was prevented: %d, want %d", me.Life, life-2)
	}
}

func TestS82OldB686ChoMannosBlessingGivesProtectionFromTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	n686Cast(t, g, "Cho-Manno's Blessing", "{W}{W}", n686ChoManno, bear)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, bear)
	if !game.ProtectedFrom(&c, &game.Characteristic{Colors: []string{"R"}}) {
		t.Error("the enchanted creature has no protection from the chosen colour")
	}
	if game.ProtectedFrom(&c, &game.Characteristic{Colors: []string{"G"}}) {
		t.Error("the enchanted creature has protection from a colour that was not chosen")
	}
}

func TestS82OldB686StaminaGrantsVigilanceAndSacrificesToRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	aura := n686Cast(t, g, "Stamina", "{2}{G}", n686Stamina, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the enchanted creature has no vigilance")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	if n := effectstest.RegenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("the creature has %d regeneration shields, want 1", n)
	}
	effectstest.DestroyIt(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regenerated creature was destroyed")
	}
}

func TestS82OldB686PretendersClaimTapsTheDefendersLandsWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.SeedBear(g, me.ID)
	land1 := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	land2 := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	mine := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	n686Cast(t, g, "Pretender's Claim", "{1}{B}", n686Claim, attacker)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, l := range []uuid.UUID{land1, land2} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, l); !tapped {
			t.Error("a land of the defending player is untapped")
		}
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, mine); tapped {
		t.Error("the Aura controller's own land was tapped")
	}
}

func TestS82OldB686CharismaTakesACreatureItsHostDamagesWhileTheAuraStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Ogre", TypeLine: n686CreatureTL, Power: 3, Toughness: 3})
	victim := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 5})
	aura := n686Cast(t, g, "Charisma", "{U}{U}{U}", n686Charisma, host)
	effectstest.DealDamageTo(t, g, host, victim, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the damaged creature is not under the Aura's controller (ok=%v)", ok)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if c, _ := effectstest.BattlefieldCard(g, victim); c.Controller != opp.ID {
		t.Error("control did not return when the Aura left")
	}
}

func n686PayPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestS82OldB686UnnaturalHungerWithNoOtherCreatureDealsDamageEqualToPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	host := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Ogre", TypeLine: n686CreatureTL, Power: 3, Toughness: 3})
	n686Cast(t, g, "Unnatural Hunger", "{3}{B}{B}", n686Hunger, host)
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p := n686PayPrompt(g, opp.ID); p != nil {
		effectstest.PassPriorityAroundTable(t, g)
		if err := g.ResolvePayUnless(p.ID, opp.ID, false); err != nil {
			t.Fatalf("decline: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opp.Life != life-3 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, life-3)
	}
}

func TestS82OldB686UnnaturalHungerLetsThemSacrificeAnotherCreatureButNotTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	host := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Ogre", TypeLine: n686CreatureTL, Power: 3, Toughness: 3})
	fodder := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Fodder", TypeLine: n686CreatureTL, Power: 1, Toughness: 1})
	n686Cast(t, g, "Unnatural Hunger", "{3}{B}{B}", n686Hunger, host)
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	p := n686PayPrompt(g, opp.ID)
	if p == nil {
		t.Fatal("no sacrifice prompt for a player with another creature")
	}
	if err := g.ResolvePayUnlessWithCards(p.ID, opp.ID, true, []uuid.UUID{host}); err == nil && g.Battlefield.Contains(host) == false {
		t.Fatal("the enchanted creature itself was accepted as the sacrifice")
	}
	if p = n686PayPrompt(g, opp.ID); p != nil {
		if err := g.ResolvePayUnlessWithCards(p.ID, opp.ID, true, []uuid.UUID{fodder}); err != nil {
			t.Fatalf("sacrifice another creature: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	if !g.Battlefield.Contains(host) {
		t.Error("the enchanted creature was sacrificed")
	}
	if opp.Life != life {
		t.Errorf("the player took damage after sacrificing: %d, want %d", opp.Life, life)
	}
}
