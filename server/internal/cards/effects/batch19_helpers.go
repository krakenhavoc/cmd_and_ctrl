package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch19_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 19 (#312, `edhrec_rank` 2032–2132). Own file per
// the #231 convention; every package-level name carries the b19
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "a creature died this turn" is
// b11CreaturesDiedThisTurn, the artifacts a player controls are
// b03ArtifactsControlled, the land cards in a graveyard are
// b17LandCardsInGraveyard, "sacrifice a land:" is b08SacrificeALand,
// the permanent a cost sacrificed is b17PermanentSacrificedToPay,
// the Gnome is b17GnomeToken, the Zombie is BlackZombieToken, the
// tapped Treasure is tappedTreasureToken, the permanents a player
// controls are b11PermanentsControlled, "not the source, by name" is
// b03NotNamed, and the Guildgate is a row in guildgates.go.

// --- token templates ---------------------------------------------

// --- board reads -------------------------------------------------

// b19ZombieCardsInGraveyard counts the Zombie cards in `player`'s
// graveyard — Diregraf Colossus's entry count. A card in a graveyard
// has no layer cache, so this is the printed type line; a changeling
// counts, as printed.
func b19ZombieCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype("Zombie") {
			n++
		}
	}
	return n
}

// --- trigger conditions ------------------------------------------

// b19ZombieSpellCastByYou is Diregraf Colossus's "whenever you cast
// a Zombie spell" — b17ElfSpellCastByYou with the tribe swapped. The
// spell is read off the stack, where its printed type line is intact;
// a changeling counts.
func b19ZombieSpellCastByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype("Zombie")
}

// b19AnotherNontokenCreatureYouControlDied is Liesa's condition:
// b17AnotherNontokenCreatureDied narrowed to the source's controller.
func b19AnotherNontokenCreatureYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && !IsToken(dead) && dead.Controller == source.Controller
}

// --- replacements ------------------------------------------------

// b19EntersWithCountersCounted is b10EntersWithCounters with the
// count computed as the permanent enters — Diregraf Colossus's "a
// +1/+1 counter for each Zombie card in your graveyard". `count`
// receives the entering card (still the stack object, so its
// Controller is the caster) and runs inside the replacement, so a
// Zombie that hit the graveyard in response is counted, as printed.
func b19EntersWithCountersCounted(kind string, count func(g *game.Game, src *game.Card) int, label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			ev.AddCounterAtETB(kind, count(g, src))
			return nil
		},
		Label: label,
	}
}

// --- effect bodies -----------------------------------------------

// b19ExileTopTwoUntilEndOfNextTurn is the shared body of Reckless
// Impulse and Wrenn's Resolve: "Exile the top two cards of your
// library. Until the end of your next turn, you may play those
// cards."
//
// The engine's impulse grant expires by ROUND number (Turn.Number
// advances when the table wraps to seat 0), and the cleanup sweep
// runs on every seat's cleanup, so no single UntilTurn value means
// "the end of your next turn" for every seat: the round of your next
// turn ends the grant at seat 0's cleanup of that round, which for
// anyone but seat 0 is BEFORE their turn, and one round later would
// leave it live through opponents' turns after yours. So the grant
// is stamped two rounds out as a backstop and a CR 603.7 delayed
// trigger at the beginning of your next upkeep re-stamps it to end
// with that turn — the first cleanup after your upkeep in a round is
// your own, which is exactly "until the end of your next turn". The
// re-stamp is ordinary catalog machinery (a delayed trigger on the
// stack in your upkeep, visible and respondable), and it touches only
// cards still in exile that still carry this grant.
func b19ExileTopTwoUntilEndOfNextTurn(label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		controller := item.Controller
		exiled, err := g.ExileTopWithPermissionForEffect(controller, controller, 2, game.ExilePlayPermission{
			UntilTurn: g.Turn.Number + 2,
		})
		if err != nil {
			return err
		}
		if len(exiled) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:                 game.StepUpkeep,
			ControllerTurnOnly: true,
			Label:              label + " — the exiled cards may be played until end of turn",
			Cards:              exiled,
			Effect:             b19EndImpulseGrantWithThisTurn,
		}.Apply(NewContext(g, item))
	}
}

// b19EndImpulseGrantWithThisTurn is the delayed-trigger body behind
// b19ExileTopTwoUntilEndOfNextTurn: every listed card that is still
// in exile under the item controller's grant has the grant re-stamped
// to lapse at this turn's cleanup. A card already played, or exiled
// again by something else since, is left alone — re-stamping goes
// through ExileCardWithPermissionForEffect, which would MOVE a card
// that is anywhere but exile, so the zone check is not optional.
// Package-level so the delayed trigger captures nothing.
func b19EndImpulseGrantWithThisTurn(g *game.Game, item *game.StackItem) error {
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		z := g.FindCardZoneForEffect(t.ID)
		if z == nil || z.Kind != game.ZoneExile {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || c.ExilePlay.Player != item.Controller || c.ExilePlay.WhileExiled {
			continue
		}
		perm := c.ExilePlay
		perm.UntilTurn = g.Turn.Number
		if err := g.ExileCardWithPermissionForEffect(t.ID, perm); err != nil {
			return err
		}
	}
	return nil
}

// b19BirthingPodSearch is Birthing Pod's body: the sacrificed
// creature's mana value plus one is the only mana value the search
// accepts. The sacrifice was paid at announce
// (b17PermanentSacrificedToPay reads it back off the log) and the
// card is looked up where it now sits — its printed mana cost
// survives the move, and a token's is zero, so a sacrificed token
// fetches a one-drop as printed. With no sacrifice on record the
// ability does nothing.
func b19BirthingPodSearch(g *game.Game, item *game.StackItem) error {
	sacrificed, ok := b17PermanentSacrificedToPay(g, item)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(sacrificed)
	if !ok {
		return nil
	}
	want := c.ManaValue() + 1
	return SearchLibrary{
		Player: item.Controller,
		Predicate: func(c game.Card) bool {
			return c.IsCreature() && c.ManaValue() == want
		},
		Dest:    game.ZoneBattlefield,
		Limit:   1,
		Shuffle: true,
		Reason:  "Birthing Pod — a creature card with mana value " + strconv.Itoa(want),
	}.Apply(NewContext(g, item))
}
