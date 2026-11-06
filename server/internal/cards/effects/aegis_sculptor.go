package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aegis Sculptor — Creature — Bird Wizard {3}{U}, 2/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 Ward {2}
//	 At the beginning of your upkeep, you may exile two cards from your
//	 graveyard. If you do, put a +1/+1 counter on this creature."
//
// Ward is the shared trigger. The upkeep trigger is Emeritus of Ideation's
// "exactly N or none" pick over your graveyard, with N two: no question is
// asked with fewer than two cards there, choosing none declines, and only
// once both cards are exiled does the counter go on. A Sculptor that has
// left the battlefield in response still lets you exile the cards, and
// gets no counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a75c56a-bdf7-490d-aa41-51d5eb8be46c",
		Name:            "Aegis Sculptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Aegis Sculptor — ward {2}"),
			AtYourUpkeep("Aegis Sculptor — you may exile two cards from your graveyard; if you do, put a +1/+1 counter on it",
				aegisSculptorAsk),
		},
	})
}

// aegisSculptorAsk offers the two-card exile when the graveyard holds two.
func aegisSculptorAsk(g *game.Game, item *game.StackItem) error {
	var cards []uuid.UUID
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) < 2 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Aegis Sculptor — exile two cards from your graveyard to put a +1/+1 counter on it (choose exactly two, or none)",
		Cards:    cards,
		Min:      0,
		Max:      2,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  aegisSculptorExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// aegisSculptorExile is the two cards exiled and then the counter, keyed.
// "Exactly two, or none" is a registered set rule. An on-disk key: never
// renamed, never reused.
var aegisSculptorExile = askedPickRule("choose-cards/aegis-sculptor-exile-two",
	func(picked []game.Card) bool { return len(picked) == 2 },
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 2 {
			return nil
		}
		for _, id := range r.Picked {
			if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	})
