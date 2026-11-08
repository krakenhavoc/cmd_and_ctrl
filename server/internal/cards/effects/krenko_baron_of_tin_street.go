package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krenko, Baron of Tin Street — Legendary Creature — Goblin {2}{R}, 3/3
// (Murders at Karlov Manor, #4469):
//
//	"Haste
//	 {T}, Sacrifice an artifact: Put a +1/+1 counter on each Goblin you
//	 control.
//	 Whenever an artifact is put into a graveyard from the battlefield,
//	 you may pay {R}. If you do, create a 1/1 red Goblin creature token.
//	 It gains haste until end of turn."
//
// The activation sacrifices an artifact as a cost (Ezrim's sacrificeSpec)
// and puts one counter on each Goblin its controller controls as it
// resolves, Krenko included; "Goblin" is read through the card's text
// (CR 612.2). The trigger is Disciple of the Vault's: any artifact, any
// player's, put into a graveyard from the battlefield, as it last
// existed there (CR 603.10a). As it resolves Krenko's controller may pay
// {R} (Digsite Conservator's MayPay), and the Goblin's haste is Mardu
// Charm's keyed grant over the token that was really made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "91b2520e-85b6-4e1f-88cf-a585feeb8e65",
		Name:            "Krenko, Baron of Tin Street",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Put a +1/+1 counter on each Goblin you control.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact", Artifact())}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				goblin := ctx.TextSubtype("Goblin")
				var goblins []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.HasSubtype(goblin) {
						goblins = append(goblins, c.InstanceID)
					}
				}
				return b13PutCounterOnEach(ctx, goblins)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b23ArtifactPutIntoGraveyardFromBattlefield(ev, g)
			}, "Krenko, Baron of Tin Street — you may pay {R} to create a 1/1 red Goblin with haste",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{R}",
						Question: "Krenko, Baron of Tin Street — pay {R} to create a 1/1 red Goblin that gains haste until end of turn?",
						Then:     krenkoBaronGoblin,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// krenkoBaronGoblin is the paid half, keyed (#4264): the Goblin, then
// haste until end of turn for the token that was made.
var krenkoBaronGoblin = PayUnlessBranches("pay-unless/krenko-baron-of-tin-street-goblin",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(),
			Source:     ctx.Source(),
			Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("1/1 red Goblin")), Count: 1}},
		}, withWord(ctx.Tail(thenCreatedGainKeywordUntilEOT), "haste"))
	}, nil)
