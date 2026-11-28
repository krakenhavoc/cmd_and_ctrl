package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b671_cards_test.go — Prophecy: Other creatures, batch 671 (#5617). Each
// test plays the real catalog entry and asserts what its text says.

const (
	b671Homunculus = "1af26f29-627c-48e5-a19d-deb43a11a846"
	b671Squad      = "2531eb1a-ae24-450a-b8e8-dbbd65eaad64"
	b671Serpent    = "6e0deb42-f876-4fb8-9536-6d9a42bd6e5e"
	b671Veteran    = "77a963a3-473a-4c34-bc46-8ef016a92c0c"
	b671Darba      = "c64c0cb5-5a8c-4a2f-a0fc-8b080aaf90bb"
	b671Efreet     = "e5cd1024-ec1a-4df9-8bc6-ad2f3f443d06"
	b671Thresher   = "e88915fd-7867-45e6-b4b9-ca0af1823c0d"
	b671Branded    = "e975a8b3-f817-42f3-84da-b03ebf95cf25"
	b671Mungha     = "fe6ff20b-5e4d-44d5-b011-c6e512169294"
)

func TestB671CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		b671Homunculus: "Hazy Homunculus", b671Squad: "Reveille Squad", b671Serpent: "Shrouded Serpent",
		b671Veteran: "Veteran Brawlers", b671Darba: "Darba", b671Efreet: "Fickle Efreet",
		b671Thresher: "Thresher Beast", b671Branded: "Branded Brawlers", b671Mungha: "Mungha Wurm",
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

func b671Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = true })
}

func TestB671HazyHomunculusCantBeBlockedWhileTheyControlAnUntappedLand(t *testing.T) {
	for _, untapped := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		hom := effectstest.PushCatalogPermanent(g, me.ID, "Hazy Homunculus", "Creature — Homunculus Illusion", b671Homunculus, false)
		wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
		land := effectstest.GaPut(g, opp.ID, "Island", "Basic Land — Island", "")
		if !untapped {
			b671Tap(g, land)
		}
		effectstest.BrAttack(t, g, hom)
		err := effectstest.ReqBlock(t, g, hom, wall)
		if untapped && err == nil {
			t.Error("blocked a Homunculus while the defender controls an untapped land")
		}
		if !untapped && err != nil {
			t.Errorf("could not block with only tapped lands: %v", err)
		}
	}
}

func TestB671BrawlersCantAttackOrBlockWhileTheLandIsUntapped(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{{"Veteran Brawlers", b671Veteran}, {"Branded Brawlers", b671Branded}} {
		// Attack: the defender's untapped land stops it.
		for _, untapped := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			brawler := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Creature — Human Soldier", tc.oracle, false)
			land := effectstest.GaPut(g, opp.ID, "Island", "Basic Land — Island", "")
			if !untapped {
				b671Tap(g, land)
			}
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			err := g.DeclareAttacker(brawler, opp.ID)
			if untapped && err == nil {
				t.Errorf("%s attacked a player with an untapped land", tc.name)
			}
			if !untapped && err != nil {
				t.Errorf("%s could not attack a player with only tapped lands: %v", tc.name, err)
			}
		}
		// Block: its own controller's untapped land stops it.
		for _, untapped := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			attacker := effectstest.ReqCreature(g, me.ID, "Ogre", "", 3, 3)
			brawler := effectstest.PushCatalogPermanent(g, opp.ID, tc.name, "Creature — Human Soldier", tc.oracle, false)
			land := effectstest.GaPut(g, opp.ID, "Mountain", "Basic Land — Mountain", "")
			if !untapped {
				b671Tap(g, land)
			}
			effectstest.BrAttack(t, g, attacker)
			err := effectstest.ReqBlock(t, g, attacker, brawler)
			if untapped && err == nil {
				t.Errorf("%s blocked while its controller controls an untapped land", tc.name)
			}
			if !untapped && err != nil {
				t.Errorf("%s could not block with only tapped lands: %v", tc.name, err)
			}
		}
	}
}

