package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tsp_b1_helpers.go — shapes shared by the Time Spiral batch 1 cards
// (#5596, S82). Each is a package-level func or a constructor whose
// closure captures only constants, so nothing here holds a game.

// tspB1PsionicPing is Psionic Sliver's granted "{T}: This creature
// deals 2 damage to any target and 3 damage to itself." The Sliver that
// has the ability is the source of both. The self-damage is dealt only
// while that Sliver is still on the battlefield: a creature that has left
// is no longer there to be dealt damage (CR 400.7).
func tspB1PsionicPing(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host := item.SourceCardID
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	if err := (DealDamage{Source: host, Target: legal[0].ID, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	if !onBattlefield(g, host) {
		return nil
	}
	return DealDamage{Source: host, Target: host, Amount: 3}.Apply(ctx)
}

// tspB1EachPlayerDiscardsACard is "Each player discards a card." as an
// ability's whole body (Mindlash Sliver's granted ability): every player
// still in the game, in APNAP order, picks one card of their own.
func tspB1EachPlayerDiscardsACard(question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
			Source: item.SourceCardID, N: 1, Question: question,
		})
		return nil
	}
}

// tspB1OpponentCastWhileSuspended is Deep-Sea Kraken's trigger
// condition: an opponent of the card's owner cast a spell (CR 109.5,
// "you" for a card in exile is its owner) while the card is suspended
// (CR 702.62b), read as it triggers (CR 603.4).
func tspB1OpponentCastWhileSuspended(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !game.CardIsSuspended(*source) {
		return false
	}
	return AnOpponentCast(nil)(ev, source, lki, g)
}

// tspB1GainTwoPerCreatureTargetPlayerControls is Divine Congregation's
// body: its controller gains 2 life for each creature the targeted
// player controls, counted as it resolves.
func tspB1GainTwoPerCreatureTargetPlayerControls(_ *game.StackItem, ctx *Context) error {
	pid, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return GainLife{Amount: 2 * creaturesControlledBy(ctx.Game, pid)}.Apply(ctx)
}

// tspB1PumpAndGrantUntilEOT is "Target creature gets +P/+T and gains
// <keyword> until end of turn." (Stonewood Invocation): the pump, then
// the keyword, on the first target still legal (CR 608.2b).
func tspB1PumpAndGrantUntilEOT(power, toughness int, keyword, name string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: power, Toughness: toughness,
			Label: name + " — pump until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword},
			Label: name + " — " + keyword + " until end of turn"}.Apply(ctx)
	}
}

// tspB1UntapGainControlAndHaste is "Untap target permanent and gain
// control of it until end of turn. It gains haste until end of turn."
// (Word of Seizing), in printed order.
func tspB1UntapGainControlAndHaste(name string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		if err := (GainControl{
			Target:   id,
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    name + " — gain control until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    name + " — haste until end of turn",
		}.Apply(ctx)
	}
}
