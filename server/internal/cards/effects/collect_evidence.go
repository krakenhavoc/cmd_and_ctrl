package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// collect_evidence.go — #5193: collect evidence N (CR 701.59), the card
// side of game/collect_evidence.go. Three forms, one clause underneath
// (game.CollectEvidenceCost, an any-number graveyard exile whose cards'
// total mana value must be N or greater), so one gate decides who may
// pay for the engine, the view's offers and the bot (ExileCost.FloorMet):
//
//	OptionalCosts: []game.AdditionalCost{OptionalCollectEvidence(6)},  // Crimestopper Sprite
//	Cost: Plus(TapCost(), CollectEvidenceN(3)),                         // Forensic Researcher
//	RegisterCollectEvidence("collect-evidence/izoni-spiders", 4, …)     // Izoni, Center of the Web
//
// "If evidence was collected" is ctx.EvidenceCollected() on a resolving
// spell and ThisEvidenceCollected on a permanent's own enters trigger.

// OptionalCollectEvidence is "As an additional cost to cast this spell,
// you may collect evidence N" (CR 701.59a). The caster who takes the offer
// names graveyard cards with total mana value N or greater on cast_spell's
// exile_ids; they are exiled with the spell on the stack and recorded on
// the paid cost. The offer can't be taken when the graveyard's whole total
// is below N (CR 701.59b).
func OptionalCollectEvidence(n int) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:   true,
		Key:        game.CollectEvidenceKey,
		ExileCards: game.CollectEvidenceCost(n, collectEvidenceLabel(n)),
		Label:      "Collect evidence " + fmt.Sprint(n),
	}
}

// CollectEvidenceN is "Collect evidence N" inside a CR 602 activated
// ability's cost, composed with Plus: "{T}, Collect evidence 3: Tap target
// creature you don't control" (Forensic Researcher). The activator names
// the cards as they activate (activate_ability's exile_ids); with too
// little in the graveyard the ability can't be activated.
func CollectEvidenceN(n int) game.AbilityCost {
	return game.AbilityCost{ExileCards: game.CollectEvidenceCost(n, collectEvidenceLabel(n))}
}

// collectEvidenceLabel is the picker's words for the clause, without the
// verb ("Exile …").
func collectEvidenceLabel(n int) string {
	return fmt.Sprintf("cards from your graveyard with total mana value %d or greater", n)
}

// EvidenceCollected is "if evidence was collected" on a spell whose
// optional additional cost is OptionalCollectEvidence: the caster took the
// offer when they cast it (CR 701.59c, 607.2i). False when they declined,
// and for every ability item. A copy of the spell reads the original's
// announcement (CR 707.10).
func (c *Context) EvidenceCollected() bool {
	return c.OptionalCostTimes(game.CollectEvidenceKey) > 0
}

// ThisEvidenceCollected is "if evidence was collected" as a permanent's
// own trigger condition — Crimestopper Sprite's and Vitu-Ghazi Inspector's
// "When this creature enters, … if evidence was collected": the spell it
// resolved from was cast with its optional collect evidence cost paid
// (game.CardEvidenceCollected).
func ThisEvidenceCollected(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && game.CardEvidenceCollected(*source)
}

// WhenThisEntersEvidence is "When this creature enters, <effect>" where
// the effect reads "if evidence was collected" (Crimestopper Sprite's "tap
// target creature. If evidence was collected, put a stun counter on it").
// The answer is read off the permanent as the trigger is put on the stack
// (game.CardEvidenceCollected, CR 400.7d) and stamped on the item, so it
// holds after the creature leaves; the effect reads it with
// ItemEvidenceCollected. Set Targets on the returned ability for a
// targeted body.
func WhenThisEntersEvidence(label string, effect Effect) game.TriggeredAbility {
	t := WhenThisEnters(label, effect)
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if source != nil && game.CardEvidenceCollected(*source) {
			item.Params.Amount = 1
		}
		return item
	}
	return t
}

// ItemEvidenceCollected is the stamp WhenThisEntersEvidence put on its
// item: the permanent's spell was cast with evidence collected.
func ItemEvidenceCollected(item *game.StackItem) bool {
	return item != nil && item.Params.Amount > 0
}

