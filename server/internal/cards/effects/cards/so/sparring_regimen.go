package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sparring Regimen — Enchantment {2}{W} (EDHREC rank 17593):
//
//	"When this enchantment enters, learn. (You may reveal a Lesson card
//	 you own from outside the game and put it into your hand, or discard
//	 a card to draw a card.)
//	 Whenever you attack, put a +1/+1 counter on target attacking
//	 creature and untap it."
//
// The enters trigger learns (CR 701.48a, learn.go). The attack trigger
// is Hollowmurk Siege's Abzan one: once per declaration of attackers
// (CR 603.2c), whatever the number of creatures, with "target attacking
// creature" picked when it goes on the stack. Untapping an attacking
// creature does not remove it from combat (CR 506.4b), and the creature
// keeps attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2aac3575-4a06-452c-bb0f-68dc34086226",
		Name:         "Sparring Regimen",
		Purpose:      game.Purpose{Learn: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sparring Regimen — learn", effects.Do(effects.Learn{})),
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Sparring Regimen — put a +1/+1 counter on target attacking creature and untap it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
						return effects.UntapTarget{Target: id}.Apply(ctx)
					}),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
		},
	})
}
