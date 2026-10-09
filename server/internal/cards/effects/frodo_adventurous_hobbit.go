package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frodo, Adventurous Hobbit — Legendary Creature — Halfling Scout
// {W}{B}, 1/3:
//
//	"Partner with Sam, Loyal Attendant
//	 Vigilance
//	 Whenever Frodo attacks, if you gained 3 or more life this turn,
//	 the Ring tempts you. Then if Frodo is your Ring-bearer and the Ring
//	 has tempted you two or more times this game, draw a card."
//
// The printing carries no reminder text, but the keyword is the whole
// of CR 702.124j: the commander pairing with Sam (internal/deck) and
// the entry search for a card named Sam (PartnerWith, #2142).
//
// The attack trigger's "if" is an intervening if (CR 603.4): checked
// when Frodo attacks and again on resolution. The draw is checked
// after the tempt, so the tempt that makes Frodo the Ring-bearer, and
// the one that brings the count to two, both count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58ee8c10-7be8-4889-944d-13cae0a4166d",
		Name:            "Frodo, Adventurous Hobbit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Frodo, Adventurous Hobbit", "Sam, Loyal Attendant"),
			On(game.EventAttack, AllOf(ThisAttacked, youGainedLifeThisTurnAtLeast(3)),
				"Frodo, Adventurous Hobbit — the Ring tempts you, then draw if Frodo is your Ring-bearer",
				func(g *game.Game, item *game.StackItem) error {
					if b15LifeGainedThisTurn(g, item.Controller) < 3 {
						return nil
					}
					return TheRingTemptsYou{Then: frodoAdventurousDraw}.Apply(NewContext(g, item))
				}),
		},
	})
}

// youGainedLifeThisTurnAtLeast is an intervening-if trigger condition:
// the source's controller gained n or more life this turn.
func youGainedLifeThisTurnAtLeast(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b15LifeGainedThisTurn(g, source.Controller) >= n
	}
}

// frodoAdventurousDraw is "Then if Frodo is your Ring-bearer and the
// Ring has tempted you two or more times this game, draw a card."
func frodoAdventurousDraw(ctx *Context, _ uuid.UUID) error {
	you := ctx.Controller()
	if game.RingBearerOf(ctx.Game, you) != ctx.Source() || game.RingTemptCount(ctx.Game, you) < 2 {
		return nil
	}
	return DrawCards{Player: you, N: 1}.Apply(ctx)
}
