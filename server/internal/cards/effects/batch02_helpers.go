package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 02 (#295, `edhrec_rank` 238–360). Own file, and
// every name prefixed b02, per the parallel-batch convention: three
// batches are being written at once on three branches, and a shared
// helper file or a bare name is a merge conflict waiting to happen.

// b02IsDesert is the sacrifice-cost predicate for "Sacrifice a
// Desert" (Scavenger Grounds). Post-layer subtype, exact match.
func b02IsDesert(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return hasSubtype(c, "Desert")
}

// b02CreatureCardEnteredGraveyardNotFromBattlefield is Syr Konrad's
// middle clause: "a creature card is put into a graveyard from
// anywhere other than the battlefield". Each path into a graveyard
// emits exactly one event kind — EventDiscardCard from a hand,
// EventMill from a library, EventZoneMove from the stack (a countered
// or resolved creature spell) — and every one of them stamps OldZone
// / NewZone, so one predicate covers them all and never counts a
// card twice. A death is EventLTB plus a ZoneMove FROM the
// battlefield, which the OldZone check excludes; the dies clause
// counts those.
func b02CreatureCardEnteredGraveyardNotFromBattlefield(ev game.Event, g *game.Game) bool {
	switch ev.Kind {
	case game.EventDiscardCard, game.EventMill, game.EventZoneMove:
	default:
		return false
	}
	if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// b02CreatureCardLeftYourGraveyard is Syr Konrad's last clause: "a
// creature card leaves your graveyard". "Your graveyard" is the one
// you own, so the card's Owner is the test, not its controller. The
// card is looked up post-move, wherever it went — exile, hand, the
// battlefield, a library — because the type line travels with it.
func b02CreatureCardLeftYourGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventZoneMove || ev.OldZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Owner == source.Controller
}

// b02UntapLandsYouControl untaps up to n tapped lands `controller`
// controls, in battlefield order. Snap's "untap up to two lands"
// prints no "target" and no "you control" — it is a resolution-time
// choice among every land at the table — but the only lands a
// player ever wants untapped are their own tapped ones, and the
// engine has no resolution-time pick-a-permanent prompt for a
// spell. Auto-picking the first n is never stronger than printed
// (the printed card allows exactly this outcome), only less
// controllable, and it is declared on the card.
func b02UntapLandsYouControl(ctx *Context, controller uuid.UUID, n int) error {
	untapped := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if untapped >= n {
			break
		}
		if c.Controller != controller || !c.IsLand() || !c.Tapped {
			continue
		}
		if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
			return err
		}
		untapped++
	}
	return nil
}

// b02CountLandsControlledBy counts the lands a player controls —
// Avenger of Zendikar's "for each land you control".
func b02CountLandsControlledBy(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsLand() {
			n++
		}
	}
	return n
}

// b02PlantToken is Avenger of Zendikar's 0/1 green Plant. Lives here
// rather than in tokens.go so a concurrent batch editing that file
// doesn't collide with this one.
func b02PlantToken() game.Card {
	return game.Card{
		Name:      "Plant",
		TypeLine:  "Token Creature — Plant",
		Power:     0,
		Toughness: 1,
	}
}

// b02ExileAllGraveyards is "exile all graveyards" — every seat's,
// the controller's included — through the same per-player helper
// Bojuka Bog and Farewell use.
func b02ExileAllGraveyards(g *game.Game, item *game.StackItem) error {
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		if err := exileGraveyardForEffect(g, item, p.ID); err != nil {
			return err
		}
	}
	return nil
}

// b02EachPlayerMills is "each player mills a card" (Syr Konrad's
// activated ability): every live seat, the controller included.
func b02EachPlayerMills(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := (MillCards{Player: p.ID, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b02CastInstantOrSorcery is the magecraft trigger condition: the
// controller cast an instant or sorcery. The "or copy" half is not
// modelled — no spell-copy event exists (storm_kiln_artist.go
// declares the same gap).
func b02CastInstantOrSorcery(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}

// b02PlantsYouControl lists the creatures with the Plant subtype
// `controller` controls — Avenger of Zendikar's landfall targets
// (not targeted: "each Plant creature you control").
func b02PlantsYouControl(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && hasSubtype(c, "Plant") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
