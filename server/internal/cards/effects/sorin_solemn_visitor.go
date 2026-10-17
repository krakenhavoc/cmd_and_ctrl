package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorin, Solemn Visitor — Legendary Planeswalker — Sorin {2}{W}{B},
// starting loyalty 4 (#4077, deck request Sorin's Birthday Party):
//
//	"+1: Until your next turn, creatures you control get +1/+0 and
//	     gain lifelink.
//	 −2: Create a 2/2 black Vampire creature token with flying.
//	 −6: You get an emblem with "At the beginning of each opponent's
//	     upkeep, that player sacrifices a creature of their choice.""
//
// THE +1 is one effect on the creatures you control as it resolves
// (CR 611.2c), lasting until your next turn begins (CR 611.2b): a
// creature that arrives later gets neither the +1/+0 nor lifelink, and
// the effect is gone before anything untaps on your following turn.
//
// THE EMBLEM (CR 114, ADR 0064) is a triggered ability that works from
// the command zone and goes on the stack like any other trigger. Its
// owner never changes (CR 114.2), so "each opponent" is read against the
// owner, and the upkeep's player is read off the trigger's event: that
// player makes the choice and sacrifices (an edict, not a target, so
// hexproof is irrelevant, and a player with no creature sacrifices
// nothing). Every opponent's upkeep triggers it once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f1f1ae7-dd00-4825-832e-3a049cf8ce73",
		Name:         "Sorin, Solemn Visitor",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Sorin, Solemn Visitor emblem",
			Text:  "At the beginning of each opponent's upkeep, that player sacrifices a creature of their choice.",
			Triggered: []game.TriggeredAbility{
				AtEachOpponentsUpkeep("Sorin, Solemn Visitor emblem — that player sacrifices a creature", sorinSolemnEdict),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Until your next turn, creatures you control get +1/+0 and gain lifelink.",
				Cost:   LoyaltyCost(1),
				Effect: sorinSolemnAnthem,
			},
			{
				Label: "−2: Create a 2/2 black Vampire creature token with flying.",
				Cost:  LoyaltyCost(-2),
				Effect: Do(CreateToken{
					Template: TokenCard("2/2 black Vampire with flying"),
					N:        1,
				}),
			},
			{
				Label: "−6: You get an emblem with \"At the beginning of each opponent's upkeep, that player sacrifices a creature of their choice.\"",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// sorinSolemnAnthem is the +1: +1/+0 and lifelink on the creatures you
// control now, until your next turn.
func sorinSolemnAnthem(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ScopedEffectFor{
		Match:    And(Creature(), YouControl()),
		Mods:     []game.Mod{game.ModifyPTMod(1, 0), game.AddKeywordsMod("lifelink")},
		Duration: DurationUntilYourNextTurn(ctx, item.Controller),
		Label:    "Sorin, Solemn Visitor — +1/+0 and lifelink until your next turn",
	}.Apply(ctx)
}

// sorinSolemnEdict is the emblem's trigger: the player whose upkeep it
// is sacrifices a creature of their choice.
func sorinSolemnEdict(g *game.Game, item *game.StackItem) error {
	player := NewContext(g, item).Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	g.PlayerSacrificesForEffect(item.SourceCardID, player, sacrificeSpec("a creature", Creature()),
		"Sorin, Solemn Visitor emblem — sacrifice a creature")
	return nil
}
