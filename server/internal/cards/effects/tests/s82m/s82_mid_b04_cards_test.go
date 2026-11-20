package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	midB04DiregrafHorde   = "d7947858-f11d-438c-8ad3-991afe950a11"
	midB04Startle         = "7996222e-68c8-4a5c-9531-c60cdd88ae05"
	midB04Procession      = "39e20898-cf01-48dc-8972-7ac500c3fa79"
	midB04FlipTheSwitch   = "a28eb946-c7f2-4c90-ab6c-b194e290e33d"
	midB04HobblingZombie  = "9d1c3ec1-00a4-4890-93a3-e3cf137ca7e4"
	midB04Harvest         = "bbef4d8e-5861-4d50-9589-14b46c71e9af"
	midB04Revenge         = "d5d869bc-8996-4d19-9e32-d1cfb968875c"
	midB04RottenReunion   = "1bde8eaa-9beb-4529-b540-5e0155c09713"
	midB04NoWayOut        = "8cbbb1f4-4654-4865-a14c-0c94cd5d8b19"
	midB04BladestitchedSk = "d72b8254-3df8-431b-955a-ec2aea493e2b"
	midB04Crawl           = "16c99713-f19e-4b25-9b03-6a8d8203adfe"
	midB04TombTyrant      = "0fb2c61a-9e74-4ee2-810a-b77a3b45f446"
	midB04EmptyTheLab     = "84f3736f-5b2a-40f9-8c3b-c18f8c637c22"
)

func TestMidB04CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		midB04DiregrafHorde:   "Diregraf Horde",
		midB04Startle:         "Startle",
		midB04Procession:      "Ghoulish Procession",
		midB04FlipTheSwitch:   "Flip the Switch",
		midB04HobblingZombie:  "Hobbling Zombie",
		midB04Harvest:         "Ghoulcaller's Harvest",
		midB04Revenge:         "Revenge of the Drowned",
		midB04RottenReunion:   "Rotten Reunion",
		midB04NoWayOut:        "No Way Out",
		midB04BladestitchedSk: "Bladestitched Skaab",
		midB04Crawl:           "Crawl from the Cellar",
		midB04TombTyrant:      "Tomb Tyrant",
		midB04EmptyTheLab:     "Empty the Laboratory",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// midB04DecayedZombies counts a controller's decayed Zombie tokens, reading
// decayed through the layered view.
func midB04DecayedZombies(t *testing.T, g *game.Game, controller uuid.UUID) int {
	t.Helper()
	n := 0
	for _, id := range effectstest.BattlefieldIDsNamed(g, "Zombie") {
		c, ok := g.LookupCardForEffect(id)
		if ok && c.Controller == controller && c.IsToken() && effectstest.HasEffectiveKeyword(t, g, id, game.KeywordDecayed) {
			n++
		}
	}
	return n
}

func midB04Zombie(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, "Creature — Zombie", 2, 2)
}

func TestMidB04DiregrafHordeMakesTwoDecayedZombiesAndExilesUpToTwoCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B17GraveyardCard(me, "My Card", "Instant", "{U}")
	theirs := effectstest.B17GraveyardCard(opp, "Their Card", "Creature — Bear", "{1}{G}")
	spare := effectstest.B17GraveyardCard(opp, "Spare Card", "Sorcery", "{R}")
	effectstest.CastCatalogSpell(t, g, "Diregraf Horde", "Creature — Zombie", midB04DiregrafHorde, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 2 {
		t.Fatalf("decayed Zombie tokens = %d, want 2", n)
	}
	effectstest.B17PickCards(t, g, me.ID, mine, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Error("both chosen cards are exiled, from two graveyards")
	}
	if g.Exile.Contains(spare) {
		t.Error("a third card was exiled; the clause is up to two")
	}
}

func TestMidB04DiregrafHordeWithEmptyGraveyardsStillMakesTheZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Diregraf Horde", "Creature — Zombie", midB04DiregrafHorde, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 2 {
		t.Fatalf("decayed Zombie tokens = %d, want 2", n)
	}
	if !effectstest.StackFullyEmpty(g) {
		t.Error("the reflexive trigger had no target and should be gone")
	}
}

func TestMidB04StartleShrinksMakesAZombieAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Startle", "Instant", midB04Startle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 3 {
		t.Errorf("target is %d/%d, want 1/3", p, tough)
	}
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Errorf("decayed Zombie tokens = %d, want 1", n)
	}
	// The spell leaves the hand as it is cast: one drawn card nets even.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

func TestMidB04StartleRefusesAnIllegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogHandCard(active, "Startle", "Instant", midB04Startle)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(land), Strict: true}); err == nil {
		t.Fatal("Startle targeted a land")
	}
}

func TestMidB04GhoulishProcessionTriggersOncePerTurnForNontokenDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ghoulish Procession", "Enchantment", midB04Procession, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	tok := effectstest.PushToken(g, opp.ID, effects.TokenCard("2/2 black Zombie with decayed"))
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b, tok} {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Fatalf("one batch of nontoken deaths made %d Zombies, want 1", n)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(c); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Errorf("a second batch the same turn made it %d Zombies, want still 1", n)
	}
}

func TestMidB04GhoulishProcessionIgnoresATokenDyingAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ghoulish Procession", "Enchantment", midB04Procession, false)
	tok := effectstest.PushToken(g, opp.ID, effects.TokenCard("2/2 black Zombie with decayed"))
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(tok); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 0 {
		t.Errorf("a token's death made %d Zombies, want 0", n)
	}
}

func TestMidB04FlipTheSwitchCountersUnlessFourIsPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
		effectstest.CastCatalogSpell(t, g, "Flip the Switch", "Instant", midB04FlipTheSwitch, effectstest.CardTarget(spell))
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.FillPool(me, 4)
		}
		if c := effectstest.AnswerPayUnless(t, g, me.ID, pay); c == nil {
			t.Fatal("no pay-unless prompt")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if countered := !g.Stack.Contains(spell) && !g.Battlefield.Contains(spell); countered == pay {
			t.Errorf("pay=%v: countered=%v", pay, countered)
		}
		if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
			t.Errorf("pay=%v: decayed Zombies = %d, want 1 either way", pay, n)
		}
	}
}

func TestMidB04HobblingZombieHasDeathtouchAndLeavesADecayedZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	z := effectstest.B12Push(g, me.ID, "Hobbling Zombie", "Creature — Zombie", midB04HobblingZombie, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, z, "deathtouch") {
		t.Error("Hobbling Zombie lacks deathtouch")
	}
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(z); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Errorf("decayed Zombies after it died = %d, want 1", n)
	}
}

func TestMidB04GhoulcallersHarvestMakesHalfRoundedUp(t *testing.T) {
	for _, tc := range []struct{ creatures, spells, want int }{{0, 2, 0}, {1, 0, 1}, {4, 3, 2}, {5, 0, 3}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < tc.creatures; i++ {
			effectstest.B17GraveyardCard(me, "Dead", "Creature — Bear", "{G}")
		}
		for i := 0; i < tc.spells; i++ {
			effectstest.B17GraveyardCard(me, "Spell", "Sorcery", "{G}")
		}
		effectstest.CastCatalogSpell(t, g, "Ghoulcaller's Harvest", "Sorcery", midB04Harvest, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if n := midB04DecayedZombies(t, g, me.ID); n != tc.want {
			t.Errorf("%d creature cards: %d Zombies, want %d", tc.creatures, n, tc.want)
		}
	}
}

func TestMidB04GhoulcallersHarvestFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B17GraveyardCard(me, "Dead", "Creature — Bear", "{G}")
	effectstest.B17GraveyardCard(me, "Dead2", "Creature — Bear", "{G}")
	id := effectstest.SeedGraveyardCard(t, g, "Ghoulcaller's Harvest", "Sorcery", midB04Harvest)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Errorf("Zombies = %d, want 1 (two creature cards, halved)", n)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back sorcery should be exiled")
	}
}

func TestMidB04RevengeOfTheDrownedTucksByOwnersChoiceThenMakesYourZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Revenge of the Drowned", "Instant", midB04Revenge, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Error("the caster was asked; the owner chooses")
	}
	if n := midB04DecayedZombies(t, g, me.ID); n != 0 {
		t.Errorf("the token came before the owner answered (%d)", n)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{theirs}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target is still on the battlefield")
	}
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Errorf("decayed Zombies = %d, want 1, under the caster's control", n)
	}
}

