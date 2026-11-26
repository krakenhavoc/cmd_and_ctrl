package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kheru Lich Lord — Creature — Zombie Wizard {3}{B}{G}{U}, 4/4 (Khans of
// Tarkir):
//
//	"At the beginning of your upkeep, you may pay {2}{B}. If you do, return a
//	 creature card at random from your graveyard to the battlefield. It gains
//	 flying, trample, and haste. Exile that card at the beginning of your next
//	 end step. If it would leave the battlefield, exile it instead of putting it
//	 anywhere else."
//
// The {2}{B} is paid as the trigger resolves (CR 118.12, Moku's MayPay). The
// card is picked at random from the creature cards in your graveyard (ADR
// 0054's keyed stream, Moldgraf Monstrosity's pick), returned, and given the
// three keywords for as long as it stays (a pinned indefinite grant, not
// until end of turn: the text names no duration). Whip of Erebos's exile at the
// next end step, which is this turn's since the upkeep is before it, and its
// "exile it instead" replacement, pinned to that object. An empty graveyard
// does nothing, the cost already paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e4cbfd1-2393-4643-bc05-022d06b30210",
		Name:         "Kheru Lich Lord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Kheru Lich Lord — you may pay {2}{B}", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}{B}",
					Question: "Pay {2}{B} to return a creature card at random from your graveyard to the battlefield?",
					Then:     kheruLichLordPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// kheruLichLordPaid is the paid half: a random creature card back, with
// flying, trample and haste, exiled at the next end step. An on-disk key.
var kheruLichLordPaid = effects.PayUnlessBranches("pay-unless/s82-old-b352-kheru-lich-lord-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		g := ctx.Game
		controller := ctx.Controller()
		ids := effects.GraveyardIDs(g, controller, func(c game.Card) bool { return c.IsCreature() })
		pick := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
		if len(pick) == 0 {
			return nil
		}
		id := pick[0]
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("flying", "trample", "haste")},
			Duration: g.PinnedTo(game.IndefiniteDuration(), id),
			Label:    "Kheru Lich Lord — the returned creature has flying, trample and haste",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.ScheduleDelayedTrigger{
			Label: "Kheru Lich Lord — exile the returned creature",
			Cards: []uuid.UUID{id},
			Body:  effects.ExileListedCardsBody,
		}).Apply(ctx); err != nil {
			return err
		}
		effects.ExileInsteadOfLeavingBattlefield(g, id, controller, "Kheru Lich Lord: if it would leave the battlefield, exile it instead")
		return nil
	}, nil)
