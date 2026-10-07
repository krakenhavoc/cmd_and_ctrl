package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyclave Relic — Artifact {3}:
//
//	"Kicker {3}
//	 Indestructible
//	 When this artifact enters, if it was kicked, create two tapped
//	 tokens that are copies of this artifact.
//	 {T}: Add one mana of any color."
//
// The token copies carry the Relic's oracle ID, so they are
// indestructible mana rocks too (every hook resolves through the
// catalog key). They were not cast, so they are not kicked and do not
// trigger again. The trigger's "if it was kicked" is an intervening if
// (CR 603.4): an unkicked Relic never puts the trigger on the stack. The
// copy is taken at resolution, from wherever the Relic is (CR 707.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb44b65d-6b56-4f20-a4fb-c5dd147a54c4",
		Name:            "Skyclave Relic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{3}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Skyclave Relic — kicked, create two tapped copies of this artifact", skyclaveRelicCopies),
		},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
	})
}

func skyclaveRelicCopies(g *game.Game, item *game.StackItem) error {
	tmpl, ok := TokenCopyTemplate(g, item.SourceCardID)
	if !ok {
		return nil
	}
	_, err := g.CreateTokensForEffect(item.Controller, tmpl, 2, game.TokenEntryOptions{Tapped: true})
	return err
}
