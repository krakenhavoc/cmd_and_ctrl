package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wanderwine Prophets — Creature — Merfolk Wizard {4}{U}{U}, 4/4 (Special
// Guests):
//
//	"Champion a Merfolk (When this enters, sacrifice it unless you exile
//	 another Merfolk you control. When this leaves the battlefield, that
//	 card returns to the battlefield.)
//	 Whenever this creature deals combat damage to a player, you may
//	 sacrifice a Merfolk. If you do, take an extra turn after this one."
//
// Champion is the shared champion helper (s82_bns_b05_helpers.go, CR
// 702.72). The damage trigger asks "you may sacrifice a Merfolk" as it
// resolves over every Merfolk its controller controls, the Prophets
// included (a changeling counts), and takes the extra turn (CR 500.7) only
// if one really left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ab36915-b600-4970-961a-403f5da0dba1",
		Name:         "Wanderwine Prophets",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			wanderwineProphetsChampion.trigger("Wanderwine Prophets"),
			TriggerWithPurpose(WheneverThisDealsCombatDamageToAPlayer("Wanderwine Prophets — you may sacrifice a Merfolk; if you do, take an extra turn after this one",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Wanderwine Prophets — sacrifice a Merfolk? If you do, you take an extra turn after this one.",
						Subtype("Merfolk"), false, wanderwineProphetsSacrificed)
				}), game.Purpose{ExtraTurns: 1}),
		},
	})
}

// wanderwineProphetsChampion is "Champion a Merfolk". Its pick key is an
// on-disk identity.
var wanderwineProphetsChampion = newChampion("Wanderwine Prophets", "Merfolk", "choose-cards/s82-bns-b05-wanderwine-prophets-champion", game.TailRef{})

// wanderwineProphetsSacrificed is the "if you do": the extra turn. An
// on-disk key.
var wanderwineProphetsSacrificed = RegisterSacrificedOne("wanderwine-prophets/extra-turn",
	func(ctx *Context, _ uuid.UUID) error { return TakeExtraTurn{}.Apply(ctx) })
