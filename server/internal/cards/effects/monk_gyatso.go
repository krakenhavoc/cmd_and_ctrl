package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Monk Gyatso — 3/3 Legendary Creature — Human Monk for {3}{W}:
//
//	"Whenever another creature you control becomes the target of a
//	spell or ability, you may airbend that creature. (Exile it.
//	While it's exiled, its owner may cast it for {2} rather than its
//	mana cost.)"
//
// A protection engine: someone points removal at your creature, you
// exile it in response, the removal fizzles for want of a target,
// and you rebuy the creature for {2}. The timing is the card — the
// trigger goes on the stack ABOVE the spell that targeted, so it
// resolves first. That falls out of emitting EventBecomesTarget at
// announce rather than at resolution; see the kind's doc comment in
// events.go for why that is also what CR 115.3 says.
//
// "That creature" is not a target — the trigger names it by
// reference to the event, so there is no target clause here and
// nothing to re-check at resolution beyond "is it still there".
//
// Shape notes and simplifications:
//
//   - "Another CREATURE YOU CONTROL" is checked against the live
//     battlefield when the event fires. A creature that is targeted
//     and then changes controller before the trigger resolves keeps
//     the trigger, which is correct (CR 603.4 has no intervening-if
//     here).
//   - Unlike the "other target" clauses elsewhere in this decklist,
//     "another" IS enforceable here, because the excluded object is
//     the trigger's own source and its InstanceID is in hand at
//     AppliesTo time.
//   - Targets chosen by an effect that CHANGES them after announce
//     (Deflecting Swat) do not fire this, because the engine has no
//     change-targets path at all. Noted on EventBecomesTarget.
//   - One prompt per target slot. A spell that targets two of your
//     creatures asks twice, which is right; a spell that targets the
//     same creature twice would also ask twice, which is also right
//     (CR 115.3 counts instances of the word "target").
func init() {
	Register(Spec{
		OracleID:     "ff92fa60-f0fe-496e-8155-d9d6f5af651b",
		Name:         "Monk Gyatso",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesTarget},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return targetedAnotherCreatureYouControl(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				victim := ev.CardID
				return game.NewTriggeredItem(source, "Monk Gyatso — airbend the targeted creature",
					func(g *game.Game, item *game.StackItem) error {
						// The creature may have been answered some
						// other way while the trigger sat on the
						// stack. Nothing to airbend is not an error
						// (CR 608.2c).
						if !onBattlefield(g, victim) {
							return nil
						}
						return Airbend{Target: victim}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Monk Gyatso — airbend the targeted creature? (Exile it; its owner may cast it for {2}.)",
			},
		}},
	})
}

// targetedAnotherCreatureYouControl is Gyatso's trigger condition:
// an EventBecomesTarget naming a battlefield creature that the
// source's controller controls and that is not the source itself.
//
// Player targets are excluded by the ev.CardID != uuid.Nil test —
// the event leaves CardID zero when the target is a player, which
// is the whole reason it carries the field twice.
func targetedAnotherCreatureYouControl(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.CardID == uuid.Nil || ev.CardID == source.InstanceID || g == nil {
		return false
	}
	if !onBattlefield(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller
}

// onBattlefield reports whether a card is currently a battlefield
// permanent. Card.Controller is only meaningful there — off the
// battlefield it tracks Owner — so any "you control" test has to
// establish the zone first.
func onBattlefield(g *game.Game, cardID uuid.UUID) bool {
	z := g.FindCardZoneForEffect(cardID)
	return z != nil && z.Kind == game.ZoneBattlefield
}
