package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Koma, Cosmos Serpent — Legendary Creature — Serpent {3}{G}{G}{U}{U},
// 6/6 (Avatar: The Last Airbender Eternal, #5075):
//
//	"This spell can't be countered.
//	 At the beginning of each upkeep, create a 3/3 blue Serpent creature
//	 token named Koma's Coil.
//	 Sacrifice another Serpent: Choose one —
//	 • Tap target permanent. Its activated abilities can't be activated
//	   this turn.
//	 • Koma gains indestructible until end of turn."
//
// The upkeep trigger fires in every player's upkeep. The sacrifice is
// "another" Serpent (the ruling): Koma can't pay with itself, but a
// Koma's Coil or any other Serpent will do. The first mode is Deadlock
// Trap's lock on any permanent, already tapped or not (the ruling), and
// it stops mana abilities too (the ruling: CantActivateMana).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ba0656f-e0d2-4735-9c66-3a00afa9fc23",
		Name:            "Koma, Cosmos Serpent",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Koma, Cosmos Serpent — create Koma's Coil",
				Do(CreateToken{Template: TokenCard("3/3 blue Serpent named Koma's Coil"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another Serpent: Choose one — Tap target permanent. Its activated abilities can't be activated this turn; or Koma gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict | game.AnswerSurvive},
			Cost:    SacrificeAnotherN(1, "another Serpent", HasSubtype("Serpent")),
			Modes: ChooseOne(
				ModeDoing("Tap target permanent. Its activated abilities can't be activated this turn.",
					TargetPermanent("target permanent"),
					komaTapAndLock),
				ModeDoing("Koma gains indestructible until end of turn.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						if !sourceIsStillThisPermanent(ctx.Game, item) {
							return nil
						}
						return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"}, Label: "Koma, Cosmos Serpent — indestructible until end of turn"}.Apply(ctx)
					}),
			),
		}},
	})
}

// komaTapAndLock taps the mode's still-legal target and stops its
// activated abilities, mana abilities among them, until end of turn.
func komaTapAndLock(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target:   t.ID,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantActivate | game.CantActivateMana)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Koma, Cosmos Serpent — its activated abilities can't be activated this turn",
	}.Apply(ctx)
}
