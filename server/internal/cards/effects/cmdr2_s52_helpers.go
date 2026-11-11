package effects

import (
	"strings"

	"github.com/google/uuid"

	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S52CombatDamageToPlayerDoubler is Felix Five-Boots' "If a creature you
// control dealing combat damage to a player causes a triggered ability of a
// permanent you control to trigger, that ability triggers an additional time"
// (CR 603.2d). The cause is a combat EventDealDamage whose source is a
// creature the doubler's controller controls (read from its last-known
// characteristics) and whose target is a player; the ability belongs to a
// permanent they control.
func cmdr2S52CombatDamageToPlayerDoubler(name string) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label:    name,
		RowLabel: "Abilities of your permanents that trigger when a creature you control deals combat damage to a player trigger an additional time.",
		Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || !q.SourceIsPermanent || q.Ability == nil || !q.HasSubject {
				return false
			}
			ev := q.Event
			if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
				return false
			}
			if q.SourceLKI.Controller != q.DoublerLKI.Controller {
				return false
			}
			return q.SubjectLKI.Controller == q.DoublerLKI.Controller && slices.Contains(q.SubjectLKI.Types, "Creature")
		},
	}
}

// cmdr2S52OpponentsBelowHalfStarting counts the source controller's living
// opponents whose life total is less than half their starting life total
// (Anya, Merciless Angel). The table's starting total is read through
// startingLifeOf; the comparison is life*2 < starting so an odd starting
// total has no rounding.
func cmdr2S52OpponentsBelowHalfStarting(g *game.Game, source *game.Card) int {
	if g == nil || source == nil {
		return 0
	}
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.ID == source.Controller || p.Eliminated {
			continue
		}
		if p.Life*2 < startingLifeOf(g) {
			n++
		}
	}
	return n
}

// cmdr2S52CardTypesSacrificedThisTurn counts the distinct card types among the
// permanents `player` sacrificed this turn (Korvold, Gleeful Glutton's cost
// reduction), from TurnTally.SacrificedCardTypes: each sacrificed permanent
// was tallied once per card type it had as it was sacrificed, tokens included.
func cmdr2S52CardTypesSacrificedThisTurn(g *game.Game, player uuid.UUID) int {
	if g == nil {
		return 0
	}
	n := 0
	for _, t := range game.ChoosableCardTypes {
		if g.SacrificedWithCardTypeThisTurn(player, strings.ToLower(t)) > 0 {
			n++
		}
	}
	return n
}

// cmdr2S52KorvoldGrow is Korvold's combat-damage effect: X is the number of
// permanent types among cards in your graveyard, read once as the trigger
// resolves; put X +1/+1 counters on Korvold and draw X cards.
func cmdr2S52KorvoldGrow(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := permanentTypesInGraveyard(g, item.Controller)
	if x <= 0 {
		return nil
	}
	if err := (CounterOnThis{Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{Player: item.Controller, N: x}.Apply(ctx)
}

// cmdr2S52DereviTapOrUntapCause is the shared condition of Derevi's one
// trigger: "When Derevi enters and whenever a creature you control deals
// combat damage to a player".
func cmdr2S52DereviTapOrUntapCause(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB {
		return Self(ev, source, lki, g)
	}
	return combatDamageToPlayerBy(ev, source.Controller, g)
}

// cmdr2S52PutThisFromCommandZone is Derevi's "Put Derevi onto the battlefield
// from the command zone": the ability's own card, if it is still the object
// whose ability this is (CR 400.7) and still in its owner's command zone, is
// put onto the battlefield under its owner's control. Putting is not casting,
// so the commander tax neither applies nor grows (CR 903.8); the commander
// designation stays with the card.
func cmdr2S52PutThisFromCommandZone(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Owner != item.Controller || z.Kind != game.ZoneCommand {
		return nil
	}
	if c, ok := g.LookupCardForEffect(id); !ok || c.ObjectEpoch != item.SourceEpoch {
		return nil
	}
	_, err := g.PutFromCommandZoneOntoBattlefieldForEffect(id, game.ZoneEntryOptions{Controller: item.Controller})
	return err
}
