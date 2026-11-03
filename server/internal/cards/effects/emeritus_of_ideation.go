package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emeritus of Ideation // Ancestral Recall — Creature — Human Wizard
// {3}{U}{U}, 5/5 // Instant {U} (Secrets of Strixhaven, #4456; preparation
// card, CR 722):
//
//	"Flying, ward {2}
//	 This creature enters prepared.
//	 Whenever this creature attacks, you may exile eight cards from your
//	 graveyard. If you do, this creature becomes prepared."
//
//	Ancestral Recall — "Target player draws three cards."
//
// "Enters prepared" is SelfEntersPrepared. The attack trigger is Sphinx's
// Approach's "exactly N or none" pick over your graveyard (a set rule on
// the prompt, so choosing fewer than eight declines): no question is asked
// with fewer than eight cards there. The eight are exiled as the cost of
// the "if you do", and only then does the creature become prepared. A
// creature that is already prepared may still exile the cards and gains
// nothing more (the ruling, CR 722.3a). The prepare spell,
// Ancestral Recall, is Ancestral Vision's body word for word, so it
// shares it (targetPlayerDrawsThree).
//
// No simplification.
func init() {
	const oracle = "24a29be4-5bc2-469b-b2ce-2fb305c3facb"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Emeritus of Ideation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Emeritus of Ideation — ward {2}"),
			WheneverThisAttacks("Emeritus of Ideation — you may exile eight cards from your graveyard; if you do, this creature becomes prepared", emeritusOfIdeationAsk),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Ancestral Recall",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Draws: 3}),
		OnResolve:    targetPlayerDrawsThree,
	})
}

// emeritusOfIdeationAsk offers the eight-card exile when the graveyard
// holds that many.
func emeritusOfIdeationAsk(g *game.Game, item *game.StackItem) error {
	var cards []uuid.UUID
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) < 8 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Emeritus of Ideation — exile eight cards from your graveyard to become prepared (choose exactly eight, or none)",
		Cards:    cards,
		Min:      0,
		Max:      8,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  emeritusOfIdeationExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// emeritusOfIdeationExile is the eight cards exiled and then the
// preparing, keyed. "Exactly eight, or none" is a registered set rule, so a
// restored prompt still refuses any other count. An on-disk key: never
// renamed, never reused.
var emeritusOfIdeationExile = askedPickRule("choose-cards/emeritus-of-ideation-exile-eight",
	func(picked []game.Card) bool { return len(picked) == 8 },
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 8 {
			return nil
		}
		for _, id := range r.Picked {
			if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return BecomePrepared{Target: ctx.Item.SourceCardID}.Apply(ctx)
	})
