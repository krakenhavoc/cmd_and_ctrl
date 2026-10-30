package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Fantasticar — {3} Legendary Artifact — Vehicle 4/4 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 Whenever you cast a noncreature spell, you may have The Fantasticar
//	 become an artifact creature until end of turn.
//	 Whenever you cast your fourth noncreature spell each turn, you may
//	 sacrifice The Fantasticar. If you do, create four 4/4 colorless
//	 Construct artifact creature tokens with flying and haste."
//
// Both are "you may" triggers answered as they resolve (CR 603.5). The first
// is the crew effect without the crew cost (BecomeCreatureUntilEOT). The
// second reads the turn's noncreature cast tally the spell just bumped, so
// the fourth such spell of the turn is the one that counts (a creature spell
// neither counts nor triggers it); its sacrifice is a state-trigger-style
// SacrificeThisThen, and the Constructs are made only if The Fantasticar
// really left the battlefield. The two triggers fire together on the fourth
// spell, in the controller's order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9b0abb58-4e2c-45dc-99b1-61a4cb79354b",
		Name:            "The Fantasticar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverYouCast(Noncreature(), "The Fantasticar — become an artifact creature until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return BecomeCreatureUntilEOT{Label: "The Fantasticar — becomes an artifact creature"}.Apply(NewContext(g, item))
				}),
				"The Fantasticar — have it become an artifact creature until end of turn?"),
			Optional(On(game.EventCast, youCastYourFourthNoncreatureSpellThisTurn,
				"The Fantasticar — sacrifice it to create four 4/4 Construct artifact creature tokens with flying and haste",
				SacrificeThisThen("then/the-fantasticar-constructs", func(ctx *Context) error {
					return CreateToken{Template: TokenCard("4/4 colorless Construct artifact with flying and haste"), N: 4}.Apply(ctx)
				})),
				"The Fantasticar — sacrifice it to create four 4/4 Construct tokens?"),
		},
	})
}

// youCastYourFourthNoncreatureSpellThisTurn is the cast of the source
// controller's fourth noncreature spell this turn (the tally is bumped
// before EventCast fires).
func youCastYourFourthNoncreatureSpellThisTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || spell.IsCreature() {
		return false
	}
	return g.CastTallyFor(ev.Actor).Noncreature == 4
}
