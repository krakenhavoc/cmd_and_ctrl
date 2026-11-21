package testst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// teamwork_modes_test.go — #1703's modal cards: "choose one. If this
// spell was cast using teamwork (or its blight was paid), choose both
// instead." The cost is #1703's; the count is #1655's InsteadIf.

const (
	hulkSmashOracle       = "5f97d49c-d0fa-4776-9455-93c1a172cd83"
	goNutsOracle          = "0efe1357-bfbb-42c0-9cc9-6a919d686c66"
	atlantisAttacksOracle = "b5c8e24a-f4d7-42dc-8117-905ad0bad888"
	murdocksCrusadeOracle = "36330947-5ef9-4bff-b541-bfbddd9715a3"
	widowsBiteOracle      = "39f2a632-30d2-4b35-9e7f-fef27713a4f7"
	pyrrhicStrikeOracle   = "5ea1d89c-8650-4373-9835-e369587020ef"
)

// TestHulkSmashModeCountFollowsTeamwork is the whole of "choose both
// instead": unteamed, both bullets is refused; teamed, one bullet is
// refused and both resolve.
func TestHulkSmashModeCountFollowsTeamwork(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		ogre := effectstest.PushVanillaCreature(g, me.ID, "Ogre", 4, 4)
		biter := effectstest.PushVanillaCreature(g, me.ID, "Bear", 3, 3)
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Elf", 2, 2)
		rock := effectstest.PushArtifactForTest(g, opp.ID, "Mind Stone")
		return g, me, ogre, biter, victim, rock
	}

	t.Run("unteamed, both is refused", func(t *testing.T) {
		g, me, _, biter, victim, rock := setup(t)
		id, err := effectstest.CastModalPaying(t, g, "HULK SMASH!", "Instant", hulkSmashOracle, []int{0, 1},
			effectstest.CardRefsOf(rock, biter, victim), nil, nil, nil)
		if err == nil {
			t.Fatalf("both bullets without teamwork was accepted")
		}
		if !me.Hand.Contains(id) {
			t.Errorf("a refused cast left its hand")
		}
	})

	t.Run("teamed, one is refused", func(t *testing.T) {
		g, _, ogre, _, _, rock := setup(t)
		if _, err := effectstest.CastModalPaying(t, g, "HULK SMASH!", "Instant", hulkSmashOracle, []int{0},
			effectstest.CardRefsOf(rock), []int{0}, []uuid.UUID{ogre}, nil); err == nil {
			t.Fatalf("one bullet with teamwork was accepted — teamwork forces both")
		}
	})

	t.Run("teamed, both resolve", func(t *testing.T) {
		g, _, ogre, biter, victim, rock := setup(t)
		if _, err := effectstest.CastModalPaying(t, g, "HULK SMASH!", "Instant", hulkSmashOracle, []int{0, 1},
			effectstest.CardRefsOf(rock, biter, victim), []int{0}, []uuid.UUID{ogre}, nil); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		if !effectstest.TwCard(g, ogre).Tapped {
			t.Errorf("the teamwork creature was not tapped")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.TwCard(g, rock) != nil {
			t.Errorf("the noncreature artifact survived")
		}
		if effectstest.TwCard(g, victim) != nil {
			t.Errorf("the 2/2 survived a 3-power bite")
		}
	})

	t.Run("unteamed, one resolves", func(t *testing.T) {
		g, _, _, biter, victim, _ := setup(t)
		if _, err := effectstest.CastModalPaying(t, g, "HULK SMASH!", "Instant", hulkSmashOracle, []int{1},
			effectstest.CardRefsOf(biter, victim), nil, nil, nil); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.TwCard(g, victim) != nil {
			t.Errorf("the bite did not kill the 2/2")
		}
	})
}

