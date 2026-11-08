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
//
// #5223 adds the payoff and the X form: every collection emits
// game.EventCollectEvidence, which WheneverYouCollectEvidence watches
// (Evidence Examiner), and "you may collect evidence X. When you do, …" is
// RegisterCollectEvidenceX (Incinerator of the Guilty), X chosen as it
// resolves and gated by game.Game.CollectEvidenceXMetForEffect.

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

// CollectEvidenceFromTargetsCost is "As an additional cost to cast this
// spell, collect evidence X, where X is the total mana value of the
// permanents this spell targets" (Urgent Necropsy, #5240, CR 701.59a): the
// mandatory slot's graveyard exile with its floor read off the targets the
// cast announces (CR 601.2c), as the total cost is determined (CR 601.2f).
// The caster names any number of graveyard cards on cast_spell's exile_ids;
// CastSpell refuses a set whose total mana value is below X, and with no
// targets X is 0 and none is a payment. Every cast collects evidence, so
// "whenever you collect evidence" sees it, collect evidence 0 included.
//
// Register holds it to a spell whose every target clause is "up to", so X
// = 0 is always a legal announcement and the spell's cast gate needs no
// target search (checkCollectEvidenceFromTargets).
func CollectEvidenceFromTargetsCost() *game.AdditionalCost {
	return &game.AdditionalCost{
		ExileCards: game.CollectEvidenceFromTargetsCost("cards from your graveyard with total mana value X or greater"),
		Label:      "Collect evidence X, where X is the total mana value of the permanents this spell targets",
	}
}

