package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b09_cards_test.go — bonus-sheet batch 9 (#5929): The List cards.

const (
	bnsB09JaceOracle     = "5277a68e-1ceb-4b40-8e10-3d0564f5d7be"
	bnsB09KuldothaOracle = "f3e28778-0149-4dbd-badb-ba5aeacacd76"
	bnsB09LeoninOracle   = "8e720216-a980-4a34-8aef-62ac3ef15a6e"
	bnsB09MagmawOracle   = "bc891258-2f22-4248-9972-4f41905744c6"
	bnsB09PsycheOracle   = "6f98c1eb-0f52-4d21-a266-e8d52f98c1b3"
	bnsB09KeyOracle      = "6a1e0dfd-2dd1-4372-966e-e28b48ae0115"
	bnsB09ShardOracle    = "c6c2f370-fc33-4907-975c-a07af3dc664d"
	bnsB09TerrainOracle  = "eced76b8-b29a-4c89-b77a-cbae99f56cb7"
)

func TestBnsB09EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB09JaceOracle: "Jace, Wielder of Mysteries", bnsB09KuldothaOracle: "Kuldotha Rebirth",
		bnsB09LeoninOracle: "Leonin Relic-Warder", bnsB09MagmawOracle: "Magmaw",
		bnsB09PsycheOracle: "Molten Psyche", bnsB09KeyOracle: "Mystery Key",
		bnsB09ShardOracle: "Shard of Broken Glass", bnsB09TerrainOracle: "Treacherous Terrain",
	} {
		cmdrB21AssertFull(t, oracle, name)
	}
}

func bnsB09Artifact(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", Owner: owner, Controller: owner,
	})
}

func TestBnsB09KuldothaRebirthSacrificesAnArtifactForThreeGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	if err := castCatalogSpellErrStrict(t, g, "Kuldotha Rebirth", "Sorcery", bnsB09KuldothaOracle, nil); err == nil {
		t.Fatal("cast without an artifact to sacrifice was legal")
	}
	rock := bnsB09Artifact(g, me.ID, "Rock")
	castWithSacrifice(t, g, "Kuldotha Rebirth", "Sorcery", bnsB09KuldothaOracle, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := len(battlefieldIDsNamed(g, "Goblin")); got != 3 {
		t.Errorf("%d Goblins, want 3", got)
	}
}

func TestBnsB09LeoninRelicWarderExilesAnArtifactAndReturnsItWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := bnsB09Artifact(g, opp.ID, "Their Rock")
	warder := castAndResolveCreature(t, g, "Leonin Relic-Warder", "Creature — Cat Cleric", bnsB09LeoninOracle)
	pickCard(t, g, me.ID, rock)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Fatal("the artifact was not exiled")
	}
	killCreature(t, g, me.ID, warder)
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Their Rock")) != 1 {
		t.Error("the artifact did not return when the Warder left")
	}
}

func TestBnsB09LeoninRelicWarderMayDeclineTheExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := bnsB09Artifact(g, opp.ID, "Their Rock")
	castAndResolveCreature(t, g, "Leonin Relic-Warder", "Creature — Cat Cleric", bnsB09LeoninOracle)
	pickCard(t, g, me.ID, rock)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(rock) || !g.Battlefield.Contains(rock) {
		t.Error("declining still exiled the artifact")
	}
}

func TestBnsB09MagmawSacrificesANonlandPermanentToPing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mag := pushCatalogPermanent(g, me.ID, "Magmaw", "Creature — Elemental", bnsB09MagmawOracle, false)
	land := pushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	rock := bnsB09Artifact(g, me.ID, "Rock")
	apaMana(me, "C", "C")
	tgt := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, mag, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}, Targets: tgt}); err == nil {
		t.Fatal("sacrificing a land was legal")
	}
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, mag, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}, Targets: tgt}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if opp.Life != before-1 {
		t.Errorf("life %d, want %d", opp.Life, before-1)
	}
}

func TestBnsB09MoltenPsycheRedrawsAndPunishesWithMetalcraft(t *testing.T) {
	for _, tc := range []struct {
		name      string
		artifacts int
		damage    bool
	}{{"two artifacts", 2, false}, {"three artifacts", 3, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			for i := 0; i < tc.artifacts; i++ {
				bnsB09Artifact(g, me.ID, "Rock")
			}
			advanceToMain(t, g)
			hands := map[uuid.UUID]int{}
			lives := map[uuid.UUID]int{}
			for _, p := range g.Seats {
				hands[p.ID] = p.Hand.Size()
				lives[p.ID] = p.Life
			}
			castCatalogSpell(t, g, "Molten Psyche", "Sorcery", bnsB09PsycheOracle, nil)
			passPriorityAroundTable(t, g)
			for _, p := range g.Seats {
				if p.Hand.Size() != hands[p.ID] {
					t.Errorf("%s hand %d, want %d", p.Name, p.Hand.Size(), hands[p.ID])
				}
				want := lives[p.ID]
				if tc.damage && p.ID != me.ID {
					want -= g.TurnTallyFor(p.ID).CardsDrawn
					if g.TurnTallyFor(p.ID).CardsDrawn < hands[p.ID] {
						t.Errorf("%s drew %d this turn, want at least %d", p.Name, g.TurnTallyFor(p.ID).CardsDrawn, hands[p.ID])
					}
				}
				if p.Life != want {
					t.Errorf("%s life %d, want %d", p.Name, p.Life, want)
				}
			}
		})
	}
}

