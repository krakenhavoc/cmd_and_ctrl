package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malicious Affliction — Instant {B}{B} (Commander 2014, #5553):
//
//	"Morbid — When you cast this spell, if a creature died this turn, you
//	 may copy Malicious Affliction and may choose a new target for the copy.
//	 Destroy target nonblack creature."
//
// The morbid copy is a cast trigger (CR 601.2i, WhenYouCastThisSpell) with
// an intervening "if" (CR 603.4) on the table-wide death tally
// (b11CreaturesDiedThisTurn): it triggers only when a creature, anyone's,
// already died this turn, and asks "you may" as it resolves. The copy is
// made from the spell's last-known information if it was countered in
// response (Demonstrate's CopySpell), and the caster may pick a new nonblack
// creature for it. The spell itself is Doom Blade.
//
// No simplification.
func init() {
	morbid := WhenYouCastThisSpell("Malicious Affliction — Morbid: you may copy it and may choose a new target for the copy",
		func(g *game.Game, item *game.StackItem) error {
			if b11CreaturesDiedThisTurn(g) == 0 {
				return nil
			}
			return CopySpell{StackID: item.SourceCardID, ChooseNewTargets: true, FromLastKnown: true}.Apply(NewContext(g, item))
		})
	morbid.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return Self(ev, source, lki, g) && b11CreaturesDiedThisTurn(g) > 0
	}
	Register(Spec{
		OracleID:     "18f22cc4-e10b-4af0-9c1f-d7d15dbcce63",
		Name:         "Malicious Affliction",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		Triggered:    []game.TriggeredAbility{Optional(morbid, "Malicious Affliction — copy it? (you may choose a new target for the copy)")},
		OnResolve:    destroyTheTargetPermanent,
	})
}
