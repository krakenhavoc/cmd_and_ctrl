package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winter, Tormented Loner — Legendary Creature — Human Warlock {2}{B}, 0/3
// (Reality Fracture, tracker #2795):
//
//	"When Winter enters, you may sacrifice a creature or planeswalker.
//	 When you do, each opponent sacrifices a creature of their choice.
//	 Winter gets +1/+0 for each creature and planeswalker card in your
//	 graveyard."
//
// The sacrifice is chosen as the trigger resolves (ChoosePermanents, min 0
// and max 1, with Winter himself a legal pick), and only when something was
// actually sacrificed does the "when you do" reflexive trigger (CR 603.12)
// go on the stack above it. The edict is "of their choice" and does not
// target. The power bonus is a layer 7c modify that counts the controller's
// graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84eaac2e-74ce-4a16-9faa-409df7c58eb3",
		Name:         "Winter, Tormented Loner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Winter, Tormented Loner — you may sacrifice a creature or planeswalker", winterMaySacrifice),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += winterGraveyardCount(g, source.Controller)
			},
		}},
	})
}

// winterGraveyardCount is the number of creature and planeswalker cards in
// `player`'s graveyard.
func winterGraveyardCount(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() || c.IsPlaneswalker() {
			n++
		}
	}
	return n
}

// winterEdictBody is the reflexive "each opponent sacrifices a creature of
// their choice".
var winterEdictBody = game.ReflexiveBody("winter-tormented-loner/edict",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		return EachPlayerSacrifices{ExceptController: true, Match: Creature(), Label: "a creature"}.Apply(NewContext(g, item))
	}), nil)

// winterMaySacrifice offers the sacrifice; a decline (or nothing to
// sacrifice) ends the ability.
func winterMaySacrifice(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question: "Winter, Tormented Loner — you may sacrifice a creature or planeswalker",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return permanentsControlledByMatching(g, of, Or(Creature(), Planeswalker())), 0, 1
		},
		Sacrifice: true,
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			if picked.Count() == 0 {
				return nil
			}
			return ReflexiveTrigger{
				Label: "Winter, Tormented Loner — each opponent sacrifices a creature of their choice",
				Cards: picked.Cards(),
				Body:  winterEdictBody,
			}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
