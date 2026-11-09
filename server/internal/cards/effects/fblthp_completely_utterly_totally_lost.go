package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fblthp: Completely, Utterly, Totally Lost — Plane — Secret Lair (pssc, #4870):
//
//	"When you planeswalk to this plane and at the beginning of your upkeep,
//	 create a Map token. ...
//	 Whenever chaos ensues, draw two cards, then discard two cards. Untap
//	 up to three lands."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first is one ability with two trigger conditions. The chaos ability
// (CR 311.7) is Frantic Search's text: the loot first, the untap choice
// after the discards are made (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec2f6e72-8a2b-44f0-8241-7fe29c601d96",
		Name:         "Fblthp: Completely, Utterly, Totally Lost",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Fblthp: Completely, Utterly, Totally Lost — create a Map token", Do(CreateToken{Template: MapToken(), N: 1})),
			WheneverChaosEnsues("Fblthp: Completely, Utterly, Totally Lost — draw two cards, then discard two cards; untap up to three lands",
				Do(
					Instruction(func(ctx *Context) error { return lootOne(ctx.Game, ctx.Item, 2) }),
					UntapUpToLands{N: 3, Question: "Fblthp: Completely, Utterly, Totally Lost — untap up to three lands"},
				)),
		},
	})
}
