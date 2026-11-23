package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dreadfulApathyOracle is Dreadful Apathy's oracle ID: it prints Cooped Up's
// text and shares its exile body (coopedUpExileEnchanted).
const DreadfulApathyOracle = "9c41b4ad-a73f-4274-9b71-9c9834910e05"

// coopedUpExileEnchanted is "Exile enchanted creature", read through the
// Aura's last-known attachment when it has left (CR 608.2h).
func CoopedUpExileEnchanted(g *game.Game, item *game.StackItem) error {
	id := CoopedUpHost(g, item.SourceCardID)
	if id == uuid.Nil {
		return nil
	}
	return ExileTarget{Target: id}.Apply(NewContext(g, item))
}

// tormentOfScarabsUpkeep is the enchanted player's choice: lose 3 life,
// sacrifice a nonland permanent, or discard a card.
func TormentOfScarabsUpkeep(g *game.Game, item *game.StackItem) error {
	victim := item.Trigger.Event.Actor
	p := g.PlayerByIDForEffect(victim)
	if p == nil || p.Eliminated {
		return nil
	}
	options := []game.ChoiceOption{{Label: tormentLoseLabel}}
	if len(NonlandPermanentsControlledBy(g, victim)) > 0 {
		options = append(options, game.ChoiceOption{Label: tormentSacLabel})
	}
	if p.Hand != nil && p.Hand.Size() > 0 {
		options = append(options, game.ChoiceOption{Label: tormentDiscardLabel})
	}
	return PickOption{
		Player:   victim,
		Question: "Torment of Scarabs — lose 3 life, sacrifice a nonland permanent, or discard a card?",
		Options:  options,
		ThenKey:  tormentOfScarabsAnswer,
	}.Apply(NewContext(g, item))
}

const (
	tormentLoseLabel    = "Lose 3 life"
	tormentSacLabel     = "Sacrifice a nonland permanent"
	tormentDiscardLabel = "Discard a card"
)

// tormentOfScarabsAnswer is the answer. On-disk keys: never renamed, never
// reused. A refusal to answer loses the life.
var tormentOfScarabsAnswer = OptionPickThen("option-pick/torment-of-scarabs-answer", func(ctx *Context, r game.OptionPicked) error {
	victim := r.Chooser
	if r.Option != nil {
		switch r.Option.Label {
		case tormentSacLabel:
			return SacrificeChoice{
				Player:     victim,
				Candidates: NonlandPermanentsControlledBy(ctx.Game, victim),
				Question:   "Torment of Scarabs — sacrifice a nonland permanent",
			}.Apply(ctx)
		case tormentDiscardLabel:
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   r.Source,
				N:        1,
				Question: "Torment of Scarabs — discard a card",
			}, tormentOfScarabsDiscarded, nil)
		}
	}
	return ctx.Game.ChangePlayerLifeForEffect(r.Source, victim, -3)
})

// tormentOfScarabsDiscarded: nothing follows the discard.
var tormentOfScarabsDiscarded = game.RegisterDiscardThen("choose-cards/torment-of-scarabs-discarded", func(*game.Game, game.DiscardsDone) error {
	return nil
})

// eidolonOfAstralWindsPump is the constellation effect: base 4/4 and flying
// until end of turn for the target creature you control.
func EidolonOfAstralWindsPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	const label = "Eidolon of Astral Winds — base power and toughness 4/4 and flying until end of turn"
	if err := UntilEndOfTurn(ctx, target, nil, label, game.SetBasePTMods(4, 4)...); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: target, Keywords: []string{"flying"}, Label: label}.Apply(ctx)
}

// smokeShroudReturn returns Smoke Shroud from the graveyard attached to the
// Ninja that entered.
func SmokeShroudReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ninja := item.Trigger.Event.CardID
	if !OnBattlefield(g, ninja) {
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	card, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !CanBeAttachedTo(g, item.Controller, card, ninja) {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{item.SourceCardID},
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: ninja},
	}.Apply(ctx)
}

// constellationTriggers is the condition of a constellation trigger:
// "whenever this creature or another enchantment you control enters".
func ConstellationTriggers(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := EnteredUnderYourControl(ev, source, g, true)
	return ok && c.IsEnchantment()
}
