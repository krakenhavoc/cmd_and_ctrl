package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// karganIntimidatorLabel is the activated ability's label, and with it
// the key its "hasn't been chosen this turn" memory is kept under.
const karganIntimidatorLabel = "{1}: Choose one that hasn't been chosen this turn"

// Kargan Intimidator — Creature — Human Warrior {1}{R}, 3/1:
//
//	"Cowards can't block Warriors.
//	 {1}: Choose one that hasn't been chosen this turn —
//	 • This creature gets +1/+1 until end of turn.
//	 • Target creature becomes a Coward until end of turn.
//	 • Target Warrior gains trample until end of turn."
//
// The first ACTIVATED ability with ADR 0097's restriction. The modes
// are announced with the activation (CR 602.2b) and recorded once the
// activation succeeds, so a used bullet is refused at activation with
// nothing paid, and the bot's enumerator never offers it. With all
// three used the ability cannot be activated again this turn.
//
// "Cowards can't block Warriors" is a block rule on the blocker's side
// of the pair, and it binds every creature at the table, not just this
// one's controller's — the printed line has no "you control". "Becomes
// a Coward" ADDS the subtype until end of turn, so the creature keeps
// its other creature types, pinned to that object (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "045bf1fd-f375-47a0-9983-7e2ebfed7854",
		Name:         "Kargan Intimidator",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnMatching(effects.OfCreatureType("Coward")), effects.OnMatching(effects.OfCreatureType("Warrior")),
				"Cowards can't block Warriors"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   karganIntimidatorLabel,
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}"),
			Modes: effects.ChooseOneNotChosenThisTurn(
				effects.ModeDoing("This creature gets +1/+1 until end of turn.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						if effects.SourceIsNewObject(ctx.Game, item) {
							return nil
						}
						return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1,
							Label: "Kargan Intimidator — +1/+1 until end of turn"}.Apply(ctx)
					}),
				effects.ModeDoing("Target creature becomes a Coward until end of turn.", effects.TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return effects.UntilEndOfTurn(ctx, t.ID, nil,
							"Kargan Intimidator — becomes a Coward until end of turn",
							game.AddSubtypesMod(ctx.TextSubtypes("Coward")...))
					}),
				effects.ModeDoing("Target Warrior gains trample until end of turn.",
					effects.TargetCreature("target Warrior", effects.OfCreatureType("Warrior")),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"},
							Label: "Kargan Intimidator — trample until end of turn"}.Apply(ctx)
					}),
			),
		}},
	})
}