func TestB671ReveilleSquadMayUntapEverythingWhenYouAreAttacked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		squadTap bool
		defender int // seat attacked
		wantAsk  bool
	}{
		{"attacked and untapped", false, 1, true},
		{"attacked but tapped", true, 1, false},
		{"another player attacked", false, 2, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
		squad := effectstest.PushCatalogPermanent(g, opp.ID, "Reveille Squad", "Creature — Human Rebel", b671Squad, false)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		b671Tap(g, bear)
		if tc.squadTap {
			b671Tap(g, squad)
		}
		atk := effectstest.ReqCreature(g, me.ID, "Ogre", "", 3, 3)
		effectstest.DeclareAttack(t, g, g.Seats[tc.defender].ID, atk)
		_ = other
		effectstest.PassPriorityAroundTable(t, g)
		asked := effectstest.LatestTriggerPrompt(g, opp.ID) != nil
		if asked != tc.wantAsk {
			t.Fatalf("%s: asked %v, want %v", tc.name, asked, tc.wantAsk)
		}
		if asked {
			effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.PermanentTapped(g, bear) {
				t.Errorf("%s: the Bear is still tapped", tc.name)
			}
		} else if !tc.squadTap && !effectstest.PermanentTapped(g, bear) {
			t.Errorf("%s: the Bear untapped without a trigger", tc.name)
		}
	}
}

func TestB671ShroudedSerpentCantBeBlockedUnlessTheDefenderPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		serpent := effectstest.PushCatalogPermanent(g, me.ID, "Shrouded Serpent", "Creature — Serpent", b671Serpent, false)
		wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 6)
		effectstest.DeclareAttack(t, g, opp.ID, serpent)
		effectstest.PassUntilTaxed(t, g, opp.ID)
		if pay {
			for i := 0; i < 4; i++ {
				opp.ManaPool.AddMana(game.ManaToken{Color: "C"})
			}
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := effectstest.ReqBlock(t, g, serpent, wall)
		if !pay && err == nil {
			t.Error("the Serpent was blocked though the defender declined to pay")
		}
		if pay && err != nil {
			t.Errorf("the defender paid but could not block: %v", err)
		}
	}
}

func b671DarbaUpkeep(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	darba := effectstest.B12Push(g, me.ID, "Darba", "Creature — Bird Beast", b671Darba, 5, 4)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	return g, me, darba
}

func TestB671DarbaIsSacrificedUnlessYouPayGG(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g, me, darba := b671DarbaUpkeep(t)
		if pay {
			effectstest.B06AddMana(me, "G", "G")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("no pay prompt for the upkeep trigger")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(darba) != pay {
			t.Errorf("pay=%v: Darba on the battlefield = %v", pay, g.Battlefield.Contains(darba))
		}
	}
}

func TestB671ThresherBeastMakesTheDefenderSacrificeALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	beast := effectstest.PushCatalogPermanent(g, me.ID, "Thresher Beast", "Creature — Beast", b671Thresher, false)
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 6)
	a := effectstest.GaPut(g, opp.ID, "Island", "Basic Land — Island", "")
	b := effectstest.GaPut(g, opp.ID, "Swamp", "Basic Land — Swamp", "")
	mine := effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	theirs := effectstest.GaPut(g, other.ID, "Plains", "Basic Land — Plains", "")
	effectstest.BrAttack(t, g, beast)
	if err := effectstest.ReqBlock(t, g, beast, wall); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the defending player was not asked to sacrifice a land")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, b)
	if g.Battlefield.Contains(b) || !g.Battlefield.Contains(a) || !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("want only the chosen land of the defending player sacrificed")
	}
}

