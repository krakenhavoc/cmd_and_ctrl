package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haunted One — Legendary Enchantment — Background {2}{B}:
//
//	"Commander creatures you own have "Whenever this creature becomes
//	 tapped, it and other creatures you control that share a creature
//	 type with it each get +2/+0 and gain undying until end of turn.""
//
// A layer-6 grant (ADR 0093) to each commander creature you own, under
// anyone's control. "Becomes tapped" is any tap: attacking, an ability
// cost or an effect (ThisBecameTapped). The creatures are found as the
// trigger resolves (CR 611.2c), sharing a type with the commander as it
// is then, or as it last existed if it has left; it gets the bonus only
// if it is still there. The undying is the engine's (#2075).
//
// No simplification.
const hauntedOneGrant = "haunted-one/tapped-pump-and-undying"

func init() {
	Register(Spec{
		OracleID:     "689cd6a8-8be0-49a6-9758-a81cc9f55cc8",
		Name:         "Haunted One",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: hauntedOneGrant,
			Triggered: []game.TriggeredAbility{
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
					"Haunted One — it and creatures sharing a type with it get +2/+0 and undying until end of turn",
					hauntedOneTapped),
			},
			Text: "Whenever this creature becomes tapped, it and other creatures you control that share a creature type with it each get +2/+0 and gain undying until end of turn.",
		}},
		Static: []game.StaticAbility{
			GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsCommander && target.Owner == source.Controller
			}, hauntedOneGrant),
		},
	})
}

func hauntedOneTapped(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	hostID := item.SourceCardID
	hostHere := !host.Left
	return ScopedEffectFor{
		Match: func(g *game.Game, caster uuid.UUID, c game.Card) bool {
			if !c.IsCreature() || c.Controller != caster {
				return false
			}
			if c.InstanceID == hostID {
				return hostHere
			}
			return sharesACreatureTypeWith(host, c)
		},
		Mods:     []game.Mod{game.ModifyPTMod(2, 0), game.AddKeywordsMod(game.KeywordUndying)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Haunted One — +2/+0 and undying until end of turn",
	}.Apply(ctx)
}
