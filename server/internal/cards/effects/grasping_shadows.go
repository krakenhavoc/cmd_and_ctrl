package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	graspingShadowsOracleID = "522a4b02-24c7-45d2-9097-2803cc9fffad"
	dreadCounter            = "dread"
)

// Grasping Shadows — Enchantment {3}{B}, the front of a transforming card
// whose back is Shadows' Lair (The Lost Caverns of Ixalan, #4470):
//
//	"Whenever a creature you control attacks alone, it gains deathtouch and
//	 lifelink until end of turn. Put a dread counter on this enchantment.
//	 Then if there are three or more dread counters on it, transform it."
//
// The three sentences are one trigger (CR 608.2c, in order): the lone
// attacker, a creature of yours, gains deathtouch and lifelink until the
// end of the turn, a dread counter goes on the enchantment, and then, with
// the counters counted after that one, it transforms at three or more. The
// attacker is read from the event; if it has left the battlefield by
// resolution it gains nothing, and the counter and the transform still
// follow.
//
// Shadows' Lair:
//
//	"{T}: Add {B}.
//	 {B}, {T}, Remove a dread counter from this land: You draw a card and
//	 you lose 1 life."
//
// A transformed permanent keeps its counters (CR 712.18), so the Lair spends
// the dread counters the enchantment gathered. Removing one is a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     graspingShadowsOracleID,
		Name:         "Grasping Shadows",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, "Grasping Shadows — the lone attacker gains deathtouch and lifelink until end of turn; put a dread counter on it; transform it at three",
				Do(Instruction(graspingShadowsGrant),
					CounterOnThis{Kind: dreadCounter, N: 1},
					Instruction(func(ctx *Context) error {
						c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
						if !ok || !onBattlefield(ctx.Game, ctx.Source()) || c.Counters[dreadCounter] < 3 {
							return nil
						}
						return TransformThis{}.Apply(ctx)
					}))),
		},
	})
	Register(Spec{
		OracleID:     graspingShadowsOracleID + "#1",
		Name:         "Shadows' Lair",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "{T}: Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}, Remove a dread counter from this land: You draw a card and you lose 1 life.",
			Cost:    Plus(ManaCost("{B}"), TapCost(), RemoveCountersFromThis(dreadCounter, 1)),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Effect:  Do(DrawCards{N: 1}, GainLife{Amount: -1}),
		}},
	})
}

// graspingShadowsGrant is "it gains deathtouch and lifelink until end of
// turn": the creature that attacked alone.
func graspingShadowsGrant(ctx *Context) error {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	attacker := ctx.Item.Trigger.Event.CardID
	if !onBattlefield(ctx.Game, attacker) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"deathtouch", "lifelink"}, Label: "Grasping Shadows — deathtouch and lifelink"}.Apply(ctx)
}
