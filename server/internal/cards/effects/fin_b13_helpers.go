package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b13_helpers.go — shared words for Final Fantasy batches 13 and 14
// (#4462).

// thisPermanentsAbility is a cost predicate for "this <permanent>'s
// <abilities>": the activated ability being priced belongs to the very
// permanent the modifier sits on. An activation's q.Card is its source
// permanent, and q.Source is the object that contributes the modifier.
func thisPermanentsAbility() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.InstanceID == q.Source.InstanceID
	}
}

// WheneverYouTapANonlandPermanentForMana is "Whenever you tap a nonland
// permanent for mana, …" (Kinnan, Bonder Prodigy): a triggered mana
// ability (CR 605.1b) over a {T} mana ability of a nonland permanent the
// source's controller controls. The permanent is read as it was when it
// produced.
func WheneverYouTapANonlandPermanentForMana(
	label string,
	produced func(prod game.ManaProduced, source *game.Card, g *game.Game) string,
) game.ManaTrigger {
	return game.ManaTrigger{
		Label: label,
		AppliesTo: func(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
			return !prod.Source.IsLand() && prod.Controller == source.Controller
		},
		Produced: produced,
	}
}

// firstLegendaryCreatureSpellYouCastThisTurn is the cost predicate for
// "the first legendary creature spell you cast each turn" (Serah Farron):
// the spell being priced is a legendary creature spell and the caster has
// cast no legendary creature spell yet this turn (the tally is bumped after
// a spell is priced, so none so far means this one is the first).
func firstLegendaryCreatureSpellYouCastThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.IsCreature() && q.Card.IsLegendary() &&
			(q.Game == nil || q.Game.CastTallyFor(q.Controller).LegendaryCreature == 0)
	}
}

// finB13EachOpponentSacrificesN is "each opponent sacrifices N creatures of
// their choice": one pick of exactly min(N, creatures) per opponent, in
// turn order, all made before any of them is sacrificed (the shape of
// Barter in Blood, over opponents only).
func finB13EachOpponentSacrificesN(n int, question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		var legs []game.ChooseCardsPrompt
		source := item.SourceCardID
		spec := sacrificeSpec("a creature", Creature())
		for _, seat := range apnapLivingSeats(g) {
			if seat == item.Controller || !g.CanEffectCauseSacrificeForEffect(seat) {
				continue
			}
			candidates := g.EffectSacrificeCandidatesForEffect(seat, spec)
			if len(candidates) == 0 {
				continue
			}
			k := min(n, len(candidates))
			legs = append(legs, game.ChooseCardsPrompt{
				Chooser: seat, FromPlayer: seat, Source: source,
				Question: question,
				Cards:    candidates, Min: k, Max: k, Zone: game.ZoneBattlefield,
				Destination: game.ChooseSacrifice,
			})
		}
		return g.ChoosePermanentsInOrderSacrificeForEffect(source, legs)
	}
}

// controlsAtLeastOtherLegendaryCreatures is "you control N or more other
// legendary creatures", read through the layered view.
func controlsAtLeastOtherLegendaryCreatures(g *game.Game, controller, except uuid.UUID, n int) bool {
	count := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != except && c.Controller == controller && c.IsCreature() && c.IsLegendary() {
			count++
		}
	}
	return count >= n
}
