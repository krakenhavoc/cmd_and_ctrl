package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashling, Rekindled // Ashling, Rimebound — {1}{R} Legendary Creature —
// Elemental Sorcerer 1/3 // Legendary Creature — Elemental Wizard 1/3
// (Lorwyn Eclipsed, #4458, #4694):
//
//	Front: "Whenever this creature enters or transforms into Ashling,
//	        Rekindled, you may discard a card. If you do, draw a card.
//	        At the beginning of your first main phase, you may pay {U}. If
//	        you do, transform Ashling."
//	Back:  "Whenever this creature transforms into Ashling, Rimebound and
//	        at the beginning of your first main phase, add two mana of any
//	        one color. Spend this mana only to cast spells with mana value
//	        4 or greater.
//	        At the beginning of your first main phase, you may pay {R}. If
//	        you do, transform Ashling."
//
// The front's first trigger is one ability with two conditions, this
// permanent entering or turning over to its front face (CR 701.27e,
// Huntmaster of the Fells' shape), and its rummage is Cool but Rude's:
// the discard first, and a draw only for a card discarded. The back's
// mana trigger is one ability with two conditions as well: turning over
// to the back face, or your first main phase beginning (CR 505.1a, the
// precombat main phase). It triggers on a transform and a phase, not on
// mana being added, so it is no mana ability (CR 605.1b) and uses the
// stack; the colour is picked as it resolves, and the mana carries
// game.ManaRestrictManaValueAtLeast(4) (CR 106.6), judged on the spell
// being paid for with {X} at its announced value (CR 202.3e). It empties
// at the end of the step like any other mana (CR 500.5).
//
// Each face's "you may pay … If you do, transform Ashling" is a MayPay
// whose paid branch is TransformThis, so it does nothing if Ashling has
// turned over since the trigger went on the stack (CR 701.27f). Both
// faces' first-main-phase triggers fire together on the back face; the
// controller orders them, and the mana trigger resolves whether or not
// Ashling has transformed back first.
//
// No simplification.
func init() {
	const oracle = "31fc5d8b-94bc-4688-8a84-f15800e717a8"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Ashling, Rekindled",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventTransform}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) || ThisTransformedInto(0)(ev, source, lki, g)
			}, ashlingRekindledRummageLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, ashlingRekindledRummageLabel,
					drawAsManyAsDiscarded)(NewContext(g, item))
			}),
			AtYourPrecombatMain(ashlingRekindledTransformLabel, func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{U}",
					Question: "Ashling, Rekindled — pay {U} to transform Ashling?",
					Then:     ashlingTransformPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Ashling, Rimebound",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTransform, game.EventBeginPrecombatMain}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if ev.Kind == game.EventBeginPrecombatMain {
					return ByYou(ev, source, lki, g)
				}
				return ThisTransformedInto(1)(ev, source, lki, g)
			}, "Ashling, Rimebound — add two mana of any one color (spend only on spells with mana value 4 or greater)",
				Do(AddMana{
					Produced:     OneColorOfAmount(2),
					Restrictions: []string{game.ManaRestrictManaValueAtLeast(4)},
				})),
			AtYourPrecombatMain(ashlingRimeboundTransformLabel, func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{R}",
					Question: "Ashling, Rimebound — pay {R} to transform Ashling?",
					Then:     ashlingTransformPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

const (
	ashlingRekindledRummageLabel   = "Ashling, Rekindled — you may discard a card; if you do, draw a card"
	ashlingRekindledTransformLabel = "Ashling, Rekindled — you may pay {U} to transform Ashling"
	ashlingRimeboundTransformLabel = "Ashling, Rimebound — you may pay {R} to transform Ashling"
)

// ashlingTransformPaid is both faces' paid branch, keyed: transform the
// asking item's source, unless it has turned over since (CR 701.27f).
var ashlingTransformPaid = PayUnlessBranches("pay-unless/ashling-transform",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return TransformThis{}.Apply(ctx)
	}, nil)
