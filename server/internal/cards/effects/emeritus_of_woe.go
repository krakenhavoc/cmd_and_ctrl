package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emeritus of Woe // Demonic Tutor — Creature — Vampire Warlock {3}{B},
// 5/4 // Sorcery {1}{B} (Secrets of Strixhaven, #4456; preparation card,
// CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 At the beginning of your end step, if two or more creatures died this
//	 turn, this creature becomes prepared."
//
//	Demonic Tutor — "Search your library for a card, put that card into
//	 your hand, then shuffle."
//
// "Enters prepared" is SelfEntersPrepared. The end-step trigger is
// Bloodline Recollector's with "your" end step and a count of two: the
// intervening if (CR 603.4) reads the table-wide per-turn tally of
// creatures that died, whoever controlled them, as the step begins and
// again as the ability resolves. A creature that is already prepared
// gains nothing (CR 722.3a). The prepare spell is Demonic Tutor word for
// word, so it shares that card's body (demonic_tutor.go).
//
// No simplification.
const emeritusOfWoeOracleID = "93056597-b964-421f-be2f-e92abef1c2a4"

func init() {
	Register(Spec{
		OracleID:     emeritusOfWoeOracleID,
		Name:         "Emeritus of Woe",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep && ev.Actor == source.Controller && b11CreaturesDiedThisTurn(g) >= 2
			}, "Emeritus of Woe — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					if b11CreaturesDiedThisTurn(g) < 2 {
						return nil
					}
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     emeritusOfWoeOracleID + "#1",
		Name:         "Demonic Tutor",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve:    demonicTutorOnResolve,
	})
}
