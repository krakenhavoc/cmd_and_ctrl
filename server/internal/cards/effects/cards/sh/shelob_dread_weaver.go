package cardssh

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shelob, Dread Weaver — {3}{B} Legendary Creature — Spider Demon 3/3
//
//	"Whenever a nontoken creature an opponent controls dies, exile it.
//	 {2}{B}, Put a creature card exiled with Shelob into its owner's
//	 graveyard: Put two +1/+1 counters on Shelob. Draw a card.
//	 {X}{1}{B}: Put target creature card with mana value X exiled with
//	 Shelob onto the battlefield tapped under your control."
//
// The three abilities are linked (CR 607.2a): the trigger exiles the dead
// creature card LINKED to this Shelob object (Game.ExileLinkedForEffect),
// and both activated abilities read only the cards so linked. A card that
// left the graveyard in response is a new object and is not exiled
// (CR 400.7, DeadCardInGraveyard). The control test is the creature's
// controller as it last existed (CR 603.10a).
//
// The second ability's cost is the #6092 seam: the processing cost
// component narrowed to the cards exiled with the paying object, any
// owner's (ProcessCardExiledWithThis). The third targets a card in exile
// exiled with this object whose mana value equals the announced X (Gelatinous
// Cube's shape) and puts it onto the battlefield tapped under Shelob's
// controller's control.
//
// No simplification.
const shelobDiesLabel = "Shelob, Dread Weaver — exile the creature that died"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "7eade76f-92e4-4d26-9eb8-d3d954845c04",
		Name:         "Shelob, Dread Weaver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			Key:     shelobDiesLabel,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B33OpponentsNontokenCreatureDied(ev, source, g)
			},
			Build:  effects.EventCardTriggerItem(shelobDiesLabel),
			Effect: shelobExileTheDead,
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label: "{2}{B}, Put a creature card exiled with Shelob into its owner's graveyard: Put two +1/+1 counters on Shelob. Draw a card.",
				Cost: effects.Plus(effects.ManaCost("{2}{B}"), effects.ProcessCardExiledWithThis("a creature card exiled with Shelob",
					func(c game.Card) bool { return c.IsCreature() })),
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerPump},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.SourceIsNewObject(g, item) && effects.OnBattlefield(g, item.SourceCardID) {
						if err := g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, 2); err != nil {
							return err
						}
					}
					return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "{X}{1}{B}: Put target creature card with mana value X exiled with Shelob onto the battlefield tapped under your control.",
				Cost:    effects.ManaCost("{X}{1}{B}"),
				Targets: effects.ExiledWithThis(effects.TargetCardInExile("target creature card with mana value X exiled with Shelob", effects.Creature()).WithManaValueEqualsX()),
				Effect:  shelobReanimate,
			},
		},
	})
}

// shelobExileTheDead exiles the creature card that died, linked to this
// Shelob, if it is still that object in a graveyard.
func shelobExileTheDead(g *game.Game, item *game.StackItem) error {
	id, ok := effects.DeadCardInGraveyard(g, item)
	if !ok {
		return nil
	}
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return g.ExileCardForEffect(id)
	}
	return g.ExileLinkedForEffect(ref, id)
}

// shelobReanimate puts the targeted exiled creature card onto the
// battlefield tapped under the ability's controller's control.
func shelobReanimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		_, err := g.ReturnFromExileToBattlefieldForEffect(t.ID, ctx.Controller(), true)
		if errors.Is(err, game.ErrCardNotFound) {
			return nil
		}
		return err
	}
	return nil
}
