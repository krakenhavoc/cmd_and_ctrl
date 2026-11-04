package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flickering Ward — Enchantment — Aura for {W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 As this Aura enters, choose a color.
//	 Enchanted creature has protection from the chosen color. This effect
//	 doesn't remove this Aura.
//	 {W}: Return this Aura to its owner's hand."
//
// The card #4843 built CR 702.16n's exemption for. A white Aura that
// chooses white would fall off its own creature under CR 704.5m, so
// the protection instance it grants is minted with the Aura's id
// (ProtectionFromChosenColorSparingThisAura) and the attachment check
// passes over that one instance for this Aura alone. Another instance
// of protection from white, from a second Ward or anything else, still
// removes it, as CR 702.16n's last sentence says.
//
// The colour is the Ward of Lights shape: an as-enters choice
// (ChooseColorAsEnters) written onto Card.ChosenColor and read by the
// static on every recompute. The bounce is Batterskull's: a mana-cost
// activation with no target, at instant speed, so it saves itself
// from a Disenchant and comes back with a new colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e2ae59a-40a1-46ca-8f24-3231ebd4e4eb",
		Name:         "Flickering Ward",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		AsEnters:     ChooseColorAsEnters(game.ColorForProtection, "Flickering Ward"),
		Static: []game.StaticAbility{
			ProtectionFromChosenColorSparingThisAura(),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{W}: Return this Aura to its owner's hand",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
