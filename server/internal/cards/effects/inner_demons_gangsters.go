package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inner Demons Gangsters — Creature — Human Rogue Villain {3}{B}, 3/4
// (Marvel Super Heroes Commander, #4455):
//
//	"Discard a card: This creature gets +1/+0 and gains menace until end
//	 of turn. Activate only as a sorcery."
//
// The cost is Dwarven Armorer's DiscardACard, the timing is Captivating
// Crew's SorcerySpeed. Both halves are until-end-of-turn effects on the
// Gangsters, which do nothing if they have left the battlefield.
//
// No simplification.
func init() {
	const label = "Inner Demons Gangsters — +1/+0 and menace until end of turn"
	Register(Spec{
		OracleID:     "40b43763-3fbf-4ecd-8187-5d2c2c9cda82",
		Name:         "Inner Demons Gangsters",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "Discard a card: This creature gets +1/+0 and gains menace until end of turn. Activate only as a sorcery.",
			Cost:         DiscardACard(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"menace"}, Label: label}.Apply(ctx)
			},
		}},
	})
}
