package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise of the Witch-king — Sorcery {2}{B}{G} (EDHREC rank 1924):
//
//	"Each player sacrifices a creature of their choice. If you
//	 sacrificed a creature this way, you may return another permanent
//	 card from your graveyard to the battlefield."
//
// An edict that pays you back. The sacrifice is EachPlayerSacrifices
// with the controller included — every player picks their own, one
// prompt each, and a player with no creature is skipped (CR
// 701.17b). "If you sacrificed a creature this way" is whether the
// controller was handed a prompt: a player with a creature MUST
// sacrifice one, so the prompt being queued is the condition.
//
// Sandbox simplification, declared (the Mount Doom / Time Wipe
// posture): "you may return another permanent card from your
// graveyard" is a resolution-time choice made AFTER the sacrifices,
// and the engine has no pick-from-graveyard prompt with a
// continuation for a spell — the sacrifice prompts carry none. So
// the card to return is picked when the spell is cast, as an
// optional target ("up to one"), and it comes back as the spell
// resolves, before the sacrifice prompts are answered. Three
// consequences, all weaker than printed: opponents see the pick
// before the spell resolves; the creature you sacrifice cannot be
// the card that comes back (it is not in the graveyard yet when you
// pick — "another" for free, but also no choosing it deliberately);
// and if the picked card left the graveyard in response the spell
// is countered by game rules (CR 608.2b) and the edict does not
// happen either, where the printed card would still make everyone
// sacrifice. The returned permanent is on the battlefield before
// you answer your own sacrifice prompt, but is not offered by it:
// the options were listed when the prompt was queued.
func init() {
	Register(Spec{
		OracleID:     "3c86541c-3601-4a38-8872-39705e41303a",
		Name:         "Rise of the Witch-king",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The permanent card to return is picked when you cast the spell rather than after the sacrifices, so opponents can respond to the choice, and the creature you sacrifice to it can't be the one that comes back.",
		},
		Targets: TargetCardInGraveyard("up to one permanent card in your graveyard to return", YouOwn(), Permanent()).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			spec := sacrificeSpec("a creature", Creature())
			ctx.Game.EachPlayerSacrificesForEffect(ctx.Source(), controller, spec, "Rise of the Witch-king — sacrifice a creature")
			if ctx.Game.PlayerSacrificesForEffect(ctx.Source(), controller, spec, "Rise of the Witch-king — sacrifice a creature") == 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
			}
			return nil
		},
	})
}
