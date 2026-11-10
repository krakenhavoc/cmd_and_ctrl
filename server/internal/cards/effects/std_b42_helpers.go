package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b42_helpers.go — shared pieces for Standard batch 42 (#4452).

// CondFaceDownEnteredThisTurn is "if a face-down creature entered the
// battlefield under your control this turn" (Tunnel Tipster, CR 708.2)
// as an intervening-if Condition, read for the item's controller as it
// resolves. A face-down permanent is always a creature, and it counts
// though it has turned face up or left since (the ruling). The tally only
// rises within a turn, so it declares a monotone footprint condition
// (game.FootprintCondFaceDownEntered).
type CondFaceDownEnteredThisTurn struct{}

func (CondFaceDownEnteredThisTurn) Holds(ctx *Context) bool {
	return ctx.Game.PermanentEnteredFaceDownThisTurn(ctx.Controller())
}

func (CondFaceDownEnteredThisTurn) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondFaceDownEntered, 0), true
}

// YouHadAFaceDownEntryThisTurn is the When for the same fact, read for the
// source's controller as the event happens, for composing with AllOf.
func YouHadAFaceDownEntryThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.PermanentEnteredFaceDownThisTurn(source.Controller)
}

// ImpDamageToken is Murders at Karlov Manor's Imp (Judith, Carnage
// Connoisseur): a 2/2 red Imp creature token with "When this creature
// dies, it deals 2 damage to each opponent." The trigger is the token's
// own (ADR 0083): the token's controller is "you", and the damage is
// dealt by the token as it last existed (CR 608.2h).
func ImpDamageToken() game.Card { return tokenFromCatalog(printedImpDamageToken) }

func printedImpDamageToken() tokenTemplate {
	return tokenTemplate{
		Slug: "imp-dies-damage",
		Card: game.Card{
			Name:      "Imp",
			TypeLine:  "Token Creature — Imp",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Imp — it deals 2 damage to each opponent", Do(DamageEachOpponent{N: 2})),
		},
		Text: "When this creature dies, it deals 2 damage to each opponent.",
	}
}

// ExileTopAnyNumberFreeCast exiles the top n cards of the controller's
// library and offers "you may cast any number of <filter> spells from
// among the exiled cards without paying their mana costs" (Kylox,
// Visionary Inventor), as the resolving ability does it (CR 608.2g):
// one free, flash-speed, cast-only permission over the cards that landed,
// closing when its holder passes, that a cast does not use up (AnyNumber),
// with no cap on a spell's mana value. A card the filter refuses (a land,
// a creature) is stranded in exile; X in a cast spell's cost is 0.
func ExileTopAnyNumberFreeCast(ctx *Context, n int, filter game.PermissionFilter) error {
	if n <= 0 {
		return nil
	}
	controller := ctx.Controller()
	perm := game.CastPermission{
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		AnyNumber:   true,
		Filter:      filter,
		Source:      ctx.Source(),
		Label:       "Cast any number of spells from among the exiled cards without paying their mana costs",
	}
	if src, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
		perm.SourceName = src.Name
	}
	return ctx.Game.ExileTopWithPermissionKeyedForEffect(controller, controller, n, perm,
		game.KeyedTail{Ref: impulseExileOfferCasts, Player: controller, Source: ctx.Source(), Word: perm.SourceName})
}
