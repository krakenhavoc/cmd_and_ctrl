package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Friendly Neighborhood — Enchantment — Aura {3}{W}:
//
//	"Enchant land
//	 When this Aura enters, create three 1/1 green and white Human
//	 Citizen creature tokens.
//	 Enchanted land has "{1}, {T}: Target creature gets +1/+1 until end
//	 of turn for each creature you control. Activate only as a
//	 sorcery.""
//
// Roadside Assistance's enters trigger (the tokens are the Aura's
// controller's) over Squirrel Nest's granted activated ability on a
// land (ADR 0093). The granted ability is the land's controller's: the
// {1}, the tap, the "you" in "each creature you control" and the
// sorcery timing are all theirs, and a Friendly Neighborhood on an
// opponent's land feeds the opponent. The count is read as the ability
// resolves (CR 608.2h's last known value is not needed: the pump is
// applied once, on resolution), the target included in it.
//
// No simplification.
const friendlyNeighborhoodGrant = "friendly-neighborhood/pump-for-each-creature"

func init() {
	Register(Spec{
		OracleID:     "10fe402e-8d6f-4058-9f42-86f229055359",
		Name:         "Friendly Neighborhood",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Grants: []AbilityGrant{{
			Key: friendlyNeighborhoodGrant,
			Activated: []ActivatedAbility{{
				Label:        "{1}, {T}: Target creature gets +1/+1 until end of turn for each creature you control. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{1}"), TapCost()),
				Targets:      TargetCreature("target creature"),
				SorcerySpeed: true,
				Effect:       friendlyNeighborhoodPump,
			}},
			Text: "{1}, {T}: Target creature gets +1/+1 until end of turn for each creature you control. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(friendlyNeighborhoodGrant)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Friendly Neighborhood — create three 1/1 green and white Human Citizen creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 green and white Human Citizen"), N: 3})),
		},
	})
}

// friendlyNeighborhoodPump is the granted ability's body: the pump is
// the number of creatures its controller has as it resolves.
func friendlyNeighborhoodPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := countControlled(g, item.Controller, MatchCreature)
	return vsPumpTheTarget(ctx, n, n, "Friendly Neighborhood — +1/+1 for each creature you control")
}