func TestMidB04RottenReunionExilesACardAndMakesAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.B17GraveyardCard(opp, "Their Card", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Rotten Reunion", "Instant", midB04RottenReunion, effectstest.CardTarget(card))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(card) {
		t.Error("the targeted card is not exiled")
	}
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Errorf("decayed Zombies = %d, want 1", n)
	}
}

func TestMidB04RottenReunionWithNoTargetStillMakesAZombieAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Rotten Reunion", "Instant", midB04RottenReunion, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Fatalf("no target: Zombies = %d, want 1", n)
	}
	id := effectstest.SeedGraveyardCard(t, g, "Rotten Reunion", "Instant", midB04RottenReunion)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 2 {
		t.Errorf("after flashback Zombies = %d, want 2", n)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back instant should be exiled")
	}
}

func TestMidB04NoWayOutDiscardsTwoThenMakesAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushCatalogHandCard(opp, "A", "Instant", "")
	b := effectstest.PushCatalogHandCard(opp, "B", "Instant", "")
	effectstest.CastCatalogSpell(t, g, "No Way Out", "Sorcery", midB04NoWayOut, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 0 {
		t.Errorf("the token came before the discard (%d)", n)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB04DecayedZombies(t, g, me.ID); n != 1 {
		t.Errorf("decayed Zombies = %d, want 1", n)
	}
}

func TestMidB04NoWayOutRefusesYourselfAsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	err := effectstest.CastCatalogSpellErrStrict(t, g, "No Way Out", "Sorcery", midB04NoWayOut, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err == nil {
		t.Fatal("No Way Out targeted its own controller")
	}
}

func TestMidB04BladestitchedSkaabPumpsOtherZombiesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skaab := effectstest.B12Push(g, me.ID, "Bladestitched Skaab", "Creature — Zombie Soldier", midB04BladestitchedSk, 2, 3)
	mine := midB04Zombie(g, me.ID, "My Zombie")
	theirs := midB04Zombie(g, opp.ID, "Their Zombie")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for id, want := range map[uuid.UUID]int{skaab: 2, mine: 3, theirs: 2, bear: 2} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("%s power %d, want %d", id, got, want)
		}
	}
	if _, tough := effectstest.EffectivePTOf(t, g, mine); tough != 2 {
		t.Errorf("toughness %d, want unchanged 2", tough)
	}
}

func TestMidB04CrawlFromTheCellarReturnsACreatureAndGrowsAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	zombie := midB04Zombie(g, me.ID, "My Zombie")
	effectstest.CastCatalogSpell(t, g, "Crawl from the Cellar", "Sorcery", midB04Crawl, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead}, {Kind: game.TargetCard, ID: zombie},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card is not back in hand")
	}
	if c, ok := effectstest.BattlefieldCard(g, zombie); !ok || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Zombie has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestMidB04CrawlFromTheCellarWithNoZombieStillReturnsAndRefusesANonZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Crawl from the Cellar", "Sorcery", midB04Crawl, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead}, {Kind: game.TargetCard, ID: bear},
	})
	if err == nil {
		t.Fatal("a non-Zombie was accepted for the counter")
	}
	effectstest.CastCatalogSpell(t, g, "Crawl from the Cellar", "Sorcery", midB04Crawl, effectstest.CardTarget(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("with no Zombie chosen the creature card is still returned")
	}
}

func TestMidB04TombTyrantAnthemAndReanimation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tyrant := effectstest.B12Push(g, me.ID, "Tomb Tyrant", "Creature — Zombie Noble", midB04TombTyrant, 3, 3)
	mine := midB04Zombie(g, me.ID, "My Zombie")
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("other Zombie is %d/%d, want 3/3", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, tyrant); p != 3 {
		t.Errorf("the Tyrant pumped itself: %d", p)
	}
	z1 := effectstest.B17GraveyardCard(me, "Dead Zombie 1", "Creature — Zombie", "{B}")
	z2 := effectstest.B17GraveyardCard(me, "Dead Zombie 2", "Creature — Zombie", "{B}")
	z3 := effectstest.B17GraveyardCard(me, "Dead Zombie 3", "Creature — Zombie", "{B}")
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.FillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, tyrant, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	back := 0
	for _, id := range []uuid.UUID{z1, z2, z3} {
		if g.Battlefield.Contains(id) {
			back++
		}
	}
	if back != 1 {
		t.Errorf("%d Zombie cards returned, want exactly 1", back)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !effectstest.B16Tapped(t, g, tyrant) {
		t.Error("the Tyrant was not tapped as the cost")
	}
}

