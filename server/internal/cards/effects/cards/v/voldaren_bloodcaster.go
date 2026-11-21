package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Voldaren Bloodcaster // Bloodbat Summoner — Creature — Vampire Wizard
// {1}{B}, 2/1 // Creature — Vampire Wizard 3/3 (Innistrad: Crimson Vow,
// #5495):
//
//	Front: "Flying
//	        Whenever this creature or another nontoken creature you control
//	        dies, create a Blood token.
//	        Whenever you create a Blood token, if you control five or more
//	        Blood tokens, transform this creature."
//	Back:  "Flying
//	        At the beginning of combat on your turn, up to one target Blood
//	        token you control becomes a 2/2 black Bat creature with flying
//	        and haste in addition to its other types."
//
// The death trigger is Avacyn, Angel of Horror's condition (this creature, or
// another nontoken creature of its controller's, last-known). The transform
// trigger is one per Blood token created (Rosie Cotton's "whenever you create
// a token", narrowed to the new token being a Blood), with an intervening
// "if" (CR 603.4) that counts your Blood tokens again as it resolves; the
// Bloodcaster turns over in place (TransformThis), and does nothing if it has
// already turned over since the trigger fired. The back face's trigger
// targets as it goes on the stack, and the Blood token becomes a Bat for good
// (no duration, Case of the Filched Falcon's pinned indefinite effect): a
// 2/2 black Bat creature with flying and haste in addition to its other
// types, so it keeps its Blood subtype and its ability, and can be sacrificed
// for its draw.
//
// No simplification.
func init() {
	const oracle = "a720f3d3-c2d7-4e50-8cc6-600120380e9c"
	fiveBlood := effects.CondBoard{
		Name: "you control five or more Blood tokens",
		Test: func(ctx *effects.Context) bool {
			return effects.VowB05BloodTokensControlledBy(ctx.Game, ctx.Controller()) >= 5
		},
	}
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Voldaren Bloodcaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return effects.ThisDied(ev, source, ch, g) || effects.B19AnotherNontokenCreatureYouControlDied(ev, source, g)
				},
				Key:    "Voldaren Bloodcaster — create a Blood token",
				Effect: effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1}),
			},
			effects.InterveningIf(fiveBlood, effects.On(game.EventTokenCreated,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Actor != source.Controller {
						return false
					}
					tok, ok := g.LookupCardForEffect(ev.CardID)
					return ok && tok.HasSubtype("Blood") && effects.VowB05BloodTokensControlledBy(g, source.Controller) >= 5
				}, "Voldaren Bloodcaster — transform this creature", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Bloodbat Summoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Bloodbat Summoner — up to one target Blood token you control becomes a 2/2 black Bat creature with flying and haste in addition to its other types",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					bat := []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Bat")...),
						game.SetColorsMod(ctx.TextColors("B")...),
						game.SetBasePowerMod(2),
						game.SetBaseToughnessMod(2),
						game.AddKeywordsMod("flying", "haste"),
					}
					return effects.ScopedEffectFor{Target: id, Mods: bat, Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
						Label: "Bloodbat Summoner — becomes a 2/2 black Bat creature with flying and haste"}.Apply(ctx)
				}),
				effects.TargetPermanent("up to one target Blood token you control", effects.YouControl(),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return effects.IsToken(c) && c.HasSubtype("Blood") },
				).WithCount(0, 1)),
		},
	})
}
