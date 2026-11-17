package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sandstalker Moloch — Creature — Lizard {1}{G}{G}, 4/2 (March of the
// Machine, S82 MOM batch 13):
//
//	"Flash
//	 When this creature enters, if an opponent cast a blue and/or black
//	 spell this turn, look at the top four cards of your library. You may
//	 reveal a permanent card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// An intervening "if" (CR 603.4): checked as it enters (the row's AppliesTo)
// and again as it resolves (CondBoard), on the per-turn cast tally of every
// other player, so a blue or black spell an opponent cast earlier this turn
// counts however it ended up. The look is Arthur's: the top four, a
// "may reveal" of one permanent card (never a token, an instant or a
// sorcery) that goes to hand, and the rest on the bottom in a random order,
// the whole four going down when nothing is taken.
//
// No simplification.
func init() {
	const label = "Sandstalker Moloch — look at the top four cards; you may reveal a permanent card and put it into your hand; the rest go on the bottom in a random order"
	cond := CondBoard{
		Name: "an opponent cast a blue and/or black spell this turn",
		Test: func(ctx *Context) bool {
			return s82AnOpponentCastBlueOrBlackSpellThisTurn(ctx.Game, ctx.Controller())
		},
	}
	trig := InterveningIf(cond, On(game.EventETB,
		AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return s82AnOpponentCastBlueOrBlackSpellThisTurn(g, source.Controller)
		}),
		label, sandstalkerMolochLook))
	Register(Spec{
		OracleID:        "f2a6dc8d-3c98-44ae-aff0-b838d7aee0b7",
		Name:            "Sandstalker Moloch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{trig},
	})
}

func sandstalkerMolochLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 4)
	if len(looked) == 0 {
		return nil
	}
	var permanents []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && !IsToken(c) && c.IsPermanent() {
			permanents = append(permanents, id)
		}
	}
	if len(permanents) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Sandstalker Moloch — you may reveal a permanent card from among them and put it into your hand (the rest go to the bottom in a random order)",
		Cards:    permanents,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  sandstalkerMolochPicked,
		Carry:    looked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var sandstalkerMolochPicked = askedPick("choose-cards/s82-mom-b13-sandstalker-moloch-permanent", func(ctx *Context, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	if picked != uuid.Nil {
		ctx.Game.RevealForEffect(game.RevealSpec{Player: ctx.Controller(), Source: ctx.Source(), Reason: "Sandstalker Moloch — reveal the permanent card", Cards: []uuid.UUID{picked}})
		ctx.Game.TakeFromLibraryToHandForEffect(ctx.Controller(), []uuid.UUID{picked})
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, rest)
})
