package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lively Dirge — Sorcery {1}{B} (Outlaws of Thunder Junction):
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Search your library for a card, put it into your graveyard, then
//	   shuffle.
//	 + {2} — Return up to two creature cards with total mana value 4 or less
//	   from your graveyard to the battlefield."
//
// The bullets run in printed order, so the card the first one puts into the
// graveyard can be returned by the second. The second is not targeted: the
// creature cards are chosen as it resolves, up to two with total mana value 4
// or less (a registered set rule over the cards picked, X counting as 0 in a
// graveyard, CR 202.3e), and enter together as one entry (CR 614.12). The
// search finds any card, or none (CR 701.23b), and always shuffles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b23d9aa-d4ed-4ccf-8810-4a4d11d6fb9c",
		Name:         "Lively Dirge",
		Completeness: CompletenessFull,
		Modes: Spree(
			ModeWithPurpose(SpreeModeDoing("Search your library for a card, put it into your graveyard, then shuffle.", "{1}", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:  item.Controller,
						Dest:    game.ZoneGraveyard,
						Limit:   1,
						Shuffle: true,
						Reason:  "Lively Dirge — put a card into your graveyard",
					}.Apply(ctx)
				}), game.Purpose{SelfMillTutor: 1}),
			SpreeModeDoing("Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield.", "{2}", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					p := ctx.PlayerByID(ctx.Controller())
					if p == nil || p.Graveyard == nil {
						return nil
					}
					var cards []uuid.UUID
					for _, c := range p.Graveyard.Cards {
						if c.IsCreature() && !c.IsToken() {
							cards = append(cards, c.InstanceID)
						}
					}
					if len(cards) == 0 {
						return nil
					}
					ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  ctx.Controller(),
						Source:   ctx.Source(),
						Question: "Lively Dirge — return up to two creature cards with total mana value 4 or less to the battlefield",
						Cards:    cards,
						Min:      0,
						Max:      2,
						LiveZone: game.ZoneGraveyard,
						ThenKey:  livelyDirgeReturn,
						Asked:    game.NewAskedItem(ctx.Item),
					})
					return nil
				}),
		),
	})
}

// livelyDirgeReturn is the pick: the cards chosen enter together under your
// control. The total mana value rule is checked as the answer is submitted. An
// on-disk key: never renamed, never reused.
var livelyDirgeReturn = askedPickRule("choose-cards/lively-dirge-return", b23TotalManaValueAtMost(4),
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		return ReturnFromGraveyardTogether{Targets: r.Picked, Controller: ctx.Controller()}.Apply(ctx)
	})
