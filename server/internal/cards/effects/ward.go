package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward.go — CR 702.21, the keyword ADR 0038 §7 deliberately left out
// of the targeting choke point, built where it actually belongs.
//
// The ADR's argument is worth keeping in front of the code, because
// the tempting implementation is the wrong one. Ward LOOKS like
// hexproof: both appear on a creature, both make it awkward to
// remove, and hexproof is enforced in ten lines at
// game.CanBeTargetedBy. But CR 702.21a makes ward a TRIGGERED
// ability, and the difference is observable on every play:
//
//   - A warded permanent is a perfectly LEGAL target. The removal
//     spell is announced, it goes on the stack, and the trigger goes
//     on top of it. Enforcing ward at the targeting gate would
//     refuse the announce outright — no trigger, no payment offered,
//     no response window, and no way for the spell's controller to
//     pay and have it resolve.
//   - The payer is not the ward permanent's controller. It is
//     whoever cast the spell, which is exactly the shape the
//     pay-unless and confirm prompts already have: a chooser who is
//     not the controller of the source.
//   - The trigger is an object on the stack in its own right. It can
//     be responded to, and countering IT leaves the removal spell
//     alive and unpaid-for.
//
// PARAMETERISATION. The other half of ADR 0038 §7 is that ward
// carries a cost — "ward {2}", "ward—pay 3 life", "ward—sacrifice a
// creature" — and Characteristic.Abilities is a []string of bare
// tokens with nowhere to put it. That is why ward does NOT join
// canonicalKeywords here: a bare "ward" token in the enforced table
// would tell the coverage signal (ADR 0037) that every ward card is
// implemented while the engine had no idea what to charge. The cost
// lives where the rest of a card's parameters live — on the catalog
// Spec, as an argument to the helper below — and a ward card that is
// not in the catalog still flags as unimplemented, which is the
// honest answer.
//
// SCOPE. Three printed cost shapes ship, and each one is charged by
// the prompt that already fits it:
//
//   - "Ward {N}" (Hulking Raptor) — PendingChoicePayUnless, which
//     parses a mana cost and auto-taps for it.
//   - "Ward—Pay N life" (Refraction Elemental, Sedgemoor Witch) — a
//     PendingChoiceConfirm with LifeCost declared, Sylvan Library's
//     "pay 4 life or put it back" shape (#552).
//   - "Ward—Sacrifice a creature" (Vein Ripper) — the same Confirm,
//     whose accept branch chains a PendingChoiceChooseCards over the
//     payer's own creatures.
//
// None of the three needed engine work: the Confirm and ChooseCards
// kinds take any Chooser, have bot enumerator cases in legal/, and are
// rendered by the client. What a ward cost still cannot be is a mix of
// components ("{1}, pay 1 life" is not a printed ward), and Ward()
// refuses one rather than charging half of it.

// WardCost is the payment a ward demands. Exactly one field is set;
// build it with WardMana, WardLife or WardSacrifice rather than by
// hand.
type WardCost struct {
	// Mana is "ward {N}": a mana cost string for the pay-unless prompt.
	Mana string
	// Life is "ward—pay N life".
	Life int
	// Sacrifice is "ward—sacrifice a <thing>".
	Sacrifice *WardSacrificeCost
}

// WardSacrificeCost is the "sacrifice a <thing>" half of a ward cost:
// the printed noun phrase for the prompt, and the predicate a permanent
// the payer controls must pass to be sacrificed for it.
type WardSacrificeCost struct {
	Label string
	OK    CardPredicate
}

// WardMana builds the "ward {N}" cost.
func WardMana(cost string) WardCost { return WardCost{Mana: cost} }

// WardLife builds the "ward—pay N life" cost.
func WardLife(n int) WardCost { return WardCost{Life: n} }

// WardSacrifice builds the "ward—sacrifice <label>" cost, where every
// predicate must pass: WardSacrifice("a creature", Creature()).
func WardSacrifice(label string, preds ...CardPredicate) WardCost {
	return WardCost{Sacrifice: &WardSacrificeCost{Label: label, OK: And(preds...)}}
}

