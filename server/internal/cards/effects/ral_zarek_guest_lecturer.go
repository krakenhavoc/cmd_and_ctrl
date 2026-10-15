package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ral Zarek, Guest Lecturer — Legendary Planeswalker — Ral {1}{B}{B},
// loyalty 3:
//
//	"+1: Surveil 2.
//	 −1: Any number of target players each discard a card.
//	 −2: Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield.
//	 −7: Flip five coins. Target opponent skips their next X turns,
//	 where X is the number of coins that came up heads."
//
// The −7 is the first card to skip a TARGET player's turns, for a
// number read at resolution (CR 614.10a, ADR 0188 Decision 1, #3565):
// five uncalled flips (CR 705.1, FlipCoinsForEffect: nobody wins or
// loses them, only the faces count), then the target skips X turns.
// Zero heads skips nothing, and a target that left the game in response
// is no longer legal (CR 608.2b). The flips happen when the ability
// resolves, not when it is activated.
//
// "Any number of target players" is the (0, 0) count Court of Cunning
// declares: zero is legal and yourself is a legal pick. Each player
// discards through the ordinary discard prompt, so a player with an
// empty hand discards nothing.
//
// Starting loyalty is printed card data stamped by the deck importer
// (ADR 0032), not declared here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0dc76b7f-6405-44a1-8c55-31fa480ff70b",
		Name:         "Ral Zarek, Guest Lecturer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Surveil 2.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Surveil{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−1: Any number of target players each discard a card.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetPlayer("any number of target players").WithCount(0, 0),
				Effect:  ralZarekDiscard,
			},
			{
				Label:   "−2: Return target creature card with mana value 3 or less from your graveyard to the battlefield.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", YouOwn(), And(Creature(), ManaValueLE(3))),
				Effect:  ralZarekReanimate,
			},
			{
				Label:   "−7: Flip five coins. Target opponent skips their next X turns, where X is the number of coins that came up heads.",
				Cost:    LoyaltyCost(-7),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect:  ralZarekFlipFive,
			},
		},
	})
}

// ralZarekDiscard queues one discard prompt per still-legal chosen
// player.
func ralZarekDiscard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: t.ID, Source: item.SourceCardID, N: 1})
	}
	return nil
}

// ralZarekReanimate puts the still-legal target onto the battlefield
// under its owner's control (it is in your graveyard, so that is you).
func ralZarekReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: ids, Controller: item.Controller}.Apply(ctx)
}

// ralZarekFlipFive flips five coins and has the target opponent skip
// that many turns as there were heads. The flips come first and only
// when the ability has a legal target, as an ability whose only target
// is gone does not resolve (CR 608.2b).
func ralZarekFlipFive(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(ctx.LegalTargets()) == 0 {
		return nil
	}
	faces, err := g.FlipCoinsForEffect(game.RandomDraw{Player: item.Controller, Source: item.SourceCardID}, 5)
	if err != nil {
		return err
	}
	heads := 0
	for _, h := range faces {
		if h {
			heads++
		}
	}
	if heads == 0 {
		return nil
	}
	return TargetPlayerSkipsNextTurns{N: heads}.Apply(ctx)
}
