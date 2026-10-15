package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flash_cleanup_auras_test.go — #3712: the nine Visions-era cards that
// print Necromancy's flash rider. Each is cast in the end step (on its
// own flash, so it is sacrificed at cleanup) and in a main phase (so it
// stays), and its effect is read off the permanent while it is there.

type flashCleanupCard struct {
	name, oracle string
	typeLine     string
	cost         string
	color        string
	generic      int
	// host puts the permanent the Aura enchants onto the battlefield; nil
	// for Parapet, which is not targeted.
	host func(g *game.Game, me, opp *game.Player) uuid.UUID
	// check reads the effect off the host (or the board, for Parapet).
	check func(t *testing.T, g *game.Game, me, opp *game.Player, host uuid.UUID)
}

func fcBear(g *game.Game, me, _ *game.Player) uuid.UUID { return seedBear(g, me.ID) }

func fcStats(t *testing.T, g *game.Game, id uuid.UUID, power, toughness int) game.Card {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	if e := c.Effective(); e.Power != power || e.Toughness != toughness {
		t.Errorf("host is %d/%d, want %d/%d", e.Power, e.Toughness, power, toughness)
	}
	return c
}

func fcKeyword(t *testing.T, c game.Card, kw string) {
	t.Helper()
	if !game.HasKeyword(&c, kw) {
		t.Errorf("host does not have %q", kw)
	}
}

var flashCleanupCards = []flashCleanupCard{
	{
		name: "Lightning Reflexes", oracle: "2b0c82f1-078a-4e8b-8e43-407f5c5d8549", cost: "{1}{R}", color: "R", generic: 1, host: fcBear,
		check: func(t *testing.T, g *game.Game, _, _ *game.Player, h uuid.UUID) {
			fcKeyword(t, fcStats(t, g, h, 3, 2), "first strike")
		},
	},
	{
		name: "Spider Climb", oracle: "2636aafa-83f9-4b72-b729-04e1c3aa0eec", cost: "{G}", color: "G", host: fcBear,
		check: func(t *testing.T, g *game.Game, _, _ *game.Player, h uuid.UUID) {
			fcKeyword(t, fcStats(t, g, h, 2, 5), "reach")
		},
	},
	{
		name: "Soar", oracle: "acee6f4f-8b66-495d-bb0e-71c339e7dc49", cost: "{1}{U}", color: "U", generic: 1, host: fcBear,
		check: func(t *testing.T, g *game.Game, _, _ *game.Player, h uuid.UUID) {
			fcKeyword(t, fcStats(t, g, h, 2, 3), "flying")
		},
	},
	{
		name: "Armor of Thorns", oracle: "31ead5a5-aa01-4601-8d3e-267b65fce3bb", cost: "{1}{G}", color: "G", generic: 1, host: fcBear,
		check: func(t *testing.T, g *game.Game, _, _ *game.Player, h uuid.UUID) { fcStats(t, g, h, 4, 4) },
	},
	{
		name: "Grave Servitude", oracle: "6f54e7ea-bb87-4ce0-96b6-4fc82605ed2a", cost: "{1}{B}", color: "B", generic: 1, host: fcBear,
		check: func(t *testing.T, g *game.Game, _, _ *game.Player, h uuid.UUID) {
			c := fcStats(t, g, h, 5, 1)
			if cols := c.Effective().Colors; len(cols) != 1 || cols[0] != "B" {
				t.Errorf("host colors = %v, want black", cols)
			}
		},
	},
	{
		name: "Mystic Veil", oracle: "43c88fcd-aac9-4741-8fd6-31ac81c5bb6f", cost: "{1}{U}", color: "U", generic: 1, host: fcBear,
		check: func(t *testing.T, g *game.Game, _, _ *game.Player, h uuid.UUID) {
			fcKeyword(t, fcStats(t, g, h, 2, 2), "shroud")
		},
	},
	{
		name: "Relic Ward", oracle: "c6321de2-80b0-4679-93cf-97df75344e97", cost: "{1}{W}", color: "W", generic: 1,
		host: func(g *game.Game, me, _ *game.Player) uuid.UUID {
			return pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
				Owner: me.ID, Controller: me.ID,
			})
		},
		check: func(t *testing.T, g *game.Game, _, _ *game.Player, h uuid.UUID) {
			c, _ := battlefieldCard(g, h)
			fcKeyword(t, c, "shroud")
		},
	},
	{
		name: "Ward of Lights", oracle: "edc05b33-2bfa-4692-bb3b-26b7cf111ce8", cost: "{W}{W}", color: "W", generic: 0, host: fcBear,
		check: func(t *testing.T, g *game.Game, _, _ *game.Player, h uuid.UUID) {
			fcKeyword(t, fcStats(t, g, h, 2, 2), game.ProtectionFromColor("R"))
		},
	},
	{
		name: "Parapet", oracle: "4e73634a-9df8-4c36-b1d2-f366eae902fc", typeLine: "Enchantment", cost: "{1}{W}", color: "W", generic: 1,
		host: func(g *game.Game, me, opp *game.Player) uuid.UUID { seedBear(g, opp.ID); return seedBear(g, me.ID) },
		check: func(t *testing.T, g *game.Game, _, opp *game.Player, h uuid.UUID) {
			fcStats(t, g, h, 2, 3)
			for _, c := range g.Battlefield.Cards {
				if c.Controller == opp.ID && c.Effective().Toughness != 2 {
					t.Errorf("an opponent's creature is %d/%d, want it unchanged", c.Effective().Power, c.Effective().Toughness)
				}
			}
		},
	},
}

