package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veiled Ascension — Enchantment {3}{W} (Murders at Karlov Manor Commander,
// #5368):
//
//	"When this enchantment enters, put a flying counter on each face-down
//	 creature you control.
//	 Face-down creatures you control enter with a flying counter on them.
//	 At the beginning of your upkeep, you may cloak the top card of your
//	 library."
//
// The entry clause is a CR 614 replacement on a face-down permanent
// entering under your control (Curator Beastie's shape): a face-down
// permanent is a creature unless a listing says otherwise. The upkeep
// cloak is a "you may" and is itself an entry the replacement sees.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbf0190a-198b-49cc-8c79-4028da07fa77",
		Name:         "Veiled Ascension",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{veiledAscensionFlyingCounter()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Veiled Ascension — put a flying counter on each face-down creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.FaceDownIsPermanent() && c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (AddCounter{Target: id, Kind: game.CounterFlying, N: 1}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
			Optional(
				AtYourUpkeep("Veiled Ascension — you may cloak the top card of your library", Do(CloakTopOfLibrary{})),
				"Veiled Ascension — cloak the top card of your library?"),
		},
	})
}

func veiledAscensionFlyingCounter() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.FaceDown == game.FaceDownNone {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || entering.Controller != src.Controller {
				return false
			}
			if ev.FaceDownListed == nil {
				return true
			}
			for _, t := range ev.FaceDownListed.Types {
				if t == "Creature" {
					return true
				}
			}
			return false
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterFlying, 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Veiled Ascension: face-down creatures you control enter with a flying counter",
	}
}
