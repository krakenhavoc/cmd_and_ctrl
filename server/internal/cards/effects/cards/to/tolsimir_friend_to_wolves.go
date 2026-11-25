package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tolsimir, Friend to Wolves — Legendary Creature — Elf Scout {2}{G}{G}{W}, 3/3
// (War of the Spark, #5520):
//
//	"When Tolsimir enters, create Voja, Friend to Elves, a legendary 3/3
//	 green and white Wolf creature token.
//	 Whenever a Wolf you control enters, you gain 3 life and that creature
//	 fights up to one target creature you don't control."
//
// Jiang Yanggu's legendary token. The second ability watches every Wolf that
// enters under your control, Voja included: the life is gained whatever the
// target, and the entering Wolf fights the target creature only while both are
// still legal and on the battlefield (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cef501b-5ce1-4699-89a5-3949f7e65bf1",
		Name:         "Tolsimir, Friend to Wolves",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tolsimir, Friend to Wolves — create Voja, Friend to Elves, a legendary 3/3 Wolf",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green and white Legendary Wolf named Voja, Friend to Elves"), N: 1})),
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.HasSubtypeWord("Wolf")
				},
				Targets: effects.TargetCreature("up to one target creature you don't control", effects.Not(effects.YouControl())).WithCount(0, 1),
				Key:     "Tolsimir, Friend to Wolves — you gain 3 life and that Wolf fights up to one target creature you don't control",
				Effect:  tolsimirWolfFights,
			},
		},
	})
}

// tolsimirWolfFights gains the life, then has the Wolf that entered fight the
// chosen creature.
func tolsimirWolfFights(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.GainLife{Player: item.Controller, Amount: 3}).Apply(ctx); err != nil {
		return err
	}
	if item.Trigger == nil {
		return nil
	}
	wolf := item.Trigger.Event.CardID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && wolf != uuid.Nil && wolf != t.ID {
			return effects.B10Fight(ctx, wolf, t.ID)
		}
	}
	return nil
}
