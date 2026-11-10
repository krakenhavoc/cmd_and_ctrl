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
			append([]CardPredicate{trekB36OwnedByNthOpponent(n)}, what...)...).WithCount(0, 1))
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

// cmdrB13PowerAndToughness is a creature's power and toughness as it is on
// the battlefield (the raw power, so a negative one is not clamped), or as it
// last existed there once it has left (CR 608.2h). ok is false when nothing
// is known of it.
func cmdrB13PowerAndToughness(g *game.Game, id uuid.UUID) (power, toughness int, ok bool) {
	if onBattlefield(g, id) {
		g.RecomputeLayersIfStaleLocked()
		if c, found := g.LookupCardForEffect(id); found {
			return c.PowerForComparison(), c.CurrentToughness(), true
		}
	}
	if info, found := g.LastKnownPermanentForEffect(id); found {
		return info.Power, info.Toughness, true
	}
	return 0, 0, false
}

// cmdrB13DragonCombatDamageToPlayer reports whether ev is combat damage dealt
// to a player by a Dragon controlled by `controller`. "Dragon" is a word of the
// reader's text (CR 612.2), so a text-changing effect on it is honoured.
func cmdrB13DragonCombatDamageToPlayer(ev game.Event, reader *game.Card, g *game.Game, controller uuid.UUID) bool {
	if !combatDamageToPlayerBy(ev, controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && game.ThroughTextOf(reader, c).HasSubtype(reader.TextSubtype("Dragon"))
}

// cmdrB13EntersTappedUnlessRevealOrControlDragon is "This land enters tapped
// unless you revealed a Dragon card this way or you control a Dragon" (Temple
// of the Dragon Queen): the reveal-land replacement
// (EntersTappedUnlessYouRevealFromHand) that does not apply at all, and so asks
// nothing, while the entering permanent's controller controls a Dragon. The
// Dragon is a word of the land's text (CR 612.2), read off the battlefield
// through the layers.
func cmdrB13EntersTappedUnlessRevealOrControlDragon(name string) game.ReplacementEffect {
	r := EntersTappedUnlessYouRevealFromHand(name, "a Dragon card", func(c game.Card) bool {
		return c.HasSubtypeWord("Dragon")
	})
	r.RowLabel = "Enters tapped unless you reveal a Dragon card from your hand or you control a Dragon."
	base := r.AppliesTo
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !base(ev, g, src) {
			return false
		}
		controller := EnteringPermanentChooser(ev, g, src)
		dragon := "Dragon"
		if src != nil {
			dragon = src.TextSubtype("Dragon")
		}
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.IsCreature() && c.HasSubtype(dragon) {
				return false
			}
		}
		return true
	}
	return r
}

// loseDefenderApply is the layer 6 Apply of "can attack as though they didn't
// have defender" (Ghalta the Immovable, Felothar the Steadfast): the affected
// creatures lose the keyword while the static applies, which is observably the
// same thing (a creature with defender still blocks).
func loseDefenderApply(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	kept := c.Abilities[:0:0]
	for _, a := range c.Abilities {
		if a != "defender" {
			kept = append(kept, a)
		}
	}
	c.Abilities = kept
}
