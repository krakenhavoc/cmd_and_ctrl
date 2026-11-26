package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b469_test.go — Scars of Mirrodin, metalcraft, S82 OLD batch 469
// (#5577): nine cards, each driven through the engine.

const (
	b469Drake    = "0700d889-d235-49fe-8b92-698d486b2539"
	b469Certarch = "337533fb-2783-493c-856f-c41a83cca443"
	b469Relic    = "57b22bc3-23e8-4632-b10a-495fac5c77c7"
	b469Berserk  = "655847c6-7c5e-4659-870e-825d0cf5e90c"
	b469Vampires = "698f3526-489e-47a1-ac10-70cca73dea10"
	b469Sphinx   = "7220c391-8d0e-4f43-94fc-ff77ffa2ee8d"
	b469Silcaw   = "7a1a4fe7-6574-4242-ae6f-108c1b9a5635"
	b469Phoenix  = "884260cf-b3f3-43a4-82b6-4f6a0b19cc75"
	b469Dispense = "a47bf064-b810-43d0-89a5-13c107a317b0"
)

func TestS82OldB469EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b469Drake: "Lumengrid Drake", b469Certarch: "Vedalken Certarch", b469Relic: "Rusted Relic",
		b469Berserk: "Blade-Tribe Berserkers", b469Vampires: "Bleak Coven Vampires", b469Sphinx: "Argent Sphinx",
		b469Silcaw: "Screeching Silcaw", b469Phoenix: "Kuldotha Phoenix", b469Dispense: "Dispense Justice",
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

// b469Artifacts puts n artifacts onto p's side of the battlefield.
func b469Artifacts(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, p.ID, "Trinket", "Artifact")
	}
}

// b469CastCreature casts a creature from the active seat's hand and lets it
// settle, leaving any prompt open.
func b469CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func b469Layered(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return effectstest.LayeredCard(t, g, id)
}

func TestS82OldB469LumengridDrakeBouncesOnlyWithMetalcraft(t *testing.T) {
	// Two artifacts: the trigger never goes on the stack, so no target is asked.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b469Artifacts(g, me, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b469CastCreature(t, g, "Lumengrid Drake", "Creature — Drake", b469Drake, 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil || !g.Battlefield.Contains(theirs) {
		t.Fatal("the Drake bounced something without metalcraft")
	}
	// Three artifacts: it returns the target creature to its owner's hand.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b469Artifacts(g, me, 3)
	theirs = effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b469CastCreature(t, g, "Lumengrid Drake", "Creature — Drake", b469Drake, 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt with metalcraft")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("their creature was not returned to its owner's hand")
	}
}

func TestS82OldB469LumengridDrakeRechecksMetalcraftOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b469Artifacts(g, me, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b469CastCreature(t, g, "Lumengrid Drake", "Creature — Drake", b469Drake, 2, 2)
	effectstest.PickCard(t, g, me.ID, theirs)
	// An artifact leaves in response: the count is two as the trigger resolves.
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.IsArtifact() {
			effectstest.KillCreature(t, g, me.ID, c.InstanceID)
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the Drake bounced a creature after metalcraft was lost")
	}
}

func TestS82OldB469VedalkenCertarchTapsOnlyWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cert := effectstest.PushCatalogPermanent(g, me.ID, "Vedalken Certarch", "Creature — Vedalken Wizard", b469Certarch, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}
	b469Artifacts(g, me, 2)
	if err := g.ActivateCatalogAbility(me.ID, cert, 0, params); err == nil {
		t.Fatal("activated with two artifacts")
	}
	b469Artifacts(g, me, 1)
	if err := g.ActivateCatalogAbility(me.ID, cert, 0, params); err != nil {
		t.Fatalf("activation with three artifacts: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FindBattlefieldCardForTest(g, bear).Tapped {
		t.Error("the target creature was not tapped")
	}
}

func TestS82OldB469RustedRelicIsAGolemOnlyWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.B12Push(g, me.ID, "Rusted Relic", "Artifact", b469Relic, 0, 0)
	b469Artifacts(g, me, 1)
	if c := b469Layered(t, g, relic); c.IsCreature() {
		t.Fatal("a creature with two artifacts")
	}
	extra := effectstest.B12Permanent(g, me.ID, "Trinket", "Artifact")
	c := b469Layered(t, g, relic)
	if !c.IsCreature() || !c.IsArtifact() || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Fatalf("with three artifacts: creature=%v artifact=%v %d/%d, want a 5/5 artifact creature",
			c.IsCreature(), c.IsArtifact(), c.CurrentPower(), c.CurrentToughness())
	}
	if !c.HasSubtype("Golem") {
		t.Error("it is not a Golem")
	}
	effectstest.KillCreature(t, g, me.ID, extra)
	if c := b469Layered(t, g, relic); c.IsCreature() {
		t.Error("still a creature after metalcraft was lost")
	}
}

