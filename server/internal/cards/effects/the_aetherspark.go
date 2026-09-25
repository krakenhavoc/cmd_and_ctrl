package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Aetherspark — Legendary Artifact Planeswalker — Equipment {4},
// starting loyalty 4:
//
//	"As long as The Aetherspark is attached to a creature, The
//	 Aetherspark can't be attacked and has 'Whenever equipped creature
//	 deals combat damage during your turn, put that many loyalty
//	 counters on The Aetherspark.'
//	 +1: Attach The Aetherspark to up to one target creature you
//	     control. Put a +1/+1 counter on that creature.
//	 −5: Draw two cards.
//	 −10: Add ten mana of any one color."
//
// The catalog's first permanent that is a planeswalker and an
// Equipment at once, and the surprise is how little of that needed
// new machinery: both halves are ordinary and they do not touch.
// CR 606 asks "is the source a planeswalker" and the type line says
// yes; CR 301.5c and the CR 704.5n state-based action ask "is this an
// Equipment attached to a creature" and the subtype says yes. One
// permanent answers both questions without either answer knowing
// about the other.
//
// THE +1 IS THE CARD AND IT IS COMPLETE. A free, repeatable, sorcery-
// speed attach that also grows the creature it lands on: the target
// clause is "up to one target creature you control", so Min is 0 and
// the picker's confirm button is live with nothing selected — a
// Aetherspark with no creature to hold it still ticks up, exactly as
// printed. When a creature IS picked the attach and the +1/+1 counter
// both happen, in that order.
//
// The counter is deliberately NOT conditional on the attach
// succeeding. "Attach … to up to one target creature you control.
// Put a +1/+1 counter on that creature" is two sentences, and if The
// Aetherspark itself has left the battlefield in response the attach
// simply does not happen (CR 701.3b) while the creature still gets
// its counter. game.AttachSourceForEffect owns that check and reports
// it as a quiet skip rather than an error, so there is nothing to
// branch on here.
//
// A target that became illegal between announce and resolution — the
// creature died, or changed controller — is skipped whole (CR 608.2b,
// through Context.LegalTargets), which also means neither half
// happens for it. That is the rule, not a shortcut: the ability's
// only target is gone, so there is no "that creature" to put a
// counter on.
//
// THE GRANTED TRIGGER IS AN ORDINARY TRIGGER ON THIS CARD. The
// printed text grants it conditionally — "as long as The Aetherspark
// is attached to a creature, [it] has '…'" — but the ability it
// grants is on The Aetherspark itself, not on another permanent, so
// this is not the "ability granted to another permanent" seam (#754).
// A plain Spec.Triggered entry whose condition asks "am I attached to
// the creature that just dealt this damage" says the same thing: an
// unattached Aetherspark matches nothing, and it starts matching the
// instant the +1 attaches it.
//
// "That many" is read off the event as the trigger is built, so a
// pump or a shrink afterwards does not change the number of loyalty
// counters. "During your turn" is the active-player check and it is a
// real restriction — the equipped creature blocking on somebody
// else's turn deals combat damage and The Aetherspark gains nothing.
// Damage to a blocker or to a planeswalker counts the same as damage
// to a player: the printed text says "deals combat damage", with no
// recipient named.
//
// THE −10 IS A LOYALTY ABILITY, NOT A MANA ABILITY. CR 605.1a
// excludes an ability with a loyalty cost from the mana-ability
// definition, so it uses the stack, it can be responded to, and the
// mana arrives when it resolves. That is why it is an ActivatedAbility
// applying AddMana rather than an entry in Spec.ManaAbilities — the
// same call Ugin, the Spirit Dragon's sibling batch makes. "Ten mana
// of any ONE color" is one colour pick worth ten tokens
// (OneColorOfAmount), never ten separate picks, which would be a
// strictly better card.
//
// WHAT IS NOT WIRED: "can't be attacked". No creature can attack a
// planeswalker in this engine at all — DeclareAttacker rejects any
// target that is not a seated player, and the client only ever offers
// seats (ADR 0032 §7, "Absent"). So the clause has nothing to
// restrict and no bit to set. Declaring a "can't be attacked" flag
// with no attack-declaration path to read it would put a promise on
// the card that nothing enforces, which is the one thing ADR 0038 §7
// says not to do. It is a caveat instead, and it flips to a real
// restriction in the same change that teaches combat about
// planeswalkers.
//
// Printed loyalty reaches the card through deck import (ADR 0032 §1),
// so Spec.StartingLoyalty is deliberately unset: it is printed card
// data, not card-effect data.
func init() {
	Register(Spec{
		OracleID:     "483c747a-b602-4747-95bc-72c2c17ed509",
		Name:         "The Aetherspark",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Nothing can attack a planeswalker in this game yet, so \"The Aetherspark can't be attacked\" never comes up.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: equippedCreatureDealtCombatDamageOnYourTurn,
			Key:       "The Aetherspark — put that many loyalty counters on it",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AddCounter{
					Target: item.SourceCardID,
					Kind:   game.CounterLoyalty,
					N:      item.Trigger.Event.Amount,
				}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Attach The Aetherspark to up to one target creature you control. Put a +1/+1 counter on that creature.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
				Effect:  theAethersparkPlusOne,
			},
			{
				Label:  "−5: Draw two cards.",
				Cost:   LoyaltyCost(-5),
				Effect: Do(DrawCards{N: 2}),
			},
			{
				Label:  "−10: Add ten mana of any one color.",
				Cost:   LoyaltyCost(-10),
				Effect: Do(AddMana{Produced: OneColorOfAmount(10)}),
			},
		},
	})
}

// equippedCreatureDealtCombatDamageOnYourTurn is the granted
// trigger's condition: the creature this Equipment is attached to
// dealt combat damage, on its controller's turn.
//
// The attachment test does double duty. It is the "as long as The
// Aetherspark is attached to a creature" clause AND the "equipped
// creature" in the granted ability, because the only creature that
// can satisfy the second is the one the first is asking about.
func equippedCreatureDealtCombatDamageOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if !source.IsAttachedTo(ev.Source) {
		return false
	}
	return isActivePlayer(g, source.Controller)
}

// theAethersparkPlusOne is the +1: attach the source to the chosen
// creature, then put a +1/+1 counter on it. No target chosen is a
// legal, complete resolution — the loyalty counter was the whole
// point of that activation.
func theAethersparkPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.AttachSourceForEffect(item, t); err != nil {
			return err
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return nil
}
