package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Mechanist, Aerial Artisan — Legendary Creature — Human Artificer
// Ally {2}{U}, 1/3:
//
//	"Whenever you cast a noncreature spell, create a Clue token.
//	 {T}: Until end of turn, target artifact token you control becomes a
//	 3/1 Construct artifact creature with flying."
//
// The animation is one until-end-of-turn record (CR 613.7): artifact and
// creature added (layer 4), its creature types SET to Construct — "a
// [creature type] artifact creature" replaces any it had and keeps every
// other type (CR 205.1b, the ruling) — base 3/1 (layer 7b, so counters
// and pumps still apply, the rulings) and flying (layer 6). Its other
// abilities stay. The effect is pinned to the token object (CR 400.7).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "bad27b43-7716-4ce1-82ba-2a6d437363e7",
		Name:         "The Mechanist, Aerial Artisan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "The Mechanist, Aerial Artisan — create a Clue token",
				Do(CreateToken{Template: ClueToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Until end of turn, target artifact token you control becomes a 3/1 Construct artifact creature with flying.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target artifact token you control", Artifact(), IsTokenPredicate(), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				mods := []game.Mod{
					game.AddTypesMod("Artifact", "Creature"),
					game.SetCreatureTypesMod(ctx.TextSubtypes("Construct")...),
				}
				mods = append(mods, game.SetBasePTMods(3, 1)...)
				mods = append(mods, game.AddKeywordsMod("flying"))
				return untilEndOfTurn(ctx, id, nil,
					"The Mechanist, Aerial Artisan — a 3/1 Construct artifact creature with flying", mods...)
			},
		}},
	})
}
