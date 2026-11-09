package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zidane, Tantalus Thief — Legendary Creature — Human Mutant Scout
// {3}{R}{W}, 3/3 (Final Fantasy, #4462):
//
//	"When Zidane enters, gain control of target creature an opponent
//	 controls until end of turn. Untap it. It gains lifelink and haste
//	 until end of turn.
//	 Whenever an opponent gains control of a permanent from you, you
//	 create a Treasure token."
//
// The entry is an Act of Treason with lifelink as well as haste: control
// lasts to the cleanup step, the creature is untapped, and both keywords
// are granted until end of turn. The second ability reads the layer-2
// control change (EventControlChanged): the permanent left the control of
// this ability's controller for a different player, whoever's it was to
// own — so Zidane's own stolen-back creature and a Mind Control both count,
// and an exchange is one trigger for each permanent that left you.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f267fb54-1881-464c-81af-9e03c3f3b41d",
		Name:         "Zidane, Tantalus Thief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Zidane, Tantalus Thief — gain control of target creature an opponent controls until end of turn; untap it; it gains lifelink and haste",
					zidaneThreaten),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			On(game.EventControlChanged, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventControlChanged && ev.Target == source.Controller &&
					ev.Actor != uuid.Nil && ev.Actor != source.Controller
			}, "Zidane, Tantalus Thief — you create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}

func zidaneThreaten(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (GainControl{Target: target, Controller: item.Controller, Label: "Zidane, Tantalus Thief — gain control until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: target, Keywords: []string{"lifelink", "haste"}, Label: "Zidane, Tantalus Thief"}.Apply(ctx)
}