func TestS82OldB469BladeTribeBerserkersPumpsAndHastesWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b469CastCreature(t, g, "Blade-Tribe Berserkers", "Creature — Human Berserker", b469Berserk, 3, 3)
	if c := b469Layered(t, g, id); c.CurrentPower() != 3 || effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
		t.Fatalf("without metalcraft it is %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	b469Artifacts(g, me, 3)
	id = b469CastCreature(t, g, "Blade-Tribe Berserkers", "Creature — Human Berserker", b469Berserk, 3, 3)
	if c := b469Layered(t, g, id); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("with metalcraft it is %d/%d, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
		t.Error("it has no haste")
	}
	effectstest.DcCleanup(g)
	if c := b469Layered(t, g, id); c.CurrentPower() != 3 || effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
		t.Error("the pump and haste outlasted the turn")
	}
}

func TestS82OldB469BleakCovenVampiresDrainsFourWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b469CastCreature(t, g, "Bleak Coven Vampires", "Creature — Vampire Warrior", b469Vampires, 4, 3)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for without metalcraft")
	}
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b469Artifacts(g, me, 3)
	mine, theirs := me.Life, opp.Life
	b469CastCreature(t, g, "Bleak Coven Vampires", "Creature — Vampire Warrior", b469Vampires, 4, 3)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt with metalcraft")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-4 || me.Life != mine+4 {
		t.Errorf("life: you %d → %d, they %d → %d; want +4 and -4", mine, me.Life, theirs, opp.Life)
	}
}

func TestS82OldB469ArgentSphinxFlickersOnlyWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sphinx := effectstest.B21Push(g, me.ID, "Argent Sphinx", "Creature — Sphinx", b469Sphinx, 4, 3, "U")
	b469Artifacts(g, me, 2)
	effectstest.B06AddMana(me, "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with two artifacts")
	}
	b469Artifacts(g, me, 1)
	if err := g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activation with three artifacts: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sphinx) {
		t.Fatal("the Sphinx was not exiled")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Argent Sphinx"); n != 1 {
		t.Errorf("%d Argent Sphinxes on the battlefield at the end step, want 1", n)
	}
}

func TestS82OldB469ScreechingSilcawMillsFourWithMetalcraft(t *testing.T) {
	for _, tc := range []struct {
		artifacts, want int
	}{{2, 0}, {3, 4}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		b469Artifacts(g, me, tc.artifacts)
		silcaw := effectstest.B12Push(g, me.ID, "Screeching Silcaw", "Creature — Bird", b469Silcaw, 1, 2)
		lib := opp.Library.Size()
		effectstest.DealCombatDamage(g, me.ID, silcaw, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if got := lib - opp.Library.Size(); got != tc.want {
			t.Errorf("%d artifacts: the damaged player milled %d, want %d", tc.artifacts, got, tc.want)
		}
	}
}

func TestS82OldB469KuldothaPhoenixReturnsOnlyInYourUpkeepWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Kuldotha Phoenix", TypeLine: "Creature — Phoenix", OracleID: b469Phoenix,
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	b469Artifacts(g, me, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with two artifacts")
	}
	b469Artifacts(g, me, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated outside the upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.B06AddMana(me, "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activation in the upkeep with three artifacts: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("the Phoenix did not return to the battlefield")
	}
}

func TestS82OldB469DispenseJusticeSacrificesOneOrTwoAttackers(t *testing.T) {
	for _, tc := range []struct {
		artifacts, left int
	}{{2, 1}, {3, 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		b469Artifacts(g, me, tc.artifacts)
		a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
		b := effectstest.B12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
		effectstest.BrAttack(t, g, a, b)
		id := effectstest.B13HandCard(me, "Dispense Justice", "Instant", b469Dispense)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSacrifice(t, g, me.ID, a)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSacrifice(t, g, me.ID, b)
		effectstest.PassPriorityAroundTable(t, g)
		left := 0
		for _, c := range []uuid.UUID{a, b} {
			if g.Battlefield.Contains(c) {
				left++
			}
		}
		if left != tc.left {
			t.Errorf("%d artifacts: %d attackers left, want %d", tc.artifacts, left, tc.left)
		}
	}
}
