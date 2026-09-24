package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cori-Steel Cutter — Artifact — Equipment {1}{R}:
//
//	"Equipped creature gets +1/+1 and has trample and haste.
//	 Flurry — Whenever you cast your second spell each turn, create a
//	 1/1 white Monk creature token with prowess. You may attach this
//	 Equipment to it.
//	 Equip {1}{R}"
//
// Batch 40 (#403) skipped it for one reason: "the Monk token has
// prowess, which the engine does not implement anywhere". Since #706
// it does — prowess is a keyword the engine turns into a trigger — so
// the Monk is the token table's "1/1 white Monk with prowess" row.
//
// Flurry is an ability word (CR 207.2c), not a keyword: it is
// YouCastYourSecondSpellEachTurn, the condition Breeches and Avatar
// Yangchen already read. Any second spell counts, creature or not.
//
// "You may attach this Equipment to it" is asked at RESOLUTION
// (MayChoice), after the Monk exists, and rides the token creation's
// continuation rather than the next line: a token doubler opens a CR
// 616 ordering prompt and the token's ID is only known once it is
// answered (living_weapon.go explains the same trap). With two Monks
// from a doubler the offer is for the first. If the Cutter has left
// the battlefield, or left and come back, the attach does nothing
// (CR 701.3b, CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46b1a169-bc32-47bf-b379-fa0c3d13878f",
		Name:         "Cori-Steel Cutter",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("trample", "haste"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Cori-Steel Cutter — flurry: create a 1/1 Monk with prowess, then you may attach this to it",
				coriSteelCutterFlurry),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}{R}"),
		},
	})
}

// coriSteelCutterFlurry is the flurry trigger's resolution. A
// package-level func: it captures nothing and reads everything off the
// item it is handed, so it survives the Clone an undo restores.
func coriSteelCutterFlurry(g *game.Game, item *game.StackItem) error {
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: TokenCard("1/1 white Monk with prowess"),
			Count:    1,
		}},
	}, func(g *game.Game, created []uuid.UUID) error {
		if len(created) == 0 {
			return nil
		}
		monk := created[0]
		return MayChoice{
			Question: "Cori-Steel Cutter — attach it to the Monk?",
			YesLabel: "Attach",
			NoLabel:  "Don't attach",
			OnYes: func(ctx *Context) error {
				return ctx.Game.AttachSourceForEffect(ctx.Item, game.TargetRef{Kind: game.TargetCard, ID: monk})
			},
		}.Apply(NewContext(g, item))
	})
}