// TestGoNutsCounterLandsBeforeTheFight: printed order, so a 2/2 that
// gets the counter fights as a 3/3, kills a 2/3 and survives it.
func TestGoNutsCounterLandsBeforeTheFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	team := effectstest.PushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Troll", 2, 3)
	if _, err := effectstest.CastModalPaying(t, g, "Go Nuts!", "Sorcery", goNutsOracle, []int{0, 1},
		effectstest.CardRefsOf(bear, bear, theirs), []int{0}, []uuid.UUID{team}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TwCard(g, theirs) != nil {
		t.Errorf("the 2/3 survived a fight with a Bear that had its counter first")
	}
	if c := effectstest.TwCard(g, bear); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Bear should have survived as a 3/3 with its counter: %+v", c)
	}
}

func TestAtlantisAttacksBothBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	team := effectstest.PushVanillaCreature(g, me.ID, "Ogre", 4, 4)
	a := effectstest.PushArtifactForTest(g, opp.ID, "Rock A")
	b := effectstest.PushVanillaCreature(g, opp.ID, "Elf", 1, 1)
	targets := append([]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, effectstest.CardRefsOf(a, b)...)
	if _, err := effectstest.CastModalPaying(t, g, "Atlantis Attacks", "Sorcery", atlantisAttacksOracle, []int{0, 1},
		targets, []int{0}, []uuid.UUID{team}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Errorf("both nonland permanents should be back in their owner's hand")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Leviathan" && c.Controller == me.ID && c.Power == 6 && c.Toughness == 5 {
			found = true
		}
	}
	if !found {
		t.Errorf("no 6/5 Leviathan under the target player's control")
	}
}

func TestMurdocksCrusadeReadsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	small := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 5, 3)
	if _, err := effectstest.CastModalPaying(t, g, "Murdock's Crusade", "Sorcery", murdocksCrusadeOracle, []int{0},
		effectstest.CardRefsOf(small), nil, nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("toughness 3 was accepted for 'toughness 4 or greater': %v", err)
	}
	big := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	if _, err := effectstest.CastModalPaying(t, g, "Murdock's Crusade", "Sorcery", murdocksCrusadeOracle, []int{0},
		effectstest.CardRefsOf(big), nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TwCard(g, big) != nil {
		t.Errorf("the toughness-4 creature was not exiled")
	}
}

func TestWidowsBiteBothBulletsOnOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	team := effectstest.PushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Troll", 4, 4)
	if _, err := effectstest.CastModalPaying(t, g, "Widow's Bite", "Instant", widowsBiteOracle, []int{0, 1},
		effectstest.CardRefsOf(mine, mine), []int{0}, []uuid.UUID{team}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 2 || tg != 2 {
		t.Errorf("Troll is %d/%d, want 2/2", p, tg)
	}
	if !game.HasKeyword(effectstest.TwCard(g, mine), "deathtouch") {
		t.Errorf("Troll did not gain deathtouch")
	}
}

// TestPyrrhicStrikeBlightForcesBoth: the blight is the cost that
// raises the count, and it may kill the creature that pays it.
func TestPyrrhicStrikeBlightForcesBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	payer := effectstest.PushVanillaCreature(g, me.ID, "Goblin", 2, 2)
	rock := effectstest.PushArtifactForTest(g, opp.ID, "Mind Stone")
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ogre",
		TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if _, err := effectstest.CastModalPaying(t, g, "Pyrrhic Strike", "Instant", pyrrhicStrikeOracle, []int{0},
		effectstest.CardRefsOf(rock), []int{0}, nil, []uuid.UUID{payer}); err == nil {
		t.Fatalf("one bullet with the blight paid was accepted")
	}
	if _, err := effectstest.CastModalPaying(t, g, "Pyrrhic Strike", "Instant", pyrrhicStrikeOracle, []int{0, 1},
		effectstest.CardRefsOf(rock, big), []int{0}, nil, []uuid.UUID{payer}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if effectstest.TwCard(g, payer) != nil {
		t.Errorf("a 2/2 blighted for 2 survived")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TwCard(g, rock) != nil || effectstest.TwCard(g, big) != nil {
		t.Errorf("both targets should be destroyed")
	}
}
