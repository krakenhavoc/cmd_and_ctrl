package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bottomless Vault — Land:
//
//	"This land enters tapped.
//	 You may choose not to untap this land during your untap step.
//	 At the beginning of your upkeep, if this land is tapped, put a
//	 storage counter on it.
//	 {T}, Remove any number of storage counters from this land: Add {B}
//	 for each storage counter removed this way."
//
// The tapped entry is the self-replacement; the untap opt-out is how
// its controller keeps it tapped through the untap step so the upkeep
// trigger has something to count. The trigger carries an intervening
// "if" (CR 603.4): it is read as the upkeep begins and again as the
// ability resolves, so a Vault untapped or tapped in response is
// judged on how it stands then. The payout is a mana ability whose
// cost removes any number of storage counters, zero included; the
// mana counts the counters that actually came off, through the
// paid-cost record (Mage-Ring Network's shape), so removing none adds
// nothing and the land still taps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e43413e4-be17-49af-978a-26210d05f52a",
		Name:         "Bottomless Vault",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Bottomless Vault — you may choose not to untap this land")},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, AllOf(ByYou, thisPermanentIsTapped),
				"Bottomless Vault — if it is tapped, put a storage counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok || info.Left || !info.Tapped {
						return nil
					}
					return putStorageCounterOnThis(g, item)
				}),
		},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				RemoveCounters: RemoveCountersXFromThis(game.CounterStorage, 0).RemoveCounters,
			},
			ProducedForPaid: ProducedPerCounterRemoved("{B}"),
			Label:           "Remove any number of storage counters: Add {B} for each",
		}},
	})
}

// thisPermanentIsTapped is an intervening "if this permanent is
// tapped" (CR 603.4) for a trigger's condition. The effect reads it
// again as the ability resolves.
func thisPermanentIsTapped(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source.Tapped
}
