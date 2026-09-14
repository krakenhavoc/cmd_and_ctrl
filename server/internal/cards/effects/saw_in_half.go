package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saw in Half — Instant {2}{B} (EDHREC rank 585):
//
//	"Destroy target creature. If that creature dies this way, its
//	 controller creates two tokens that are copies of that creature,
//	 except their power is half that creature's power and their
//	 toughness is half that creature's toughness. Round up each
//	 time."
//
// Removal that is really an ETB doubler: point it at your own
// Mulldrifter and draw four. The shape is Beast Within's — read the
// controller and the creature's power and toughness BEFORE the
// destroy (afterwards the card is in a graveyard and both are stale
// for a permanent that had changed hands or been pumped) — then, if
// the creature really is in a graveyard, two CreateTokenCopy tokens
// under the victim's controller with the halved, rounded-up stats
// stamped as their base P/T. CreateTokenCopy copies from any zone,
// so the graveyard card is the right source, and its oracle ID rides
// onto the tokens so their triggers fire again.
//
// "If that creature dies this way" is read literally: an
// indestructible creature, a regenerated one, or a creature whose
// death was replaced by exile makes no tokens. Today only the
// graveyard check exists, which is the same test.
//
// Sandbox simplification, inherited from CreateTokenCopy (Hashaton's
// posture): a copied card's Spec.OnETB hook does not fire on the
// tokens — its Triggered EventETB abilities do, which is what the
// modern catalog uses, so a Sawed Mulldrifter still draws four.
func init() {
	Register(Spec{
		OracleID:     "eea18c55-8695-4ba1-9b38-3e7638692f5f",
		Name:         "Saw in Half",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Token copies skip the enters-the-battlefield effect on some cards."},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			victim, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			controller := victim.Controller
			power, toughness := victim.CurrentPower(), victim.CurrentToughness()
			if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if z := ctx.Game.FindCardZoneForEffect(target); z == nil || z.Kind != game.ZoneGraveyard {
				return nil // it did not die this way
			}
			return CreateTokenCopy{
				Controller: controller,
				Copy:       target,
				N:          2,
				Except: func(t *game.Card) {
					t.Power = (power + 1) / 2
					t.Toughness = (toughness + 1) / 2
				},
			}.Apply(ctx)
		},
	})
}
