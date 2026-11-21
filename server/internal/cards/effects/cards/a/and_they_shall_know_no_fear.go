package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// And They Shall Know No Fear — Instant {1}{W} (Warhammer 40,000 Commander,
// #5948):
//
//	"Choose a creature type. Creatures you control of the chosen type get
//	 +1/+0 and gain indestructible until end of turn."
//
// The type is asked while the spell resolves (Kindred Charge's prompt); the
// creatures you control of it as it is answered are the set (CR 611.2c), so
// one that enters later is not affected. A changeling counts for every type
// (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ea6d165-2c30-41a4-8829-6d0510afc03f",
		Name:         "And They Shall Know No Fear",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID,
				"And They Shall Know No Fear — choose a creature type", ctx.ItemTail(andTheyShallKnowNoFearTypeChosen))
			return nil
		},
	})
}

// andTheyShallKnowNoFearTypeChosen gives the chosen type's creatures +1/+0
// and indestructible. An on-disk key: never renamed, never reused.
var andTheyShallKnowNoFearTypeChosen = effects.CreatureTypeChosen("then/and-they-shall-know-no-fear-type-chosen",
	func(ctx *effects.Context, t string, _ game.TailOutcome) error {
		match := effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype(t))
		if err := (effects.BoostUntilEOT{Match: match, Power: 1, Label: "And They Shall Know No Fear — +1/+0"}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"indestructible"},
			Label: "And They Shall Know No Fear — indestructible"}.Apply(ctx)
	})
