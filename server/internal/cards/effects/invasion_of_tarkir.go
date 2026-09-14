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
// still satisfies "when you do", so the damage is never skipped and
// its floor is 2. That is why the target is a real, mandatory target
// clause on the ETB trigger rather than something conditional.
//
// TWO SANDBOX SIMPLIFICATIONS, both weaker than printed, both
// declared.
//
//  1. The reveal and the damage happen in ONE resolution instead of
//     two. Printed, "when you do" is a reflexive trigger: it goes on
//     the stack above the ETB trigger after the reveal, and its
//     target is chosen then, knowing X. Here the target is chosen as
//     the ETB trigger goes on the stack, before the player has
//     decided what to reveal, and nobody gets priority between the
//     reveal and the damage. Choosing a target with less information
//     is strictly worse for the caster, and losing the priority
//     window means an opponent cannot respond to the reflexive half
//     — which they also cannot benefit from, since the only thing
//     they could do with it is save the target they already knew
//     about. The engine has no reflexive-trigger constructor; the
//     nearest shape is Ziatora's two-ability gate on a resolving
//     ability's label, which is a heavier mechanism than this card
//     earns. (riveteers_overlook.go and the b08 land family made the
//     same fold for the same reason.)
//
//  2. The reveal is a CHOICE from the Dragon cards in hand rather
//     than from the whole hand. Printed you may reveal any number of
//     Dragon cards, so the legal picks are identical — what is lost
//     is the bluff of revealing nothing while holding Dragons, which
//     the prompt still permits (Min 0), and the ability to reveal a
//     card that is a Dragon only because of a type-changing effect
//     the prompt's candidate scan reads through HasSubtype anyway.
func init() {
	Register(Spec{
		OracleID:     invasionOfTarkirOracleID,
		Name:         "Invasion of Tarkir",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			SiegeTransformedCastCaveat,
			"You pick the target before you choose which Dragons to reveal, and nobody gets to respond in between — the two halves happen as one.",
		},
		Battle: &BattleSpec{
			Defense: 5,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				// "any OTHER target" is enforced at RESOLUTION rather
				// than in the clause, which is Screaming Nemesis's
				// posture (b35TargetAnyOther): a target spec is built
				// once per card and cannot name the instance it hangs
				// off. A Siege chosen as its own target is skipped
				// rather than damaged.
				Targets: tarkirAnyOtherTarget(),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Invasion of Tarkir — reveal Dragons, then X+2 damage",
						invasionOfTarkirReveal)
				},
			},
			DefeatedTrigger("Invasion of Tarkir — defeated: exile it, then cast Defiant Thundermaw", SiegeDefeated()),
		},
	})
}

// invasionOfTarkirOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfTarkirOracleID = "5c7f02ad-1daf-4d1a-bef0-0b2064f9b67e"

// tarkirAnyOtherTarget is TargetAny with the card's own label.
func tarkirAnyOtherTarget() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any other target"
	return spec
}

// invasionOfTarkirReveal is the ETB trigger's body: offer the Dragon
// cards in the controller's hand, reveal what they pick, then deal
// that many plus two.
//
// Package-level rather than a closure so it captures nothing — the
// prompt's continuation outlives this call, and a closure over a
// *Card would be a pointer into a zone slice that reallocates.
//
// A controller with no Dragon cards in hand skips the prompt and goes
// straight to the two damage. QueueChooseCardsForEffect deliberately
// does not short-circuit an empty candidate set (its doc says why),
// so the skip has to happen here — and it is safe precisely because
// the continuation is a named function both paths call rather than a
// body one of them would have to duplicate.
//
// Caller holds g.mu (this runs from a resolving trigger).
func invasionOfTarkirReveal(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	dragons := dragonCardsInHand(g, controller)
	if len(dragons) == 0 {
		return invasionOfTarkirDamage(g, item, nil)
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
			return invasionOfTarkirDamage(g, item, picked)
		},
	})
	return nil
}

// invasionOfTarkirDamage reveals the picks and deals len(picked)+2 to
// the chosen target.
//
// Caller holds g.mu.
func invasionOfTarkirDamage(g *game.Game, item *game.StackItem, revealed []uuid.UUID) error {
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
		Amount: len(revealed) + 2,
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
