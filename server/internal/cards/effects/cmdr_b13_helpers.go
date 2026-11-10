package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdrB13CastSpellPower is "that spell's power" for a trigger that watched
// a spell being cast: the power it has where it is now (on the stack, or in
// the graveyard if it was countered).
func cmdrB13CastSpellPower(ctx *Context) int {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return 0
	}
	spell, ok := ctx.Game.LookupCardForEffect(ctx.Item.Trigger.Event.CardID)
	if !ok {
		return 0
	}
	return ctx.Game.PowerAnywhereForEffect(spell)
}

// cmdrB13CreatureFromGraveyard is "a creature you control entered from a
// graveyard or was cast from a graveyard" for one entering creature
// (Celes, Rune Knight's reading, the source itself included).
func cmdrB13CreatureFromGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	return ev.EnteredFrom == game.ZoneGraveyard || c.Provenance.FromZone == game.ZoneGraveyard
}

// cmdrB13SpellTargetsPermanentOrPlayer reports whether the spell on the
// stack has a target that is a player or a permanent (Shiko and Narset,
// Unified). The cast path announces targets before the cast event
// (CR 601.2c), so the item already carries them. A card target counts only
// while it is on the battlefield, which at the cast is what makes it a
// permanent rather than a card in a graveyard or a spell.
func cmdrB13SpellTargetsPermanentOrPlayer(g *game.Game, spellID uuid.UUID) bool {
	item := g.StackItemForEffect(spellID)
	if item == nil {
		return false
	}
	for _, t := range item.Targets {
		switch t.Kind {
		case game.TargetPlayer, game.TargetSelf:
			return true
		case game.TargetCard:
			if z := g.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneBattlefield {
				return true
			}
		}
	}
	return false
}

// cmdrB13TotalPowerAtMost is "with total power n or less", the set rule
// TotalManaValueAtMost is for mana value (Reunion of the House).
func cmdrB13TotalPowerAtMost(n int) *game.TargetTotal {
	return &game.TargetTotal{
		Label: "have total power " + strconv.Itoa(n) + " or less",
		Stat:  game.TargetTotalPower,
		Max:   n,
	}
}

// cmdrB13OwnedByNthOpponent passes for a card owned by the n-th seat
// (1-based) after the chooser's in turn order: controlledByNthOpponent's
// twin for a card in a graveyard, which has an owner and no controller to
// speak of. Eliminated seats keep their place.
func cmdrB13OwnedByNthOpponent(n int) CardPredicate {
	return func(g *game.Game, chooser uuid.UUID, c game.Card) bool {
		p := g.OpponentSeatAfter(chooser, n)
		return p != nil && c.Owner == p.ID
	}
}

// cmdrB13PerPlayerGraveyardClauses is "for each player, choose up to one
// target <what> card in that player's graveyard" (Afterlife from the Loam):
// one "up to one" clause for the caster's own graveyard, then one per
// possible opponent, each bound to the n-th seat after the caster's. A seat
// that is not at the table, or is eliminated, offers nothing.
func cmdrB13PerPlayerGraveyardClauses(noun string, what ...CardPredicate) *game.TargetSpec {
	own := TargetCardInGraveyard("up to one target "+noun+" card in your graveyard",
		append([]CardPredicate{YouOwn()}, what...)...).WithCount(0, 1)
	var rest []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		rest = append(rest, TargetCardInGraveyard("up to one target "+noun+" card in the "+nthOpponentWords[n]+" opponent's graveyard",
			append([]CardPredicate{cmdrB13OwnedByNthOpponent(n)}, what...)...).WithCount(0, 1))
	}
	return Clauses(own, rest...)
}

// cmdrB13GreatestManaValueYouControl is the greatest mana value among
// permanents `player` controls, zero with none ({X} in a cost is 0 on the
// battlefield, CR 202.3e).
func cmdrB13GreatestManaValueYouControl(g *game.Game, player uuid.UUID) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player {
			if mv := c.ManaValue(); mv > best {
				best = mv
			}
		}
	}
	return best
}
