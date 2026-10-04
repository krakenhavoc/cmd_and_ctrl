package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apocalypse Demon — Creature — Demon {4}{B}{B}, */*:
//
//	"Flying
//	 Apocalypse Demon's power and toughness are each equal to the
//	 number of cards in your graveyard.
//	 At the beginning of your upkeep, tap this creature unless you
//	 sacrifice another creature."
//
// A layer 7a CDA over the controller's graveyard, and The Gitrog
// Monster's "unless" shape with a creature to sacrifice and a tap in
// place of the sacrifice. The sacrifice is the player's choice: with
// no other creature the Demon simply taps.
func init() {
	Register(Spec{
		OracleID:        "5b8084b2-0c3c-4a54-86dc-41f4fe513e99",
		Name:            "Apocalypse Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
					n = p.Graveyard.Size()
				}
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Apocalypse Demon — tap this unless you sacrifice another creature", apocalypseDemonUpkeep),
		},
	})
}

func apocalypseDemonUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := item.Controller
	source := item.SourceCardID
	var others []uuid.UUID
	for _, id := range permanentsControlledByMatching(g, controller, Creature()) {
		if id != source {
			others = append(others, id)
		}
	}
	tap := func(ctx *Context) error { return TapTarget{Target: source}.Apply(ctx) }
	if len(others) == 0 {
		return tap(ctx)
	}
	return PickOption{
		Question: "Apocalypse Demon — sacrifice another creature, or tap Apocalypse Demon",
		Options: []game.ChoiceOption{
			{Label: "Tap Apocalypse Demon"},
			{Label: "Sacrifice another creature"},
		},
		Then: func(ctx *Context, index int) error {
			if index == 1 {
				return SacrificeChoice{
					Player:     controller,
					Candidates: others,
					Question:   "Apocalypse Demon — sacrifice another creature",
				}.Apply(ctx)
			}
			return tap(ctx)
		},
	}.Apply(ctx)
}
