package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// anyPlayerControlsColor reports whether any player controls a
// permanent of the given wire colour ("B"), read through the layered
// view. The Knights of Grace and Malice gate on it (#4465).
func anyPlayerControlsColor(g *game.Game, color string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.HasColor(color) {
			return true
		}
	}
	return false
}

// registerHexproofKnight registers "First strike; hexproof from
// <hexColor>; this creature gets +1/+0 as long as any player controls a
// <otherColor> permanent" (Knight of Grace, Knight of Malice). The two
// cards share the one body, so this is the one helper.
func registerHexproofKnight(oracleID, name, hexColor, otherColor, otherWord string) {
	pump := game.StaticAbility{
		Label:    "Gets +1/+0 as long as any player controls a " + otherWord + " permanent.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return selfCreature(target, g, source) && anyPlayerControlsColor(g, otherColor)
		},
		Apply: pumpBy(1, 0),
	}
	Register(Spec{
		OracleID:        oracleID,
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", game.HexproofFromColor(hexColor)},
		Static:          []game.StaticAbility{pump},
	})
}

// registerReturnWithCounterSpell registers "Until end of turn, target
// creature gains 'When this creature dies, return it to the battlefield
// tapped under its owner's control with a +1/+1 counter on it'" (Feign
// Death, Undying Malice): one instant, one duration grant under `key`.
func registerReturnWithCounterSpell(oracleID, name, key string) {
	Register(Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Grants: []AbilityGrant{{
			Key: key,
			Triggered: []game.TriggeredAbility{
				WhenThisDies(name+" — return it with a +1/+1 counter",
					returnThisCreatureFromGraveyard(map[string]int{game.CounterPlusOne: 1}, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control with a +1/+1 counter on it.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{key},
				Label:  name + " — until end of turn, it returns when it dies",
			}.Apply(ctx)
		},
	})
}

// edictGainToughness is "Target player sacrifices a creature of their
// choice. You gain life equal to that creature's toughness" (Consuming
// Vapors, Tribute to Hunger): the target player is the first legal player
// target, and the life waits for the sacrifice (consumingVaporsGain).
// The carry is the caster, then the victim.
func edictGainToughness(ctx *Context, name string) error {
	victim, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return ctx.Game.PlayerSacrificesThenKeyedForEffect(ctx.Source(), victim,
		sacrificeSpec("a creature", Creature()),
		name+" — sacrifice a creature", 1,
		consumingVaporsGain, []uuid.UUID{ctx.Controller(), victim})
}
