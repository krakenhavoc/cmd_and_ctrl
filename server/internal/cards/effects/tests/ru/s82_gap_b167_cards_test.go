package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB167Taranika   = "1e3c1920-7bf5-44f9-b1ff-b17d6420e69b"
	gapB167TaskForce  = "b80415cd-5887-4146-852e-73119fcd3ec6"
	gapB167Wispweaver = "f8a88521-80e5-419b-ab4e-f4b3aba8afa0"
	gapB167Feast      = "8ce9afd9-0c96-42a2-a534-d92ae04f0151"
	gapB167Festering  = "0581b047-ffa1-42cc-a175-b7bf928e279e"
	gapB167GrafHarv   = "d3ba6922-c2f7-45ab-87a3-d4bbd770d1ba"
	gapB167MireBlight = "fcfbe3b8-a955-45d8-9329-3af565a3b53c"
	gapB167FeedPack   = "48981194-6465-4900-8cea-206f8ab5d00e"
)

func TestGapB167EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB167Taranika: "Taranika, Akroan Veteran", gapB167TaskForce: "Task Force",
		gapB167Wispweaver: "Wispweaver Angel", gapB167Feast: "Feast on the Fallen",
		gapB167Festering: "Festering Evil", gapB167GrafHarv: "Graf Harvest",
		gapB167MireBlight: "Mire Blight", gapB167FeedPack: "Feed the Pack",
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

func TestGapB167TaranikaUntapsAnotherCreatureAndMakesItAFourByFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	tar := effectstest.B12Push(g, me.ID, "Taranika, Akroan Veteran", "Legendary Creature — Human Soldier", gapB167Taranika, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tar, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(tar)); err == nil {
		t.Error("Taranika could target itself")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, bear) {
		t.Error("the bear is still tapped")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the bear is not indestructible")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("indestructible outlasted the turn")
	}
}

func TestGapB167TaskForceGetsPlusZeroPlusThreeWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	tf := effectstest.B12Push(g, opp.ID, "Task Force", "Creature — Human Rebel", gapB167TaskForce, 1, 3)
	_ = me
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardRefs(tf))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tf) {
		t.Fatal("Task Force died to a Lightning Bolt it should have outgrown")
	}
	if _, tough := effectstest.PtOf(t, g, tf); tough != 6 {
		t.Errorf("toughness %d, want 6", tough)
	}
}

func TestGapB167WispweaverAngelMayBlinkAnotherCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		card := game.NewCard("My Bear", me.ID)
		card.TypeLine, card.Power, card.Toughness = "Creature — Bear", 2, 2
		me.Hand.PushTop(card)
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, card.InstanceID); err != nil {
			t.Fatalf("enter: %v", err)
		}
		bear := card.InstanceID
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
		angel := effectstest.CastHoldCreature(t, g, "Wispweaver Angel", "Creature — Angel", gapB167Wispweaver, 4, 4)
		effectstest.B04WaitForPick(t, g, me.ID)
		if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(angel)); err == nil {
			t.Error("the Angel could target itself")
		}
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		// A blinked creature is a new object (CR 400.7): find it by name.
		n, counters := 0, 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Name == "My Bear" {
				n++
				counters = c.Counters[game.CounterPlusOne]
			}
		}
		if n != 1 {
			t.Fatalf("yes=%v: %d bears on the battlefield, want 1", yes, n)
		}
		want := 1
		if yes {
			want = 0 // it returned as a new object
		}
		if counters != want {
			t.Errorf("yes=%v: %d counters, want %d", yes, counters, want)
		}
	}
}

func TestGapB167GrafHarvestGivesZombiesMenaceAndMakesZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Graf Harvest", "Enchantment", gapB167GrafHarv, 0, 0)
	z := effectstest.B12Creature(g, me.ID, "Zed", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Zed", "Creature — Zombie", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, z, "menace") {
		t.Error("my Zombie has no menace")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "menace") || effectstest.HasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("a non-Zombie or an opponent's Zombie got menace")
	}
	harvest := effectstest.B12Permanent(g, me.ID, "Graf Harvest", "Enchantment")
	_ = harvest
}

