package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frenzied Fugue — Enchantment — Aura {3}{R} (Commander 2016, #5537):
//
//	"Enchant permanent
//	 When this Aura enters and at the beginning of your upkeep, gain control of
//	 enchanted permanent until end of turn. Untap that permanent. It gains haste
//	 until end of turn."
//
// One body on two triggers (CR 603.2): the enchanted permanent is read as the
// trigger resolves (nothing if the Aura has gone), control changes until end of
// turn (CR 611.2b), then it untaps and gains haste. Assault Suit's lend with the
// Aura's controller as the taker.
//
// No simplification.
func init() {
	const label = "Frenzied Fugue — gain control of enchanted permanent until end of turn, untap it, it gains haste"
	steal := func(g *game.Game, item *game.StackItem) error {
		host := effects.AttachedHostFor(g, item.SourceCardID)
		if host == nil {
			return nil
		}
		ctx := effects.NewContext(g, item)
		if err := (effects.GainControl{Target: host.InstanceID, Duration: effects.DurationUntilEndOfTurn(ctx),
			Label: "Frenzied Fugue — gain control until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.UntapTarget{Target: host.InstanceID}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: host.InstanceID, Keywords: []string{"haste"},
			Label: "Frenzied Fugue — haste until end of turn"}.Apply(ctx)
	}
	effects.Register(effects.Spec{
		OracleID:     "b3f9bd29-8dd6-4915-8ffd-02e02f26fed8",
		Name:         "Frenzied Fugue",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, steal),
			effects.AtYourUpkeep(label, steal),
		},
	})
}
