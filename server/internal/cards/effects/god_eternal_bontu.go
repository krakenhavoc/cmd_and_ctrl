package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// God-Eternal Bontu — Legendary Creature — Zombie God {3}{B}{B}, 5/6
// (EDHREC rank 3546):
//
//	"Menace
//	 When God-Eternal Bontu enters, sacrifice any number of other
//	 permanents, then draw that many cards.
//	 When God-Eternal Bontu dies or is put into exile from the
//	 battlefield, you may put it into its owner's library third from
//	 the top."
//
// The aristocrats deck's refill that will not stay dead. The return
// is God-Eternal Oketra's, unchanged — the same two-condition
// EventLTB, the same "you may", the same tuck under the top two.
//
// The entry ability is where the sandbox differs, and the difference
// is declared. "Sacrifice any number of other permanents" is a
// choice made as the ability RESOLVES; the engine's only
// self-choose-permanents prompt picks one permanent and carries no
// continuation, so "then draw that many" could not follow it. The
// permanents are therefore chosen as the ability's TARGETS instead —
// "any number of other target permanents you control", Appa,
// Steadfast Guardian's Min-0 unbounded clause, Bontu himself kept
// out by name — and every announced permanent that is still legal
// and still the controller's when the ability resolves is
// sacrificed, then that many cards are drawn. Lich-Knights'
// Conquest's trade, on a trigger.
//
// What that costs, all weaker than printed and never stronger:
// opponents see the picks before the ability resolves and can
// respond to them; a permanent of the controller's with shroud
// cannot be chosen; a permanent that became illegal in response is
// skipped and draws nothing; and "becomes the target" triggers on
// the controller's own permanents fire, which the printed sacrifice
// never does. With nothing else on the battlefield the trigger is
// dropped before the prompt (CR 603.3d) — the printed card would put
// it on the stack to do nothing.
func init() {
	Register(Spec{
		OracleID:        "183891b0-b5ec-47f4-8d09-b9d3cfc4e7f1",
		Name:            "God-Eternal Bontu",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"You choose the permanents to sacrifice when the entry trigger goes on the stack rather than as it resolves, so opponents can respond to the picks and a permanent with shroud can't be chosen."},
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets: TargetPermanent("any number of other target permanents you control",
					YouControl(), b03NotNamed("God-Eternal Bontu")).WithCount(0, 0),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "God-Eternal Bontu — sacrifice the chosen permanents, then draw that many cards",
						b33SacrificeChosenThenDrawThatMany)
				},
			},
			Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b22SelfDiedOrWasExiledFromBattlefield(ev, source)
			}, "God-Eternal Bontu — put it into its owner's library third from the top", b33TuckSelfThirdFromTop), "God-Eternal Bontu — put it into its owner's library third from the top?"),
		},
	})
}
