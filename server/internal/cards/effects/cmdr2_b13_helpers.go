package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2B13DesertLand is the shared front of the Commander Deserts with a
// "{T}, Pay 1 life: Add <colour>" mana ability and a "Sacrifice a Desert"
// ability: the printed {T}: Add {C}, the life-cost coloured mana
// (CR 118.4, a cost paid before the land taps), and the cash-in row the
// caller supplies. The Desert sacrificed may be this land or another.
func cmdr2B13DesertLand(oracleID, name, colour string, cashIn ActivatedAbility) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "Add {C}"},
			{Cost: ManaAbilityCost{Tap: true, Life: 1}, Produced: "{" + colour + "}", Label: "{T}, Pay 1 life: Add {" + colour + "}"},
		},
		Activated: []ActivatedAbility{cashIn},
	}
}

// cmdr2B13SacrificeADesert is the "Sacrifice a Desert" cost component.
func cmdr2B13SacrificeADesert() game.AbilityCost {
	return SacrificeN(1, "a Desert", HasSubtype("Desert"))
}

// cmdr2B13Animate is the shared body of "this land becomes a P/T <colours>
// <subtype> creature [with <keywords>] until end of turn. It's still a
// land." (CR 611.2a, CR 205.1b): ONE continuous effect from the resolving
// ability pinned to the land (CR 611.2c, CR 400.7): the creature type and
// subtype (layer 4), colours when printed (layer 5), keywords (layer 6)
// and the base power and toughness (layer 7b). Nothing is removed, so the
// land stays a land and keeps its mana ability, which the engine refuses
// while it is a summoning-sick creature (CR 302.6).
func cmdr2B13Animate(ctx *Context, target uuid.UUID, label string, power, toughness int, subtype string, colours []string, keywords ...string) error {
	mods := []game.Mod{
		game.AddTypesMod("Creature"),
		game.AddSubtypesMod(ctx.TextSubtypes(subtype)...),
	}
	if len(colours) > 0 {
		mods = append(mods, game.SetColorsMod(ctx.TextColors(colours...)...))
	}
	if len(keywords) > 0 {
		mods = append(mods, game.AddKeywordsMod(keywords...))
	}
	return ScopedEffectFor{
		Target:   target,
		Mods:     append(mods, game.SetBasePTMods(power, toughness)...),
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    label,
	}.Apply(ctx)
}

// cmdr2B13LandCardToYourGraveyard is "a land card is put into your
// graveyard from anywhere": any move into a graveyard, from any zone (a
// battlefield exit, a discard, a mill), of a land card its controller
// owns. The watcher listens to both zone-motion event kinds the engine
// emits for a graveyard landing (EventLTB for the battlefield,
// EventZoneMove for every other zone) and reads the card after it lands.
func cmdr2B13LandCardToYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand() && c.Owner == source.Controller
}

// cmdr2B13OneOrMoreLandCardsToYourGraveyard is the "one or more land
// cards" form of the same trigger: one batch fires it once (CR 603.2c).
func cmdr2B13OneOrMoreLandCardsToYourGraveyard(label string, effect Effect) game.TriggeredAbility {
	return OncePerBatch(OnAny([]game.EventKind{game.EventMill, game.EventZoneMove, game.EventLTB}, cmdr2B13LandCardToYourGraveyard, label, effect))
}
