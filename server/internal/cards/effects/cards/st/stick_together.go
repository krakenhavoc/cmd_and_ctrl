package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stick Together — Sorcery {3}{W}{W} (Commander Legends: Battle for Baldur's
// Gate, #5948):
//
//	"Each player chooses a party from among creatures they control, then
//	 sacrifices the rest. (To choose a party, choose up to one each of
//	 Cleric, Rogue, Warrior, and Wizard.)"
//
// Each player (APNAP, CR 101.4) who controls a creature with a party type is
// asked for their party: any set of creatures that can each stand for a
// different one of the four types (a changeling stands for any one). When
// everyone has answered, every creature each player controls outside their
// party is sacrificed at once, including creatures with no party type. A
// player with none of those is not asked and sacrifices all their creatures.
// Nothing targets.
//
// ADR 0126: it reads as a board wipe, declared as a partial creature
// sacrifice sweep (a party survives).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4abda0ad-0a2b-4bd9-81c7-56329e62311d",
		Name:         "Stick Together",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			var legs []game.ChooseCardsPrompt
			var seats []uuid.UUID
			for _, seat := range effects.ApnapLivingSeats(g) {
				if len(effects.PermanentsControlledByMatching(g, seat, effects.Creature())) == 0 {
					continue
				}
				seats = append(seats, seat)
				ids := effects.GapB233PartyCandidates(g, seat)
				if len(ids) == 0 {
					continue
				}
				legs = append(legs, game.ChooseCardsPrompt{
					Chooser:    seat,
					FromPlayer: seat,
					Source:     item.SourceCardID,
					Question:   "Stick Together — choose a party (up to one each of Cleric, Rogue, Warrior, and Wizard) from among creatures you control; you sacrifice the rest",
					Cards:      ids,
					Min:        0,
					Max:        4,
					Zone:       game.ZoneBattlefield,
					ThenKey:    effects.GapB233PartyLeg,
				})
			}
			if len(seats) == 0 {
				return nil
			}
			if len(legs) == 0 {
				return stickTogetherSacrificeRest(g, item.SourceCardID, seats, nil)
			}
			return g.ChoosePermanentsInOrderKeyedForEffect(legs, stickTogetherKeep, item.SourceCardID, game.RunData{Carry: seats})
		},
	})
}

// stickTogetherKeep is "then sacrifices the rest". Carry is every seat that
// controlled a creature. An on-disk key: never renamed, never reused.
var stickTogetherKeep = game.RegisterPromptRunThen("prompt-run/stick-together-keep", func(g *game.Game, r game.RunDone) error {
	return stickTogetherSacrificeRest(g, r.Source, r.Carry, r.Picks().Cards())
})

// stickTogetherSacrificeRest sacrifices, at once, every creature the seats
// control that is not in a party chosen.
func stickTogetherSacrificeRest(g *game.Game, source uuid.UUID, seats, kept []uuid.UUID) error {
	keep := map[uuid.UUID]bool{}
	for _, id := range kept {
		keep[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range seats {
		for _, id := range effects.PermanentsControlledByMatching(g, seat, effects.Creature()) {
			if !keep[id] {
				rest = append(rest, id)
			}
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(source, rest, game.KeyedTail{})
}
