package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// uds_b3_helpers.go — the shapes Urza's Destiny batch 3 (#5621) prints
// more than once: the "{2}, Sacrifice this Aura: Draw a card." rows and
// Mantis Engine's two keyword rows. Its own file, as #231 asks of a
// batch, so a concurrent batch never collides here.
//
// Append-only: add a builder, never change what one means.

// udsB3SacrificeThisDrawRow is "{2}, Sacrifice this <what>: Draw a
// card." (Capashen Standard and Illuminated Wings print it on an Aura;
// Marker Beetles and Plague Dogs spell the same row out inline). The
// permanent is sacrificed as part of the cost (CR 701.21a), so the draw
// happens even if it is gone by resolution.
func UdsB3SacrificeThisDrawRow(what string) ActivatedAbility {
	return ActivatedAbility{
		Label:   "{2}, Sacrifice this " + what + ": Draw a card.",
		Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
		Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
		Effect:  EbDrawACard,
	}
}

// udsB3ThisGainsRow is "<cost>: This creature gains <keyword> until end
// of turn." (Mantis Engine's flying and first strike rows): a layer-6
// grant pinned to the ability's source that ends in the cleanup step
// (CR 514.2). With no {T} in the cost it can be activated the turn the
// creature arrives.
func UdsB3ThisGainsRow(cost, keyword, name string) ActivatedAbility {
	return ActivatedAbility{
		Label:   cost + ": This creature gains " + keyword + " until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
		Cost:    ManaCost(cost),
		Effect:  ThisGainsKeywordUntilEndOfTurn(keyword, name+" — "+keyword+" until end of turn"),
	}
}
