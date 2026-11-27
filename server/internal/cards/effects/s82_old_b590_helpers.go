package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b590_helpers.go — Saviors of Kamigawa, batch 590 (#5601): the
// stop clause of Rally the Horde's repeated triple exile and the edict that
// takes several lands from every player.

// untilAGroupOfThreeEndsOnALandRef is Rally the Horde's "exile the top card of
// your library three times. If the last card exiled isn't a land card, repeat
// this process": the run stops once a whole group of three has landed and the
// last card of it is a land. Cards are counted as they landed in exile, so the
// groups line up unless a replacement keeps one out of exile.
var untilAGroupOfThreeEndsOnALandRef = game.RegisterMillUntil("until/s82-old-b590-group-of-three-ends-on-a-land", func(landed []game.Card, _ int) bool {
	n := len(landed)
	return n > 0 && n%3 == 0 && landed[n-1].IsLand()
})

// OldB590UntilAGroupOfThreeEndsOnALand is that stop clause.
var OldB590UntilAGroupOfThreeEndsOnALand = game.MillUntil{Ref: untilAGroupOfThreeEndsOnALandRef}

// OldB590EachPlayerSacrificesN is "each player sacrifices N <things> of their
// choice" (CR 701.21a): every living player in APNAP order makes ONE choice
// of N matching permanents they control (fewer if they have fewer; none is
// not asked), and all the choices are sacrificed together. JmpB1EachPlayerSacrificesN
// with the kind of permanent as a parameter; the choice is not a target.
func OldB590EachPlayerSacrificesN(n int, label string, match CardPredicate, question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if n <= 0 {
			return nil
		}
		var legs []game.ChooseCardsPrompt
		source := item.SourceCardID
		spec := SacrificeSpec(label, match)
		for _, seat := range ApnapLivingSeats(g) {
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

// OldB590CardsExiledThatAreNonland counts the cards among `ids` that are not
// land cards, as they now sit in exile.
func OldB590CardsExiledThatAreNonland(g *game.Game, ids []uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			n++
		}
	}
	return n
}
