package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bonder's Ornament — Artifact {3} (Commander 2020, #5513):
//
//	"{T}: Add one mana of any color.
//	 {4}, {T}: Each player who controls a permanent named Bonder's Ornament
//	 draws a card."
//
// The draw is decided as the ability resolves (CR 608.2h): every player who
// then controls a permanent with that name, the activator included only if
// they still do, draws one card, in turn order from the active player (CR
// 101.4). Each player draws once however many Ornaments they control. A
// face-down permanent has no name (CR 708.2a) and does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "634a61a1-47a4-4a23-b138-fb487e6a7cba",
		Name:         "Bonder's Ornament",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Each player who controls a permanent named Bonder's Ornament draws a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Instruction(bondersOrnamentDraws)),
		}},
	})
}

// bondersOrnamentDraws draws a card for each player who controls a
// permanent named Bonder's Ornament.
func bondersOrnamentDraws(ctx *Context) error {
	holders := map[uuid.UUID]bool{}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if !c.FaceDown && c.Name == "Bonder's Ornament" {
			holders[c.Controller] = true
		}
	}
	for _, p := range gapB34TableInTurnOrder(ctx) {
		if !holders[p] {
			continue
		}
		if err := (DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
