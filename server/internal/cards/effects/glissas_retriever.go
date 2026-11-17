package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glissa's Retriever — Creature — Phyrexian Beast {5}{G}, 6/6:
//
//	"Haste, toxic 3
//	 This creature can't be blocked by creatures with power 2 or less.
//	 Corrupted — When this creature dies, exile it. When you do, return
//	 up to X target cards from your graveyard to your hand, where X is the
//	 number of opponents who have three or more poison counters."
//
// The evasion is Bristlebane Outrider's block rule. The dies trigger
// exiles the card only while it is still the object that died in the
// graveyard (CR 400.7), and "when you do" is a CR 603.12 reflexive
// trigger raised only once it really reached exile. X is counted as the
// reflexive trigger is put on the stack, when its targets are chosen
// (Teferi Akosa's Params.Amount, which builds the clause), and it is a
// ceiling: "up to X" is any number of cards from none to X. With X of 0
// the reflexive trigger would have no targets and do nothing, so it is
// not raised. The Retriever is in exile, so it is not one of the cards
// that can be chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d099be2e-1879-4bb0-a27d-a386308387c0",
		Name:            "Glissa's Retriever",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", "toxic 3"},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Corrupted — Glissa's Retriever — exile it; when you do, return up to X target cards from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					dead := item.Trigger.Event.CardID
					if !diedCardStillInGraveyard(ctx, dead) {
						return nil
					}
					return ExileTarget{Target: dead, ThenTail: ctx.ItemTail(glissasRetrieverExiled)}.Apply(ctx)
				}),
		},
	})
}

// glissasRetrieverExiled is "when you do": the reflexive trigger, raised
// only for a card that really reached exile and an X above zero. An
// on-disk key.
var glissasRetrieverExiled = CardTail("glissas-retriever/exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	x := s82OneB16CorruptedOpponents(ctx.Game, ctx.Controller())
	if x <= 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Glissa's Retriever — return up to X target cards from your graveyard to your hand",
		Body:   glissasRetrieverReturnBody,
		Params: game.EffectParams{Amount: x},
	}.Apply(ctx)
})

// glissasRetrieverReturnBody is the reflexive trigger's registration: up to
// Params.Amount target cards in your graveyard. An on-disk key.
var glissasRetrieverReturnBody = game.ReflexiveBody("glissas-retriever/return-to-hand",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCardInGraveyard("up to "+strconv.Itoa(p.Amount)+" target cards from your graveyard", YouOwn()).WithCount(0, p.Amount)
	})
