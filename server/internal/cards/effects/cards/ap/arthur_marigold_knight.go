package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arthur, Marigold Knight — Legendary Creature — Mouse Knight {2}{U}{R}{W},
// 4/5 (Bloomburrow Commander, #5368):
//
//	"Haste
//	 Whenever Arthur and at least one other creature attack, look at the
//	 top six cards of your library. You may put a creature card from among
//	 them onto the battlefield tapped and attacking. Put the rest on the
//	 bottom of your library in a random order. Return that creature to its
//	 owner's hand at end of combat."
//
// Winota's look-and-put with the creature filter widened, behind a
// battalion-style attack condition (Arthur plus at least one other attacker
// you control). The creature is not declared an attacker (CR 506.3c). The
// controller picks what it attacks when there is a choice (CR 508.4). A
// delayed trigger at the end of combat returns it to its owner's hand if it
// is still the same object on the battlefield.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8d6ce07f-b21b-42eb-a3d6-df810664e2be",
		Name:            "Arthur, Marigold Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclared(ev, source) && effects.AttackersYouControlOtherThan(g, source.Controller, source.InstanceID) >= 1
			}, "Arthur, Marigold Knight — look at the top six cards; you may put a creature card onto the battlefield tapped and attacking; return it to its owner's hand at end of combat",
				arthurLook),
		},
	})
}

func arthurLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 6)
	if len(looked) == 0 {
		return nil
	}
	var creatures []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() && !effects.IsToken(c) {
			creatures = append(creatures, id)
		}
	}
	if len(creatures) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Arthur, Marigold Knight — you may put a creature card from among them onto the battlefield tapped and attacking (the rest go to the bottom in a random order)",
		Cards:    creatures,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  arthurPicked,
		Carry:    looked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var arthurPicked = effects.AskedPick("choose-cards/cmdr2-s18-arthur-marigold-knight-creature", func(ctx *effects.Context, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if len(rest) > 0 {
		if err := ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, rest); err != nil {
			return err
		}
	}
	if picked == uuid.Nil {
		return nil
	}
	return effects.FinB11PutAttackingThen(ctx, picked, game.ZoneLibrary, "Arthur, Marigold Knight", arthurEntered, arthurAttackPicked)
})

var arthurAttackPicked = effects.OptionPickThen("option-pick/cmdr2-s18-arthur-marigold-knight-attack", func(ctx *effects.Context, r game.OptionPicked) error {
	card, target, err := effects.FinB11PickedAttack(r)
	if err != nil {
		return nil
	}
	return effects.FinB11PutAttacking(ctx.Game, ctx.Controller(), card, game.ZoneLibrary, target, ctx.Tail(arthurEntered))
})

// arthurEntered schedules the end-of-combat return of what entered.
var arthurEntered = effects.EnteredTail("cmdr2-s18-arthur-marigold-knight/entered", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(entered)
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:     game.StepEndCombat,
		Label:  "Arthur, Marigold Knight — return that creature to its owner's hand",
		Cards:  []uuid.UUID{entered},
		Body:   effects.ZaraReturnToHandBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})
