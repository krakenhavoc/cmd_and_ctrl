package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b157_helpers.go — shared bodies of the GAP batch 157 cards
// (#5948): Mardu Heart-Piercer's raid damage, Plundering Predator's loot and
// Magmatic Channeler's exile-two-and-choose.

// gapB157RaidDamage is "Raid — When this creature enters, if you attacked this
// turn, this creature deals N damage to any target." Raid is an intervening
// "if" (CR 603.4): checked as the trigger would go on the stack and again as
// it resolves; the creature is the source of the damage, as it last existed
// if it has left (CR 608.2h).
func gapB157RaidDamage(name string, n int) game.TriggeredAbility {
	return TriggerWithPurpose(Targeting(
		On(game.EventETB, AllOf(Self, tmcB6YouAttackedThisTurn),
			name+" — Raid: damage to any target",
			Do(If{Cond: avatarB11CondYouAttacked{}, Then: []Applier{
				Instruction(func(ctx *Context) error {
					return sourceDealsDamageToEachLegalTarget(n)(ctx.Game, ctx.Item)
				}),
			}})),
		TargetAny()), ForTargets(DamageToTarget(0, n)))
}

// gapB157MayDiscardThenDraw is "you may discard a card. If you do, draw a
// card": the discard is asked first, and only a card actually discarded draws.
func gapB157MayDiscardThenDraw(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
	}
}

// gapB157ExileTwoChooseOnePlay is "Exile the top two cards of your library,
// then choose one of them. You may play that card this turn.": both cards are
// exiled face up, the controller picks one, and only it gets the window.
func gapB157ExileTwoChooseOnePlay(ctx *Context) error {
	ids := ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), 2)
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(gapB157ExileChoose))
}

// gapB157ExileChoose is "choose one of them" over the cards that reached
// exile. An on-disk key: never renamed, never reused.
var gapB157ExileChoose = CardTail("then/s82-gap-b157-choose-one-play", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	name := "Choose one of them"
	if src := ctx.Source(); src != uuid.Nil {
		if c, ok := ctx.Game.LookupCardForEffect(src); ok {
			name = c.Name
		}
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: name + " — choose one of them (you may play it this turn)",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  gapB157ExilePlayGrant,
	})
	return nil
})

// gapB157ExilePlayGrant gives the chosen card its window: its chooser may play
// it this turn. An on-disk key.
var gapB157ExilePlayGrant = game.RegisterChooseCardsThen("choose-cards/s82-gap-b157-play-this-turn", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   r.Chooser,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfTurnDuration(),
			Label:    "Play it this turn",
		})
	}
	return nil
})

// gapB157TargetsCantBlockThisTurn is "<up to N> target creatures can't block
// this turn": every target still legal as the ability resolves (CR 608.2b)
// is restricted until end of turn.
func gapB157TargetsCantBlockThisTurn(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
