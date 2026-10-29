package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const corpseHarvesterOracle = "f02ad6fa-7a82-471a-979e-0bb9ba866ef6"

func TestCorpseHarvesterSearchSlotsSurviveTextChangesAndRestore(t *testing.T) {
	for _, words := range []struct{ name, zombie, swamp string }{
		{"printed", "Zombie", "Swamp"},
		{"Magical Hack", "Zombie", "Island"},
		{"Artificial Evolution", "Elf", "Swamp"},
		{"both", "Elf", "Island"},
	} {
		for _, picks := range []string{"none", "creature", "land", "both", "dual-first", "dual-last", "two-duals"} {
			t.Run(words.name+"/"+picks, func(t *testing.T) {
				g := newCatalogGame(t)
				me := g.Seats[0]
				advanceToMain(t, g)
				source := b12Push(g, me.ID, "Corpse Harvester", "Creature — Zombie Wizard", corpseHarvesterOracle, 3, 3)
				if words.swamp != "Swamp" {
					twhHack(t, g, source, twhSwamp, twhIsland)
				}
				if words.zombie != "Zombie" {
					twxEvolve(t, g, source, "Zombie", "Elf")
				}
				ids := seedSearchLibrary(me,
					game.Card{Name: "Tribe A", TypeLine: "Creature — " + words.zombie},
					game.Card{Name: "Tribe B", TypeLine: "Kindred Enchantment — " + words.zombie},
					game.Card{Name: "Land A", TypeLine: "Basic Land — " + words.swamp},
					game.Card{Name: "Land B", TypeLine: "Land — " + words.swamp},
					game.Card{Name: "Dual A", TypeLine: "Kindred Land — " + words.zombie + " " + words.swamp},
					game.Card{Name: "Dual B", TypeLine: "Kindred Land — " + words.zombie + " " + words.swamp},
					game.Card{Name: "Unrelated", TypeLine: "Artifact"},
					game.Card{Name: "Old tribe", TypeLine: "Creature — Zombie"},
					game.Card{Name: "Old land", TypeLine: "Basic Land — Swamp"},
				)
				if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{C}"); err != nil {
					t.Fatal(err)
				}
				if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{source}, Strict: true}); err != nil {
					t.Fatal(err)
				}
				if battlefieldHas(g, source) {
					t.Fatal("self-sacrifice was not paid")
				}
				// Preserve both the ability's catalog identity and departed source's
				// words before it resolves; then preserve the captured prompt lens.
				g = restoreRoundTrip(t, g, true)
				passPriorityAroundTable(t, g)
				g = restoreRoundTrip(t, g, true)
				me = g.Seats[0]
				c := searchChoiceFor(g, me.ID)
				if c == nil {
					t.Fatal("missing search")
				}
				beginSearchIfAsked(t, g, c)
				invalid := [][]uuid.UUID{{ids[0], ids[1]}, {ids[2], ids[3]}, {ids[4], ids[4]}, {ids[6]}}
				if words.zombie != "Zombie" {
					invalid = append(invalid, []uuid.UUID{ids[7]})
				}
				if words.swamp != "Swamp" {
					invalid = append(invalid, []uuid.UUID{ids[8]})
				}
				for _, bad := range invalid {
					var legal bool
					g.ReadSnapshot(func() { legal = g.SearchPickLegalLocked(c, bad) })
					if legal {
						t.Fatalf("bot may choose invalid set %v", bad)
					}
					if err := g.ResolveSearchLibrary(c.ID, me.ID, bad); err == nil {
						t.Fatalf("accepted invalid set %v", bad)
					}
					if searchChoiceFor(g, me.ID) == nil {
						t.Fatal("rejection removed search prompt")
					}
				}
				wanted := map[string][]uuid.UUID{"none": nil, "creature": {ids[1]}, "land": {ids[2]}, "both": {ids[0], ids[2]}, "dual-first": {ids[4], ids[0]}, "dual-last": {ids[2], ids[4]}, "two-duals": {ids[4], ids[5]}}[picks]
				g.ReadSnapshot(func() {
					if !g.SearchPickLegalLocked(c, wanted) {
						t.Error("bot rejected legal assignment")
					}
				})
				answerSearchByID(t, g, me.ID, wanted...)
				for _, id := range wanted {
					if !me.Hand.Contains(id) {
						t.Fatal("found card missing from hand")
					}
				}
				searches := 0
				for _, ev := range g.Events {
					if ev.Kind == game.EventSearchLibrary && ev.Actor == me.ID {
						searches++
						if ev.Label != game.SearchShuffledLabel {
							t.Fatal("search did not shuffle")
						}
					}
				}
				if searches != 1 || searchChoiceFor(g, me.ID) != nil {
					t.Fatalf("searches=%d; want one completed search and shuffle", searches)
				}
			})
		}
	}
}

func TestCorpseHarvesterActivationCosts(t *testing.T) {
	for _, tc := range []struct {
		name, mana          string
		tapped, sick, noSac bool
	}{
		{name: "mana", mana: "{B}"},
		{name: "tap", mana: "{C}{B}", tapped: true},
		{name: "summoning sickness", mana: "{C}{B}", sick: true},
		{name: "sacrifice", mana: "{C}{B}", noSac: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			source := b12Push(g, me.ID, "Corpse Harvester", "Creature — Zombie Wizard", corpseHarvesterOracle, 3, 3)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == source {
						g.Battlefield.Cards[i].Tapped = tc.tapped
						g.Battlefield.Cards[i].SummonedThisTurn = tc.sick
					}
				}
			})
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.mana); err != nil {
				t.Fatal(err)
			}
			sacrifice := []uuid.UUID{source}
			if tc.noSac {
				sacrifice = nil
			}
			if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{SacrificeIDs: sacrifice, Strict: true}); err == nil {
				t.Fatalf("accepted unpaid %s cost", tc.name)
			}
			if !battlefieldHas(g, source) {
				t.Fatal("failed activation sacrificed source")
			}
		})
	}
}
