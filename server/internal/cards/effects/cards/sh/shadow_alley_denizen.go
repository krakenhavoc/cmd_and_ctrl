package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow Alley Denizen — Creature — Vampire Rogue {B}, 1/1 (Ravnica
// Remastered, #5948):
//
//	"Whenever another black creature you control enters, target creature
//	 gains intimidate until end of turn."
//
// Ivy Lane Denizen's shape: the trigger reads the entering creature's
// effective colour (CR 105.2), "another" is object identity, and the target
// is chosen as the trigger is put on the stack. The grant is a layer 6
// keyword until end of turn on the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3dd2f3de-98cc-4263-bd16-b5803e48e892",
		Name:         "Shadow Alley Denizen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B13AnotherCreatureYouControlEntered(ev, source, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.HasColor(source.TextColor("B"))
			},
			Targets: effects.TargetCreature("target creature"),
			Key:     "Shadow Alley Denizen — target creature gains intimidate until end of turn",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"intimidate"}, Label: "Shadow Alley Denizen — intimidate"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
