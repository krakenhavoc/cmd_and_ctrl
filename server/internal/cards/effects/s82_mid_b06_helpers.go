package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b06_helpers.go — shared pieces of the Innistrad: Midnight Hunt
// vampires (S82 batch 6, #5497).

// s82MidB06VampiresYouControl lists the Vampire creatures a player controls,
// read through the layered view as the caller runs.
//
// vampire is the word "Vampire" as the source's text reads it (CR 612.2).
func s82MidB06VampiresYouControl(g *game.Game, player uuid.UUID, vampire string) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.HasSubtype(vampire) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// s82MidB06OtherVampiresYouControlEnterWithACounter is Vampire Socialite's
// "as long as an opponent lost life this turn, each other Vampire you control
// enters with an additional +1/+1 counter on it": a CR 614.1c replacement on
// the entry of another Vampire under the source's controller, applying only
// while an opponent of that controller has lost life this turn.
func s82MidB06OtherVampiresYouControlEnterWithACounter(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			if opponentLifeLostThisTurn(g, src.Controller) == 0 {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature() && entering.HasSubtype(src.TextSubtype("Vampire")) && entering.Controller == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB("+1/+1", 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// s82MidB06CondAnOpponentLostLife is avB34CondAnOpponentLostLife with its
// read declared (FootprintCondOpponentLostLife), for a one-sentence trigger
// whose row must carry a footprint (Arrogant Outlaw's drain).
type s82MidB06CondAnOpponentLostLife struct{}

func (s82MidB06CondAnOpponentLostLife) Holds(ctx *Context) bool {
	return opponentLifeLostThisTurn(ctx.Game, ctx.Controller()) > 0
}

func (s82MidB06CondAnOpponentLostLife) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondOpponentLostLife, 1), true
}
