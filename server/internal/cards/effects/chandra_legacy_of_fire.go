package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Legacy of Fire — Legendary Planeswalker — Chandra {4}{R},
// loyalty 3:
//
//	"At the beginning of your end step, Chandra deals X damage to each
//	 opponent, where X is the number of planeswalkers you control.
//	 +1: Add {R} for each planeswalker you control.
//	 0: Remove a loyalty counter from each of any number of permanents
//	    you control. Exile that many cards from the top of your library.
//	    You may play them this turn."
//
// The end-step trigger counts the planeswalkers you control as it
// resolves, and Chandra (or her last-known self) deals the damage.
//
// +1 is a loyalty ability, not a mana ability (CR 605.1a): it uses the
// stack, and the mana arrives as it resolves, one {R} for each
// planeswalker you control then (ManaN), Chandra included.
//
// 0 asks, as it resolves, which permanents you control to take a loyalty
// counter from: any number of them, among those with a loyalty counter
// to remove (a permanent with none can't have one removed, so it is not
// offered). Chandra herself and a Gatewatch Beacon are both legal picks.
// The answer is a keyed continuation (choose-cards/chandra-legacy-of-fire-
// remove), which checks each pick again, removes one counter from each,
// and exiles that many cards from the top of your library with a
// play-this-turn permission (the impulse exile Jeska's Will uses). "Play",
// so a land among them can be your land drop. Choosing none exiles
// nothing.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "8dfa789d-ae1c-4f18-a55c-f49a363a5e33",
		Name:         "Chandra, Legacy of Fire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Chandra, Legacy of Fire — X damage to each opponent, X the planeswalkers you control",
				func(g *game.Game, item *game.StackItem) error {
					x := chandraLegacyWalkers(g, item.Controller)
					if x <= 0 {
						return nil
					}
					return damageToEachOpponent(g, item, x)
				}),
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Add {R} for each planeswalker you control.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := chandraLegacyWalkers(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return AddMana{Player: item.Controller, Amount: ManaN(n, "R")}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "0: Remove a loyalty counter from each of any number of permanents you control. Exile that many cards from the top of your library. You may play them this turn.",
				Cost:   LoyaltyCost(0),
				Effect: chandraLegacyAskWhichCounters,
			},
		},
	})
}

// chandraLegacyWalkers is the number of planeswalkers `you` control.
func chandraLegacyWalkers(g *game.Game, you uuid.UUID) int {
	return countControlled(g, you, func(c game.Card) bool { return c.IsPlaneswalker() })
}

// chandraLegacyCanGive reports whether c is a permanent `you` control
// with a loyalty counter to remove.
func chandraLegacyCanGive(c game.Card, you uuid.UUID) bool {
	return c.Controller == you && c.Counters[game.CounterLoyalty] > 0
}

// chandraLegacyAskWhichCounters is the 0: the controller picks any
// number of their permanents with a loyalty counter.
func chandraLegacyAskWhichCounters(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if chandraLegacyCanGive(c, item.Controller) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Chandra, Legacy of Fire — remove a loyalty counter from each of any number of permanents you control; exile that many cards from the top of your library, and you may play them this turn",
		Cards:    candidates,
		Min:      0,
		Max:      len(candidates), // "any number"
		Zone:     game.ZoneBattlefield,
		ThenKey:  chandraLegacyChosen,
	})
	return nil
}

// chandraLegacyChosen is the keyed answer. Its key is an on-disk
// identity: never renamed, never reused.
var chandraLegacyChosen = game.RegisterChooseCardsThen("choose-cards/chandra-legacy-of-fire-remove", chandraLegacyRemoveThenExile)

// chandraLegacyRemoveThenExile removes one loyalty counter from each
// pick that still has one and is still the chooser's, then exiles that
// many cards from the top of the chooser's library, playable this turn.
func chandraLegacyRemoveThenExile(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	removed := 0
	for _, id := range r.Picked {
		if !onBattlefield(g, id) {
			continue
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok || !chandraLegacyCanGive(c, r.Chooser) {
			continue
		}
		if err := g.AddCounterForEffect(id, game.CounterLoyalty, -1); err != nil {
			return err
		}
		removed++
	}
	if removed == 0 {
		return nil
	}
	_, err := g.ExileTopWithPermissionForEffect(r.Chooser, r.Chooser, removed, game.CastPermission{})
	return err
}
