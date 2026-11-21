package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b11_helpers.go — the shared pieces of the Phyrexia: All Will Be
// One oil-counter creatures (S82, batch 11).

// s82OneOil is the oil counter's kind. It is the same on-disk identity
// Vat of Rebirth and Mindsplice Apparatus use.
const S82OneOil = "oil"

// s82HasOilCounter passes for a card with one or more oil counters on it.
func S82HasOilCounter(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.Counters[S82OneOil] > 0
}

// s82OilPermanentsYouControl is "the number of permanents you control
// with oil counters on them", read as the effect resolves.
func S82OilPermanentsYouControl(ctx *Context) int {
	return len(MatchingBattlefield(ctx, And(YouControl(), S82HasOilCounter)))
}

// s82YouControlAnOilPermanent is "you control a permanent with an oil
// counter on it".
var S82YouControlAnOilPermanent = CondBoard{
	Name: "you control a permanent with an oil counter on it",
	Test: func(ctx *Context) bool { return S82OilPermanentsYouControl(ctx) > 0 },
}

// s82OilPump is "This creature gets +P/+T for each oil counter on it": a
// layer 7c modifier over the source alone, read live on every pass.
func S82OilPump(power, toughness int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			n := source.Counters[S82OneOil]
			c.Power += power * n
			c.Toughness += toughness * n
		},
	}
}

// s82OilOnDeparture is the first half of "Whenever another creature or
// artifact you control is put into a graveyard from the battlefield"
// (Vat of Rebirth's condition): types and controller as the permanent
// last existed.
func S82OilOnDeparture(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return vatOfRebirthCondition(ev, source, lki, g)
}
