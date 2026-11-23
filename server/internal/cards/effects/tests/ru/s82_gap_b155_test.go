package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b155_test.go — GAP batch 155 (Jumpstart 2022 red creatures,
// #5948), played through the catalog and the real stack.

const (
	gapB155Pyreling      = "5291deb9-0b89-42c2-bde6-26f88931c3f4"
	gapB155Coalborn      = "49f1dd11-aa10-442d-aaca-dd94d1ca108a"
	gapB155Stinger       = "edd1727e-3a6e-4d0e-b7ae-fae2108d2147"
	gapB155Fireslinger   = "0e5a5cee-d068-42cf-bf58-8e1360720be0"
	gapB155LavaSerpent   = "4d94d807-97d9-4cdf-a63c-d8b82c2acc6e"
	gapB155LightningBolt = "4457ed35-7c10-48c8-9776-456485fdf070"
	gapB155Bear          = "Creature — Bear"
)

func TestGapB155FireslingerHitsTheTargetAndYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	slinger := effectstest.B12Push(g, me.ID, "Fireslinger", "Creature — Human Wizard", gapB155Fireslinger, 1, 1)
	myLife, oppLife := me.Life, opp.Life
	effectstest.ApaActivate(t, g, me, slinger, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != oppLife-1 || me.Life != myLife-1 {
		t.Errorf("target at %d (want %d), you at %d (want %d)", opp.Life, oppLife-1, me.Life, myLife-1)
	}
	// A tap ability: tapped, it cannot do it again.
	if err := g.ActivateCatalogAbility(me.ID, slinger, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a tapped Fireslinger activated again")
	}
}

func TestGapB155PyrelingGrowsFromNoncombatDamageToAnOpponentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pyreling := effectstest.B12Push(g, me.ID, "Chandra's Pyreling", "Creature — Elemental Lizard", gapB155Pyreling, 1, 3)
	// Damage to me: nothing.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB155LightningBolt, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, pyreling); live.CurrentPower() != 1 || effectstest.P1g2HasKeyword(g, pyreling, "double strike") {
		t.Fatalf("damage to its controller pumped it: %d/%d", live.CurrentPower(), live.CurrentToughness())
	}
	// Noncombat damage to an opponent, twice: two triggers.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB155LightningBolt, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, pyreling); live.CurrentPower() != 2 || !effectstest.P1g2HasKeyword(g, pyreling, "double strike") {
		t.Fatalf("after one Bolt it is %d/%d (double strike %v), want 2/3 with double strike",
			live.CurrentPower(), live.CurrentToughness(), effectstest.P1g2HasKeyword(g, pyreling, "double strike"))
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB155LightningBolt, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.ApaLive(g, pyreling); live.CurrentPower() != 3 {
		t.Errorf("after two Bolts it is %d/%d, want 3/3", live.CurrentPower(), live.CurrentToughness())
	}
}

func TestGapB155StingerPingsEachOpponentWhenYouCycleAnotherCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Drannith Stinger", "Creature — Human Wizard", gapB155Stinger, 2, 2)
	serpent := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: serpent, Name: "Lava Serpent", TypeLine: "Creature — Elemental Serpent",
			OracleID: gapB155LavaSerpent, ManaCost: "{5}{R}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	})
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, serpent, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle the Serpent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("%s is at %d, want %d", p.Name, p.Life, want)
		}
	}
}

func TestGapB155StingerCyclesForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	stinger := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: stinger, Name: "Drannith Stinger", TypeLine: "Creature — Human Wizard",
			OracleID: gapB155Stinger, ManaCost: "{1}{R}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, stinger, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle the Stinger: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand is %d after cycling, want %d (discard one, draw one)", me.Hand.Size(), hand)
	}
}

func TestGapB155CoalbornEntityTargetsTokensPlayersAndPlaneswalkersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	entity := effectstest.B12Push(g, me.ID, "Coalborn Entity", "Creature — Elemental", gapB155Coalborn, 4, 4)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	nontoken := effectstest.B12Creature(g, opp.ID, "Bear", gapB155Bear, 2, 2)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3}, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatForTest(g, me, "RRR")
	if err := g.ActivateCatalogAbility(me.ID, entity, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(nontoken)}); err == nil {
		t.Fatal("targeted a nontoken creature")
	}
	effectstest.ApaActivate(t, g, me, entity, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(token)})
	if g.Battlefield.Contains(token) {
		t.Error("the token survived 1 damage")
	}
	effectstest.FloatForTest(g, me, "RRR")
	oppLife := opp.Life
	effectstest.ApaActivate(t, g, me, entity, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != oppLife-1 {
		t.Errorf("the player is at %d, want %d", opp.Life, oppLife-1)
	}
	effectstest.FloatForTest(g, me, "RRR")
	effectstest.ApaActivate(t, g, me, entity, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(walker)})
	if got := effectstest.CounterCount(g, walker, game.CounterLoyalty); got != 2 {
		t.Errorf("the planeswalker has %d loyalty, want 2", got)
	}
}
