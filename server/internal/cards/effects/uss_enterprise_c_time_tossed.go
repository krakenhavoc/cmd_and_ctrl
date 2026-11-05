package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Enterprise-C, Time-Tossed — {5} Legendary Artifact — Spacecraft
// 6/6 (Star Trek Commander, #4974):
//
//	"When U.S.S. Enterprise-C enters or dies, you may discard your hand.
//	 If you do, draw a card for each card discarded this way.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 3+.)
//	 3+ | Flying"
//
// "Enters or dies" is one printed ability with two trigger conditions,
// so one row watching both events (Patron of the Arts' shape). It is
// above the station line, so it exists at any charge count (CR 721.4),
// but "dies" means put into a graveyard from the battlefield as a
// creature (CR 700.4): the Spacecraft is one only at three or more
// counters, so the last-known types are read and a destroyed
// Spacecraft below three counters did not die. The discard is the whole
// hand at once, and the draw counts the cards really discarded
// (Forgotten Creation's tail); an empty hand discards and draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2768bab-a6a4-490a-a676-fbfbfefdd8b8",
		Name:         "U.S.S. Enterprise-C, Time-Tossed",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventLTB},
				func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
					if ev.Kind == game.EventETB {
						return ev.CardID == source.InstanceID
					}
					return cardDied(ev, source) && keywordSliceContains(lki.Types, "Creature")
				},
				"U.S.S. Enterprise-C — you may discard your hand; if you do, draw a card for each card discarded this way",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "U.S.S. Enterprise-C — discard your hand? If you do, draw a card for each card discarded this way.",
						Then:     trekB43EnterpriseCDiscardHand,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(3, 6, 6),
			ThresholdKeywords(3, "flying"),
		),
	})
}
