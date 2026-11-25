package cardsj

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kadena's Silencer — Creature — Snake Wizard {1}{U}, 2/1 (Commander 2019):
//
//	"When this creature is turned face up, counter all abilities your
//	 opponents control.
//	 Megamorph {1}{U} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its megamorph cost and put a +1/+1
//	 counter on it.)"
//
// Megamorph is the shared constructor; the trigger is Belltoll Dragon's
// turned-face-up shape. Summary Dismissal's ability half: every activated or
// triggered ability on the stack controlled by an opponent of the trigger's
// controller, read once as it resolves, is countered (CR 701.6); the trigger
// itself is yours and stays, as does one that can't be countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "02095128-82cd-4531-b47e-471307746f81",
		Name:             "Kadena's Silencer",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Kadena's Silencer — counter all abilities your opponents control", kadenasSilencerCounter),
		},
	})
}

func kadenasSilencerCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var abilities []uuid.UUID
	for id, other := range g.StackMeta {
		if other == nil || id == item.ID || other.Controller == item.Controller {
			continue
		}
		switch other.Kind {
		case game.StackItemActivated, game.StackItemTriggered:
			abilities = append(abilities, id)
		}
	}
	sort.Slice(abilities, func(i, j int) bool { return abilities[i].String() < abilities[j].String() })
	for _, id := range abilities {
		if err := (effects.CounterTarget{StackID: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
