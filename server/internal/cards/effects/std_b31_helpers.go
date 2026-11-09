package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b31_helpers.go — the shared pieces behind the Standard leftovers
// batches 31–34 (#4452). Every package-level name carries the stdB31
// prefix, or is a Cond/When named for the rule it reads.

// CondSpellsCast is "if you've cast N or more spells this turn" (Loan
// Shark: two): the item's controller's cast tally this turn, which only
// rises and which no step in the footprint vocabulary moves
// (game.FootprintCondSpellsCast). AtLeast 0 means 1.
type CondSpellsCast struct{ AtLeast int }

func (c CondSpellsCast) Holds(ctx *Context) bool {
	return ctx.Game.CastTallyFor(ctx.Controller()).Total >= max(c.AtLeast, 1)
}

func (c CondSpellsCast) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondSpellsCast, max(c.AtLeast, 1)), true
}

// YouHaveCastSpellsThisTurn is the trigger-time half of CondSpellsCast
// (CR 603.4): the source's controller has cast n or more spells this turn.
func YouHaveCastSpellsThisTurn(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return source != nil && g.CastTallyFor(source.Controller).Total >= n
	}
}

// stdB31Threaten is "gain control of target creature until end of turn.
// Untap that creature. It gains haste until end of turn" (Take for a Ride):
// the control change, the untap and the haste are one effect, in printed
// order, each skipped quietly if the creature has gone (CR 608.2b).
func stdB31Threaten(ctx *Context, target game.TargetRef, label string) error {
	if target.Kind != game.TargetCard {
		return nil
	}
	if err := (GainControl{Target: target.ID, Duration: DurationUntilEndOfTurn(ctx), Label: label + " — gain control until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: target.ID}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: target.ID, Keywords: []string{"haste"}, Label: label + " — haste until end of turn"}.Apply(ctx)
}

// stdB31CoinCounter is Wishing Well's coin counter. Counter kinds are free-form
// strings (game/counter_types.go); this one has no meaning beyond its name.
const stdB31CoinCounter = "coin"

// stdB31WishingWellBody is Wishing Well's reflexive trigger: "you may cast
// target instant or sorcery card with mana value equal to the number of coin
// counters from your graveyard without paying its mana cost. If that spell
// would be put into your graveyard, exile it instead." The number of counters is
// fixed when the trigger is made and rides Params.Amount (Teferi Akosa's
// ceiling is carried the same way), so the target clause is the same at announce,
// at resolution and after a restore. An on-disk key.
var stdB31WishingWellBody = game.ReflexiveBody("wishing-well/free-cast", simpleBody(freeCastTargetFromYourGraveyardExilingIt("Wishing Well")),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCardInGraveyard("target instant or sorcery card with mana value equal to the number of coin counters from your graveyard",
			YouOwn(), Or(Instant(), Sorcery()), b03ManaValueIs(p.Amount))
	})
