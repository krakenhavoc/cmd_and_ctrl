package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Radiant Grace // Radiant Restraints (Innistrad: Crimson Vow, #5677).
//
// Front face, Enchantment — Aura {W}:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 and has vigilance.
//	 When enchanted creature dies, return this card to the battlefield
//	 transformed under your control attached to target opponent."
//
// Back face, Enchantment — Aura Curse:
//
//	"Enchant player
//	 Creatures enchanted player controls enter tapped."
//
// The dies trigger is harvested while the Aura is still attached, and
// its opponent is chosen as it goes on the stack. It resolves with the
// Aura in a graveyard, where the state-based action put it (CR 704.5m),
// and returns that card back face up, already attached to the target
// (ReturnThisCardTransformedAttachedToTarget): the effect names the
// player, so nobody chooses one as it enters (CR 303.4f) and the Curse
// can't land on its own controller. A target that has become illegal,
// or a card that left the graveyard in the meantime, leaves it where it
// is.
//
// Radiant Restraints is a CR 614 replacement on every creature entering
// under the enchanted player's control: cast, reanimated, put, flickered
// or created as a token. It has no way back to its front face; in a
// graveyard it is Radiant Grace again (CR 712.8a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     radiantGraceOracleID,
		Name:         "Radiant Grace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenEnchantedCreatureDies(
				"Radiant Grace — return this card to the battlefield transformed attached to target opponent",
				ReturnThisCardTransformedAttachedToTarget),
				TargetPlayer("target opponent", Opponent())),
		},
	})
	Register(Spec{
		OracleID:     radiantGraceOracleID + "#1",
		Name:         "Radiant Restraints",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Replacements: []game.ReplacementEffect{radiantRestraintsEnterTapped()},
	})
}

const radiantGraceOracleID = "93cddbad-92ef-4b30-a0e7-981aee714bce"

// radiantRestraintsEnterTapped is "Creatures enchanted player controls
// enter tapped": Authority of the Consuls' replacement, asked of the
// player the Curse is attached to in place of the opponents of its
// controller.
func radiantRestraintsEnterTapped() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src.AttachedTo.Kind != game.TargetPlayer {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.Controller == src.AttachedTo.ID && entering.IsCreature()
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Radiant Restraints: enters tapped",
	}
}
