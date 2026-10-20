package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Estinien Varlineau — Legendary Creature — Elf Warrior {2}{W}{B}, 3/3:
//
//	"Whenever you cast a noncreature spell, put a +1/+1 counter on
//	 Estinien Varlineau. It gains flying until end of turn.
//	 At the beginning of your second main phase, you draw X cards and
//	 lose X life, where X is the number of your opponents who were
//	 dealt combat damage by Estinien Varlineau or a Dragon this turn."
//
// The first half is the spellslinger pump, with the counter and the
// flying grant both pinned to Estinien's own battlefield object: an
// Estinien who was flickered in response gains nothing (CR 400.7).
//
// The second half is a count of DISTINCT opponents (CR 510), however
// many creatures hit each of them, and it is two reads unioned:
//
//   - "by Estinien Varlineau" is the per-object damage record read by
//     the ability's stamped object ref (ObjectRefDealtCombatDamage
//     ToPlayerThisTurn), not by the card's current identity, so an
//     Estinien who died or was exiled in response is still the object
//     that dealt the damage (CR 608.2h) while a re-entered one is a
//     stranger.
//   - "or a Dragon" is the subtype tally read over EVERY controller
//     (PlayersDealtCombatDamageThisTurnByAnySubtype): "a Dragon", not
//     "a Dragon you control", so an opponent's Dragon that hit another
//     opponent counts. It is recorded at the damage, so a Dragon that
//     traded in combat still counts, and a changeling is a Dragon.
//
// "Your opponents" excludes the controller (a Dragon that hit you
// does not count) and an eliminated player. X is counted as the
// ability resolves, so it reads the whole combat, including a second
// combat phase. The draw comes before the life loss, in printed order;
// X of zero does nothing. "Your second main phase" is the turn's
// second main phase (YourSecondMainPhaseBegan), not an added third.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29cf1dc2-3a6a-4a8b-8d61-4b6f0e74f271",
		Name:         "Estinien Varlineau",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Estinien Varlineau — +1/+1 counter and flying",
				// The keyword waits for a prompt the counter asks (a CR 616 order; CR
				// 608.2c, #4155).
				Do(
					Instruction(func(ctx *Context) error {
						return AddCounter{Target: ctx.Item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						return GrantKeywordUntilEOT{
							Target:   ctx.Item.SourceCardID,
							Keywords: []string{"flying"},
							Label:    "Estinien Varlineau — flying until end of turn",
						}.Apply(ctx)
					}),
				)),
			AtYourSecondMainPhase("Estinien Varlineau — draw X cards and lose X life", estinienDrawAndLose),
		},
	})
}

// estinienDrawAndLose is the second-main-phase ability: X is the number
// of the controller's opponents dealt combat damage this turn by
// Estinien (the object that has the ability) or by any Dragon.
// estinienDrawAndLose is "draw X cards and lose X life": the loss waits for a
// prompt the draw asks (a dredge offer; CR 608.2c, #4155), and X is the count
// the draw step filed (#4163).
var estinienDrawAndLose = Do(
	Instruction(estinienDraw),
	Instruction(func(ctx *Context) error {
		if x := ctx.StepRecord(estinienRecord).N(); x > 0 {
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -x)
		}
		return nil
	}),
)

// estinienRecord is X.
const estinienRecord = "estinien-varlineau/x"

func estinienDraw(ctx *Context) error {
	g, item := ctx.Game, ctx.Item
	dragons := g.PlayersDealtCombatDamageThisTurnByAnySubtype(ctx.TextSubtype("Dragon"))
	x := 0
	for _, opp := range ctx.Opponents() {
		var byEstinien bool
		if item.SourceObject.ID != uuid.Nil {
			byEstinien = g.ObjectRefDealtCombatDamageToPlayerThisTurn(item.SourceObject, opp)
		} else {
			byEstinien = g.ObjectDealtCombatDamageToPlayerThisTurn(item.SourceCardID, opp)
		}
		if byEstinien || dragons[opp] {
			x++
		}
	}
	if x == 0 {
		return nil
	}
	ctx.RecordStep(estinienRecord, nil, x)
	return DrawCards{N: x}.Apply(ctx)
}
