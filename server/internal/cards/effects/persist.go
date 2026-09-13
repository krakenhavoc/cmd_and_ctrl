package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Persist — Sorcery {1}{B} (EDHREC rank 1347):
//
//	"Return target nonlegendary creature card from your graveyard to
//	 the battlefield with a -1/-1 counter on it."
//
// Two-mana reanimation with a legend clause and a tax: the
// nonlegendary restriction is the target predicate (Not(Legendary),
// on the printed supertype — a card in a graveyard has no layer
// cache), "your graveyard" is ownership, and the card lands under
// its owner's control, which for a card from your own graveyard is
// you (Zombify's shape, not Reanimate's).
//
// Sandbox simplification, declared (the Victimize posture): the
// engine's graveyard-return path has no "enters with counters" hook
// for the RETURNING effect, so the creature enters without the
// counter and receives it a beat later inside the same resolution.
// The state check that follows resolution then does what the
// printed card does — a 1-toughness creature dies at once. What
// differs is the beat in between: an enters-the-battlefield trigger
// on the creature sees it at printed size. Weaker, never stronger:
// nothing can act in that window.
func init() {
	Register(Spec{
		OracleID:     "367d4cf2-270f-4236-af10-7d5e8ea8c9fe",
		Name:         "Persist",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The returned creature enters without the -1/-1 counter and receives it a moment later, so its own enters-the-battlefield abilities see it at printed size."},
		Targets: TargetCardInGraveyard("target nonlegendary creature card from your graveyard",
			YouOwn(), Creature(), Not(b05Legendary())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			targets := ctx.Targets()
			if len(targets) == 0 || targets[0].Kind != game.TargetCard {
				return nil
			}
			id := targets[0].ID
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
				return err
			}
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
				return nil
			}
			return AddCounter{Target: id, Kind: "-1/-1", N: 1}.Apply(ctx)
		},
	})
}
