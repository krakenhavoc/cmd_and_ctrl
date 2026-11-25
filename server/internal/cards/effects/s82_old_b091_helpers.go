package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b091_helpers.go — Zendikar Rising damage cards, S82 OLD batch 91 (#5508).

// S82OldB091PartyCostReduction is "This spell costs {1} less to cast for
// each creature in your party." (CR 700.8), counted as the spell is cast.
func S82OldB091PartyCostReduction() game.CostModifier {
	return CostsLessEach(func(q game.CostQuery) int {
		if q.Game == nil {
			return 0
		}
		return FinB9PartySize(q.Game, q.Controller)
	}, "This spell costs {1} less to cast for each creature in your party.")
}

// S82OldB091CreatureYouControlDealtCombatDamageToAPlaneswalker is "whenever
// a creature you control deals combat damage to a planeswalker" (Zagras).
func S82OldB091CreatureYouControlDealtCombatDamageToAPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	if !ok || !dealer.IsCreature() || dealer.Controller != source.Controller {
		return false
	}
	pw, ok := g.LookupCardForEffect(ev.Target)
	return ok && pw.IsPlaneswalker() && OnBattlefield(g, ev.Target)
}

// S82OldB091ZarethSanAbility is Zareth San's "{2}{U}{B}, Return an
// unblocked attacking Rogue you control to its owner's hand: Put this card
// from your hand onto the battlefield tapped and attacking." — ninjutsu's
// body and timing-is-the-cost posture (the cost is the restriction: only an
// unblocked attacker can be returned), with the returned creature narrowed
// to a Rogue, and without being a ninjutsu ability (CR 702.49).
func S82OldB091ZarethSanAbility() ActivatedAbility {
	return ActivatedAbility{
		Label: "{2}{U}{B}, Return an unblocked attacking Rogue you control to its owner's hand: " +
			"Put this card from your hand onto the battlefield tapped and attacking.",
		Cost: Plus(ManaCost("{2}{U}{B}"),
			ReturnAPermanentToHand("an unblocked attacking Rogue you control", Creature(), HasSubtype("Rogue"), UnblockedAttacker())),
		Zones:   []game.ZoneKind{game.ZoneHand},
		Purpose: game.Purpose{Answers: game.AnswerSurvive},
		Effect:  ninjutsuEnter,
	}
}

// S82OldB091DamageEachOpponentEffect deals n damage to each opponent from
// the source, as one damage instance.
func S82OldB091DamageEachOpponent(ctx *Context, source uuid.UUID, n int) error {
	return OriginOfThorDealToEachOpponent(ctx, source, n)
}

// S82OldB091FifteenAboveStartingLife is "you have at least 15 life more than
// your starting life total" (Angel of Destiny), the table's starting total
// being its configured one.
func S82OldB091FifteenAboveStartingLife(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Life >= StartingLifeOf(g)+15
}

// S82OldB091DestroyTheDamagedPermanent is "destroy that permanent" for the
// permanent a damage event named as its target (Termination Facilitator,
// Zagras): one that has left the battlefield is not there to destroy.
func S82OldB091DestroyTheDamagedPermanent(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.Target
	if !OnBattlefield(g, id) {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(NewContext(g, item))
}
