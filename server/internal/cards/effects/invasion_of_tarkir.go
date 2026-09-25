package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of Tarkir — Battle — Siege, defense 5, for {1}{R}:
//
//	"When this Siege enters, reveal any number of Dragon cards from
//	 your hand. When you do, this Siege deals X plus 2 damage to any
//	 other target, where X is the number of cards revealed this way.
//	 (X can be 0.)
//	 (As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)"
//
// Two mana for a Shock that scales with a Dragon hand, on a body the
// table has to chew through before you get the 4/4 flier back. The
// bracketed reminder text is engine behaviour keyed on the card type
// (server/internal/game/battle.go), not card-effect data.
//
// "(X can be 0.)" is the load-bearing parenthesis: revealing nothing
// still satisfies "when you do", so the reflexive trigger is created
// either way and the damage's floor is 2. That is why the target
// clause is mandatory rather than conditional.
//
// TWO STACK ITEMS, as printed (#636). "When you do" is a CR 603.12
// reflexive trigger: the entry trigger resolves, the controller
// reveals what they like, and only then does the damage go on the
// stack — with its target chosen at that point (CR 603.3d), knowing
// X, and with a window for the table to respond to it on its own.
// (riveteers_overlook.go and the b08 land family print the same
// shape.)
//
// ONE SANDBOX SIMPLIFICATION, declared: the reveal is a CHOICE from
// the Dragon cards in hand rather than from the whole hand. Printed
// you may reveal any number of Dragon cards, so the legal picks are
// identical — what is lost is the bluff of revealing nothing while
// holding Dragons, which the prompt still permits (Min 0), and the
// ability to reveal a card that is a Dragon only because of a
// type-changing effect the prompt's candidate scan reads through
// HasSubtype anyway.
func init() {
	Register(Spec{
		OracleID:     invasionOfTarkirOracleID,
		Name:         "Invasion of Tarkir",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			SiegeTransformedCastCaveat,
		},
		Battle: &BattleSpec{
			Defense: 5,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Tarkir — reveal any number of Dragon cards", invasionOfTarkirReveal),
			DefeatedTrigger("Invasion of Tarkir — defeated: exile it, then cast Defiant Thundermaw", SiegeDefeated()),
		},
	})
}

// invasionOfTarkirOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfTarkirOracleID = "5c7f02ad-1daf-4d1a-bef0-0b2064f9b67e"

// tarkirAnyOtherTarget is TargetAny with the card's own label.
//
// "any OTHER target" is enforced at RESOLUTION rather than in the
// clause, which is Screaming Nemesis's posture (b35TargetAnyOther): a
// target spec is built once per card and cannot name the instance it
// hangs off. A Siege chosen as its own target is skipped rather than
// damaged.
func tarkirAnyOtherTarget() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any other target"
	return spec
}

// invasionOfTarkirReveal is the ETB trigger's body: offer the Dragon
// cards in the controller's hand, reveal what they pick, then create
// the reflexive trigger that deals the damage.
//
// Package-level rather than a closure so it captures nothing — the
// prompt's continuation outlives this call, and a closure over a
// *Card would be a pointer into a zone slice that reallocates.
//
// A controller with no Dragon cards in hand skips the prompt and goes
// straight to the trigger with X = 0. QueueChooseCardsForEffect
// deliberately does not short-circuit an empty candidate set (its doc
// says why), so the skip has to happen here — and it is safe
// precisely because the continuation is a named function both paths
// call rather than a body one of them would have to duplicate.
//
// Caller holds g.mu (this runs from a resolving trigger).
func invasionOfTarkirReveal(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	dragons := dragonCardsInHand(g, controller)
	if len(dragons) == 0 {
		return invasionOfTarkirQueueDamage(g, item, nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   item.SourceCardID,
		Question: "Invasion of Tarkir — reveal any number of Dragon cards from your hand",
		Cards:    dragons,
		Min:      0,
		// Max 0 means "all of them" — which is what "any number" is.
		Max:  0,
		Zone: game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return invasionOfTarkirQueueDamage(g, item, picked)
		},
	})
	return nil
}

// invasionOfTarkirQueueDamage reveals the picks and puts the printed
// "when you do" on the stack, carrying the revealed cards as its
// payload — X is their count, and it is fixed here even though the
// damage happens later.
//
// It runs unconditionally, including for an empty reveal: "(X can be
// 0.)" is the card saying so in as many words.
//
// Caller holds g.mu.
func invasionOfTarkirQueueDamage(g *game.Game, item *game.StackItem, revealed []uuid.UUID) error {
	ctx := NewContext(g, item)
	if len(revealed) > 0 {
		if err := (RevealCards{
			Player: item.Controller,
			Cards:  revealed,
			Reason: "Invasion of Tarkir — Dragon cards revealed",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return ReflexiveTrigger{
		Label: "Invasion of Tarkir — X plus 2 damage to any other target",
		Cards: revealed,
		Body:  tarkirDamageBody,
	}.Apply(ctx)
}

// invasionOfTarkirDamage is the reflexive half: X plus 2 to the
// target chosen when this trigger went on the stack, where X is the
// number of cards its payload records as revealed.
//
// Caller holds g.mu.
func invasionOfTarkirDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 {
		return nil
	}
	t := item.Targets[0]
	// "any OTHER target": the Siege cannot damage itself, and the
	// check is here because the clause could not make it.
	if t.ID == item.SourceCardID {
		return nil
	}
	if !ctx.IsTargetLegal(t) {
		return nil
	}
	return DealDamage{
		Source: item.SourceCardID,
		Target: t.ID,
		Amount: len(ctx.PayloadCards()) + 2,
	}.Apply(ctx)
}

// dragonCardsInHand returns the Dragon cards in a player's hand, in
// hand order. Subtypes are read through game.Card.HasSubtype, so a
// changeling in hand counts — which is what the printed card says,
// since a card in hand has its printed characteristics and
// changeling is one of them.
//
// Caller holds g.mu.
func dragonCardsInHand(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(playerID)
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasSubtype("Dragon") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
