package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth Resplendent — Legendary Planeswalker — Elspeth {3}{W}{W}, loyalty 5
// (Streets of New Capenna, #5491):
//
//	"+1: Choose up to one target creature. Put a +1/+1 counter and a counter
//	 from among flying, first strike, lifelink, or vigilance on it.
//	 −3: Look at the top seven cards of your library. You may put a permanent
//	 card with mana value 3 or less from among them onto the battlefield with
//	 a shield counter on it. Put the rest on the bottom of your library in a
//	 random order.
//	 −7: Create five 3/3 white Angel creature tokens with flying."
//
// The +1 targets up to one creature; the +1/+1 counter goes on at once and the
// keyword counter is the controller's pick as the ability resolves (CR
// 608.2d). The −3 is one pick over the permanent cards of mana value 3 or
// less among the seven (none is a legal answer) put onto the battlefield
// with a shield counter as it enters; whatever is not put goes to the bottom
// in a random order.
//
// No simplification.
func init() {
	const name = "Elspeth Resplendent"
	Register(Spec{
		OracleID:        "24591330-d038-478f-b89d-b7b66c0b5cc3",
		Name:            name,
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Choose up to one target creature. Put a +1/+1 counter and a counter from among flying, first strike, lifelink, or vigilance on it.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  elspethResplendentPlusOne,
			},
			{
				Label:  "−3: Look at the top seven cards of your library. You may put a permanent card with mana value 3 or less from among them onto the battlefield with a shield counter on it. Put the rest on the bottom of your library in a random order.",
				Cost:   LoyaltyCost(-3),
				Effect: elspethResplendentDig,
			},
			{
				Label:   "−7: Create five 3/3 white Angel creature tokens with flying.",
				Purpose: game.Purpose{Tokens: 5},
				Cost:    LoyaltyCost(-7),
				Effect:  Do(CreateToken{Template: TokenCard("3/3 white Angel with flying"), N: 5}),
			},
		},
	})
}

// elspethResplendentKeywordCounters are the four kinds on the +1, in printed
// order, by the label the pick shows.
var elspethResplendentKeywordCounters = []struct{ label, kind string }{
	{"Flying", game.CounterFlying},
	{"First strike", game.CounterFirstStrike},
	{"Lifelink", game.CounterLifelink},
	{"Vigilance", game.CounterVigilance},
}

func elspethResplendentPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	options := make([]game.ChoiceOption, 0, len(elspethResplendentKeywordCounters))
	for _, k := range elspethResplendentKeywordCounters {
		options = append(options, game.ChoiceOption{Label: k.label})
	}
	return PickOption{
		Question: "Elspeth Resplendent — put a flying, first strike, lifelink or vigilance counter on the creature",
		Options:  options,
		ThenKey:  elspethResplendentCounterPicked,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// elspethResplendentCounterPicked puts the chosen keyword counter on the
// creature. An on-disk key: never renamed, never reused.
var elspethResplendentCounterPicked = OptionPickThen("option-pick/elspeth-resplendent-keyword-counter", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	for _, k := range elspethResplendentKeywordCounters {
		if k.label == r.Option.Label {
			return AddCounter{Target: r.Carry[0], Kind: k.kind, N: 1}.Apply(ctx)
		}
	}
	return nil
})

func elspethResplendentDig(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 7)
	if len(looked) == 0 {
		return nil
	}
	var eligible []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsPermanent() && !IsToken(c) && c.ManaValue() <= 3 {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Elspeth Resplendent — you may put a permanent card with mana value 3 or less onto the battlefield with a shield counter (the rest go to the bottom in a random order)",
		Cards:    eligible,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  elspethResplendentDigPicked,
		Carry:    looked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// elspethResplendentDigPicked puts the chosen card onto the battlefield with
// a shield counter and the rest on the bottom. An on-disk key: never
// renamed, never reused.
var elspethResplendentDigPicked = askedPick("choose-cards/elspeth-resplendent-dig", func(ctx *Context, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if picked != uuid.Nil {
		if _, err := ctx.Game.PutFromLibraryOntoBattlefieldForEffect(picked, game.LibraryEntryOptions{
			Controller:         ctx.Controller(),
			EntersWithCounters: map[string]int{game.CounterShield: 1},
		}); err != nil {
			return err
		}
	}
	if len(rest) > 0 {
		return ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, rest)
	}
	return nil
})
