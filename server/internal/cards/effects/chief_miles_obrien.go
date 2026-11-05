package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chief Miles O'Brien — Legendary Creature — Human Artificer {2}{U}, 2/4
// (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, double the
//	 number of each kind of counter on target artifact or creature you
//	 control."
//
// "Seek out" is an ability word (CR 207.2c). The trigger is Thopter
// Fabricator's "your second card each turn" (b41DrewYourSecondCardThisTurn,
// once a turn), and the body is Vorel of the Hull Clade's: each kind of
// counter on the still-legal target is doubled, every kind snapshotted
// before the first placement, through AddCounter so a counter doubler
// applies (b30DoubleCountersOnFirstLegalTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c98d476a-2646-4171-bab7-4a4624fce917",
		Name:         "Chief Miles O'Brien",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b41DrewYourSecondCardThisTurn(ev, source, g)
				}, "Chief Miles O'Brien — double the number of each kind of counter on target artifact or creature you control",
					b30DoubleCountersOnFirstLegalTarget),
				TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl())),
		},
	})
}