func TestGapB167GrafHarvestExilesACreatureCardForATwoTwoZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	harvest := effectstest.PushCatalogPermanent(g, me.ID, "Graf Harvest", "Enchantment", gapB167GrafHarv, false)
	spell := effectstest.PushCatalogGraveyardCard(me, "Shock", "Instant", "", 0, 0)
	bear := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "BBBB")
	if err := g.ActivateCatalogAbility(me.ID, harvest, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Error("exiled a noncreature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, harvest, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the creature card was not exiled")
	}
	toks := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Zombie" && c.IsToken() {
			toks++
			if c.Power != 2 || c.Toughness != 2 {
				t.Errorf("token is %d/%d", c.Power, c.Toughness)
			}
		}
	}
	if toks != 1 {
		t.Errorf("%d Zombie tokens, want 1", toks)
	}
}

func TestGapB167FesteringEvilDamagesEverythingAtYourUpkeepAndWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	evil := effectstest.B12Permanent(g, me.ID, "Festering Evil", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == evil {
				g.Battlefield.Cards[i].OracleID = gapB167Festering
			}
		}
	})
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	myLife, oppLife := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Battlefield.Contains(big) {
		t.Errorf("upkeep damage: small alive=%v big alive=%v", g.Battlefield.Contains(small), g.Battlefield.Contains(big))
	}
	if me.Life != myLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife-1, oppLife-1)
	}
	if err := effectstest.BroB01Activate(t, g, me, evil, 0, "{B}{B}"); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(evil) || g.Battlefield.Contains(big) {
		t.Error("the Evil or the 3/3 survived the 3 damage")
	}
	if me.Life != myLife-4 || opp.Life != oppLife-4 {
		t.Errorf("life after sacrifice %d/%d, want %d/%d", me.Life, opp.Life, myLife-4, oppLife-4)
	}
}

func TestGapB167MireBlightDestroysTheEnchantedCreatureWhenItIsDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Big Host", "Creature — Giant", 5, 5)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Giant", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Mire Blight", "Enchantment — Aura", gapB167MireBlight, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DamageAndSettle(t, g, other, other, 1)
	if !g.Battlefield.Contains(host) {
		t.Fatal("the host died to another creature being damaged")
	}
	effectstest.DamageAndSettle(t, g, other, host, 1)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature survived being dealt 1 damage")
	}
	_ = me
}

func TestGapB167FeedThePackMakesWolvesEqualToTheSacrificedToughness(t *testing.T) {
	for _, tc := range []struct {
		name  string
		yes   bool
		token bool
		want  int
	}{{"nontoken", true, false, 4}, {"decline", false, false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.B12Permanent(g, me.ID, "Feed the Pack", "Enchantment")
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].Name == "Feed the Pack" {
						g.Battlefield.Cards[i].OracleID = gapB167FeedPack
					}
				}
			})
			wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
			tok := effectstest.B12Creature(g, me.ID, "Token Ox", "Token Creature — Ox", 9, 9)
			effectstest.AdvanceToEndStep(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if tc.yes {
				effectstest.AnswerOwnPermanents(t, g, me.ID, wall)
			} else {
				effectstest.AnswerOwnPermanents(t, g, me.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
			wolves := 0
			for _, c := range g.Battlefield.Cards {
				if c.Controller == me.ID && c.Name == "Wolf" && c.IsToken() {
					wolves++
				}
			}
			if wolves != tc.want {
				t.Errorf("%d Wolves, want %d", wolves, tc.want)
			}
			if tc.yes && g.Battlefield.Contains(wall) {
				t.Error("the Wall was not sacrificed")
			}
			if !g.Battlefield.Contains(tok) {
				t.Error("the token was sacrificed")
			}
		})
	}
}

func TestGapB167FeastOnTheFallenCountsLifeLostLastTurnByAnOpponent(t *testing.T) {
	for _, tc := range []struct {
		name              string
		oppLoses, meLoses bool
		want              int
	}{{"an opponent lost life", true, false, 1}, {"only I lost life", false, true, 0}, {"nobody lost life", false, false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Permanent(g, me.ID, "Feast on the Fallen", "Enchantment")
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].Name == "Feast on the Fallen" {
						g.Battlefield.Cards[i].OracleID = gapB167Feast
					}
				}
			})
			bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			// The turn in progress is "last turn" at the next upkeep.
			if tc.oppLoses {
				effectstest.GainLifeNow(t, g, bear, opp, -2)
			}
			if tc.meLoses {
				effectstest.GainLifeNow(t, g, bear, me, -2)
			}
			effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
			if tc.want == 1 {
				effectstest.B04WaitForPick(t, g, me.ID)
				effectstest.PickTriggerTarget(t, g, me.ID, bear)
			} else if effectstest.LatestPickTarget(g, me.ID) != nil {
				t.Fatal("the trigger went on the stack with no opponent having lost life last turn")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d counters, want %d", got, tc.want)
			}
		})
	}
}
