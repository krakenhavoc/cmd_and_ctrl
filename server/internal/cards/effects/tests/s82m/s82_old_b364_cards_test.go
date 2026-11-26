package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 364 (Conspiracy, #5557): the parley cards, Flamewright,
// Spontaneous Combustion, Vent Sentinel, Cinder Wall, Sky Spirit, Provoke and
// Howling Wolf.

const (
	oldB364Charge      = "7e6478c1-e5b3-4641-baa7-7cdeab042569"
	oldB364Enforcer    = "8ee2dd43-cda2-4cb7-9e4b-1946d5b7467e"
	oldB364Woodvine    = "b265ac18-3923-46ba-b7b6-4190855815e1"
	oldB364Rousing     = "e5e42293-ce01-4872-be2a-a79ff4312992"
	oldB364Flamewright = "0c73f566-3b8a-4d57-b7a9-7437b6f1f66d"
	oldB364Combustion  = "0f81064d-01cd-4a61-8aea-2544e4d14b70"
	oldB364Vent        = "38013607-dd80-4eb9-b0b6-f07e6023c0d0"
	oldB364CinderWall  = "16128b90-9b81-476e-8999-3ef498728c39"
	oldB364SkySpirit   = "fc7e288f-ade2-4997-b3fe-f873c26990b0"
	oldB364Provoke     = "731821b1-89fd-4166-94b8-35a1364e4d9d"
	oldB364Wolf        = "d95d0b24-2e6f-4f61-8668-26effeaaf2fb"
)

func TestOldB364EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB364Charge: "Selvala's Charge", oldB364Enforcer: "Selvala's Enforcer", oldB364Woodvine: "Woodvine Elemental",
		oldB364Rousing: "Rousing of Souls", oldB364Flamewright: "Flamewright", oldB364Combustion: "Spontaneous Combustion",
		oldB364Vent: "Vent Sentinel", oldB364CinderWall: "Cinder Wall", oldB364SkySpirit: "Sky Spirit",
		oldB364Provoke: "Provoke", oldB364Wolf: "Howling Wolf",
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

// oldB364Tops puts a nonland and a land on top of the libraries of seats 0
// and 1 so a parley sees exactly one nonland card each (seat 0: nonland, seat
// 1: land, the rest lands) and returns the nonland count.
func oldB364Tops(g *game.Game, nonland map[int]bool) int {
	n := 0
	for i, p := range g.Seats {
		p.Library.Cards = nil
		effectstest.SeedLibraryTop(p, "Filler", "Basic Land — Forest", "")
		if nonland[i] {
			effectstest.SeedLibraryTop(p, "Spell", "Sorcery", "{1}")
			n++
		} else {
			effectstest.SeedLibraryTop(p, "Forest", "Basic Land — Forest", "")
		}
	}
	return n
}

func oldB364Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, targets)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB364SelvalasChargeMakesAnElephantPerNonlandAndEveryoneDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB364Tops(g, map[int]bool{0: true, 1: true, 2: false, 3: true})
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = len(p.Hand.Cards)
	}
	oldB364Cast(t, g, "Selvala's Charge", "Sorcery", oldB364Charge, nil)
	if n := effectstest.NamedOnBattlefield(g, "Elephant", me.ID); n != 3 {
		t.Errorf("%d Elephants, want 3 (three nonland cards revealed)", n)
	}
	for i, p := range g.Seats {
		want := hands[i] + 1
		if len(p.Hand.Cards) != want {
			t.Errorf("seat %d hand %d → %d, want %d (everyone draws a card)", i, hands[i], len(p.Hand.Cards), want)
		}
	}
}

func TestOldB364RousingOfSoulsMakesFlyingSpiritsAndNoneForLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB364Tops(g, map[int]bool{})
	oldB364Cast(t, g, "Rousing of Souls", "Sorcery", oldB364Rousing, nil)
	if n := effectstest.NamedOnBattlefield(g, "Spirit", me.ID); n != 0 {
		t.Errorf("%d Spirits after four lands were revealed, want 0", n)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	oldB364Tops(g2, map[int]bool{0: true, 1: true})
	oldB364Cast(t, g2, "Rousing of Souls", "Sorcery", oldB364Rousing, nil)
	if n := effectstest.NamedOnBattlefield(g2, "Spirit", me2.ID); n != 2 {
		t.Fatalf("%d Spirits, want 2", n)
	}
	for _, c := range g2.Battlefield.Cards {
		if c.Name == "Spirit" && !effectstest.HasEffectiveKeyword(t, g2, c.InstanceID, "flying") {
			t.Error("a Spirit lacks flying")
		}
	}
}

