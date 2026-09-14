package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extract from Darkness — Sorcery {3}{U}{B} (EDHREC rank 3237):
//
//	"Each player mills two cards. Then you put a creature card from
//	 a graveyard onto the battlefield under your control."
//
// A five-mana Reanimate that mills the table first. The mill is
// every live seat, the caster included, two cards each; the
// reanimation is the shared "from A graveyard … under YOUR control"
// body, so a creature taken from an opponent's pile changes hands.
//
// One declared simplification, weaker than printed: the creature
// card is chosen when the spell is CAST, as a target, rather than
// after the mill as the spell resolves. A resolution-time "choose a
// creature card from a graveyard" has no prompt — the pick_target
// continuation belongs to triggered abilities and spell copies, and
// the search prompt is library-only — so the choice rides the
// announce-time target clause instead. Two consequences, both
// weaker: the spell cannot pick a card it milled itself, and it
// needs a creature card in some graveyard to be cast at all (printed
// it can be cast into empty graveyards and simply mill). Targeting
// also means the CR 608.2b re-check: a target that left its
// graveyard in response fizzles the reanimation, and the mill still
// happens.
func init() {
	Register(Spec{
		OracleID:     "e597d8a1-3bbc-4001-b642-f4421447970f",
		Name:         "Extract from Darkness",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You choose the creature card when you cast the spell, as a target, rather than after the mill — so it can't be one of the cards the spell itself mills, and there has to be a creature card in some graveyard to cast it."},
		Targets:      targetCreatureInAnyGraveyard(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := b02EachPlayerMills(ctx.Game, item, 2); err != nil {
				return err
			}
			reanimateSingleTarget(ctx, item.Controller)
			return nil
		},
	})
}
