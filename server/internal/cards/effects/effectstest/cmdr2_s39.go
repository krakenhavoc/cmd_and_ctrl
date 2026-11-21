package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s39Split builds a split card owned by p, in the given zone ("hand" or
// "graveyard"), and returns its instance ID.
func S39Split(p *game.Player, oracle, zone string, faces ...game.Face) uuid.UUID {
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit,
		Owner: p.ID, Controller: p.ID, Faces: faces,
	}
	card.SettleImported()
	if zone == "graveyard" {
		p.Graveyard.PushTop(card)
	} else {
		p.Hand.PushTop(card)
	}
	return card.InstanceID
}

// s39Fund fills p's pool with plenty of every colour but white.
func S39Fund(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		p.ManaPool = nil
		for _, c := range []string{"U", "B", "R", "G"} {
			for i := 0; i < 8; i++ {
				p.ManaPool.AddMana(game.ManaToken{Color: c})
			}
		}
	})
}

func S39Aftermath(name, cost, text string) game.Face {
	return game.Face{Name: name, TypeLine: "Sorcery", ManaCost: cost,
		OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)\n" + text}
}
