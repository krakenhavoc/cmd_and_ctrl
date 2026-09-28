package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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

// castModalPaying is castPaying with modes.
func castModalPaying(t *testing.T, g *game.Game, name, typeLine, oracleID string, modes []int, targets []game.TargetRef,
	optional []int, teamwork, blight []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{
		Modes:         modes,
		Targets:       targets,
		OptionalCosts: optional,
		TeamworkIDs:   teamwork,
		BlightIDs:     blight,
	})
}

func cardRefsOf(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func pushArtifactForTest(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name,
		TypeLine: "Artifact", ManaCost: "{2}", Owner: owner, Controller: owner})
}

// TestHulkSmashModeCountFollowsTeamwork is the whole of "choose both
// instead": unteamed, both bullets is refused; teamed, one bullet is
// refused and both resolve.
func TestHulkSmashModeCountFollowsTeamwork(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		ogre := pushVanillaCreature(g, me.ID, "Ogre", 4, 4)
		biter := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
		victim := pushVanillaCreature(g, opp.ID, "Elf", 2, 2)
		rock := pushArtifactForTest(g, opp.ID, "Mind Stone")
		return g, me, ogre, biter, victim, rock
	}

	t.Run("unteamed, both is refused", func(t *testing.T) {
		g, me, _, biter, victim, rock := setup(t)
		id, err := castModalPaying(t, g, "HULK SMASH!", "Instant", hulkSmashOracle, []int{0, 1},
			cardRefsOf(rock, biter, victim), nil, nil, nil)
		if err == nil {
			t.Fatalf("both bullets without teamwork was accepted")
		}
		if !me.Hand.Contains(id) {
			t.Errorf("a refused cast left its hand")
		}
	})

	t.Run("teamed, one is refused", func(t *testing.T) {
		g, _, ogre, _, _, rock := setup(t)
		if _, err := castModalPaying(t, g, "HULK SMASH!", "Instant", hulkSmashOracle, []int{0},
			cardRefsOf(rock), []int{0}, []uuid.UUID{ogre}, nil); err == nil {
			t.Fatalf("one bullet with teamwork was accepted — teamwork forces both")
		}
	})

	t.Run("teamed, both resolve", func(t *testing.T) {
		g, _, ogre, biter, victim, rock := setup(t)
		if _, err := castModalPaying(t, g, "HULK SMASH!", "Instant", hulkSmashOracle, []int{0, 1},
			cardRefsOf(rock, biter, victim), []int{0}, []uuid.UUID{ogre}, nil); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		if !twCard(g, ogre).Tapped {
			t.Errorf("the teamwork creature was not tapped")
		}
		passPriorityAroundTable(t, g)
		if twCard(g, rock) != nil {
			t.Errorf("the noncreature artifact survived")
		}
		if twCard(g, victim) != nil {
			t.Errorf("the 2/2 survived a 3-power bite")
		}
	})

	t.Run("unteamed, one resolves", func(t *testing.T) {
		g, _, _, biter, victim, _ := setup(t)
		if _, err := castModalPaying(t, g, "HULK SMASH!", "Instant", hulkSmashOracle, []int{1},
			cardRefsOf(biter, victim), nil, nil, nil); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if twCard(g, victim) != nil {
			t.Errorf("the bite did not kill the 2/2")
		}
	})
}

// TestGoNutsCounterLandsBeforeTheFight: printed order, so a 2/2 that
// gets the counter fights as a 3/3, kills a 2/3 and survives it.
func TestGoNutsCounterLandsBeforeTheFight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	team := pushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Troll", 2, 3)
	if _, err := castModalPaying(t, g, "Go Nuts!", "Sorcery", goNutsOracle, []int{0, 1},
		cardRefsOf(bear, bear, theirs), []int{0}, []uuid.UUID{team}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if twCard(g, theirs) != nil {
		t.Errorf("the 2/3 survived a fight with a Bear that had its counter first")
	}
	if c := twCard(g, bear); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Bear should have survived as a 3/3 with its counter: %+v", c)
	}
}

func TestAtlantisAttacksBothBullets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	team := pushVanillaCreature(g, me.ID, "Ogre", 4, 4)
	a := pushArtifactForTest(g, opp.ID, "Rock A")
	b := pushVanillaCreature(g, opp.ID, "Elf", 1, 1)
	targets := append([]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, cardRefsOf(a, b)...)
	if _, err := castModalPaying(t, g, "Atlantis Attacks", "Sorcery", atlantisAttacksOracle, []int{0, 1},
		targets, []int{0}, []uuid.UUID{team}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	_, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	small := pushVanillaCreature(g, opp.ID, "Bear", 5, 3)
	if _, err := castModalPaying(t, g, "Murdock's Crusade", "Sorcery", murdocksCrusadeOracle, []int{0},
		cardRefsOf(small), nil, nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("toughness 3 was accepted for 'toughness 4 or greater': %v", err)
	}
	big := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	if _, err := castModalPaying(t, g, "Murdock's Crusade", "Sorcery", murdocksCrusadeOracle, []int{0},
		cardRefsOf(big), nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if twCard(g, big) != nil {
		t.Errorf("the toughness-4 creature was not exiled")
	}
}

func TestWidowsBiteBothBulletsOnOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	team := pushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	mine := pushVanillaCreature(g, me.ID, "Troll", 4, 4)
	if _, err := castModalPaying(t, g, "Widow's Bite", "Instant", widowsBiteOracle, []int{0, 1},
		cardRefsOf(mine, mine), []int{0}, []uuid.UUID{team}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 2 || tg != 2 {
		t.Errorf("Troll is %d/%d, want 2/2", p, tg)
	}
	if !game.HasKeyword(twCard(g, mine), "deathtouch") {
		t.Errorf("Troll did not gain deathtouch")
	}
}

// TestPyrrhicStrikeBlightForcesBoth: the blight is the cost that
// raises the count, and it may kill the creature that pays it.
func TestPyrrhicStrikeBlightForcesBoth(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	payer := pushVanillaCreature(g, me.ID, "Goblin", 2, 2)
	rock := pushArtifactForTest(g, opp.ID, "Mind Stone")
	big := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ogre",
		TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if _, err := castModalPaying(t, g, "Pyrrhic Strike", "Instant", pyrrhicStrikeOracle, []int{0},
		cardRefsOf(rock), []int{0}, nil, []uuid.UUID{payer}); err == nil {
		t.Fatalf("one bullet with the blight paid was accepted")
	}
	if _, err := castModalPaying(t, g, "Pyrrhic Strike", "Instant", pyrrhicStrikeOracle, []int{0, 1},
		cardRefsOf(rock, big), []int{0}, nil, []uuid.UUID{payer}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if twCard(g, payer) != nil {
		t.Errorf("a 2/2 blighted for 2 survived")
	}
	passPriorityAroundTable(t, g)
	if twCard(g, rock) != nil || twCard(g, big) != nil {
		t.Errorf("both targets should be destroyed")
	}
}
