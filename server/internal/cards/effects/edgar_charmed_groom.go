package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Edgar, Charmed Groom // Edgar Markov's Coffin — {2}{W}{B} Legendary
// Creature — Vampire Noble 4/4 // Legendary Artifact (Innistrad: Crimson
// Vow, #5495):
//
//	Front: "Other Vampires you control get +1/+1.
//	        When Edgar dies, return it to the battlefield transformed under
//	        its owner's control."
//	Back:  "At the beginning of your upkeep, create a 1/1 white and black
//	        Vampire creature token with lifelink and put a bloodline
//	        counter on Edgar Markov's Coffin. Then if there are three or
//	        more bloodline counters on it, remove those counters and
//	        transform it."
//
// The dies trigger is the Ojer gods' shape (the card is returned from the
// graveyard as the object that died; one that has left is not found). The
// Coffin's upkeep is one resolution in printed order: token, counter,
// then the count read as it is now; "those counters" is every bloodline
// counter on it.
//
// No simplification.
func init() {
	const oracle = "96cb97ca-4ce0-4d33-86fb-4db4ce1956be"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Edgar, Charmed Groom",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Vampire"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			}, "Edgar, Charmed Groom — return it to the battlefield transformed under its owner's control",
				func(g *game.Game, item *game.StackItem) error {
					return ReturnFromGraveyard{
						Target: item.SourceCardID, Dest: game.ZoneBattlefield,
						Controller: uuid.Nil, Transformed: true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Edgar Markov's Coffin",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Edgar Markov's Coffin — create a 1/1 Vampire with lifelink and put a bloodline counter on this; at three, remove them and transform",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white and black Vampire with lifelink"), N: 1}).Apply(ctx); err != nil {
						return err
					}
					if err := (AddCounter{Target: ctx.Source(), Kind: "bloodline", N: 1}).Apply(ctx); err != nil {
						return err
					}
					c, ok := g.LookupCardForEffect(ctx.Source())
					if !ok || c.Counters["bloodline"] < 3 {
						return nil
					}
					if err := (AddCounter{Target: ctx.Source(), Kind: "bloodline", N: -c.Counters["bloodline"]}).Apply(ctx); err != nil {
						return err
					}
					return TransformThis{}.Apply(ctx)
				}),
		},
	})
}
