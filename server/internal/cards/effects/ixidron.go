package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ixidron — */* Illusion for {3}{U}{U}:
//
//	"As this creature enters, turn all other nontoken creatures face
//	 down. (They're 2/2 creatures.)
//	 Ixidron's power and toughness are each equal to the number of
//	 face-down creatures on the battlefield."
//
// The board wipe that leaves the board standing, and the card #1209's
// batch form exists for. Two clauses, one each from the two halves of
// the engine:
//
//  1. The sweep is `Spec.AsEntersOthers` — a CR 614.1c "as this
//     enters" clause, not a trigger, so it uses no stack and nothing
//     can be done about it. It is part of the entry (#3862): the engine
//     runs it after Ixidron has landed and before anything announces
//     the arrival, so there is never a moment when Ixidron is on the
//     battlefield and the creatures are face up, and a creature's own
//     "whenever another creature enters" ability does not trigger on
//     Ixidron (its ruling). Every other nontoken creature that was on
//     the battlefield BEFORE this entry is collected and then turned
//     over in ONE batch, so no "whenever a creature is turned face
//     down" trigger can read a half-turned board
//     (game.TurnFaceDownForEffect mutates everything before it emits
//     anything). A creature entering at the same time as Ixidron enters
//     face up (its ruling; game.ArrivingForEffect). A creature that is
//     ALREADY face down is skipped by CR 708.2b, which is Ixidron's own
//     ruling in as many words: "turning a face-down creature face-down
//     typically has no effect".
//
//  2. The size is a layer-7a characteristic-defining ability counting
//     the face-down CREATURES on the battlefield — which is why
//     Ixidron on an empty board is a 0/0 and dies to CR 704.5f, as
//     printed, and why killing one of the creatures it hid shrinks
//     it.
//
// CR 708.7 is the rest of the card: nothing here gives the creatures
// a way back up, so a hidden Sheoldred stays a 2/2 for as long as it
// is on the battlefield — unless its own card happens to print morph
// or disguise, which CR 702.37e and CR 702.168d let its controller
// use whatever put it face down.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a86587ea-14ec-45db-8bac-10b6b9eaeb25",
		Name:         "Ixidron",
		Completeness: CompletenessFull,
		AsEntersOthers: func(self *game.Card, ctx *Context) error {
			return TurnFaceDown{
				Source:  self.InstanceID,
				Targets: otherNontokenCreaturesOnBattlefield(ctx.Game, self.InstanceID),
			}.Apply(ctx)
		},
		Static: []game.StaticAbility{{
			Label:    "Power and toughness are each equal to the number of face-down creatures on the battlefield.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(ch *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := faceDownCreaturesOnBattlefield(g)
				ch.Power = n
				ch.Toughness = n
			},
		}},
	})
}
