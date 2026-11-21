package testsi

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func kuonRow() cards.Card {
	return effectstest.FlipRow(cardskr.KuonOracleID, "{B}{B}{B}", []string{"B"}, nil,
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
		id := effectstest.ApaPush(g, p.ID, p.ID, effectstest.ApaCreature("Doomed", "oracle-doomed", 1, 1))
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
			g, me, opp := effectstest.FlipMain(t)
			id := effectstest.FlipEntered(t, g, kuonRow(), me)
			killCreatures(t, g, me, tc.mine)
			killCreatures(t, g, opp, tc.theirs)
			effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
			effectstest.PassPriorityAroundTable(t, g)
			if tc.wantFlip {
				effectstest.WantFlipped(t, g, id, "Kuon's Essence", []string{"B"}, 3)
			} else {
				effectstest.WantUnflipped(t, g, id, "Kuon, Ogre Ascendant")
			}
		})
	}
}

// Kuon's Essence makes each player sacrifice a creature of their choice
// in their own upkeep, the Essence's controller included.
func TestKuonsEssenceMakesEachPlayerSacrificeAtUpkeep(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, kuonRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	keep := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Keeper", "oracle-keeper", 2, 2))
	lose := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Loser", "oracle-loser", 1, 1))

	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, lose) || !effects.OnBattlefield(g, keep) {
		t.Errorf("sacrificed: Loser still in play %v, Keeper in play %v", effects.OnBattlefield(g, lose), effects.OnBattlefield(g, keep))
	}
}
