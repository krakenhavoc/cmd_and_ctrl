package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b1_helpers.go — the shared shapes of the Jumpstart Dimir deck
// request's first batch (#4892): Barter in Blood's "each player
// sacrifices two creatures", Kels's "whenever you sacrifice a creature",
// and the Bogbrew Witch / Festering Newt / Bubbling Cauldron names.

// jmpB1EachPlayerSacrificesN is "each player sacrifices N creatures of
// their choice" (CR 701.21a), Liliana, Dreadhorde General's −4 shape:
// every living player in APNAP order makes ONE choice of N creatures, and
// each player's choice is sacrificed together. A player with fewer than N
// creatures sacrifices what they have; one with none is not asked. The
// choice is not a target, so hexproof is irrelevant.
func jmpB1EachPlayerSacrificesN(n int, question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		var legs []game.ChooseCardsPrompt
		source := item.SourceCardID
		spec := sacrificeSpec("a creature", Creature())
		for _, seat := range apnapLivingSeats(g) {
			if !g.CanEffectCauseSacrificeForEffect(seat) {
				continue
			}
			candidates := g.EffectSacrificeCandidatesForEffect(seat, spec)
			if len(candidates) == 0 {
				continue
			}
			k := min(n, len(candidates))
			legs = append(legs, game.ChooseCardsPrompt{
				Chooser: seat, FromPlayer: seat, Source: source,
				Question: question,
				Cards:    candidates, Min: k, Max: k, Zone: game.ZoneBattlefield,
				Destination: game.ChooseSacrifice,
			})
		}
		return g.ChoosePermanentsInOrderSacrificeForEffect(source, legs)
	}
}

// jmpB1YouSacrificedACreature is "whenever you sacrifice a creature": the
// sacrifice event's actor is the source's controller and the sacrificed
// permanent is a creature. EventSacrifice fires while the permanent is
// still on the battlefield, so its type is read as it last stood, and a
// source that sacrifices itself sees its own sacrifice (a leaves-the-
// battlefield trigger looks back in time).
func jmpB1YouSacrificedACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// jmpB1YouControlACreatureNamed reports whether `controller` controls a
// creature named `name` right now.
func jmpB1YouControlACreatureNamed(g *game.Game, controller uuid.UUID, name string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && game.HasName(c, name) {
			return true
		}
	}
	return false
}

// jmpB1NamedFesteringNewt is "named Festering Newt" as a card predicate.
func jmpB1NamedFesteringNewt(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return game.HasName(c, "Festering Newt")
}
