package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b126_helpers.go — S82 set GAP, batch 126 (#5948): the Lord of the
// Rings red rows. Append-only.

// gapB126ThisOrAnotherEntered is "whenever this creature or another
// <subtype> you control enters": the source's own entry, or another
// permanent of the subtype (read through the source's text) entering under
// its controller's control.
func GapB126ThisOrAnotherEntered(subtype string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.CardID == source.InstanceID ||
			B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype(subtype))
	}
}

// gapB126YouControlNoFood is "you don't control a Food".
func GapB126YouControlNoFood(g *game.Game, you uuid.UUID) bool {
	return CountControlled(g, you, func(c game.Card) bool { return c.HasSubtype("Food") }) == 0
}

// gapB126TribeDealtCombatDamageToAPlayer is the When of "whenever a
// <Tribe> or <Tribe> you control deals combat damage to a player".
func GapB126TribeDealtCombatDamageToAPlayer(tribes ...string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !CombatDamageToPlayerBy(ev, source.Controller, g) {
			return false
		}
		dealer, ok := g.LookupCardForEffect(ev.Source)
		if !ok {
			return false
		}
		for _, t := range tribes {
			if dealer.HasSubtype(source.TextSubtype(t)) {
				return true
			}
		}
		return false
	}
}

// gapB126GrishnakhTail is the amass's reflexive half of Grishnákh: the
// amassed Army's power at the end of the amass rides the reflexive
// trigger as its amount.
var GapB126GrishnakhTail = AmassTail("grishnakh-brash-instigator/steal", func(ctx *Context, army uuid.UUID) error {
	power := 0
	if army != uuid.Nil {
		power = B29PowerAsItLastStood(ctx.Game, army)
	}
	return ReflexiveTrigger{
		Label:  "Grishnákh, Brash Instigator — until end of turn, gain control of target nonlegendary creature an opponent controls with power less than or equal to the amassed Army's power, untap it, and it gains haste",
		Cards:  []uuid.UUID{army},
		Body:   gapB126GrishnakhStealBody,
		Params: game.EffectParams{Amount: power},
	}.Apply(ctx)
})

var gapB126GrishnakhStealBody = game.ReflexiveBody("grishnakh-brash-instigator/steal-body", SimpleBody(gapB126GrishnakhSteal),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCreature("target nonlegendary creature an opponent controls with power less than or equal to the amassed Army's power",
			OpponentControls(), Not(Legendary()), PowerLE(p.Amount))
	})

func gapB126GrishnakhSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (GainControl{Target: id, Duration: DurationUntilEndOfTurn(ctx), Label: "Grishnákh, Brash Instigator — gain control until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Grishnákh, Brash Instigator — haste until end of turn"}.Apply(ctx)
}

// gapB126RidersTail is Riders of the Mark's "if you do" half: the tokens
// equal to the returned creature's toughness as it last stood on the
// battlefield, carried as the tail's N.
var GapB126RidersTail = CardTail("riders-of-the-mark/human-soldiers", func(ctx *Context, o game.TailOutcome) error {
	if !OneWentThisWay(o) || o.Tail.N <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: o.Tail.N}.Apply(ctx)
})

// ControlsGreatestPowerAmongCreatures reports whether `controller` controls a
// creature whose power is the greatest, or tied for the greatest, among
// every creature on the battlefield.
func ControlsGreatestPowerAmongCreatures(g *game.Game, controller uuid.UUID) bool {
	if g.Battlefield == nil {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	best, found, mine := 0, false, false
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() {
			continue
		}
		p := c.PowerForComparison()
		switch {
		case !found || p > best:
			best, found, mine = p, true, c.Controller == controller
		case p == best && c.Controller == controller:
			mine = true
		}
	}
	return mine
}
