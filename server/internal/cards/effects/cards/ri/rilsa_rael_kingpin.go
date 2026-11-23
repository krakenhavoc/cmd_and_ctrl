package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rilsa Rael, Kingpin — {3}{U}{B} Legendary Creature — Human Rogue 2/5
//
//	"Deathtouch
//	 When Rilsa Rael enters, you take the initiative.
//	 Whenever you attack, target attacking creature gains deathtouch until end
//	 of turn. If you've completed a dungeon, that creature also gets +5/+0 and
//	 gains first strike and menace until end of turn."
//
// One trigger per declaration of attackers (CR 508.1m), whatever the number of
// creatures. The target is any attacking creature, chosen as the trigger goes
// on the stack and re-checked as it resolves (CR 608.2b). "If you've completed
// a dungeon" is a plain resolution check, not an intervening "if": the first
// sentence happens either way.
func init() {
	attack := effects.TriggerWithPurpose(effects.OncePerBatch(effects.Targeting(
		effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return effects.AttackDeclaredByYou(ev, source.Controller)
		}, "Rilsa Rael — target attacking creature gains deathtouch until end of turn; if you've completed a dungeon, it also gets +5/+0 and gains first strike and menace",
			rilsaRaelGrant),
		effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
		effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")))
	effects.Register(effects.Spec{
		OracleID:        "57f6a8c2-9d84-4640-9b92-8b8611e05773",
		Name:            "Rilsa Rael, Kingpin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Rilsa Rael"),
			attack,
		},
	})
}

func rilsaRaelGrant(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"deathtouch"}, Label: "Rilsa Rael — deathtouch"}).Apply(ctx); err != nil {
		return err
	}
	if !effects.YouCompletedADungeon(g, ctx.Controller()) {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: target, Power: 5, Label: "Rilsa Rael — +5/+0"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"first strike", "menace"}, Label: "Rilsa Rael — first strike and menace"}.Apply(ctx)
}
