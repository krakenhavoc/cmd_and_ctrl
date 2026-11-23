package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func orochiEggwatcherRow() cards.Card {
	return effectstest.FlipRow(cardso.OrochiEggwatcherOracleID, "{2}{G}", []string{"G"}, nil,
		cards.CardFace{Name: "Orochi Eggwatcher", TypeLine: "Creature — Snake Shaman",
			OracleText: "{2}{G}, {T}: Create a 1/1 green Snake creature token. If you control ten or more creatures, flip this creature.",
			Power:      "1", Toughness: "1"},
		cards.CardFace{Name: "Shidako, Broodmistress", TypeLine: "Legendary Creature — Snake Shaman",
			OracleText: "{G}, Sacrifice a creature: Target creature gets +3/+3 until end of turn.",
			Power:      "3", Toughness: "3"})
}

func pushBears(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.ApaPush(g, p.ID, p.ID, effectstest.ApaCreature("Bear", "oracle-bear", 2, 2))
	}
}

// The Eggwatcher makes a Snake, then flips if its controller has ten or
// more creatures counting the Snake and the Eggwatcher itself.
func TestOrochiEggwatcherFlipsAtTenCreatures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		others   int
		wantFlip bool
	}{{"eight others plus the Snake", 8, true}, {"seven others plus the Snake", 7, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.FlipMain(t)
			id := effectstest.FlipEntered(t, g, orochiEggwatcherRow(), me)
			pushBears(g, me, tc.others)
			effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
			var snakes int
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Snake" && c.Controller == me.ID {
					snakes++
					if e := c.Effective(); e.Power != 1 || e.Toughness != 1 || len(e.Colors) != 1 || e.Colors[0] != "G" {
						t.Errorf("Snake is %d/%d %v, want a 1/1 green", e.Power, e.Toughness, e.Colors)
					}
				}
			}
			if snakes != 1 {
				t.Fatalf("%d Snakes, want 1", snakes)
			}
			if tc.wantFlip {
				effectstest.WantFlipped(t, g, id, "Shidako, Broodmistress", []string{"G"}, 3)
			} else {
				effectstest.WantUnflipped(t, g, id, "Orochi Eggwatcher")
			}
		})
	}
}

// Shidako sacrifices any creature (itself included) for +3/+3.
func TestShidakoSacrificesACreatureForAPump(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, orochiEggwatcherRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	fodder := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Fodder", "oracle-fodder", 1, 1))
	target := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Target", "oracle-target", 2, 2))
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.CardRefs(target)})
	if effects.OnBattlefield(g, fodder) {
		t.Error("the sacrificed creature is still in play")
	}
	if e := effectstest.FlipState(t, g, target).Effective(); e.Power != 5 || e.Toughness != 5 {
		t.Errorf("Target is %d/%d, want 5/5", e.Power, e.Toughness)
	}
}
