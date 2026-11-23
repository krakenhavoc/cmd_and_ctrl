package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mor_b1_helpers.go — shared rows for the Morningtide batch (#5592).

// morReinforce is "Reinforce N—<cost> (<cost>, Discard this card: Put N
// +1/+1 counters on target creature.)" (CR 702.77a): Fowl Strike's
// reinforce row, an activated ability that works only from the hand.
// Discarding the card is part of the cost, so it is in the graveyard
// before the ability resolves, and the counters go on a target still
// legal at resolution (CR 608.2b). label is the printed line.
func MorReinforce(n int, cost, label string) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    Plus(ManaCost(cost), DiscardThis()),
		Zones:   []game.ZoneKind{game.ZoneHand},
		Targets: TargetCreature("target creature"),
		Effect:  RenewCountersOnTarget(game.CounterPlusOne, n),
	}
}

// morDrawIfYouControlAWizard is the Wizard spells' "If you control a
// Wizard, draw a card", read as the spell resolves (Mana Sculpt's
// ControlsA: effective subtypes, so a changeling counts).
func MorDrawIfYouControlAWizard(ctx *Context) error {
	if !ControlsA(ctx.TextSubtype("Wizard"))(ctx.Game, ctx.Controller()) {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
}

// morDrawIfProwled is "If this spell's prowl cost was paid, draw a card":
// the alternative cost announced with the cast (CR 601.2b), read back as
// the spell resolves.
func MorDrawIfProwled(ctx *Context) error {
	if !ctx.PaidAltCost("prowl") {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
}

// morMoreLifeThanAnOpponent is "if you have more life than an opponent":
// more than at least one opponent still in the game.
func MorMoreLifeThanAnOpponent(g *game.Game, you uuid.UUID) bool {
	me := g.PlayerByIDForEffect(you)
	if me == nil {
		return false
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if me.Life > p.Life {
			return true
		}
	}
	return false
}

// morCastWithProwl is "if its prowl cost was paid" on a permanent: its
// cast provenance names the prowl cost (CR 400.7d), the way Spiders-Man,
// Heroic Horde reads web-slinging.
func MorCastWithProwl(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.AltCost == "prowl"
}
