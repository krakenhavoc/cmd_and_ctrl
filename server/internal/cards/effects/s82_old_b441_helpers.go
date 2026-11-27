package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Ascension batch 441 helpers: fateful hour ("if you have 5 or less
// life", CR 207.2c: an ability word with no rules meaning of its own) and
// the flashback sorceries that read how they were cast.

// OldB441FatefulHour reports whether `player` has 5 or less life: the
// condition every fateful hour clause reads.
func OldB441FatefulHour(g *game.Game, player uuid.UUID) bool {
	if g == nil {
		return false
	}
	p := g.PlayerByIDForEffect(player)
	return p != nil && p.Life <= 5
}

// OldB441OtherCreaturesWhileFatefulHour is the static scope "other
// creatures you control" that applies only while its controller has 5 or
// less life (Gavony Ironwright, Village Survivors).
func OldB441OtherCreaturesWhileFatefulHour(target *game.Card, g *game.Game, source *game.Card) bool {
	return OtherCreaturesYouControl(target, g, source) && OldB441FatefulHour(g, source.Controller)
}

// OldB441OtherCreaturesPumpWhileFatefulHour is "Fateful hour — As long as
// you have 5 or less life, other creatures you control get +p/+t".
func OldB441OtherCreaturesPumpWhileFatefulHour(power, toughness int) game.StaticAbility {
	ab := LifeGatedPump(OldB441OtherCreaturesWhileFatefulHour, power, toughness)
	ab.Label = "Fateful hour — other creatures you control get " + SignedPT(power, toughness) + " while you have 5 or less life."
	return ab
}

// OldB441OtherCreaturesKeywordWhileFatefulHour is "Fateful hour — As long
// as you have 5 or less life, other creatures you control have <keyword>".
func OldB441OtherCreaturesKeywordWhileFatefulHour(keyword string) game.StaticAbility {
	ab := LifeGatedKeyword(OldB441OtherCreaturesWhileFatefulHour, keyword)
	ab.Label = "Fateful hour — other creatures you control have " + keyword + " while you have 5 or less life."
	return ab
}

// OldB441DiscardHandThenDrawThatMany is "Discard all the cards in your
// hand, then draw that many cards". The draw waits for a prompt the
// discard asks (a madness offer, CR 608.2c) and reads the count from the
// step record.
func OldB441DiscardHandThenDrawThatMany(record string) func(*game.StackItem, *Context) error {
	return ResolveSteps(
		Instruction(func(ctx *Context) error {
			n, err := DiscardWholeHand(ctx.Game, ctx.Controller())
			ctx.RecordStep(record, nil, n)
			return err
		}),
		Instruction(func(ctx *Context) error {
			return DrawCards{N: ctx.StepRecord(record).N()}.Apply(ctx)
		}),
	)
}

// OldB441FaithsShieldFatefulHour is the colour answer to Faith's Shield's
// fateful hour clause: you and each permanent you control gain protection
// from that colour until end of turn. An on-disk key.
var OldB441FaithsShieldFatefulHour = CardTail("then/faiths-shield-fateful-hour-protection", func(ctx *Context, o game.TailOutcome) error {
	token := game.ProtectionFromColor(o.Tail.Word)
	if token == "" {
		return nil
	}
	me := ctx.Controller()
	for _, id := range PermanentsControlledByMatching(ctx.Game, me, nil) {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{token}, Label: "Faith's Shield — " + token}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainPlayerKeyword{
		Player:   me,
		Keyword:  token,
		Label:    "Faith's Shield — " + token,
		Duration: DurationUntilEndOfTurn(ctx),
	}.Apply(ctx)
})