func TestOldB364SelvalasEnforcerGrowsPerNonlandOnEntering(t *testing.T) {
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	oldB364Tops(g2, map[int]bool{0: true, 3: true})
	hand := len(me2.Hand.Cards)
	c := effectstest.CastCatalogSpell(t, g2, "Selvala's Enforcer", "Creature — Elf Warrior", oldB364Enforcer, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.PlusOnes(g2, c); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
	if len(me2.Hand.Cards) != hand+1 {
		t.Errorf("hand %d → %d, want one drawn", hand, len(me2.Hand.Cards))
	}
}

func TestOldB364WoodvineElementalPumpsAttackersPerNonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB364Tops(g, map[int]bool{0: true, 1: true})
	elemental := effectstest.B12Push(g, me.ID, "Woodvine Elemental", "Creature — Elemental", oldB364Woodvine, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, elemental, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, elemental); p != 6 || tt != 6 {
		t.Errorf("Woodvine Elemental is %d/%d, want 6/6", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the other attacker is %d/%d, want 4/4", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, idle); p != 2 || tt != 2 {
		t.Errorf("a creature that didn't attack is %d/%d, want 2/2", p, tt)
	}
}

func TestOldB364FlamewrightMakesDefendersAndSacrificesThemToPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fw := effectstest.PushCatalogPermanent(g, me.ID, "Flamewright", "Creature — Human Artificer", oldB364Flamewright, false)
	effectstest.FillPool(me, 1)
	if err := effectstest.InvB1Activate(g, me.ID, fw, 0); err != nil {
		t.Fatalf("make a Construct: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Construct", me.ID); n != 1 {
		t.Fatalf("%d Constructs, want 1", n)
	}
	// Tapped: the second ability needs {T} too.
	if err := g.ActivateCatalogAbility(me.ID, fw, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Fatal("the second ability was activated while Flamewright was tapped")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fw {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	life := opp.Life
	var construct uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" {
			construct = c.InstanceID
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, fw, 1, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{construct}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("sacrifice the Construct: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Construct", me.ID); n != 0 {
		t.Errorf("%d Constructs left, want the sacrificed one gone", n)
	}
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, opp.Life)
	}
}

func TestOldB364FlamewrightCannotSacrificeACreatureWithoutDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fw := effectstest.PushCatalogPermanent(g, me.ID, "Flamewright", "Creature — Human Artificer", oldB364Flamewright, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, fw, 1, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{bear}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Error("sacrificed a creature without defender")
	}
}

func TestOldB364SpontaneousCombustionSacrificesAndDamagesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	// A creature must be sacrificed as the spell is cast.
	if _, err := effectstest.CastWithTapParams(t, g, "Spontaneous Combustion", "Instant", "{1}{B}{R}", oldB364Combustion,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast without sacrificing a creature")
	}
	effectstest.PaCast(t, g, "Spontaneous Combustion", "Instant", "{1}{B}{R}", oldB364Combustion,
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived 3 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Fatal("the 4/4 died to 3 damage")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 3 {
		t.Errorf("the 4/4 has %d damage marked, want 3", got)
	}
}

func TestOldB364VentSentinelCountsDefendersYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vent := effectstest.PushCatalogPermanent(g, me.ID, "Vent Sentinel", "Creature — Elemental", oldB364Vent, false)
	for i := 0; i < 2; i++ {
		c := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
		g.WithWriteLock(func() {
			for j := range g.Battlefield.Cards {
				if g.Battlefield.Cards[j].InstanceID == c {
					g.Battlefield.Cards[j].Keywords = append(g.Battlefield.Cards[j].Keywords, "defender")
				}
			}
		})
	}
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "R", 1)
	life := opp.Life
	if err := effectstest.InvB1Activate(g, me.ID, vent, 0, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The Sentinel itself and the two Walls: three creatures you control with defender.
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
}

func TestOldB364CinderWallIsDestroyedAtEndOfCombatAfterBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Cinder Wall", "Creature — Wall", oldB364CinderWall, false)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, opp, wall); zone != "battlefield" {
		t.Fatalf("the Wall is in %s before end of combat", zone)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, opp, wall); zone != "graveyard" {
		t.Errorf("the Wall is in %s after end of combat, want the graveyard", zone)
	}
}

func TestOldB364CinderWallStaysWhenItDoesNotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Cinder Wall", "Creature — Wall", oldB364CinderWall, false)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, opp, wall); zone != "battlefield" {
		t.Errorf("the Wall that never blocked is in %s", zone)
	}
}

func TestOldB364SkySpiritHasFlyingAndFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Sky Spirit", "Creature — Spirit", oldB364SkySpirit, false)
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Sky Spirit lacks %s", kw)
		}
	}
}

func TestOldB364ProvokeUntapsForcesABlockAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	effectstest.AuraSeedLibrary(me, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	hand := len(me.Hand.Cards)
	oldB364Cast(t, g, "Provoke", "Instant", oldB364Provoke, effectstest.CardTarget(theirs))
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Tapped {
		t.Error("the target is still tapped")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d → %d, want one drawn", hand, len(me.Hand.Cards))
	}
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementBlocks) {
				n++
			}
		}
	}
	if n != 1 {
		t.Errorf("%d block requirements recorded for the target, want 1", n)
	}
}

func TestOldB364ProvokeCannotTargetYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Provoke", "Instant", oldB364Provoke, effectstest.CardTarget(mine)); err == nil {
		t.Error("Provoke targeted a creature its caster controls")
	}
}

func TestOldB364HowlingWolfFetchesUpToThreeCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var cards []game.Card
	for i := 0; i < 4; i++ {
		cards = append(cards, game.Card{Name: "Howling Wolf", TypeLine: "Creature — Wolf", OracleID: oldB364Wolf})
	}
	cards = append(cards, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.SeedSearchLibrary(me, cards...)
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Howling Wolf", "Creature — Wolf", oldB364Wolf, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a card with another name was offered")
	}
	if len(c.SearchCards) != 4 {
		t.Fatalf("%d cards offered, want the four copies", len(c.SearchCards))
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:4]); err == nil {
		t.Error("four cards were accepted for an up-to-three search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:2]); err != nil {
		t.Fatalf("taking two copies: %v", err)
	}
	if len(me.Hand.Cards) != hand+2 {
		t.Errorf("hand %d → %d, want two copies in", hand, len(me.Hand.Cards))
	}
}
