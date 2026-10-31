package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ultron, Artificial Malevolence — {3} Legendary Artifact Creature — Robot
// Villain 2/4 (Marvel Super Heroes, #4455):
//
//	"Whenever another nontoken artifact you control enters, you may pay {2}.
//	 If you do, create a token that's a copy of it. If the token isn't a
//	 creature, it becomes a 2/2 Robot Villain creature in addition to its
//	 other types."
//
// The offer is Ant-Man, Colony Commander's MayPay with the entered artifact
// riding the prompt as data; the paid branch creates the copy through the
// keyed token creation (a copy is made from the artifact wherever it sits
// when the payment is made, CR 707.2) and a registered tail then looks at
// the token that entered. A token that is not a creature becomes a 2/2
// Robot Villain artifact creature for as long as it remains the same object
// (BecomeArtifactCreature: an indefinite layer 4 and 7b effect, pinned to
// the entry, CR 400.7), so the copy is a copy first and a creature only by
// that effect. A token that already is a creature is left as the copy made
// it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8f25e67-2112-4355-9298-363ed04a5cb8",
		Name:         "Ultron, Artificial Malevolence",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact() && !c.IsToken()
			}, ultronLabel, ultronOffer),
		},
	})
}

const ultronLabel = "Ultron, Artificial Malevolence — you may pay {2} to create a token copy of that artifact"

// ultronOffer asks for the {2}.
func ultronOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Ultron, Artificial Malevolence — pay {2} to create a token that's a copy of that artifact?",
		Then:     ultronPaid,
		Carry:    []uuid.UUID{item.Trigger.Event.CardID},
	}.Apply(NewContext(g, item))
}

// ultronPaid is the paid half, keyed: the copy is created and the tail
// reads what entered.
var ultronPaid = PayUnlessBranches("pay-unless/ultron-artificial-malevolence-copy",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		tmpl, ok := TokenCopyTemplate(ctx.Game, r.Carry[0])
		if !ok {
			return nil
		}
		return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(),
			Source:     ctx.Source(),
			Groups:     []game.TokenGroup{{Template: copiedToken(tmpl), Count: 1}},
		}, ctx.Tail(ultronRobotVillain))
	}, nil)

// ultronRobotVillain is "if the token isn't a creature, it becomes a 2/2
// Robot Villain creature in addition to its other types", keyed. An on-disk
// key.
var ultronRobotVillain = CardTail("msh-a9/ultron-robot-villain", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		token, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || token.IsCreature() {
			continue
		}
		if err := (BecomeArtifactCreature{
			Target:       id,
			Types:        []string{"Creature"},
			Subtypes:     []string{"Robot", "Villain"},
			SetPower:     2,
			SetToughness: 2,
			Label:        "Ultron, Artificial Malevolence — becomes a 2/2 Robot Villain creature",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
