package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 ONE batch 1: keyword-only creatures of Phyrexia: All Will Be One.

var s82OneB01Cards = []struct {
	name, oracle, typeLine string
	power, toughness       int
	keywords               []string
	toxic                  int
}{
	{"Swooping Lookout", "0594f19c-c4c2-4812-8f4d-5630fc24c47a", "Artifact Creature — Phyrexian Construct", 1, 2, []string{"flying", "vigilance"}, 0},
	{"Resistance Skywarden", "419e9ab0-fe79-4a4f-8bcb-fe4e5373e292", "Creature — Ogre Rebel", 5, 5, []string{"reach", "menace"}, 0},
	{"Bilious Skulldweller", "83bf5610-a798-4508-88d2-c60d96f60d23", "Creature — Phyrexian Insect", 1, 1, []string{"deathtouch"}, 1},
	{"Branchblight Stalker", "ae685a9b-4cb8-4a8a-935d-378102431c51", "Creature — Phyrexian Elf Scout", 3, 1, nil, 2},
	{"Ichorspit Basilisk", "5a2f8474-b032-4567-a569-3f9725805f10", "Creature — Phyrexian Basilisk", 1, 3, []string{"deathtouch"}, 1},
	{"Jawbone Duelist", "14af3ec5-d90a-4560-a70a-8645765f1253", "Creature — Phyrexian Soldier", 1, 1, []string{"double strike"}, 1},
	{"Pestilent Syphoner", "1b27303b-756e-45b9-956f-6bbbd83d31ab", "Creature — Phyrexian Insect", 1, 1, []string{"flying"}, 1},
	{"Sheoldred's Headcleaver", "a7df9a4e-d556-4345-9884-0724a3f9d5f8", "Creature — Phyrexian Warrior", 2, 4, []string{"menace"}, 2},
	{"Tyrranax Atrocity", "08261b68-b261-4e47-a085-c489ba394108", "Creature — Phyrexian Dinosaur", 4, 4, []string{"haste"}, 3},
	{"Skyscythe Engulfer", "b7cc831c-c1a1-4ea4-ae78-5be2b6173407", "Creature — Phyrexian Beast", 6, 5, []string{"reach", "trample"}, 0},
	{"Paladin of Predation", "fb5a02ce-6080-47f9-815a-fef3758d753b", "Creature — Phyrexian Knight", 6, 7, nil, 6},
	{"Oxidda Finisher", "f308a1de-94a6-4f9f-8f46-6f2018959b40", "Creature — Ogre Rebel", 7, 5, []string{"trample"}, 0},
	{"Duelist of Deep Faith", "30d69e88-2bf6-43b8-8678-2cf37f6a5cf5", "Creature — Phyrexian Soldier", 2, 2, nil, 1},
}

const (
	s82OneSkyscytheOracle = "b7cc831c-c1a1-4ea4-ae78-5be2b6173407"
	s82OnePaladinOracle   = "fb5a02ce-6080-47f9-815a-fef3758d753b"
	s82OneOxiddaOracle    = "f308a1de-94a6-4f9f-8f46-6f2018959b40"
	s82OneDuelistOracle   = "30d69e88-2bf6-43b8-8678-2cf37f6a5cf5"
)

func TestS82OneB01CardsAreFullAndCarryTheirKeywords(t *testing.T) {
	for _, c := range s82OneB01Cards {
		t.Run(c.name, func(t *testing.T) {
			spec, ok := Lookup(c.oracle)
			if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
				t.Fatalf("%s: %+v %v", c.name, spec, ok)
			}
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			card, _ := battlefieldCard(g, id)
			for _, kw := range c.keywords {
				if !game.HasKeyword(&card, kw) {
					t.Errorf("%s lacks %s", c.name, kw)
				}
			}
			if got := game.ToxicTotal(&card); got != c.toxic {
				t.Errorf("%s has toxic %d, want %d", c.name, got, c.toxic)
			}
		})
	}
}

func TestS82OneB01ToxicGivesPoisonWithCombatDamage(t *testing.T) {
	for _, c := range s82OneB01Cards {
		if c.toxic == 0 || c.name == "Jawbone Duelist" || c.name == "Pestilent Syphoner" {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, me.ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			attackWith(t, g, opp.ID, id)
			if opp.Life != game.StartingLife-c.power || poisonOn(opp) != c.toxic {
				t.Errorf("%s: %d life, %d poison; want %d and %d", c.name, opp.Life, poisonOn(opp), game.StartingLife-c.power, c.toxic)
			}
		})
	}
}

func TestS82OneB01JawboneDuelistDoubleStrikePoisonsTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, "Jawbone Duelist", "Creature — Phyrexian Soldier", "14af3ec5-d90a-4560-a70a-8645765f1253", 1, 1)
	attackWith(t, g, opp.ID, id)
	advanceTo(t, g, game.StepEndCombat)
	if opp.Life != game.StartingLife-2 || poisonOn(opp) != 2 {
		t.Errorf("%d life, %d poison; want %d and 2 (toxic 1 on each strike)", opp.Life, poisonOn(opp), game.StartingLife-2)
	}
}

func TestS82OneB01SkyscytheEngulferCantBeBlockedByFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engulfer := b12Push(g, me.ID, "Skyscythe Engulfer", "Creature — Phyrexian Beast", s82OneSkyscytheOracle, 6, 5)
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	reacher := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Reacher", TypeLine: "Creature — Spider", Keywords: []string{"reach"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(engulfer, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(flyer, engulfer); err == nil {
		t.Error("a flying creature blocked the Engulfer")
	}
	if err := g.DeclareBlocker(reacher, engulfer); err != nil {
		t.Errorf("a reach creature without flying could not block: %v", err)
	}
}

func TestS82OneB01PaladinOfPredationCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	paladin := b12Push(g, me.ID, "Paladin of Predation", "Creature — Phyrexian Knight", s82OnePaladinOracle, 6, 7)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Beast", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(paladin, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, paladin); err == nil {
		t.Error("a power 2 creature blocked the Paladin")
	}
	if err := g.DeclareBlocker(big, paladin); err != nil {
		t.Errorf("a power 3 creature could not block: %v", err)
	}
}

func TestS82OneB01OxiddaFinisherHasAffinityForEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, s82OneOxiddaOracle, "Creature — Ogre Rebel", "{5}{R}{R}"); got != 7 {
		t.Errorf("no Equipment: it costs %d, want 7", got)
	}
	b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	b12Permanent(g, me.ID, "Blade", "Artifact — Equipment")
	b12Permanent(g, me.ID, "Plain Rock", "Artifact")
	if got := selfPricedMV(t, g, me, s82OneOxiddaOracle, "Creature — Ogre Rebel", "{5}{R}{R}"); got != 5 {
		t.Errorf("two Equipment: it costs %d, want 5", got)
	}
	for i := 0; i < 6; i++ {
		b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	}
	if got := selfPricedMV(t, g, me, s82OneOxiddaOracle, "Creature — Ogre Rebel", "{5}{R}{R}"); got != 2 {
		t.Errorf("eight Equipment: it costs %d, want 2 (only generic is reduced)", got)
	}
}

func TestS82OneB01DuelistOfDeepFaithHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Push(g, me.ID, "Duelist of Deep Faith", "Creature — Phyrexian Soldier", s82OneDuelistOracle, 2, 2)
	theirs := b12Push(g, opp.ID, "Duelist of Deep Faith", "Creature — Phyrexian Soldier", s82OneDuelistOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "first strike") {
		t.Error("no first strike during its controller's turn")
	}
	if hasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("first strike on another player's turn")
	}
}
