package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b679_helpers.go — S82 OLD batch 679 (Nemesis, #5618): the shapes
// Topple and Stronghold Gambit need.

// OldB679GreatestPowerAmongAllCreatures is "creature with the greatest power
// among creatures on the battlefield" (Topple): a creature whose current
// power no other creature on the battlefield, anyone's, exceeds. Ties all
// match — the reminder text says to target any one of them — and it is read
// again as the spell resolves (CR 608.2b), so a pump in response can take
// the target out of the set.
func OldB679GreatestPowerAmongAllCreatures() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		if !c.IsCreature() {
			return false
		}
		power := c.CurrentPower()
		for _, o := range g.BattlefieldCardsForEffect() {
			if o.IsCreature() && o.CurrentPower() > power {
				return false
			}
		}
		return true
	}
}

// OldB679StrongholdGambit is Stronghold Gambit's body: every player in the
// game, in APNAP order, chooses a card in their hand (a player with an empty
// hand is not asked), then the chosen cards are revealed together, and the
// owner of each creature card among them with the lowest mana value puts it
// onto the battlefield. Nothing is revealed until the last player has chosen,
// so no choice is made knowing another.
//
// Caller must hold g.mu in write mode (a resolution frame).
func OldB679StrongholdGambit(ctx *Context) error {
	g := ctx.Game
	var legs []game.ChooseCardsPrompt
	players := ApnapOrder(g, TablePlayers(ctx))
	for _, id := range players {
		hand := HandCardIDs(g, id)
		if len(hand) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    id,
			FromPlayer: id,
			Source:     ctx.Source(),
			Question:   "Stronghold Gambit — choose a card in your hand",
			Cards:      hand,
			Min:        1,
			Max:        1,
		})
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, oldB679GambitRevealed, ctx.Source(), ctx.RunData(players))
	return err
}

// oldB679GambitRevealed reveals every chosen card together, then puts the
// creature cards with the lowest mana value among the creature cards
// revealed onto the battlefield under their owners' control, as one entry.
// An on-disk key: never renamed, never reused.
var oldB679GambitRevealed = RunThen("prompt-run/old-b679-stronghold-gambit", func(ctx *Context, r game.RunDone) error {
	g := ctx.Game
	picks := r.Picks()
	var revealed []uuid.UUID
	for _, id := range r.Carry {
		revealed = append(revealed, picks.By(id)...)
	}
	if len(revealed) == 0 {
		return nil
	}
	g.RevealForEffect(game.RevealSpec{
		Source: ctx.Source(),
		Reason: "Stronghold Gambit — each player reveals their chosen card",
		Cards:  revealed,
	})
	lowest := -1
	var winners []uuid.UUID
	for _, id := range revealed {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.IsCreature() {
			continue
		}
		mv := c.ManaValue()
		switch {
		case lowest < 0 || mv < lowest:
			lowest, winners = mv, []uuid.UUID{id}
		case mv == lowest:
			winners = append(winners, id)
		}
	}
	if len(winners) == 0 {
		return nil
	}
	return g.PutCardsFromHandOntoBattlefieldKeyedForEffect(winners, game.HandEntryOptions{PutByEnteringPlayer: true}, game.KeyedTail{})
})

// OldB679YouControlAForest is "if you control a Forest": a permanent the
// controller controls with the subtype Forest, whatever else it is (CR
// 205.3i), read through the layered view so a changeling or a land that was
// given the type counts.
func OldB679YouControlAForest(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtypeWord("Forest") {
			return true
		}
	}
	return false
}
