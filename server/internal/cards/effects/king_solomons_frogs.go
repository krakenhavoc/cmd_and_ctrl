package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// King Solomon's Frogs — Legendary Artifact {3}{W} (Marvel Super Heroes
// Commander):
//
//	"Flash
//	 When King Solomon's Frogs enters, if you cast it, for each opponent, exile
//	 up to one target permanent that player controls with mana value 3 or
//	 greater. For each permanent exiled this way, its controller draws a card.
//	 {3}, {T}, Exile King Solomon's Frogs: You become the monarch."
//
// The enter trigger is Celebrate the Mountain-king's per-opponent clause list
// (perOpponentClauses, each pick bound to its opponent, CR 608.2b) gated on
// "if you cast it" (CR 603.4, the catalog's one reading of it, checked again as
// the trigger resolves). Each still-legal pick is exiled one after another
// through its own keyed continuation; the continuation draws for the pick's
// controller as the permanent was on the battlefield, only when it really
// reached exile (a replacement that sends it elsewhere draws nothing, the
// "exiled this way" wording).
//
// No simplification.
func init() {
	kingSolomonsFrogsExiledOne = CardTail("then/king-solomons-frogs-exiled-one", func(ctx *Context, o game.TailOutcome) error {
		carry := o.Tail.Carry
		if len(carry) == 0 {
			return nil
		}
		if len(o.IDs) > 0 {
			if err := (DrawCards{Player: carry[0], N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return kingSolomonsFrogsExileNext(ctx, carry[1:])
	})
	enters := On(game.EventETB, AllOf(Self, crestingMosasaurusYouCastIt),
		"King Solomon's Frogs — for each opponent, exile up to one target permanent with mana value 3 or greater",
		func(g *game.Game, item *game.StackItem) error {
			if !b16EnteredFromStack(g, item.SourceCardID) {
				return nil
			}
			ctx := NewContext(g, item)
			var picks []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				card, ok := g.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				picks = append(picks, t.ID, card.Controller)
			}
			return kingSolomonsFrogsExileNext(ctx, picks)
		})
	enters.Targets = perOpponentClauses(game.MaxPlayers-1, "permanent with mana value 3 or greater", ManaValueGE(3))
	Register(Spec{
		OracleID:        "36396a14-9d0b-4670-bf8a-5c346fdda197",
		Name:            "King Solomon's Frogs",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
		Activated: []ActivatedAbility{{
			Label:  "{3}, {T}, Exile King Solomon's Frogs: You become the monarch.",
			Cost:   Plus(ManaCost("{3}"), TapCost(), ExileThis()),
			Effect: Do(BecomeTheMonarch{}),
			// Becoming the monarch answers nothing on the stack.
			Purpose: game.Purpose{Answers: game.AnswerValue},
		}},
	})
}

// kingSolomonsFrogsExileNext exiles the first pick of `picks`, a flat list of
// (permanent, its controller) pairs; the controller and the rest ride the
// exile's continuation as its Carry.
func kingSolomonsFrogsExileNext(ctx *Context, picks []uuid.UUID) error {
	if len(picks) < 2 {
		return nil
	}
	t := ctx.ItemTail(kingSolomonsFrogsExiledOne)
	t.Carry = append([]uuid.UUID{picks[1]}, picks[2:]...)
	return ExileTarget{Target: picks[0], ThenTail: t}.Apply(ctx)
}

// kingSolomonsFrogsExiledOne is one exile settled: draw for its controller,
// then the next pick. An on-disk key, registered in init (it refers to
// kingSolomonsFrogsExileNext, which refers to it).
var kingSolomonsFrogsExiledOne game.TailRef
