package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spirit of Resilience — Creature — Spirit Warrior {2}{R}, 2/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Whenever one or more cards leave your graveyard, put a +1/+1 counter
//	 on this creature, then you may have this creature become a copy of an
//	 artifact or creature card from among those cards until end of turn."
//
// One trigger per batch of cards leaving the graveyard together (CR 603.2c,
// OncePerBatch over whenCardsLeaveYourGraveyard, a cast from it too); the
// 2026-03-20 ruling says several cards leaving at once trigger it once. The
// counter goes on first, and only while the Spirit is still on the
// battlefield. The "then you may" is a pick of up to one of the artifact or
// creature cards of that batch (socB13ArtifactOrCreatureCardsThatLeftYourGraveyard),
// wherever each now is, and it is offered only when there is such a card
// and the Spirit is still the object that triggered.
//
// The copy is BecomeCopy until end of turn (CR 707.2, CR 611.2): the card's
// printed copiable values taken as the choice is made, with the Spirit's
// counters kept on top. Becoming a copy is neither entering nor leaving the
// battlefield, so nothing triggers. A mana cost with {X} reads X as 0
// (CR 107.3m on a permanent), which BecomeCopy's stored values already do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ddc85b23-113a-46e9-ab09-fae9312c72fa",
		Name:         "Spirit of Resilience",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(leftGraveyardEvents, whenCardsLeaveYourGraveyard,
				"Spirit of Resilience — put a +1/+1 counter on this creature, then you may have it become a copy of an artifact or creature card that left your graveyard until end of turn",
				spiritOfResilienceResolve)),
		},
	})
}

func spiritOfResilienceResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if item.Trigger == nil || !onBattlefield(g, ctx.Source()) {
		return nil
	}
	cards := socB13ArtifactOrCreatureCardsThatLeftYourGraveyard(g, item.Controller, item.Trigger.Event)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Spirit of Resilience — you may have it become a copy of one of these cards until end of turn",
		Cards:    cards,
		Min:      0,
		Max:      1,
		ThenKey:  spiritOfResilienceCopyPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// spiritOfResilienceCopyPick is the copy made of the chosen card. The asking
// item rides the prompt as data, so a table waiting on the pick is a restore
// point; BecomeCopy leaves a Spirit that has since left or returned alone
// (CR 400.7).
var spiritOfResilienceCopyPick = askedPick("choose-cards/spirit-of-resilience-copy", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return BecomeCopy{
		Targets: []uuid.UUID{ctx.Source()},
		Of:      r.Picked[0],
		Label:   "Spirit of Resilience — becomes a copy of the chosen card until end of turn",
	}.Apply(ctx)
})
