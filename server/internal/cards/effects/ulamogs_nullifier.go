package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ulamog's Nullifier — Creature — Eldrazi Processor {2}{U}{B}, 2/3:
//
//	"Devoid (This card has no color.)
//	 Flash
//	 Flying
//	 When this creature enters, you may put two cards your opponents own
//	 from exile into their owners' graveyards. If you do, counter target
//	 spell."
//
// The target spell is chosen as the trigger goes on the stack (CR 603.3d),
// so with no spell to target the trigger is never put on the stack, as
// printed. As it resolves the controller is asked "you may" only when at
// least two cards their opponents own sit in exile, then picks exactly two
// of them; they go to their owners' graveyards together (a replacement or a
// commander's choice may pause it), and only once they have landed is the
// spell countered, "if you do" (CR 608.2c). The target is re-checked as the
// counter happens (CR 608.2b), and a spell that can't be countered stays.
func init() {
	const name = "Ulamog's Nullifier"
	Register(Spec{
		OracleID:        "bd213210-42d1-49af-9959-08188abe692a",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(name+" — you may put two cards your opponents own from exile into their owners' graveyards; if you do, counter target spell",
				func(g *game.Game, item *game.StackItem) error {
					if len(ulamogsNullifierCandidates(g, item.Controller)) < 2 {
						return nil
					}
					return MayChoice{
						Question: name + " — put two cards your opponents own from exile into their owners' graveyards? If you do, counter target spell.",
						Then:     ulamogsNullifierYes,
					}.Apply(NewContext(g, item))
				}), TargetSpell("target spell")),
		},
	})
}

// ulamogsNullifierCandidates are the cards in exile an opponent of `you`
// owns.
func ulamogsNullifierCandidates(g *game.Game, you uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Owner != uuid.Nil && c.Owner != you && !c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// ulamogsNullifierYes asks for the two cards. An on-disk key.
var ulamogsNullifierYes = MayYes("confirm/ulamogs-nullifier-put", func(ctx *Context, r game.ConfirmAnswer) error {
	cards := ulamogsNullifierCandidates(ctx.Game, ctx.Controller())
	if len(cards) < 2 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Ulamog's Nullifier — choose two cards your opponents own in exile to put into their owners' graveyards",
		Cards:    cards,
		Min:      2,
		Max:      2,
		LiveZone: game.ZoneExile,
		ThenKey:  ulamogsNullifierPicked,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// ulamogsNullifierPicked puts the two cards into their owners' graveyards.
// An on-disk key.
var ulamogsNullifierPicked = askedPick("choose-cards/ulamogs-nullifier-picked", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 2 {
		return nil
	}
	return ctx.Game.PutCardsIntoGraveyardKeyedForEffect(r.Picked, ctx.ItemTail(ulamogsNullifierCounter))
})

// ulamogsNullifierCounter is "if you do, counter target spell", once the
// cards have landed in the graveyards. An on-disk key.
var ulamogsNullifierCounter = CardTail("then/ulamogs-nullifier-counter", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || ctx.Item == nil || len(ctx.Item.Targets) == 0 {
		return nil
	}
	return CounterTarget{StackID: ctx.Item.Targets[0].ID}.Apply(ctx)
})