// validate panics on a cost with zero or several components. It runs
// from Ward() at catalog init, so a malformed card fails at boot rather
// than charging the first component it happens to check.
func (c WardCost) validate() {
	n := 0
	if c.Mana != "" {
		n++
	}
	if c.Life > 0 {
		n++
	}
	if c.Sacrifice != nil {
		n++
	}
	if n != 1 || (c.Sacrifice != nil && c.Sacrifice.OK == nil) {
		panic(fmt.Sprintf("effects: a ward cost needs exactly one of Mana, Life or Sacrifice: %+v", c))
	}
}

// Ward builds the CR 702.21a triggered ability: "Whenever this
// permanent becomes the target of a spell or ability an opponent
// controls, counter that spell or ability unless its controller pays
// <cost>."
//
// Two details in the AppliesTo that are the rule rather than
// defensive coding:
//
//   - `ev.Actor != source.Controller` is "an OPPONENT controls".
//     Your own Giant Growth on your own warded creature does not
//     trigger it, which is the difference between ward and shroud.
//   - The trigger fires per target INSTANCE, because
//     EventBecomesTarget is emitted per target slot (CR 115.7). A
//     spell that targets the same warded creature twice triggers
//     ward twice, and the controller pays twice or the spell is
//     countered. That is correct and it falls out of the event
//     shape rather than needing anything here.
//
// The Effect reads the targeting spell off the event's StackItemID
// rather than its Source, because for an activated or triggered
// ability those differ — Source is the permanent the ability came
// from and would counter the wrong object (or nothing at all).
func Ward(cost WardCost, label string) game.TriggeredAbility {
	cost.validate()
	if label == "" {
		label = "Ward"
	}
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecomesTarget},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && ev.Actor != source.Controller
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			// Capture the two UUIDs, never the *Card or the *Game —
			// StackItem.Effect resolves against whichever Game it is
			// restored into.
			itemID, payer := ev.StackItemID, ev.Actor
			if itemID == uuid.Nil || payer == uuid.Nil {
				return nil
			}
			// The label is the card's own, unadorned: the cost is
			// already in it ("Hulking Raptor — ward {2}"), and the
			// pay-or-counter wording belongs on the PROMPT, which is
			// where the decision is actually made.
			return game.NewTriggeredItem(source, label,
				func(g *game.Game, item *game.StackItem) error {
					return wardPayOrCounter(g, item, itemID, payer, cost)
				})
		},
	}
}

// wardPayOrCounter is the trigger's resolution: offer the payment to
// the spell's controller, counter on a decline.
//
// Every cost shape treats the two ways a decline happens — answering
// "no", and being unable to pay (no mana in pool plus untapped
// sources, too little life, nothing to sacrifice) — identically, which
// is CR 118.12.
//
// A spell that has already left the stack when the trigger resolves
// (countered in response, or its own controller responded by
// sacrificing the ward permanent and the spell then fizzled) makes
// the counter a no-op: CounterTargetForEffect returns
// ErrCardNotOnStack and there is nothing to charge for either, so
// the prompt is skipped entirely rather than asking a player to pay
// for a spell that no longer exists.
func wardPayOrCounter(g *game.Game, item *game.StackItem, targetingItem, payer uuid.UUID, cost WardCost) error {
	if g.StackItemForEffect(targetingItem) == nil {
		return nil
	}
	counter := func(g *game.Game) error {
		if g.StackItemForEffect(targetingItem) == nil {
			return nil
		}
		return g.CounterTargetForEffect(targetingItem)
	}
	switch {
	case cost.Life > 0:
		return wardPayLifeOrCounter(g, item.SourceCardID, payer, cost.Life, targetingItem, counter)
	case cost.Sacrifice != nil:
		return wardSacrificeOrCounter(g, item.SourceCardID, payer, *cost.Sacrifice, targetingItem, counter)
	}
	return g.QueuePayUnlessForEffect(payer, item.SourceCardID, cost.Mana,
		"Ward — pay "+cost.Mana+" or the spell is countered", counter)
}

