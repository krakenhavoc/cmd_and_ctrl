package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kiora of Salt and Sand — Legendary Creature — Merfolk Noble {1}{G}{U}
// (Reality Fracture, tracker #2795, #2797):
//
//	"Whenever you attack, if you've activated a loyalty ability this turn,
//	 untap target attacking creature. It can't be blocked this turn.
//	 Planeswalkers you control have "[−8]: Create an 8/8 blue Leviathan
//	 creature token with hexproof.""
//
// The second line is the first user of the class grant (ADR 0093, with
// ADR 0109 §2's granted loyalty rows): a layer-6 bundle whose row costs
// −8, handed to every planeswalker its controller controls. The walker
// pays the 8 (CR 606.6) and the row shares the walker's one loyalty
// activation a turn with its own and with any other grant (CR 606.3).
// The token is the activator's, not Kiora's controller's, which only
// differs for a planeswalker you control that someone else activates, and
// nobody else may.
//
// "Whenever you attack" is one trigger per declaration (OncePerBatch, the
// Hollowmurk Siege shape: the engine emits EventAttack per creature), with
// a target clause on the trigger so it is chosen when the trigger goes on
// the stack. The intervening "if" (CR 603.4) is checked when the trigger
// would fire and again as it resolves, so a loyalty ability activated
// in the same turn counts however long ago, and a trigger whose condition
// has stopped being true does nothing. "You've activated a loyalty ability"
// reads the turn's activation events, a printed row or a granted one; the
// sandbox's manual loyalty verb announces none and is not counted.
//
// The target is any attacking creature, as printed, not only yours. A
// target that has left combat by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e70bb7f8-8098-49c5-929c-68f63cd821c5",
		Name:         "Kiora of Salt and Sand",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: kioraSaltAndSandGrant,
			Activated: []ActivatedAbility{{
				Label: "−8: Create an 8/8 blue Leviathan creature token with hexproof.",
				Cost:  LoyaltyCost(-8),
				// The bot reads this to know the row makes something worth
				// the eight loyalty it costs (ADR 0126 §6).
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("8/8 blue Leviathan with hexproof"),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			}},
			Text: "[−8]: Create an 8/8 blue Leviathan creature token with hexproof.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers(kioraSaltAndSandGrant)},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller) &&
						youActivatedALoyaltyAbilityThisTurn(g, source.Controller)
				}, kioraSaltAndSandTriggerLabel, kioraUntapAndUnblockable),
				TargetCreature("target attacking creature", AttackingCreature()))),
		},
	})
}

const (
	kioraSaltAndSandGrant        = "kiora-of-salt-and-sand/leviathan"
	kioraSaltAndSandTriggerLabel = "Kiora of Salt and Sand — untap target attacking creature; it can't be blocked this turn"
)

// kioraUntapAndUnblockable untaps the chosen attacker and makes it
// unblockable for the turn. The intervening "if" is re-checked here
// (CR 603.4).
func kioraUntapAndUnblockable(g *game.Game, item *game.StackItem) error {
	if !youActivatedALoyaltyAbilityThisTurn(g, item.Controller) {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return RestrictUntilEOT{
			Target:       t.ID,
			Restrictions: game.CantBeBlocked,
			Label:        "Kiora of Salt and Sand — can't be blocked",
		}.Apply(ctx)
	}
	return nil
}
