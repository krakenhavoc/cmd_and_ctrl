package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Memory Vampire — Creature — Vampire Detective {4}{U}{B}, 4/4 (Ravnica:
// Clue Edition, Murders at Karlov Manor):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, any number of
//	 target players each mill that many cards. Then you may collect
//	 evidence 9. When you do, you may cast target nonland card from
//	 defending player's graveyard without paying its mana cost."
//
// "That many" is the combat damage dealt. The targets are any number of
// players (zero is legal), chosen as the trigger goes on the stack and
// re-checked as it resolves (CR 608.2b); each mills in turn order, and the
// collect-evidence offer waits for any prompt a mill asks (CR 608.2c). The
// collection is the effect form (RegisterCollectEvidence, CR 701.59): the
// controller picks graveyard cards of total mana value 9 or more, which may
// include cards just milled, or declines. "When you do" is a reflexive
// trigger (CR 603.12) whose target, a nonland card in the graveyard of the
// player who was dealt the damage, is chosen as it goes on the stack; the
// cast is free and optional, offered as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6031d854-10a8-43da-b5c4-b561194bcd4e",
		Name:            "Memory Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisDealsCombatDamageToAPlayer("Memory Vampire — any number of target players each mill that many cards; then you may collect evidence 9, and when you do, you may cast target nonland card from defending player's graveyard without paying its mana cost",
					effects.Do(
						effects.Instruction(memoryVampireMill),
						effects.Instruction(func(ctx *effects.Context) error {
							return memoryVampireEvidence.Ask(ctx, "Memory Vampire — collect evidence 9 (exile cards with total mana value 9 or more from your graveyard) to cast a nonland card from defending player's graveyard for free? Choose none to decline.")
						}),
					)),
				effects.TargetPlayer("any number of target players").WithCount(0, 0)),
		},
	})
}

func memoryVampireMill(ctx *effects.Context) error {
	n := ctx.Trigger().Event.Amount
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (effects.MillCards{Player: t.ID, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// memoryVampireEvidence is "you may collect evidence 9. When you do, …". Its
// key is an on-disk identity: never renamed, never reused.
var memoryVampireEvidence = effects.RegisterCollectEvidence("choose-cards/memory-vampire-collect-evidence", 9,
	func(ctx *effects.Context, _ []uuid.UUID) error {
		return effects.ReflexiveTrigger{
			Label:    "Memory Vampire — you may cast target nonland card from defending player's graveyard without paying its mana cost",
			Optional: "Memory Vampire — cast a nonland card from defending player's graveyard without paying its mana cost?",
			Players:  []uuid.UUID{ctx.Trigger().Event.Target},
			Body:     memoryVampireCastBody,
			Params:   game.EffectParams{Player: ctx.Trigger().Event.Target},
		}.Apply(ctx)
	})

// memoryVampireCastBody is the reflexive half. The target clause names the
// defending player's graveyard from Params.Player, fixed when evidence was
// collected. An on-disk key.
var memoryVampireCastBody = game.ReflexiveBody("memory-vampire/cast-from-graveyard",
	effects.BnsB13FreeCastFromGraveyard("Memory Vampire"),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCardInGraveyard("target nonland card from defending player's graveyard",
			effects.Nonland(), effects.OwnedBy(p.Player))
	})
