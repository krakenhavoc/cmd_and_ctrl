package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar Aang // Aang, Master of Elements — {R}{G}{W}{U} Legendary
// Creature — Human Avatar Ally 4/4, a transforming double-faced card
// (Avatar: The Last Airbender, #4459, seam #5276):
//
//	"Flying, firebending 2
//	 Whenever you waterbend, earthbend, firebend, or airbend, draw a
//	 card. Then if you've done all four this turn, transform Avatar
//	 Aang."
//
// The four bending actions are engine events (game/bending.go), each
// raised where the CR says the action happens: paying a waterbend cost
// (CR 701.67c), creating an earthbend's return trigger (CR 701.66b), a
// firebending ability resolving (CR 702.189b), and an airbend
// instruction exiling something (CR 701.65b). One trigger watches all
// four, for its controller's actions only. Aang's own firebending 2 is
// one of them, so his attack is a firebend.
//
// "Then if you've done all four this turn" is checked as the trigger
// resolves, after the draw, against the player's per-turn record of each
// action (Game.AllFourBendsThisTurn), so a bend made before Aang was on
// the battlefield counts. It is not an intervening "if" (CR 603.4): the
// card is drawn either way. The transform is the permanent turning over
// in place (CR 701.27a), and only if it hasn't transformed since the
// trigger was put on the stack (CR 701.27f, TransformThis).
//
// The back face is aang_master_of_elements.go.
//
// No simplification.
func init() {
	Register(Spec{
		// Face 0's catalog key is the bare oracle ID; the back face
		// registers under avatarAangOracle + "#1".
		OracleID:        avatarAangOracle,
		Name:            "Avatar Aang",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Avatar Aang — firebending 2", Do(firebending(2))),
			OnAny(append([]game.EventKind(nil), game.BendKinds...), ByYou,
				"Avatar Aang — you bent: draw a card, then if you've done all four this turn, transform Avatar Aang",
				Do(
					// The transform check waits for a prompt the draw asks
					// (CR 608.2c, #4155).
					DrawCards{N: 1},
					If{Cond: allFourBendsThisTurn{}, Then: []Applier{TransformThis{}}},
				)),
		},
	})
}

const avatarAangOracle = "b4872bac-5822-4c35-9b73-38c4e3ffa477"

// allFourBendsThisTurn is "if you've done all four this turn": the
// resolving item's controller has waterbent, earthbent, firebent and
// airbent this turn (game.Game.AllFourBendsThisTurn).
type allFourBendsThisTurn struct{}

func (allFourBendsThisTurn) Holds(ctx *Context) bool {
	return ctx.Game.AllFourBendsThisTurn(ctx.Controller())
}

// condition declares no footprint: the per-turn bending record is not a
// state the trigger-order check reads, so the row keeps its prompt.
func (allFourBendsThisTurn) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