// fcCast casts the card from the active seat's hand in the current step
// and settles it onto the battlefield.
func fcCast(t *testing.T, g *game.Game, tc flashCleanupCard, host uuid.UUID, targeted bool) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	typeLine := tc.typeLine
	if typeLine == "" {
		typeLine = "Enchantment — Aura"
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: tc.name, TypeLine: typeLine, OracleID: tc.oracle,
		ManaCost: tc.cost, Colors: []string{tc.color}, Owner: me.ID, Controller: me.ID,
	})
	fillPoolColored(me, tc.color, 2+tc.generic)
	params := game.CastSpellParams{}
	if targeted {
		params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: host}}
	}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", tc.name, err)
	}
	return id
}

// fcCleanup advances into the cleanup step (an empty hand, so the CR 514.1
// discard does not hold it) and settles the sacrifice trigger it owes.
func fcCleanup(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	me.Hand.Cards = nil
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.Step != game.StepCleanup {
		t.Fatalf("at %s, want the cleanup step", g.Turn.Step)
	}
	passPriorityAroundTable(t, g)
}

func fcSettle(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	adResolveSpell(t, g, spell)
	if pendingOfKind(g, game.PendingChoiceColor) != nil {
		answerColor(t, g, g.Seats[g.Turn.ActiveSeat].ID, "R")
	}
	passPriorityAroundTable(t, g)
}

func TestFlashCleanupCardsAreSacrificedWhenFlashedIn(t *testing.T) {
	for _, tc := range flashCleanupCards {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			host := tc.host(g, me, opp)
			nmAdvanceTo(t, g, game.StepEnd)
			id := fcCast(t, g, tc, host, tc.name != "Parapet")
			if item := g.StackMeta[id]; item == nil || !item.OwnFlashCast {
				t.Fatalf("an end-step cast of %s is not marked as cast on its own flash", tc.name)
			}
			fcSettle(t, g, id)
			if adZone(g, id) != game.ZoneBattlefield {
				t.Fatalf("%s is in %v, want the battlefield", tc.name, adZone(g, id))
			}
			tc.check(t, g, me, opp, host)
			if nmCleanupTriggers(g) != 1 {
				t.Fatalf("%d cleanup sacrifices scheduled, want 1", nmCleanupTriggers(g))
			}

			fcCleanup(t, g, me)
			if !me.Graveyard.Contains(id) {
				t.Errorf("%s is in %v after cleanup, want the graveyard", tc.name, adZone(g, id))
			}
			if adZone(g, host) != game.ZoneBattlefield {
				t.Errorf("the host left the battlefield: %v", adZone(g, host))
			}
		})
	}
}

func TestFlashCleanupCardsStayWhenCastAtSorcerySpeed(t *testing.T) {
	for _, tc := range flashCleanupCards {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			host := tc.host(g, me, opp)
			nmAdvanceTo(t, g, game.StepPrecombatMain)
			id := fcCast(t, g, tc, host, tc.name != "Parapet")
			if item := g.StackMeta[id]; item == nil || item.OwnFlashCast {
				t.Fatalf("a main-phase cast of %s is marked as cast on its own flash", tc.name)
			}
			fcSettle(t, g, id)
			tc.check(t, g, me, opp, host)
			if nmCleanupTriggers(g) != 0 {
				t.Fatalf("a sorcery-speed cast owes %d cleanup sacrifices", nmCleanupTriggers(g))
			}
			nmAdvanceTo(t, g, game.StepEnd)
			advanceToUpkeepOf(t, g, 1)
			if adZone(g, id) != game.ZoneBattlefield {
				t.Errorf("%s was sacrificed at cleanup after a sorcery-speed cast: %v", tc.name, adZone(g, id))
			}
		})
	}
}

// Armor of Thorns enchants a nonblack creature only.
func TestArmorOfThornsRefusesABlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	black := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gravecrawler", TypeLine: testCreatureTypeLine, Colors: []string{"B"},
		Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	nmAdvanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Armor of Thorns", TypeLine: "Enchantment — Aura",
		OracleID: "31ead5a5-aa01-4601-8d3e-267b65fce3bb", ManaCost: "{1}{G}", Colors: []string{"G"},
		Owner: me.ID, Controller: me.ID,
	})
	fillPoolColored(me, "G", 2)
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: black}}})
	if err == nil {
		t.Fatal("Armor of Thorns was cast on a black creature")
	}
}