// wardPayLifeOrCounter is "counter it unless that player pays N life".
//
// CR 119.4: a player may pay life only if their life total is at least
// the payment. Below that the counter is the only legal outcome, so it
// happens without a prompt — an offer the engine would have to refuse
// is how a seat wedges (#544). At exactly N the payment is offered, and
// the payer loses to the state-based action if they take it; that is a
// legal play.
//
// LifeCost is declared on the Confirm so the bot's move list prices the
// accept branch; without it a bot at 3 life answers "pay" and dies
// (#547). The branch re-checks life when the answer arrives rather than
// trusting the prompt-time total.
func wardPayLifeOrCounter(g *game.Game, source, payer uuid.UUID, life int, targetingItem uuid.UUID, counter func(*game.Game) error) error {
	if p := g.PlayerByIDForEffect(payer); p == nil || p.Life < life {
		return counter(g)
	}
	g.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      payer,
		Source:       source,
		Question:     fmt.Sprintf("Ward — pay %d life, or it is countered", life),
		AcceptLabel:  fmt.Sprintf("Pay %d life", life),
		DeclineLabel: "Let it be countered",
		LifeCost:     life,
		OnAccept: func(g *game.Game) error {
			if g.StackItemForEffect(targetingItem) == nil {
				// Nothing left to pay for.
				return nil
			}
			if p := g.PlayerByIDForEffect(payer); p == nil || p.Life < life {
				return counter(g)
			}
			return g.ChangePlayerLifeForEffect(source, payer, -life)
		},
		OnDecline: counter,
	})
	return nil
}

// wardSacrificeOrCounter is "counter it unless that player sacrifices
// <a creature>": a Confirm whose accept branch chains a one-card pick
// over the payer's own matching permanents, then sacrifices the pick.
//
// Two prompts rather than one ChooseCards with a floor of zero, because
// "choose nothing" is a poor way to say "let my spell be countered",
// and the bot enumerator already treats a Confirm's decline as the
// always-legal answer. A payer with nothing to sacrifice cannot pay, so
// the spell is countered without a prompt.
//
// The candidates are the payer's permanents, not targets: hexproof and
// protection don't stop a player sacrificing their own creature, so the
// predicate is the whole test. Both links re-read the battlefield when
// their answer arrives, and a pick that is no longer a matching
// permanent the payer controls counts as not paying.
func wardSacrificeOrCounter(g *game.Game, source, payer uuid.UUID, sac WardSacrificeCost, targetingItem uuid.UUID, counter func(*game.Game) error) error {
	if len(wardSacrificeCandidates(g, payer, sac)) == 0 {
		return counter(g)
	}
	g.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      payer,
		Source:       source,
		Question:     "Ward — sacrifice " + sac.Label + ", or it is countered",
		AcceptLabel:  "Sacrifice " + sac.Label,
		DeclineLabel: "Let it be countered",
		OnAccept: func(g *game.Game) error {
			if g.StackItemForEffect(targetingItem) == nil {
				return nil
			}
			candidates := wardSacrificeCandidates(g, payer, sac)
			if len(candidates) == 0 {
				return counter(g)
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  payer,
				Source:   source,
				Question: "Ward — choose " + sac.Label + " to sacrifice",
				Cards:    candidates,
				Min:      1,
				Max:      1,
				Zone:     game.ZoneBattlefield,
				Then: func(g *game.Game, picked []uuid.UUID) error {
					if g.StackItemForEffect(targetingItem) == nil {
						return nil
					}
					if len(picked) != 1 || !wardSacrificeLegal(g, payer, sac, picked[0]) {
						return counter(g)
					}
					return g.SacrificePermanentForEffect(picked[0])
				},
			})
			return nil
		},
		OnDecline: counter,
	})
	return nil
}

// wardSacrificeCandidates lists the permanents payer controls that the
// sacrifice cost accepts, in battlefield order.
func wardSacrificeCandidates(g *game.Game, payer uuid.UUID, sac WardSacrificeCost) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == payer && sac.OK(g, payer, c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// wardSacrificeLegal re-checks one pick against the live battlefield.
func wardSacrificeLegal(g *game.Game, payer uuid.UUID, sac WardSacrificeCost, id uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id {
			return c.Controller == payer && sac.OK(g, payer, c)
		}
	}
	return false
}
