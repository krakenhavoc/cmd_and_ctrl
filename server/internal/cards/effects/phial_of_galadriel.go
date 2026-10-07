package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phial of Galadriel — Legendary Artifact {3}:
//
//	"If you would draw a card while you have no cards in hand, draw two
//	 cards instead.
//	 If you would gain life while you have 5 or less life, you gain
//	 twice that much life instead.
//	 {T}: Add one mana of any color."
//
// Two replacements, built from the shared DrawBecomes / LifeGainBecomes
// descriptions with a state gate on top. The gate is read when the
// event is offered (CR 614.1): an empty hand for the draw, the life
// total BEFORE the gain for the life.
//
// A multi-card draw is a run of single draws, and only the first one
// happens with an empty hand: the hand has a card in it for the rest.
// So the draw count is n+1, not 2n — the Phial turns "draw three" from
// an empty hand into four cards, not six.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08f17ebc-c0fd-493f-84b3-e9250694543e",
		Name:         "Phial of Galadriel",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			phialGated(
				DrawBecomes{Count: phialExtraFirstDraw, Scope: DrawsByController, Label: "Phial of Galadriel: draw two cards instead"}.Build(),
				func(ev *game.ReplacementEvent) uuid.UUID { return ev.DrawPlayer },
				func(p *game.Player) bool { return p.Hand == nil || p.Hand.Size() == 0 },
			),
			phialGated(
				YouGainTwiceThatMuchLife("Phial of Galadriel: gain twice that much life"),
				func(ev *game.ReplacementEvent) uuid.UUID { return ev.LifePlayer },
				func(p *game.Player) bool { return p.Life <= 5 },
			),
		},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
	})
}

// phialExtraFirstDraw is the draw count under an empty hand: the first
// of n draws becomes two, the rest are ordinary.
func phialExtraFirstDraw(n int) int { return n + 1 }

// phialGated narrows a replacement to the moment its affected player
// satisfies `state`.
func phialGated(r game.ReplacementEffect, who func(*game.ReplacementEvent) uuid.UUID, state func(*game.Player) bool) game.ReplacementEffect {
	applies := r.AppliesTo
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !applies(ev, g, src) {
			return false
		}
		p := g.PlayerByIDForEffect(who(ev))
		return p != nil && state(p)
	}
	return r
}
