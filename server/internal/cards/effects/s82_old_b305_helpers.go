package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b305_helpers.go — shared shapes for the Battle for Zendikar Devoid
// cards of S82 batch 305 (#5547).

// OldB305OpponentExileCards are the cards an opponent of `you` owns that sit
// in exile: what "a card an opponent owns from exile" can name (a token has
// ceased to exist, CR 111.7).
func OldB305OpponentExileCards(g *game.Game, you uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Owner != uuid.Nil && c.Owner != you && !c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// OldB305Ingest is the ingest keyword (CR 702.115a): "whenever this creature
// deals combat damage to a player, that player exiles the top card of their
// library".
func OldB305Ingest(name string) game.TriggeredAbility {
	return WheneverThisDealsCombatDamageToAPlayer(name+" — ingest: that player exiles the top card of their library",
		func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			return MillToZone{Player: item.Trigger.Event.Target, N: 1, To: game.ZoneExile}.Apply(NewContext(g, item))
		})
}

// OldB305WhenThisEntersMayProcess is the Eldrazi Processor trigger "When this
// creature enters, you may put a card an opponent owns from exile into that
// player's graveyard. If you do, <then>". `targets` is the trigger's target
// clause when `then` has one (chosen as the ability goes on the stack, CR
// 603.3d, so with no legal target it is never put on the stack); nil when it
// has none. The question is asked only when such a card exists; the pick
// lands in its owner's graveyard (a replacement or a commander's choice may
// pause it) and `then` runs only for a card that did land (CR 608.2c).
// `keyBase` names the three on-disk keys ("<keyBase>-confirm", "-picked",
// "-then"), unique to the card.
func OldB305WhenThisEntersMayProcess(name, keyBase, thenText string, targets *game.TargetSpec, then func(ctx *Context) error) game.TriggeredAbility {
	tail := CardTail("then/"+keyBase+"-then", func(ctx *Context, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return then(ctx)
	})
	picked := AskedPick("choose-cards/"+keyBase+"-picked", func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 1 {
			return nil
		}
		return ctx.Game.PutCardsIntoGraveyardKeyedForEffect(r.Picked, ctx.ItemTail(tail))
	})
	yes := MayYes("confirm/"+keyBase+"-confirm", func(ctx *Context, r game.ConfirmAnswer) error {
		cards := OldB305OpponentExileCards(ctx.Game, ctx.Controller())
		if len(cards) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: name + " — choose a card an opponent owns in exile to put into its owner's graveyard",
			Cards:    cards,
			Min:      1,
			Max:      1,
			LiveZone: game.ZoneExile,
			ThenKey:  picked,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	})
	t := WhenThisEnters(name+" — you may put a card an opponent owns from exile into that player's graveyard. If you do, "+thenText,
		func(g *game.Game, item *game.StackItem) error {
			if len(OldB305OpponentExileCards(g, item.Controller)) == 0 {
				return nil
			}
			return MayChoice{
				Question: name + " — put a card an opponent owns from exile into that player's graveyard? If you do, " + thenText + ".",
				Then:     yes,
			}.Apply(NewContext(g, item))
		})
	if targets != nil {
		t = Targeting(t, targets)
	}
	return t
}

// OldB305ControlsAnotherColorlessCreature is "you control another colorless
// creature": a creature other than `self` the controller controls that has no
// colour, read through the layers.
func OldB305ControlsAnotherColorlessCreature(g *game.Game, you, self uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.InstanceID != self && c.IsCreature() && c.IsColorless() {
			return true
		}
	}
	return false
}

// OldB305OtherColorlessAnthem is "Other colorless creatures you control get
// +p/+t": a layer 7c static over the source's other colorless creatures.
func OldB305OtherColorlessAnthem(p, t int, label string) game.StaticAbility {
	return Labelled(B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller &&
			target.IsColorless() && target.InstanceID != source.InstanceID
	}, p, t), label)
}

// OldB305IfYouControlAnotherColorless narrows a trigger on the source to the
// intervening "if you control another colorless creature" (CR 603.4): checked
// as it triggers and again as it resolves.
func OldB305IfYouControlAnotherColorless(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && OldB305ControlsAnotherColorlessCreature(g, source.Controller, source.InstanceID) &&
			(when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: "you control another colorless creature", Test: func(ctx *Context) bool {
		return OldB305ControlsAnotherColorlessCreature(ctx.Game, ctx.Controller(), ctx.Source())
	}}, t)
}
