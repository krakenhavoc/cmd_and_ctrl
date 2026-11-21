package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vraska, the Silencer — Legendary Creature — Gorgon Assassin {1}{B}{G},
// 2/2 (Outlaws of Thunder Junction, #4468):
//
//	"Deathtouch
//	 Whenever a nontoken creature an opponent controls dies, you may pay
//	 {1}. If you do, return that card to the battlefield tapped under your
//	 control. It's a Treasure artifact with '{T}, Sacrifice this artifact:
//	 Add one mana of any color,' and it loses all other card types."
//
// The dying creature is read as it last existed (CR 603.10a): its
// controller was an opponent and it was a nontoken creature. The payment is
// the trigger's own "you may pay" (MayPay). The return is a put with the
// type line set by the putting effect (EntersAs, ADR 0210): the card is only
// an artifact on the battlefield, never a creature, so "whenever a creature
// enters" does not see it, and it keeps its own abilities beside the granted
// mana ability. A card that left the graveyard in response is not returned.
//
// No simplification.
func init() {
	const grant = "vraska-the-silencer/treasure-mana"
	effects.Register(effects.Spec{
		OracleID:        "e7c5a568-3855-4a20-9edf-01a8a6a4ba97",
		Name:            "Vraska, the Silencer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
			}},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead) && effects.LeftUnderControlOf(ev, dead) != source.Controller
			}, "Vraska, the Silencer — you may pay {1}; if you do, return that card to the battlefield tapped under your control as a Treasure artifact",
				func(g *game.Game, item *game.StackItem) error {
					dead := effects.NewContext(g, item).Trigger().Event.CardID
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Vraska, the Silencer — pay {1} to return that card as a Treasure artifact under your control?",
						Then:     vraskaPaid,
						Carry:    []uuid.UUID{dead},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// vraskaPaid is the paid half: the dead creature card returns tapped under
// its payer's control as a Treasure artifact. An on-disk key.
var vraskaPaid = effects.PayUnlessBranches("pay-unless/vraska-the-silencer-return-as-treasure",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		if z := ctx.Game.FindCardZoneForEffect(r.Carry[0]); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return effects.ReturnFromGraveyardTogether{
			Targets:    []uuid.UUID{r.Carry[0]},
			Controller: ctx.Controller(),
			Tapped:     true,
			EntersAs: &game.EntersAs{
				Types:    []string{"Artifact"},
				Subtypes: []string{"Treasure"},
				Grants:   []string{"vraska-the-silencer/treasure-mana"},
				Label:    "Vraska, the Silencer — a Treasure artifact",
			},
		}.Apply(ctx)
	}, nil)