func TestB671MunghaWurmCapsOnlyItsControllersLandsUntapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, other := g.Seats[1], g.Seats[2]
	forest := effectstest.PushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := effectstest.PushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	bear := effectstest.PushTappedForTest(g, owner.ID, "Grizzly Bears", "", "Creature — Bear")
	effectstest.PushPermanentForTest(g, owner.ID, "Mungha Wurm", b671Mungha, "Creature — Wurm")
	theirs1 := effectstest.PushTappedForTest(g, other.ID, "Plains", "", "Basic Land — Plains")
	theirs2 := effectstest.PushTappedForTest(g, other.ID, "Swamp", "", "Basic Land — Swamp")

	choice := effectstest.AdvanceToUntapChoiceOf(t, g, 1)
	if choice.ChooseMin != 1 || choice.ChooseMax != 1 || len(choice.ChooseCards) != 2 {
		t.Fatalf("bounds %d..%d over %d permanents, want 1..1 over the two lands", choice.ChooseMin, choice.ChooseMax, len(choice.ChooseCards))
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("answer the Wurm's cap: %v", err)
	}
	if effectstest.TappedForTest(t, g, forest) || !effectstest.TappedForTest(t, g, island) || effectstest.TappedForTest(t, g, bear) {
		t.Error("want exactly the chosen land and the Bear untapped")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	if effectstest.TappedForTest(t, g, theirs1) || effectstest.TappedForTest(t, g, theirs2) {
		t.Error("another player's lands were capped by the Wurm")
	}
}

func TestB671FickleEfreetFlipsAtEndOfCombat(t *testing.T) {
	for _, blocks := range []bool{false, true} {
		sawKeep, sawLose := false, false
		for i := 0; i < 80 && !(sawKeep && sawLose); i++ {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			var efreet, attacker uuid.UUID
			caller := me
			if blocks {
				attacker = effectstest.ReqCreature(g, me.ID, "Ogre", "", 1, 5)
				efreet = effectstest.B12Push(g, opp.ID, "Fickle Efreet", "Creature — Efreet", b671Efreet, 5, 2)
				caller = opp
				effectstest.BrAttack(t, g, attacker)
				if err := effectstest.ReqBlock(t, g, attacker, efreet); err != nil {
					t.Fatal(err)
				}
				effectstest.LockInBlocks(t, g)
			} else {
				efreet = effectstest.B12Push(g, me.ID, "Fickle Efreet", "Creature — Efreet", b671Efreet, 5, 2)
				effectstest.DeclareAttack(t, g, opp.ID, efreet)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceTo(t, g, game.StepEndCombat)
			effectstest.PassPriorityAroundTable(t, g)
			flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
			if flip == nil {
				t.Fatalf("blocks=%v: no coin call at end of combat", blocks)
			}
			if err := g.ResolveCoinCall(flip.ID, caller.ID, "heads"); err != nil {
				t.Fatalf("coin: %v", err)
			}
			flips := effectstest.RandomEvents(g, game.EventFlipCoin)
			won := len(flips) > 0 && flips[len(flips)-1].Won
			if won {
				sawKeep = true
			} else {
				sawLose = true
				other := g.Seats[2]
				effectstest.AnswerChoosePlayer(t, g, caller.ID, other)
			}
			effectstest.PassPriorityAroundTable(t, g)
			c, ok := effectstest.BattlefieldCard(g, efreet)
			if !ok {
				t.Fatal("the Efreet left the battlefield")
			}
			if won && c.Controller != caller.ID {
				t.Errorf("blocks=%v: won the flip but control changed to %s", blocks, c.Controller)
			}
			if !won && c.Controller != g.Seats[2].ID {
				t.Errorf("blocks=%v: lost the flip but the chosen opponent does not control it (%s)", blocks, c.Controller)
			}
		}
		if !sawKeep || !sawLose {
			t.Fatalf("blocks=%v: never saw both outcomes (keep %v, lose %v)", blocks, sawKeep, sawLose)
		}
	}
}
