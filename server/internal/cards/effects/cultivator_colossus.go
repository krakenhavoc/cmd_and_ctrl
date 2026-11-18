package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cultivator Colossus — {4}{G}{G}{G} Creature — Plant Beast */* (Innistrad:
// Crimson Vow, S82):
//
//	"Trample
//	 Cultivator Colossus's power and toughness are each equal to the
//	 number of lands you control.
//	 When this creature enters, you may put a land card from your hand
//	 onto the battlefield tapped. If you do, draw a card and repeat this
//	 process."
//
// The characteristic-defining P/T is Beanstalk Giant's (landsYouControlPT).
// The enters trigger is Arboreal Grazer's tapped put from hand with a
// registered tail: when a land arrived the tail draws a card and asks
// again, so the loop ends on the first declined prompt, the first put the
// engine cancels, or the first hand with no land card in it (CR 608.2c).
// The draw comes before the next question, so the card drawn can be the
// next land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4d8505d3-c8a1-4850-95d4-5ddb8ad49770",
		Name:            "Cultivator Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Lands: 1, Draws: 1},
		Static: []game.StaticAbility{
			landsYouControlPT("Cultivator Colossus's power and toughness are each equal to the number of lands you control."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cultivator Colossus — you may put a land card from your hand onto the battlefield tapped; if you do, draw a card and repeat this process",
				func(g *game.Game, item *game.StackItem) error {
					return vowCultivatorColossusPut(NewContext(g, item))
				}),
		},
	})
}

// vowCultivatorColossusPut is one pass of "you may put a land card from
// your hand onto the battlefield tapped", with the repeat as its tail.
func vowCultivatorColossusPut(ctx *Context) error {
	put := MayPutALandFromHandTapped("Cultivator Colossus")
	put.ThenTail = vowCultivatorColossusRepeat
	put.ThenItem = true
	return put.Apply(ctx)
}

// vowCultivatorColossusRepeat is "if you do, draw a card and repeat this
// process". An on-disk key. Registered in init: it runs the put that names it.
var vowCultivatorColossusRepeat game.TailRef

func init() {
	vowCultivatorColossusRepeat = PutFromHandTail("cultivator-colossus/draw-and-repeat", func(ctx *Context, res PutFromHandResult) error {
		if res.Entered == uuid.Nil {
			return nil
		}
		if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
			return err
		}
		return vowCultivatorColossusPut(ctx)
	})
}
