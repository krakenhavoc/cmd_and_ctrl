package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func kuonRow() cards.Card {
	return flipRow(kuonOracleID, "{B}{B}{B}", []string{"B"}, nil,
		cards.CardFace{Name: "Kuon, Ogre Ascendant", TypeLine: "Legendary Creature — Ogre Monk",
			OracleText: "At the beginning of the end step, if three or more creatures died this turn, flip Kuon.",
			Power:      "2", Toughness: "4"},
		cards.CardFace{Name: "Kuon's Essence", TypeLine: "Legendary Enchantment",
			OracleText: "At the beginning of each player's upkeep, that player sacrifices a creature of their choice."})
}

// killCreatures destroys n fresh creatures of the given player.
func killCreatures(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		id := apaPush(g, p.ID, p.ID, apaCreature("Doomed", "oracle-doomed", 1, 1))
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Kuon flips at the beginning of the end step if three or more
// creatures died this turn, anyone's, and not with two.
func TestKuonFlipsWhenThreeCreaturesDied(t *testing.T) {
	for _, tc := range []struct {
		name     string
		mine     int
		theirs   int
		wantFlip bool
	}{{"two died", 1, 1, false}, {"three died", 1, 2, true}, {"three of mine", 3, 0, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := flipMain(t)
			id := flipEntered(t, g, kuonRow(), me)
			killCreatures(t, g, me, tc.mine)
			killCreatures(t, g, opp, tc.theirs)
			advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
			passPriorityAroundTable(t, g)
			if tc.wantFlip {
				wantFlipped(t, g, id, "Kuon's Essence", []string{"B"}, 3)
			} else {
				wantUnflipped(t, g, id, "Kuon, Ogre Ascendant")
			}
		})
	}
}

// Kuon's Essence makes each player sacrifice a creature of their choice
// in their own upkeep, the Essence's controller included.
func TestKuonsEssenceMakesEachPlayerSacrificeAtUpkeep(t *testing.T) {
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, kuonRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	keep := apaPush(g, opp.ID, opp.ID, apaCreature("Keeper", "oracle-keeper", 2, 2))
	lose := apaPush(g, opp.ID, opp.ID, apaCreature("Loser", "oracle-loser", 1, 1))

	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, lose)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, lose) || !onBattlefield(g, keep) {
		t.Errorf("sacrificed: Loser still in play %v, Keeper in play %v", onBattlefield(g, lose), onBattlefield(g, keep))
	}
}
