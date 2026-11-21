package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Within Range — Enchantment {3}{B} (Tarkir: Dragonstorm Commander):
//
//	"When this enchantment enters, create two 1/1 red Warrior creature
//	 tokens.
//	 Whenever you attack, each opponent loses life equal to the number of
//	 creatures attacking them."
//
// The enters trigger is the ordinary token row (the Warrior from Mobilize).
// "Whenever you attack" is Adeline's batch trigger (once per attack
// declaration, however many creatures go, CR 508.2a). Each opponent's count
// is the creatures attacking THAT PLAYER, read as the trigger resolves: a
// creature attacking a planeswalker or battle isn't attacking them, and a
// player nobody attacks loses nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc8c1e21-4540-4e74-a513-eb6c17372de1",
		Name:         "Within Range",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Within Range — create two 1/1 red Warrior creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Warrior"), N: 2})),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Within Range — each opponent loses life equal to the number of creatures attacking them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						n := 0
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.IsCreature() && c.AttackingTarget == opp {
								n++
							}
						}
						if n <= 0 {
							continue
						}
						if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -n); err != nil {
							return err
						}
					}
					return nil
				})),
		},
	})
}
