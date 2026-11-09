package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eirdu, Carrier of Dawn // Isilu, Carrier of Twilight — {3}{W}{W}
// Legendary Creature — Elemental God 5/5 // Legendary Creature — Elemental
// God 5/5 (Lorwyn Eclipsed, #4458):
//
//	Front: "Flying, lifelink
//	        Creature spells you cast have convoke.
//	        At the beginning of your first main phase, you may pay {B}. If
//	        you do, transform Eirdu."
//	Back:  "Flying, lifelink
//	        Each other nontoken creature you control has persist.
//	        At the beginning of your first main phase, you may pay {W}. If
//	        you do, transform Isilu."
//
// The convoke grant is Dazzling Theater's (SpellsYouCastHaveConvoke over
// creature spells); the persist grant is a layer 6 keyword grant to every
// other nontoken creature its controller has, read through the layers, so a
// creature that dies while Isilu is out had persist as it last existed and
// returns with a -1/-1 counter (CR 702.79a). The transform is Ashling,
// Rekindled's MayPay whose paid branch is TransformThis: it does nothing if
// the permanent has turned over since the trigger went on the stack (CR
// 701.27f). Each face's trigger is its own, and only the face that is up
// is on the battlefield.
//
// No simplification.
func init() {
	const oracle = "36af3c2c-49d7-46ea-ab02-c254b332448e"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Eirdu, Carrier of Dawn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		GrantedTapCosts: []game.GrantedTapCost{SpellsYouCastHaveConvoke("Creature spells you cast have convoke.", SpellIs(Creature()))},
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Eirdu, Carrier of Dawn — you may pay {B} to transform Eirdu", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{B}",
					Question: "Eirdu, Carrier of Dawn — pay {B} to transform Eirdu?",
					Then:     eirduTransformPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Isilu, Carrier of Twilight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() &&
					target.InstanceID != source.InstanceID && !game.IsTokenPtr(target)
			}, game.KeywordPersist),
		},
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Isilu, Carrier of Twilight — you may pay {W} to transform Isilu", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{W}",
					Question: "Isilu, Carrier of Twilight — pay {W} to transform Isilu?",
					Then:     eirduTransformPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// eirduTransformPaid is both faces' paid branch, keyed: transform the asking
// item's source, unless it has turned over since (CR 701.27f). An on-disk
// key.
var eirduTransformPaid = PayUnlessBranches("pay-unless/eirdu-transform",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return TransformThis{}.Apply(ctx)
	}, nil)
