package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Setessan Champion — Creature — Human Warrior {2}{G}, 1/3 (EDHREC
// rank 1018):
//
//	"Constellation — Whenever an enchantment you control enters, put
//	 a +1/+1 counter on this creature and draw a card."
//
// The enchantress deck's growing body. Eidolon of Blossoms' trigger
// with a counter alongside the draw — and unlike the Eidolon it does
// not count itself, because it is not an enchantment. The counter is
// skipped when the Champion has left the battlefield by the time the
// trigger resolves (AddCounter does not gate on zone); the draw still
// happens, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38513660-260f-4663-b199-6473fa004a3b",
		Name:         "Setessan Champion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsEnchantment()
			}, "Setessan Champion — a +1/+1 counter and a card (constellation)", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if b09SourceStillOnBattlefield(g, item) {
					if err := (AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
	})
}
