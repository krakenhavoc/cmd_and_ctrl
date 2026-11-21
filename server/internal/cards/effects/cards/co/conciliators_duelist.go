package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conciliator's Duelist — Creature — Kor Warlock {W}{W}{B}{B}, 4/3 (Secrets
// of Strixhaven, #4456):
//
//	"When this creature enters, draw a card. Each player loses 1 life.
//	 Repartee — Whenever you cast an instant or sorcery spell that targets
//	 a creature, exile up to one target creature. Return that card to the
//	 battlefield under its owner's control at the beginning of the next
//	 end step."
//
// The enters trigger draws first, then every player (you included) loses 1
// life, the second as its own step so a draw that waits on a question still
// gets it (CR 608.2c).
//
// Repartee is an ability word (CR 207.2c); the ability is one trigger on
// your instant and sorcery casts that target a creature, read from the
// targets the spell was cast with (informedInkwrightCast, the heroic
// reading): a target that is a creature on the battlefield as the spell is
// cast. It resolves before the spell. Its own target is any creature, up to
// one, chosen as it goes on the stack. The exile is the shared delayed blink
// (exileTargetsThenScheduleReturn, Angel of Condemnation): the card returns
// as a new object under its OWNER's control at the beginning of the next end
// step (CR 603.7), so a creature exiled during an end step comes back at the
// following one, and a token exiled this way is gone (CR 111.8).
//
// No simplification.
func init() {
	const returnLabel = "Conciliator's Duelist — return the exiled creature"
	effects.Register(effects.Spec{
		OracleID:     "d6ea0e58-7b14-425b-9135-57e149c0e47e",
		Name:         "Conciliator's Duelist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Conciliator's Duelist — draw a card; each player loses 1 life", effects.Do(
				effects.DrawCards{N: 1},
				effects.Instruction(effects.SocB22EachPlayerLosesOneLife),
			)),
			effects.Targeting(effects.On(game.EventCast, effects.InformedInkwrightCast,
				"Conciliator's Duelist — Repartee: exile up to one target creature; return it at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), returnLabel)
				}),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}
