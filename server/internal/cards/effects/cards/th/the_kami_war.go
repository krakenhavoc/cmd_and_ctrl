package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Kami War // O-Kagachi Made Manifest — an Enchantment — Saga
// {1}{W}{U}{B}{R}{G} (Kamigawa: Neon Dynasty, S82):
//
//	I   — Exile target nonland permanent an opponent controls.
//	II  — Return up to one other target nonland permanent to its owner's
//	      hand. Then each opponent discards a card.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// Back face (registered under "<oracle_id>#1"), O-Kagachi Made Manifest, an
// Enchantment Creature — Dragon Spirit 6/6 that is all colors (its printed
// colors, a characteristic-defining ability, CR 604.3):
//
//	Flying, trample
//	Whenever this creature attacks, defending player chooses a nonland card
//	in your graveyard. Return that card to your hand. This creature gets
//	+X/+0 until end of turn, where X is the mana value of that card.
//
// II's "other" is the Saga itself (it is not a legal target). The attack
// trigger is Rejoin the Fight's opponent-chooses-from-your-graveyard shape:
// the defending player of THIS attack (the player attacked, or the
// controller of the planeswalker or battle attacked) is the chooser, with
// no candidate meaning no choice and no pump. X is read off the card while it
// is still in the graveyard.
//
// No simplification.
func init() {
	const name = "The Kami War"
	effects.Register(effects.Spec{
		OracleID:     TheKamiWarOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "exile target nonland permanent an opponent controls"),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls()),
				effects.ExileEveryLegalCardTarget),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "return up to one other target nonland permanent to its owner's hand, then each opponent discards a card"),
				effects.Another(effects.TargetPermanent("up to one other target nonland permanent", effects.Nonland()).WithCount(0, 1)),
				theKamiWarBounceAndDiscard),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile this Saga, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        TheKamiWarOracleID + "#1",
		Name:            "O-Kagachi Made Manifest",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("O-Kagachi Made Manifest — defending player chooses a nonland card in your graveyard; return it to your hand; this creature gets +X/+0 until end of turn",
				oKagachiAttack),
		},
	})
}

// theKamiWarOracleID is shared with the back face.
const TheKamiWarOracleID = "4d23b22b-31a5-4eec-b894-2dd5a7138c5a"

// theKamiWarBounceAndDiscard is chapter II.
func theKamiWarBounceAndDiscard(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		break
	}
	return effects.EachOpponentDiscardsOne(g, item)
}

// oKagachiAttack asks the defending player to choose a nonland card in the
// attacker's graveyard.
func oKagachiAttack(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
	if defender == uuid.Nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	candidates := effects.GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool { return !c.IsLand() })
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    defender,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "O-Kagachi Made Manifest — choose a nonland card in their graveyard; they return it to their hand",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    oKagachiPick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// oKagachiPick returns the chosen card to its owner's hand and pumps the
// attacker by its mana value.
var oKagachiPick = effects.AskedPick("choose-cards/o-kagachi-made-manifest-pick", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	x := card.ManaValue()
	if err := (effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}).Apply(ctx); err != nil {
		return err
	}
	if x <= 0 {
		return nil
	}
	return effects.BoostUntilEOT{
		Target: ctx.Source(),
		Power:  x,
		Label:  "O-Kagachi Made Manifest — +X/+0 until end of turn",
	}.Apply(ctx)
})
