package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// storage_lands.go — the five storage lands that print one colour:
// Bottomless Vault and Fallen Empires' Dwarven Hold, Hollow Trees,
// Icatian Store and Sand Silos (#5637).
//
//	"This land enters tapped.
//	 You may choose not to untap this land during your untap step.
//	 At the beginning of your upkeep, if this land is tapped, put a
//	 storage counter on it.
//	 {T}, Remove any number of storage counters from this land: Add {C}
//	 for each storage counter removed this way."
//
// The tapped entry is the self-replacement; the untap opt-out is how
// its controller keeps it tapped through the untap step so the upkeep
// trigger has something to count. The trigger carries an intervening
// "if" (CR 603.4): it is read as the upkeep begins and again as the
// ability resolves, so a land untapped or tapped in response is judged
// on how it stands then. The payout is a mana ability whose cost
// removes any number of storage counters, zero included; the mana
// counts the counters that actually came off, through the paid-cost
// record (Mage-Ring Network's shape), so removing none adds nothing and
// the land still taps.
//
// Append-only: add a builder, never change what one means.

// storageLand is the whole Spec of a one-colour storage land. `color`
// is the mana symbol's letter ("B", "R", …). The labels and the
// trigger's key carry the card's name, so each land's rows are its own.
func storageLand(oracleID, name, color string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf(name + " — you may choose not to untap this land")},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, AllOf(ByYou, thisPermanentIsTapped),
				name+" — if it is tapped, put a storage counter on it",
				putStorageCounterOnThisIfTapped),
		},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				RemoveCounters: RemoveCountersXFromThis(game.CounterStorage, 0).RemoveCounters,
			},
			ProducedForPaid: ProducedPerCounterRemoved(color),
			Label:           "Remove any number of storage counters: Add {" + color + "} for each",
		}},
	}
}

// putStorageCounterOnThisIfTapped is the storage land's upkeep trigger
// as it resolves: the intervening "if" read again (CR 603.4), so a land
// that left, or that was untapped in response, gets nothing.
func putStorageCounterOnThisIfTapped(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Left || !info.Tapped {
		return nil
	}
	return putStorageCounterOnThis(g, item)
}

// thisPermanentIsTapped is an intervening "if this permanent is
// tapped" (CR 603.4) for a trigger's condition. The effect reads it
// again as the ability resolves.
func thisPermanentIsTapped(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source.Tapped
}
