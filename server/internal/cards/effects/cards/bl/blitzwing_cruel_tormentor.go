package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blitzwing, Cruel Tormentor // Blitzwing, Adaptive Assailant (#3195,
// ADR 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {5}{B}, 6/5:
//
//	"More Than Meets the Eye {3}{B}
//	 At the beginning of your end step, target opponent loses life equal
//	 to the life that player lost this turn. If no life is lost this
//	 way, convert Blitzwing."
//
// Back face, Legendary Artifact — Vehicle, 3/5:
//
//	"Living metal
//	 At the beginning of combat on your turn, choose flying or
//	 indestructible at random. Blitzwing gains that ability until end of
//	 turn.
//	 Whenever Blitzwing deals combat damage to a player, convert it."
//
// "The life that player lost this turn" is read off the per-turn tally as
// the trigger resolves (it already counts damage). "If no life is lost
// this way" is judged on the life total itself, so a loss that was
// prevented or replaced to nothing converts it too. The random choice
// uses the game's keyed random stream (ADR 0054), so a replay picks the
// same ability.
//
// No simplification.
const BlitzwingOracleID = "9a10668a-df42-40c2-88d1-ad3b2aba5b03"

// blitzwingRandomAbilities are the two stand-ins the keyed stream picks
// between; their order is fixed so a replay draws the same one.
var blitzwingRandomAbilities = []struct {
	id      uuid.UUID
	keyword string
}{
	{uuid.MustParse("5f1a5b7e-3c1a-4b43-8d58-0b4f6a7a0001"), "flying"},
	{uuid.MustParse("5f1a5b7e-3c1a-4b43-8d58-0b4f6a7a0002"), "indestructible"},
}

func init() {
	effects.Register(effects.Spec{
		OracleID:         BlitzwingOracleID,
		Name:             "Blitzwing, Cruel Tormentor",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.MoreThanMeetsTheEye("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourEndStep("Blitzwing, Cruel Tormentor — target opponent loses life equal to the life they lost this turn",
				blitzwingEndStep), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     BlitzwingOracleID + "#1",
		Name:         "Blitzwing, Adaptive Assailant",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.LivingMetal()},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Blitzwing, Adaptive Assailant — flying or indestructible at random until end of turn",
				blitzwingRandomAbility),
			effects.WheneverThisDealsCombatDamageToAPlayer("Blitzwing, Adaptive Assailant — convert it", effects.Do(effects.ConvertThis{})),
		},
	})
}

func blitzwingEndStep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		lost := effects.B18LifeLostThisTurn(g, t.ID)
		convert := ctx.ItemTail(blitzwingConvert)
		if lost <= 0 {
			return convert.Run(g, nil)
		}
		return g.ChangePlayerLifeKeyedForEffect(item.SourceCardID, t.ID, -lost, convert)
	}
	// With no legal target the ability is removed from the stack
	// (CR 608.2b) and nothing, the convert included, happens.
	return nil
}

// blitzwingConvert is "If no life is lost this way, convert Blitzwing", told
// the life that moved. An on-disk key: never renamed, never reused.
var blitzwingConvert = effects.CardTail("blitzwing/convert-if-none-lost", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Amount < 0 {
		return nil // life was lost this way
	}
	return effects.ConvertThis{}.Apply(ctx)
})

func blitzwingRandomAbility(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ids := make([]uuid.UUID, len(blitzwingRandomAbilities))
	for i, a := range blitzwingRandomAbilities {
		ids[i] = a.id
	}
	pick := ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
	if len(pick) == 0 {
		return nil
	}
	for _, a := range blitzwingRandomAbilities {
		if a.id == pick[0] {
			return effects.GrantKeywordUntilEOT{
				Target:   ctx.Source(),
				Keywords: []string{a.keyword},
				Label:    "Blitzwing, Adaptive Assailant — gains " + a.keyword + " until end of turn",
			}.Apply(ctx)
		}
	}
	return nil
}
