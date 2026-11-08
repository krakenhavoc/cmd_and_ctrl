package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aurelia's Vindicator — Creature — Angel {2}{W}{W}, 4/2 (Murders at
// Karlov Manor, #4469):
//
//	"Flying, lifelink, ward {2}
//	 Disguise {X}{3}{W}
//	 When this creature is turned face up, exile up to X other target
//	 creatures from the battlefield and/or creature cards from
//	 graveyards.
//	 When this creature leaves the battlefield, return the exiled cards
//	 to their owners' hands."
//
// The X is chosen as the disguise special action is taken (CR 107.3d,
// #5249), and CR 702.168e hands that X to the turned-face-up trigger:
// the special action puts it on EventTurnedFaceUp's Amount, and the
// target clause reads it off the trigger context, so a restore
// re-derives the same count. X = 0 targets nothing. Turned face up by
// an effect instead, nothing chose an X and the trigger exiles nothing.
//
// The exile and the return are Angel of Serenity's, whose two triggers
// read the same: the exile is linked to this object (ExileTarget.Linked,
// ADR 0155 §8) and the leave trigger returns every card still in exile
// linked to it to its owner's hand. If the Vindicator leaves before the
// exile resolves, the leave trigger finds nothing and the cards stay
// exiled (the 2024-02-02 ruling). The targets span the battlefield and
// the graveyards, as Angel of Serenity's do (#3172).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4f452ce7-1166-4810-afed-ea5e156a6f52",
		Name:             "Aurelia's Vindicator",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "lifelink"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{X}{3}{W}")},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Aurelia's Vindicator — ward {2}"),
			aureliasVindicatorTurnedUp(),
			On(game.EventLTB, Self, "Aurelia's Vindicator — return the exiled cards to their owners' hands", b36ReturnCardsExiledWithToOwnersHands),
		},
	})
}

// aureliasVindicatorTurnedUp is the turned-face-up trigger, its target
// count the X the disguise cost was paid with.
func aureliasVindicatorTurnedUp() game.TriggeredAbility {
	t := WhenThisIsTurnedFaceUp("Aurelia's Vindicator — exile up to X other target creatures", b36ExileChosenCreatures)
	t.TargetsFrom = aureliasVindicatorTargets
	return t
}

// aureliasVindicatorTargets is "up to X other target creatures from the
// battlefield and/or creature cards from graveyards", X the event's
// Amount (CR 702.168e). A pure read of the trigger context, so restore
// can re-derive it (no TargetsFromReadsBoard).
func aureliasVindicatorTargets(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	x := tc.Amount()
	if x <= 0 {
		return nil
	}
	spec := Another(TargetCreature("up to X other target creatures and/or creature cards in graveyards")).WithCount(0, x)
	spec.Zones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard}
	return spec
}