func TestBnsB09MysteryKeyIsSacrificedForThreeCardsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	key := seedEquipment(g, me.ID, "Mystery Key", bnsB09KeyOracle)
	equipTo(t, g, me.ID, key, bear)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(key) {
		t.Error("the Key was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

func TestBnsB09MysteryKeyDoesNothingWhenUnattachedCreatureHits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	key := seedEquipment(g, me.ID, "Mystery Key", bnsB09KeyOracle)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(key) || me.Hand.Size() != hand {
		t.Error("an unequipped creature's damage triggered the Key")
	}
}

func TestBnsB09ShardOfBrokenGlassPumpsAndMayMillTwoOnAttack(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		bear := seedBear(g, me.ID)
		shard := seedEquipment(g, me.ID, "Shard of Broken Glass", bnsB09ShardOracle)
		equipTo(t, g, me.ID, shard, bear)
		if got := effectivePower(t, g, bear); got != 3 {
			t.Errorf("power %d, want 3", got)
		}
		lib := me.Library.Size()
		declareAttack(t, g, opp.ID, bear)
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, yes)
		want := lib
		if yes {
			want -= 2
		}
		if me.Library.Size() != want {
			t.Errorf("yes=%v: library %d, want %d", yes, me.Library.Size(), want)
		}
	}
}

func TestBnsB09TreacherousTerrainDealsEachOpponentTheirLandCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	counts := []int{0, 2, 3, 5}
	for i, p := range g.Seats {
		for j := 0; j < counts[i]; j++ {
			pushCatalogPermanent(g, p.ID, "Forest", "Basic Land — Forest", "", false)
		}
	}
	pushCatalogPermanent(g, g.Seats[1].ID, "Rock", "Artifact", "", false)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Treacherous Terrain", "Sorcery", bnsB09TerrainOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= counts[i]
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, want)
		}
	}
}

func TestBnsB09JaceWielderOfMysteries(t *testing.T) {
	t.Run("plus one mills the target and draws", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		jace := s82Walker(g, me.ID, "Jace, Wielder of Mysteries", bnsB09JaceOracle, 4)
		advanceToMain(t, g)
		hand, grave := me.Hand.Size(), opp.Graveyard.Size()
		s82Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
		passPriorityAroundTable(t, g)
		if got := opp.Graveyard.Size() - grave; got != 2 {
			t.Errorf("milled %d, want 2", got)
		}
		if got := me.Hand.Size() - hand; got != 1 {
			t.Errorf("drew %d, want 1", got)
		}
		if got := loyaltyOf(g, jace); got != 5 {
			t.Errorf("loyalty %d, want 5", got)
		}
	})
	t.Run("drawing from an empty library wins", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		jace := s82Walker(g, me.ID, "Jace, Wielder of Mysteries", bnsB09JaceOracle, 4)
		advanceToMain(t, g)
		g.WithWriteLock(func() {
			for me.Library.Size() > 0 {
				me.Library.Cards = me.Library.Cards[:0]
			}
		})
		s82Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
		passPriorityAroundTable(t, g)
		if !gameHasWinner(g) {
			t.Error("drawing with an empty library did not win the game")
		}
	})
	t.Run("minus eight draws seven and wins on an empty library", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		jace := s82Walker(g, me.ID, "Jace, Wielder of Mysteries", bnsB09JaceOracle, 8)
		advanceToMain(t, g)
		g.WithWriteLock(func() { me.Library.Cards = me.Library.Cards[:7] })
		hand := me.Hand.Size()
		s82Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != 7 {
			t.Errorf("drew %d, want 7", got)
		}
		if !gameHasWinner(g) {
			t.Error("an empty library after the draw did not win the game")
		}
	})
	t.Run("minus eight with cards left does not win", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		jace := s82Walker(g, me.ID, "Jace, Wielder of Mysteries", bnsB09JaceOracle, 8)
		advanceToMain(t, g)
		s82Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
		if gameHasWinner(g) {
			t.Error("won with cards left in the library")
		}
	})
}

func gameHasWinner(g *game.Game) bool {
	_, ok := g.WinnerSeat()
	return ok
}
