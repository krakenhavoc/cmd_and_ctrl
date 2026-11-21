package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noctis, Heir Apparent — Legendary Creature — Human Noble {W}{U}{B}, 2/3
// (Final Fantasy, #4462):
//
//	"Whenever a creature you control enters during combat, you may attach
//	 target Equipment you control to target creature you control.
//	 Warp-Strike — {3}: Exile Noctis. Return it to the battlefield under
//	 its owner's control tapped and attacking at the beginning of that
//	 player's next declare attackers step. It can't be blocked that
//	 combat."
//
// The trigger is Brass Squire's two-clause attach behind a CR 603.5 "you
// may", on a creature you control entering during the combat phase (CR 506.1,
// The Blue Spirit's reading, Noctis itself included). The attach is
// AttachClauseTargets: an Equipment that can't equip that creature stays put.
//
// Warp-Strike exiles Noctis and schedules a CR 603.7 delayed trigger for the
// beginning of the owner's next declare attackers step, controlled by the
// owner so the return is theirs. It brings back the card it exiled only
// while it is still in exile, tapped and attacking a player, planeswalker or
// battle of the owner's choosing (CR 508.4): never declared as an attacker, so
// no "whenever attacks" triggers. It returns even when the owner declares no
// attackers (the ruling). The new object can't be blocked until end of that
// combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7abe351e-0889-4fc1-91ae-8372bcd28acb",
		Name:         "Noctis, Heir Apparent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.On(game.EventETB, noctisCreatureEnteredDuringCombat,
				"Noctis, Heir Apparent — you may attach target Equipment you control to target creature you control",
				effects.AttachClauseTargets),
				effects.TwoSlotAttachTargets(
					"target Equipment you control", []effects.CardPredicate{effects.YouControl()},
					"target creature you control", []effects.CardPredicate{effects.YouControl()},
				)),
				"Noctis, Heir Apparent — attach target Equipment you control to target creature you control?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Warp-Strike — {3}: Exile Noctis. Return it to the battlefield under its owner's control tapped and attacking at the beginning of that player's next declare attackers step. It can't be blocked that combat.",
			Cost:    effects.ManaCost("{3}"),
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				if err := (effects.ExileTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
					return err
				}
				return effects.ScheduleDelayedTrigger{
					At:         game.StepDeclareAttackers,
					Label:      "Noctis, Heir Apparent — return it tapped and attacking; it can't be blocked that combat",
					Controller: c.Owner,
					TurnOf:     c.Owner,
					Cards:      []uuid.UUID{item.SourceCardID},
					Body:       noctisWarpStrikeReturnBody,
				}.Apply(ctx)
			},
		}},
	})
}

// noctisCreatureEnteredDuringCombat is "a creature you control enters
// during combat", Noctis itself included.
func noctisCreatureEnteredDuringCombat(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && g.Turn.Phase == game.PhaseCombat
}

// noctisWarpStrikeReturnBody returns the exiled card tapped and attacking.
// An on-disk key: never renamed, never reused.
var noctisWarpStrikeReturnBody = game.SimpleDelayedBody("noctis/warp-strike-return", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			return effects.FinB11PutAttackingThen(ctx, t.ID, game.ZoneExile, "Noctis, Heir Apparent", noctisEnteredUnblockable, noctisAttackPicked)
		}
	}
	return nil
})

// noctisAttackPicked is the answer to "what does it attack?". An on-disk
// key: never renamed, never reused.
var noctisAttackPicked = effects.OptionPickThen("option-pick/noctis-warp-strike-attack", func(ctx *effects.Context, r game.OptionPicked) error {
	card, target, err := effects.FinB11PickedAttack(r)
	if err != nil {
		return nil
	}
	return effects.FinB11PutAttacking(ctx.Game, ctx.Controller(), card, game.ZoneExile, target, ctx.Tail(noctisEnteredUnblockable))
})

// noctisEnteredUnblockable makes the new object unblockable for the rest of
// the combat. An on-disk key: never renamed, never reused.
var noctisEnteredUnblockable = effects.EnteredTail("noctis/entered-unblockable", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	d, ok := effects.DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBeBlocked)},
		Duration: d,
		Label:    "Noctis, Heir Apparent — can't be blocked this combat",
	}.Apply(ctx)
})
