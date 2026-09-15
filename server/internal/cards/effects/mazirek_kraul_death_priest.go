package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mazirek, Kraul Death Priest — Legendary Creature — Insect Shaman
// {3}{B}{G}, 2/2:
//
//	"Flying
//	 Whenever a player sacrifices another permanent, put a +1/+1
//	 counter on each creature you control."
//
// The board-wide aristocrats payoff: not a drain, not a draw, but
// permanent stats that stack. Six tokens fed to a free outlet is
// +6/+6 spread across everything that survived.
//
// Two clauses to read carefully:
//
//   - "A PLAYER" — any player, not just you. An opponent cracking
//     their own fetchland or Treasure grows YOUR board, which is
//     what makes Mazirek good at a four-player table rather than
//     merely fine.
//   - "ANOTHER permanent" — Mazirek sacrificed to your own outlet
//     does not trigger himself.
//
// The counters go on creatures you control AT RESOLUTION, so a token
// created in response gets one and a creature that died in response
// does not. Mazirek counts himself.
func init() {
	Register(Spec{
		OracleID:        "e0420f2c-d578-421e-ae75-e7dc5f70661a",
		Name:            "Mazirek, Kraul Death Priest",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != source.InstanceID
			}, "Mazirek — +1/+1 counter on each creature you control", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller != item.Controller || !c.IsCreature() {
						continue
					}
					if err := (AddCounter{
						Target: c.InstanceID,
						Kind:   game.CounterPlusOne,
						N:      1,
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
