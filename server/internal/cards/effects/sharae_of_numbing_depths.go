package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sharae of Numbing Depths — Legendary Creature — Merfolk Wizard
// {2}{W}{U}, 3/4 (Wilds of Eldraine, #4471):
//
//	"When Sharae enters, tap target creature an opponent controls and put
//	 a stun counter on it.
//	 Whenever you tap one or more untapped creatures your opponents
//	 control, draw a card. This ability triggers only once each turn."
//
// The enters trigger taps (an instruction of yours, so the second ability
// sees it) and then puts the stun counter, whether or not the creature was
// already tapped (the ruling). The draw reads the tap an effect of yours
// instructs (YouTappedAnOpponentsCreature): not a cost, an attack, or an
// opponent told to tap their own creature. "One or more" in one tap
// instruction is one trigger, and "only once each turn" caps it at one a
// turn, counted per object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c736935a-e006-4a67-b942-e8ccf1cd9abb",
		Name:         "Sharae of Numbing Depths",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sharae of Numbing Depths — tap target creature an opponent controls and put a stun counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
				}), TargetCreature("target creature an opponent controls", OpponentControls())),
			TriggersOnlyOnceEachTurn(On(game.EventTapCard, YouTappedAnOpponentsCreature,
				"Sharae of Numbing Depths — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
