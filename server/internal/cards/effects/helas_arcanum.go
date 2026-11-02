package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hela's Arcanum — Instant — Arcane {3}{B}{B} (Marvel, mdd, #4460):
//
//	"Choose up to five {P} worth of modes. You may choose the same mode
//	 more than once.
//	 {P} — Target creature gets +1/+0 and gains deathtouch until end of turn.
//	 {P}{P} — Until end of turn, target creature gains "When this creature
//	 dies, return it to the battlefield tapped under its owner's control."
//	 {P}{P}{P} — Destroy each non-Zombie creature with power 3 or less."
//
// A pawprint spell (CR 700.2i, #4786): the chosen modes' {P} may add up
// to five, so the {P}{P}{P} sweep fits beside one {P}{P} grant or two
// {P} pumps, and the pump alone may be taken five times. Each occurrence
// has its own target (CR 700.2d) and the bodies run in printed order
// (CR 608.2c), so a creature pumped to power 4 by the first mode
// survives the sweep, and one given the return ability by the second
// comes back from it.
//
// The {P}{P} grant is Fake Your Own Death's: a dies trigger that
// returns the card it died as, tapped, under its owner's control. The
// sweep reads power as it resolves and spares Zombies by their
// effective types.
const helasArcanumReturn = "helas-arcanum/return"

func init() {
	Register(Spec{
		OracleID:     "4332a59b-f4b5-45c9-8a59-5baa59b46153",
		Name:         "Hela's Arcanum",
		Completeness: CompletenessFull,
		Modes: ChooseUpToPawprints(5, true,
			Pawprint(1, ModeWithPurpose(ModeDoing("Target creature gets +1/+0 and gains deathtouch until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return untilEndOfTurn(ctx, t.ID, nil, "Hela's Arcanum — +1/+0 and deathtouch",
						game.ModifyPTMod(1, 0), game.AddKeywordsMod("deathtouch"))
				}), ForTargets(BenefitUntilEOT(0, 1, 0, "deathtouch")))),
			Pawprint(2, ModeDoing("Until end of turn, target creature gains \"When this creature dies, return it to the battlefield tapped under its owner's control.\"",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return GrantAbilitiesFor{
						Target: t.ID,
						Keys:   []string{helasArcanumReturn},
						Label:  "Hela's Arcanum — until end of turn, it returns when it dies",
					}.Apply(ctx)
				})),
			Pawprint(3, ModeWithPurpose(ModeDoing("Destroy each non-Zombie creature with power 3 or less.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					// The pump mode ran earlier in this resolution; read
					// power through the layers it changed.
					ctx.Game.RecomputeLayersIfStaleLocked()
					return DestroyAllMatching{Match: And(Creature(), Not(Subtype("Zombie")), PowerLE(3))}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}})),
		),
		Grants: []AbilityGrant{{
			Key: helasArcanumReturn,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Hela's Arcanum — return it to the battlefield tapped",
					returnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
	})
}