// checkCollectEvidenceFromTargets refuses the shapes of a target-sized
// collect evidence cost (CollectEvidenceFromTargetsCost) the cast path does
// not pay as printed: anywhere but alone on the mandatory slot (an optional
// cost, an either/or branch and an ability have no targets to read it
// from, or none settled before their cost), on a spell with no target
// clause (X would always be 0), or on one with a clause that must be
// filled (the cast gate would then have to search the targets for a
// payable set).
func checkCollectEvidenceFromTargets(spec Spec) {
	for _, oc := range spec.OptionalCosts {
		if oc.ExileCards != nil && oc.ExileCards.ManaValueFromTargets {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q collects evidence sized by the targets — the mandatory slot's only (#5240)", spec.Name, oc.Key))
		}
	}
	ac := spec.AdditionalCost
	if ac == nil {
		return
	}
	for i, b := range ac.Either {
		if b.ExileCards != nil && b.ExileCards.ManaValueFromTargets {
			panic(fmt.Sprintf("effects.Register: %q either/or branch %d collects evidence sized by the targets — the mandatory slot's only (#5240)", spec.Name, i))
		}
	}
	ec := ac.ExileCards
	if ec == nil || !ec.ManaValueFromTargets {
		return
	}
	rest := *ac
	rest.ExileCards, rest.Label = nil, ""
	if !ec.IsCollectEvidenceFromTargets() || !rest.Empty() || rest.Key != "" || rest.Targets != nil {
		panic(fmt.Sprintf("effects.Register: %q declares a malformed target-sized collect evidence cost — build it with CollectEvidenceFromTargetsCost, alone on the mandatory slot (#5240)", spec.Name))
	}
	clauses := spec.Targets.Clauses()
	if len(clauses) == 0 || spec.Modes != nil {
		panic(fmt.Sprintf("effects.Register: %q sizes a collect evidence cost by its targets but has no card-level target clause (#5240)", spec.Name))
	}
	for i, c := range clauses {
		if c.Min != 0 || c.CountFromX {
			panic(fmt.Sprintf("effects.Register: %q target clause %d must be filled, but a target-sized collect evidence cost needs X = 0 to be reachable (every clause \"up to\", #5240)", spec.Name, i))
		}
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

// WheneverYouCollectEvidence is "Whenever you collect evidence, <effect>"
// (Evidence Examiner, Surveillance Monitor, #5223): one trigger per
// collection its controller makes, whichever form made it — a spell's
// optional cost, an ability's cost, or a resolving "you may collect
// evidence" of any card, collect evidence 0 included — off
// game.EventCollectEvidence.
func WheneverYouCollectEvidence(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventCollectEvidence, ByYou, label, effect)
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
// evidence was collected; `exiled` is the cards, in the order chosen. A nil
// `then` is "you may collect evidence N." with nothing after it (Evidence
// Examiner, Surveillance Monitor): the collection is the point, and
// "whenever you collect evidence" pays it off.
//
// The collection emits game.EventCollectEvidence once the cards are exiled
// and before `then` runs (#5223).
func RegisterCollectEvidence(key string, n int, then func(ctx *Context, exiled []uuid.UUID) error) CollectEvidence {
	if n <= 0 {
		panic(fmt.Sprintf("effects.RegisterCollectEvidence: %q needs N ≥ 1", key))
	}
	return CollectEvidence{n: n, pick: askedPickRule(key,
		func(picked []game.Card) bool { return game.CollectEvidenceMet(picked, n) },
		func(ctx *Context, r game.CardsChosen) error {
			if len(r.Picked) == 0 {
				return nil
			}
			exiled := append([]uuid.UUID(nil), r.Picked...)
			if err := collectEvidenceWith(ctx, r.Source, exiled); err != nil {
				return err
			}
			if then == nil {
				return nil
			}
			return then(ctx, exiled)
		})}
}

// collectEvidenceWith is the effect form's collection, N's and X's alike:
// each chosen card is exiled (not a target), and then the collection is
// announced for "whenever you collect evidence" (game.EventCollectEvidence),
// with the cards, before the rest of the card runs.
func collectEvidenceWith(ctx *Context, source uuid.UUID, exiled []uuid.UUID) error {
	for _, id := range exiled {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	ctx.Game.NoteCollectEvidenceForEffect(ctx.Controller(), source, exiled)
	return nil
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

// CollectEvidenceX is "you may collect evidence X. When you do, …" as the
// EFFECT of a resolving ability, X chosen as it resolves (CR 107.3f, 608.2d;
// Incinerator of the Guilty, #5223). Hold one only from
// RegisterCollectEvidenceX.
//
// Ask offers the controller, in order:
//
//  1. a yes/no: collect evidence at all. Collect evidence 0 is always
//     possible, an empty graveyard included, so the offer is always made
//     (and "whenever you collect evidence" triggers on it);
//  2. one choose_cards prompt over the whole graveyard
//     (game.CollectEvidenceXOptionsForEffect), any number, none included;
//  3. X, from 0 to the chosen cards' total mana value
//     (game.CollectEvidenceXCeilingForEffect), opening on the total; not
//     asked when the total is 0.
//
// Choosing the cards before X is the same set of outcomes the rule allows
// (any X with cards of total mana value X or greater, CR 701.59a): nothing
// can happen between the two answers. The cards are exiled once X is
// chosen, game.EventCollectEvidence is announced, and then the rest of the
// card runs, told X and the cards. Every prompt is keyed, so a table
// waiting on any of them is a restore point.
type CollectEvidenceX struct {
	may game.ConfirmThen
}

// RegisterCollectEvidenceX registers "you may collect evidence X. When you
// do, <then>" for one card. `slug` names its three on-disk keys
// ("confirm/<slug>-collect-evidence-x", "choose-cards/…" and "number/…",
// ledgered: never renamed, never reused) and `name` is the card's name for
// the prompts' headers. Call it once, from a package-level var. `then` runs
// once the cards are exiled, with X and the cards in the order chosen.
func RegisterCollectEvidenceX(slug, name string, then func(ctx *Context, x int, exiled []uuid.UUID) error) CollectEvidenceX {
	if then == nil {
		panic(fmt.Sprintf("effects.RegisterCollectEvidenceX: %q needs a body", slug))
	}
	finish := func(ctx *Context, x int, picked []uuid.UUID) error {
		if !ctx.Game.CollectEvidenceXMetForEffect(ctx.Controller(), picked, x) {
			// The cards are no longer all in the graveyard (their owner
			// left the game while asked): nothing is collected.
			return nil
		}
		if err := collectEvidenceWith(ctx, ctx.Source(), picked); err != nil {
			return err
		}
		return then(ctx, x, picked)
	}
	number := NumberThen("number/"+slug+"-collect-evidence-x", func(ctx *Context, r game.NumberAnswer) error {
		return finish(ctx, r.Number, append([]uuid.UUID(nil), r.Carry...))
	})
	pick := askedPick("choose-cards/"+slug+"-collect-evidence-x", func(ctx *Context, r game.CardsChosen) error {
		if !r.Answered {
			// The chooser left, or the pick was withdrawn: nothing is
			// collected.
			return nil
		}
		picked := append([]uuid.UUID(nil), r.Picked...)
		ceiling, ok := ctx.Game.CollectEvidenceXCeilingForEffect(ctx.Controller(), picked)
		if !ok {
			return nil
		}
		if ceiling == 0 {
			return finish(ctx, 0, picked)
		}
		return ChooseNumber{
			Min:      0,
			Max:      ceiling,
			Goal:     func(*Context) int { return ceiling },
			Question: fmt.Sprintf("%s — collect evidence X: choose X, up to %d (the total mana value of the cards you chose)", name, ceiling),
			Then:     number,
			Carry:    picked,
		}.Apply(ctx)
	})
	may := MayYes("confirm/"+slug+"-collect-evidence-x", func(ctx *Context, _ game.ConfirmAnswer) error {
		cards := ctx.Game.CollectEvidenceXOptionsForEffect(ctx.Controller())
		if len(cards) == 0 {
			return finish(ctx, 0, nil)
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: name + " — collect evidence X: choose the cards to exile from your graveyard (none is X = 0)",
			Cards:    cards,
			Min:      0,
			Max:      len(cards),
			LiveZone: game.ZoneGraveyard,
			ThenKey:  pick,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	})
	return CollectEvidenceX{may: may}
}

// Ask is the offer, made as the ability resolves (CR 608.2c). `question` is
// the yes/no's header.
func (ce CollectEvidenceX) Ask(ctx *Context, question string) error {
	return MayChoice{Question: question, Then: ce.may}.Apply(ctx)
}
