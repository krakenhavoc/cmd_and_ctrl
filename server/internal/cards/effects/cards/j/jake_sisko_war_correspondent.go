package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jake Sisko, War Correspondent — {1}{G} Legendary Creature — Human
// Citizen, 2/2 (Star Trek, #4973):
//
//	"Landfall — Whenever a land you control enters, you gain 1 life. If
//	 this is the second time this ability has resolved this turn,
//	 investigate."
//
// "The second time this ability has resolved this turn" is Elrond, Lord
// of Rivendell's count: the per-object resolution tally
// (Game.ResolvedThisTurn), which includes the resolution in progress, so
// only the second resolution investigates, not the third. It is per
// object (CR 400.7), and a landfall trigger that never resolved does not
// count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0c1c97f-b8e1-49ec-b369-3d2694c33f29",
		Name:         "Jake Sisko, War Correspondent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying(jakeSiskoLandfallLabel, game.LandfallPayoff{LifeGain: 1}, jakeSiskoLandfall),
		},
	})
}

const jakeSiskoLandfallLabel = "Jake Sisko, War Correspondent — landfall: gain 1 life, then investigate the second time this turn"

func jakeSiskoLandfall(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.GainLife{Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, jakeSiskoLandfallLabel) != 2 {
		return nil
	}
	return effects.Investigate{N: 1}.Apply(ctx)
}
