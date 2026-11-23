package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Codie, Vociferous Codex — Legendary Artifact Creature — Book Construct
// {3}, 1/4 (Special Guests, S82 BNS):
//
//		"You can't cast permanent spells.
//		 {4}, {T}: Add {W}{U}{B}{R}{G}. When you next cast a spell this turn,
//		 exile cards from the top of your library until you exile an instant or
//		 sorcery card with lesser mana value. Until end of turn, you may cast that
//		 card without paying its mana cost. Put each other card exiled this way on
//		 the bottom of your library in a random order."
//
//	  - "You can't cast permanent spells" is a cast restriction on Codie's
//	    controller only (Grid Monitor's shape), read from the battlefield. The
//	    "can't" wins over any permission (CR 101.2), so Codie's own free cast
//	    of a permanent is moot (it only ever finds instants and sorceries).
//	  - The second ability is a mana ability with a mana cost (CR 605) whose
//	    rider schedules a delayed triggered ability (CR 603.7), Ether's shape.
//	    It belongs to the ability, so it outlives Codie, ends at cleanup, and
//	    does not care whether the mana was spent on that spell.
//	  - "Lesser mana value" is lesser than the spell that was cast, X counted
//	    (CR 202.3e), read off the triggering spell. The walk, the random bottom
//	    of the rest, and the hit's {0} cast-only permission until end of turn are
//	    game.ExileUntilCastFreeThisTurnForEffect: the hit stays in exile at its
//	    ordinary timing, so a sorcery waits for the stack to empty.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67a9357c-4713-4e01-a60d-532bf0dd80b6",
		Name:         "Codie, Vociferous Codex",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.YouCantCastUnless("You can't cast permanent spells.", effects.Permanent(), nil),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{4}"},
			Produced: "{W}{U}{B}{R}{G}",
			Label:    "{4}, {T}: Add {W}{U}{B}{R}{G}. When you next cast a spell this turn, exile cards from the top of your library until you exile an instant or sorcery card with lesser mana value. Until end of turn, you may cast that card without paying its mana cost. Put each other card exiled this way on the bottom of your library in a random order.",
			Rider:    codieVociferousNextSpell,
		}},
	})
}

// codieVociferousExileBody is the delayed trigger's body.
var codieVociferousExileBody = game.SimpleDelayedBody("codie-vociferous-codex/exile-until-lesser-instant-or-sorcery", codieVociferousExile)

// codieVociferousNextSpell schedules the delayed trigger: the next spell of
// any kind its controller casts this turn.
func codieVociferousNextSpell(g *game.Game, controller, source uuid.UUID) error {
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: source,
		Label:        "Codie, Vociferous Codex — exile until an instant or sorcery with lesser mana value",
		On:           []game.EventKind{game.EventCast},
		Condition:    effects.YouNextCastCondition,
		Body:         codieVociferousExileBody,
	})
	return nil
}

// codieVociferousExile is the body: the walk, bottoming and permission.
func codieVociferousExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	cast := ctx.PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	spell, ok := g.LookupCardForEffect(cast[0])
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(spell)
	return g.ExileUntilCastFreeThisTurnForEffect(item.Controller, item.SourceCardID, func(c game.Card) bool {
		if !c.IsInstant() && !c.IsSorcery() {
			return false
		}
		cmv, ok := c.ParsedManaValue()
		return ok && cmv < mv
	})
}
