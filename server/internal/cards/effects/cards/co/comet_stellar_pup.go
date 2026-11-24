package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Comet, Stellar Pup — Legendary Planeswalker — Comet {2}{R}{W}, loyalty 5
// (Unfinity, #6100):
//
//	"0: Roll a six-sided die.
//	 1 or 2 — [+2], then create two 1/1 green Squirrel creature tokens.
//	 They gain haste until end of turn.
//	 3 — [−1], then return a card with mana value 2 or less from your
//	 graveyard to your hand.
//	 4 or 5 — Comet deals damage equal to the number of loyalty counters on
//	 him to a creature or player, then [−2].
//	 6 — [+1], and you may activate Comet's loyalty ability two more times
//	 this turn."
//
// The roll goes through the engine's one dice door, so a replacement of
// the roll applies, and the result picks the row (CR 706.3a). A loyalty
// symbol in an effect means "put N loyalty counters on this permanent" or
// "remove N" (CR 107.7): an effect, not a cost, so it does nothing once
// Comet has left the battlefield, and a [−N] on a Comet with fewer
// counters removes what is there.
//
//   - 1 or 2: two loyalty counters, then two Squirrels that gain haste
//     until end of turn.
//   - 3: one loyalty counter off, then you choose a card with mana value 2
//     or less in your graveyard (not a target) and return it to your hand.
//   - 4 or 5: you choose a creature or a player (not a target: hexproof
//     does not stop it), first which kind, then which one; Comet deals it
//     damage equal to his loyalty counters as the damage is dealt (his
//     last-known loyalty if he has left, CR 608.2h), then two loyalty
//     counters come off.
//   - 6: one loyalty counter, and two more activations of his loyalty
//     ability this turn beyond CR 606.3's limit, at the usual timing; a
//     second 6 grants two more again.
//
// No simplification.
const cometName = "Comet, Stellar Pup"

var (
	// cometRolled is the roll's result. An on-disk key: never renamed,
	// never reused.
	cometRolled = effects.CardTail("dice/comet-stellar-pup-roll", cometResult)

	// cometGraveyardPick is row 3's card. An on-disk key.
	cometGraveyardPick = effects.ItemChoice("choose-cards/comet-stellar-pup-return-to-hand", func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
	})

	// cometCreatureOrPlayer is row 4–5's "a creature or player": yes is a
	// creature. An on-disk key.
	cometCreatureOrPlayer = effects.MayYesNo("confirm/comet-stellar-pup-creature-or-player",
		func(ctx *effects.Context, _ game.ConfirmAnswer) error { return cometChooseCreature(ctx) },
		func(ctx *effects.Context, _ game.ConfirmAnswer) error { return cometChoosePlayer(ctx) })

	// cometCreaturePick is the creature chosen. An on-disk key.
	cometCreaturePick = effects.ItemChoice("choose-cards/comet-stellar-pup-damage-creature", func(ctx *effects.Context, r game.CardsChosen) error {
		victim := uuid.Nil
		if len(r.Picked) > 0 {
			victim = r.Picked[0]
		}
		return cometDamageThenMinusTwo(ctx, victim)
	})

	// cometPlayerPick is the player chosen. An on-disk key.
	cometPlayerPick = effects.ChoosePlayerThen("option-pick/comet-stellar-pup-damage-player", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
		return cometDamageThenMinusTwo(ctx, chosen)
	})
)

func init() {
	effects.Register(effects.Spec{
		OracleID:        "e7c67480-1816-4071-9eab-a1d1b535a344",
		Name:            cometName,
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{{
			Label: "0: Roll a six-sided die.",
			Cost:  effects.LoyaltyCost(0),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RollDice{Sides: 6, N: 1, Then: cometRolled}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// cometLoyalty puts n loyalty counters on Comet, or removes -n, while he is
// still the same object on the battlefield.
func cometLoyalty(ctx *effects.Context, n int) error {
	src := ctx.Item.SourceCardID
	if !effects.OnBattlefield(ctx.Game, src) || ctx.IsNewSourceObject(src) {
		return nil
	}
	return effects.AddCounter{Target: src, Kind: game.CounterLoyalty, N: n}.Apply(ctx)
}

// cometResult is the results table.
func cometResult(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	switch r := effects.DiceResult(o); {
	case r <= 2:
		if err := cometLoyalty(ctx, 2); err != nil {
			return err
		}
		return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(),
			Source:     ctx.Source(),
			Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("1/1 green Squirrel")), Count: 2}},
		}, effects.WithWord(ctx.ItemTail(effects.ThenCreatedGainKeywordUntilEOT), "haste"))
	case r == 3:
		if err := cometLoyalty(ctx, -1); err != nil {
			return err
		}
		cands := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
			mv, ok := ctx.Game.ManaValueForEffect(c)
			return ok && mv <= 2
		})
		if len(cands) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: cometName + " — return a card with mana value 2 or less from your graveyard to your hand",
			Cards:    cands,
			Min:      1,
			Max:      1,
			LiveZone: game.ZoneGraveyard,
			ThenKey:  cometGraveyardPick,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	case r <= 5:
		if len(cometCreatures(ctx)) == 0 {
			return cometChoosePlayer(ctx)
		}
		return effects.MayChoice{
			Question: cometName + " — deal the damage to a creature or to a player?",
			YesLabel: "A creature",
			NoLabel:  "A player",
			Then:     cometCreatureOrPlayer,
		}.Apply(ctx)
	default:
		if err := cometLoyalty(ctx, 1); err != nil {
			return err
		}
		src := ctx.Item.SourceCardID
		if effects.OnBattlefield(ctx.Game, src) && !ctx.IsNewSourceObject(src) {
			ctx.Game.GrantThisPermanentMoreLoyaltyActivationsForEffect(src, 2)
		}
		return nil
	}
}

// cometCreatures is every creature on the battlefield.
func cometCreatures(ctx *effects.Context) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func cometChooseCreature(ctx *effects.Context) error {
	cands := cometCreatures(ctx)
	if len(cands) == 0 {
		return cometChoosePlayer(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: cometName + " — choose a creature to deal the damage to",
		Cards:    cands,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  cometCreaturePick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

func cometChoosePlayer(ctx *effects.Context) error {
	return effects.ChoosePlayer{
		Question: cometName + " — choose a player to deal the damage to",
		ThenKey:  cometPlayerPick,
	}.Apply(ctx)
}

// cometDamageThenMinusTwo deals the damage, then takes two loyalty
// counters off.
func cometDamageThenMinusTwo(ctx *effects.Context, victim uuid.UUID) error {
	src := ctx.Item.SourceCardID
	loyalty := 0
	if effects.OnBattlefield(ctx.Game, src) && !ctx.IsNewSourceObject(src) {
		if c, ok := ctx.Game.LookupCardForEffect(src); ok {
			loyalty = c.Counters[game.CounterLoyalty]
		}
	} else if info, ok := ctx.Game.LastKnownPermanentForEffect(src); ok {
		loyalty = info.Counters[game.CounterLoyalty]
	}
	if victim != uuid.Nil && loyalty > 0 {
		if err := (effects.DealDamage{Source: src, Target: victim, Amount: loyalty}).Apply(ctx); err != nil {
			return err
		}
	}
	return cometLoyalty(ctx, -2)
}
