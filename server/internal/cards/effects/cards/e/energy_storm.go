package cardse

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Energy Storm — Enchantment {1}{W} (Ice Age, S82 OLD batch 838, #5636):
//
//	"Cumulative upkeep {1}
//	 Prevent all damage that would be dealt by instant and sorcery spells.
//	 Creatures with flying don't untap during their controllers' untap steps."
//
// The prevention is a standing CR 615 prevention over every recipient, read off
// the source as the damage would be dealt: it must be an instant or sorcery
// SPELL (a card in a graveyard dealing damage through a triggered ability is not
// one). The untap lock is Blizzard's, a layered flying test (CR 502.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "567d3de7-8d56-4b6d-a59a-f8674172f595",
		Name:         "Energy Storm",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Energy Storm — cumulative upkeep {1}", "{1}")},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.SourceLKI == nil {
					return false
				}
				if _, isSpell := g.DamageSourceSpellControllerForEffect(ev.DamageSource); !isSpell {
					return false
				}
				return slices.Contains(ev.SourceLKI.Types, "Instant") || slices.Contains(ev.SourceLKI.Types, "Sorcery")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			Label:      "Energy Storm — prevent all damage that would be dealt by instant and sorcery spells",
		}},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.And(effects.Creature(), effects.HasKeyword("flying")))},
	})
}
