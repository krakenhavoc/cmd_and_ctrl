package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// uds_b2_helpers.go — the shapes Urza's Destiny batch 2 (#5621) prints
// more than once: the "{2}, Sacrifice this <permanent>: Draw a card."
// rows, the dies triggers that pump or shrink a target creature, Mantis
// Engine's two keyword rows, and the "Tap an untapped creature you
// control" cost Opposition and Field Surgeon share. Its own file, as
// #231 asks of a batch, so a concurrent batch never collides here.
//
// Append-only: add a builder, never change what one means.

// udsB2SacrificeThisDrawRow is "{2}, Sacrifice this <what>: Draw a
// card." (Marker Beetles and Plague Dogs print it on a creature,
// Capashen Standard and Illuminated Wings on an Aura). The permanent is
// sacrificed as part of the cost (CR 701.21a), so the draw happens even
// if it is gone by resolution; with no {T} in the cost, a creature that
// arrived this turn can pay it (CR 302.6).
func udsB2SacrificeThisDrawRow(what string) ActivatedAbility {
	return ActivatedAbility{
		Label:   "{2}, Sacrifice this " + what + ": Draw a card.",
		Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
		Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
		Effect:  ebDrawACard,
	}
}

// udsB2WhenThisDiesTargetCreatureGets is "When this creature dies,
// target creature gets +P/+T until end of turn." (Marker Beetles'
// +1/+1, Disease Carriers' -2/-2). It is a leaves-the-battlefield
// trigger (CR 603.6c, 603.10a), mandatory, and any creature may be its
// target, the controller's own included when nothing else is there. A
// target gone by resolution is skipped (CR 608.2b). The purpose is a
// benefit for a pump and a harm for a shrink, so the bot aims each at
// the right side of the table.
func udsB2WhenThisDiesTargetCreatureGets(name string, power, toughness int) game.TriggeredAbility {
	pt := signedPT(power, toughness)
	purpose := ForTargets(BenefitUntilEOT(0, power, toughness))
	if power < 0 || toughness < 0 {
		purpose = ForTargets(HarmUntilEOT(0, power, toughness))
	}
	return TriggerWithPurpose(Targeting(
		WhenThisDies(name+" — target creature gets "+pt+" until end of turn",
			boostTargetUntilEOT(name+" — "+pt+" until end of turn", power, toughness)),
		TargetCreature("target creature")), purpose)
}

// udsB2ThisGainsRow is "<cost>: This creature gains <keyword> until end
// of turn." (Mantis Engine's flying and first strike rows): a layer-6
// grant pinned to the ability's source that ends in the cleanup step
// (CR 514.2). With no {T} in the cost it can be activated the turn the
// creature arrives.
func udsB2ThisGainsRow(cost, keyword, name string) ActivatedAbility {
	return ActivatedAbility{
		Label:   cost + ": This creature gains " + keyword + " until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
		Cost:    ManaCost(cost),
		Effect:  thisGainsKeywordUntilEndOfTurn(keyword, name+" — "+keyword+" until end of turn"),
	}
}

// udsB2TapAnUntappedCreatureYouControl is the cost "Tap an untapped
// creature you control" (Opposition, Field Surgeon). It is not the {T}
// symbol (CR 302.6), so a creature that arrived this turn may pay it,
// and a creature card with the ability may tap itself to pay it.
func udsB2TapAnUntappedCreatureYouControl() game.AbilityCost {
	return TapAnUntapped("an untapped creature you control", Creature())
}

// udsB2TapTheTarget is "Tap target <permanent>." as an activated
// ability's whole effect (Opposition). A target gone by resolution is
// skipped (CR 608.2b).
var udsB2TapTheTarget = Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
	return TapTarget{Target: id}.Apply(ctx)
}))
