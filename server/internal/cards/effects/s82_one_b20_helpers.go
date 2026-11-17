package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b20_helpers.go — shared pieces of Phyrexia: All Will Be One
// batch 20 (#5479): the planeswalkers and the toxic Beast token.

// s82OneB20BeastToken is the 3/3 green Phyrexian Beast creature token with
// toxic 1 that Goliath Hatchery, Viral Spawning and Lukka, Bound to Ruin
// make.
func s82OneB20BeastToken() game.Card { return TokenCard("3/3 green Phyrexian Beast with toxic 1") }

// s82OneB20CreateBeasts creates n Beast tokens under the effect's controller.
func s82OneB20CreateBeasts(g *game.Game, item *game.StackItem, n int) error {
	return CreateToken{Controller: item.Controller, Template: s82OneB20BeastToken(), N: n}.Apply(NewContext(g, item))
}

// s82OneB20LandsControlled counts the permanents `controller` controls with
// the land subtype `sub` (a word of the source's text, CR 612.2).
func s82OneB20LandsControlled(g *game.Game, controller uuid.UUID, sub string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype(sub) {
			n++
		}
	}
	return n
}

// printedS82OneB20DroneToken is Kaito, Dancing Shadow's 2/2 colourless Drone
// artifact creature with deathtouch and "When this token leaves the
// battlefield, each opponent loses 2 life and you gain 2 life." The slug is
// an on-disk identity: never renamed, never reused.
func printedS82OneB20DroneToken() tokenTemplate {
	return tokenTemplate{
		Slug: "drone-deathtouch-drain",
		Card: game.Card{
			Name:      "Drone",
			TypeLine:  "Token Artifact Creature — Drone",
			Power:     2,
			Toughness: 2,
			Keywords:  []string{"deathtouch"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Drone — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
		Text: "Deathtouch\nWhen this token leaves the battlefield, each opponent loses 2 life and you gain 2 life.",
	}
}

// s82OneB20DroneToken is the Drone's card value for CreateToken.
func s82OneB20DroneToken() game.Card { return tokenFromCatalog(printedS82OneB20DroneToken) }

// s82OneB20ManaValueBelowLoyalty is "with mana value less than <this
// planeswalker>'s loyalty" (Nahiri, the Unforgiving): judged at targeting
// and again on resolution, against the loyalty the source has then (its
// last-known count once it has left, CR 608.2h).
func s82OneB20ManaValueBelowLoyalty() game.SourceComparison {
	return game.SourceComparison{Stat: game.SourceManaValueVsLoyalty, Cmp: game.CmpLess}
}

// s82OneB20SourceLoyalty is the loyalty of the planeswalker whose ability is
// resolving: its loyalty counters now while it is on the battlefield, and
// as it last existed there once it has left (CR 608.2h). Zero when neither
// is known.
func s82OneB20SourceLoyalty(ctx *Context) int {
	ref, ok := ctx.SourceRef()
	if !ok {
		return 0
	}
	info, ok := ctx.Game.PermanentForEffect(ref)
	if !ok {
		return 0
	}
	return info.Counters[game.CounterLoyalty]
}

// s82OneB20DivideGreatestPower is "X damage divided as you choose …, where X
// is the greatest power among creatures you control as you activate this
// ability" (Lukka, Bound to Ruin): read once, at announce, and the number
// divided whatever happens to the board afterwards.
var s82OneB20DivideGreatestPower = game.RegisterDivideAmount("greatest-power-among-your-creatures",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		best := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == a.Controller && c.IsCreature() && c.CurrentPower() > best {
				best = c.CurrentPower()
			}
		}
		return game.DivideSpec{Total: best}
	})
