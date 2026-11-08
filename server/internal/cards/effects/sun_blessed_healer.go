package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sun-Blessed Healer — Creature — Human Cleric {1}{W}, 3/1 (Foundations,
// #4465):
//
//	"Kicker {1}{W} (You may pay an additional {1}{W} as you cast this
//	 spell.)
//	 Lifelink (Damage dealt by this creature also causes you to gain that
//	 much life.)
//	 When this creature enters, if it was kicked, return target nonland
//	 permanent card with mana value 2 or less from your graveyard to the
//	 battlefield."
//
// Nullpriest of Oblivion's kicked enters trigger (ThisKickedAtLeast, an
// intervening "if" read off the permanent's carried kicker record, so an
// unkicked Healer puts nothing on the stack and asks for no target) with
// Primary Research's target: a nonland permanent card you own with mana
// value 2 or less, returned under your control if it is still there as
// the ability resolves. Lifelink rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1729c6f2-2452-4fbc-a1bc-3e625f2ad8e1",
		Name:            "Sun-Blessed Healer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Sun-Blessed Healer — kicked, return target nonland permanent card with mana value 2 or less from your graveyard to the battlefield",
					returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target nonland permanent card with mana value 2 or less from your graveyard",
					YouOwn(), NonlandPermanentCard(), ManaValueLE(2))),
		},
	})
}