func TestMidB04TombTyrantNeedsThreeZombieCardsAndYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyrant := effectstest.B12Push(g, me.ID, "Tomb Tyrant", "Creature — Zombie Noble", midB04TombTyrant, 3, 3)
	effectstest.B17GraveyardCard(me, "Dead Zombie 1", "Creature — Zombie", "{B}")
	effectstest.B17GraveyardCard(me, "Dead Zombie 2", "Creature — Zombie", "{B}")
	effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{B}")
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.FillPoolColored(me, "B", 3)
	params := game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}
	if err := g.ActivateCatalogAbility(me.ID, tyrant, 0, params); err == nil {
		t.Fatal("activated with only two Zombie cards in the graveyard")
	}
	effectstest.B17GraveyardCard(me, "Dead Zombie 3", "Creature — Zombie", "{B}")
	// Another player's turn: the Tyrant's controller may not activate.
	g.Turn.ActiveSeat = 1
	if err := g.ActivateCatalogAbility(me.ID, tyrant, 0, params); err == nil {
		t.Fatalf("activated during %s's turn", opp.Name)
	}
}

func TestMidB04EmptyTheLaboratorySacrificesXZombiesAndRevealsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	z1 := midB04Zombie(g, me.ID, "Zombie 1")
	z2 := midB04Zombie(g, me.ID, "Zombie 2")
	z3 := midB04Zombie(g, me.ID, "Zombie 3")
	effectstest.SeedLibraryTop(me, "Filler Top", "Creature — Bear", "{G}")
	hit2 := effectstest.SeedLibraryTop(me, "Library Zombie B", "Creature — Zombie", "{B}")
	effectstest.SeedLibraryTop(me, "Middle", "Sorcery", "{R}")
	hit1 := effectstest.SeedLibraryTop(me, "Library Zombie A", "Creature — Zombie", "{B}")
	effectstest.PaCast(t, g, "Empty the Laboratory", "Sorcery", "{X}{U}{U}", midB04EmptyTheLab, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("sacrifice bounds %d..%d, want exactly 2", c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{z1, z3}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(z1) || g.Battlefield.Contains(z3) || !g.Battlefield.Contains(z2) {
		t.Error("exactly the two chosen Zombies are sacrificed")
	}
	if !g.Battlefield.Contains(hit1) || !g.Battlefield.Contains(hit2) {
		t.Error("the two Zombie creature cards revealed did not enter the battlefield")
	}
	foundMiddle := false
	for i := 0; i < 3 && i < len(me.Library.Cards); i++ {
		if me.Library.Cards[i].Name == "Middle" {
			foundMiddle = true
		}
	}
	if !foundMiddle {
		t.Error("the non-Zombie revealed card is not on the bottom of the library")
	}
}

func TestMidB04EmptyTheLaboratoryWithFewerZombiesThanXSacrificesThemAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	z1 := midB04Zombie(g, me.ID, "Zombie 1")
	effectstest.SeedLibraryTop(me, "Library Zombie", "Creature — Zombie", "{B}")
	effectstest.SeedLibraryTop(me, "Library Zombie 2", "Creature — Zombie", "{B}")
	effectstest.PaCast(t, g, "Empty the Laboratory", "Sorcery", "{X}{U}{U}", midB04EmptyTheLab, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("sacrifice bounds %d..%d, want 1..1 (the only Zombie)", c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{z1}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(z1) {
		t.Error("the Zombie was not sacrificed")
	}
	entered := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name != "Zombie 1" && c.TypeLine == "Creature — Zombie" {
			entered++
		}
	}
	if entered != 1 {
		t.Errorf("%d library Zombies entered, want exactly 1 (the number sacrificed)", entered)
	}
}