// checkOptionalCollectEvidence refuses the shapes of an optional collect
// evidence cost that would compile and then charge something the card
// does not print:
//
//   - an optional cost keyed CollectEvidenceKey that is not exactly a
//     collect evidence clause (OptionalCollectEvidence), or one with any
//     other component beside it;
//   - two of them, or one beside a mandatory exile-cards cost or branch:
//     a cast names its graveyard cards on one exile_ids list.
func checkOptionalCollectEvidence(spec Spec) {
	seen := 0
	for _, oc := range spec.OptionalCosts {
		if oc.Key != game.CollectEvidenceKey {
			continue
		}
		seen++
		rest := oc
		rest.ExileCards, rest.Key, rest.Label, rest.Optional = nil, "", "", false
		if !oc.ExileCards.IsCollectEvidence() || oc.Repeat > 1 || oc.MinX != 0 || !rest.Empty() ||
			rest.Targets != nil || rest.AsThoughFlash || !rest.Waterbend.Empty() {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q is a malformed collect evidence cost — build it with OptionalCollectEvidence", spec.Name, oc.Key))
		}
	}
	if seen > 1 {
		panic(fmt.Sprintf("effects.Register: %q declares %d optional collect evidence costs — a cast names its cards on one exile_ids list", spec.Name, seen))
	}
	if seen == 0 || spec.AdditionalCost == nil {
		return
	}
	for _, b := range append([]game.AdditionalCost{*spec.AdditionalCost}, spec.AdditionalCost.Either...) {
		if b.ExileCards != nil {
			panic(fmt.Sprintf("effects.Register: %q pairs an optional collect evidence with a mandatory graveyard exile — they share exile_ids", spec.Name))
		}
	}
}

// CostsLessIfEvidenceCollected is "This spell costs {n} less to cast if
// evidence was collected" (Bite Down on Crime), a SelfCostModifiers entry:
// the collect evidence offer is announced at CR 601.2b, so the total at
// 601.2f reads it off the announcement (game.CostQuery.OptionalCosts).
func CostsLessIfEvidenceCollected(n int, label string) game.CostModifier {
	return CostsLess(n, label, func(q game.CostQuery) bool {
		return q.OptionalCostTimes(game.CollectEvidenceKey) > 0
	})
}

// CollectEvidence is "you may collect evidence N. If you do, …" as the
// EFFECT of a resolving ability (Izoni, Center of the Web; Sample
// Collector's "When you do"). Hold one only from RegisterCollectEvidence.
//
// Ask offers the controller one choose_cards prompt over their graveyard
// (game.CollectEvidenceOptionsForEffect): any number of cards, none
// included, whose total mana value must be N or greater — a registered set
// rule (game.CollectEvidenceMet), so a restored prompt refuses a pick that
// falls short. Choosing none declines. A graveyard whose whole total is
// below N asks nothing (CR 701.59b: the player can't choose to collect
// evidence). The chosen cards are exiled, and then the rest of the card
// runs, told which cards were exiled. The choice is not a target.
type CollectEvidence struct {
	n    int
	pick game.ChooseCardsThen
}

// RegisterCollectEvidence registers "you may collect evidence N. If you do,
// <then>" under key ("choose-cards/<card>-collect-evidence"; an on-disk
// identity, ledgered: never renamed, never reused). Call it once, from a
// package-level var. `then` runs after the cards are exiled and only when
// evidence was collected; `exiled` is the cards, in the order chosen.
func RegisterCollectEvidence(key string, n int, then func(ctx *Context, exiled []uuid.UUID) error) CollectEvidence {
	if n <= 0 || then == nil {
		panic(fmt.Sprintf("effects.RegisterCollectEvidence: %q needs N ≥ 1 and a body", key))
	}
	return CollectEvidence{n: n, pick: askedPickRule(key,
		func(picked []game.Card) bool { return game.CollectEvidenceMet(picked, n) },
		func(ctx *Context, r game.CardsChosen) error {
			if len(r.Picked) == 0 {
				return nil
			}
			for _, id := range r.Picked {
				if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return then(ctx, append([]uuid.UUID(nil), r.Picked...))
		})}
}

// Ask is the offer, made as the ability resolves (CR 608.2c). `question` is
// the prompt's header.
func (ce CollectEvidence) Ask(ctx *Context, question string) error {
	if ce.n <= 0 {
		return fmt.Errorf("%w: collect evidence of %d", game.ErrInvalidParam, ce.n)
	}
	cards := ctx.Game.CollectEvidenceOptionsForEffect(ctx.Controller(), uuid.Nil, ce.n)
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Item.SourceCardID,
		Question: question,
		Cards:    cards,
		Min:      0,
		Max:      len(cards),
		LiveZone: game.ZoneGraveyard,
		ThenKey:  ce.pick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}
