package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coveted Jewel — Artifact {6}:
//
//	"When this artifact enters, draw three cards.
//	 {T}: Add three mana of any one color.
//	 Whenever one or more creatures an opponent controls attack you
//	 and aren't blocked, that player draws three cards and gains
//	 control of this artifact. Untap it."
//
// The mana ability is Gilded Lotus's OneColorOfAmount(3).
//
// The steal trigger watches EventBlockersDeclared, the event #1279
// added for CR 509.3's "attacks and isn't blocked" (attacks_unblocked.go):
// one per defending player per combat, emitted once that player's
// blocks are locked in. The Jewel's controller is the defender whose
// declaration it watches, and it triggers when at least one creature an
// opponent controls is attacking THAT PLAYER — not a planeswalker or a
// battle they protect — and nothing was declared blocking it. One event
// per declaration is the "one or more": three unblocked attackers are
// one trigger.
//
// "That player" is the attacking player, which is the active player
// (CR 506.2) — read off the turn as the ability resolves, since the
// event names only the defender. It draws even if the Jewel has left
// by then; the control change and the untap need the Jewel still to be
// the object that triggered (CR 400.7). The control change has no
// stated duration, so it lasts until the Jewel leaves (CR 611.2a), and
// a later Jewel trigger under the new controller takes it on again.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "98492d7d-3b9e-4ae1-ac45-1b508d6d2670",
		Name:         "Coveted Jewel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Coveted Jewel — draw three cards", Do(DrawCards{N: 3})),
			{
				Watches: []game.EventKind{game.EventBlockersDeclared},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return covetedJewelAttackedAndUnblocked(ev, source, g)
				},
				Key:    "Coveted Jewel — that player draws three cards and gains control of this artifact",
				Effect: covetedJewelSteal,
			},
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: OneColorOfAmount(3),
			Label:    "Add three mana of any one color",
		}},
	})
}

// covetedJewelAttackedAndUnblocked is the trigger condition: `ev` is the
// block declaration of the Jewel's controller, and at least one
// creature an opponent controls is attacking that player directly and
// was not blocked.
//
// Caller holds the game lock (the harvester's contract).
func covetedJewelAttackedAndUnblocked(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventBlockersDeclared || source == nil {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor != you {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.AttackingTarget != you || c.Controller == you || !c.IsCreature() {
			continue
		}
		if g.UnblockedAttackerForEffect(c.InstanceID) {
			return true
		}
	}
	return false
}

// covetedJewelSteal is the trigger's resolution: the attacking player
// draws three, then takes the Jewel and untaps it.
func covetedJewelSteal(g *game.Game, item *game.StackItem) error {
	if len(g.Seats) == 0 {
		return nil
	}
	attacker := g.Seats[g.Turn.ActiveSeat]
	if attacker == nil || attacker.Eliminated {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: attacker.ID, N: 3}).Apply(ctx); err != nil {
		return err
	}
	// CR 400.7: a Jewel that left and came back is a new object the
	// trigger knows nothing about. GainControlForEffect refuses a Jewel
	// that is not on the battlefield at all.
	jewel := item.SourceCardID
	if sourceIsNewObject(g, item) {
		return nil
	}
	if !g.GainControlForEffect(jewel, jewel, attacker.ID, game.IndefiniteDuration(),
		"Coveted Jewel — gains control of this artifact") {
		return nil
	}
	return g.UntapTargetForEffect(jewel)
}
