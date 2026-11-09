package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jalira's Show — Plane — Las Vegas (punk, #4870):
//
//	"Whenever you attack, create a 1/1 red Performer creature token with
//	 flying.
//	 Whenever you planeswalk here or chaos ensues, destroy up to one target
//	 creature. When you do, its controller reveals cards from the top of
//	 their library until they reveal a creature card. The player puts that
//	 card onto the battlefield, then puts the rest on the bottom of their
//	 library in a random order."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// attack trigger is one per declaration (OncePerBatch). The second is one
// ability with two trigger conditions: the planeswalk to this plane and a
// chaos roll (CR 311.7). "When you do" is a reflexive trigger (CR 603.12)
// raised only when the creature really left the battlefield, so an
// indestructible or regenerated one reveals nothing; the victim's controller,
// read before the destruction, is the player who reveals, puts the creature
// card onto the battlefield and bottoms the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eea700a1-a5f5-4b9e-aa03-8e092deb7396",
		Name:         "Jalira's Show",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Jalira's Show — create a 1/1 red Performer creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 red Performer with flying"), N: 1}))),
			Targeting(OnAny([]game.EventKind{game.EventPlaneswalkTo, game.EventRollDie}, planesB11PlaneswalkedHereOrChaos,
				"Jalira's Show — destroy up to one target creature; when you do, its controller reveals until a creature card and puts it onto the battlefield",
				jalirasShowDestroy), TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}

// jalirasShowDestroy destroys the target and, if it really left the
// battlefield (the keyed tail is told what was destroyed), raises the
// reflexive trigger naming its controller (read before the destruction).
func jalirasShowDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	victim, ok := controllerOfTarget(ctx, t)
	if !ok {
		return nil
	}
	return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{t}, withCarry(ctx.ItemTail(jalirasShowDestroyed), victim))
}

// jalirasShowDestroyed raises the "when you do" reflexive trigger once the
// creature has really been destroyed, naming its controller. An on-disk key:
// never renamed, never reused.
var jalirasShowDestroyed = CardTail("jaliras-show/when-you-do-reveal", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	trigger := WhenYouDo("Jalira's Show — its controller reveals cards until a creature card and puts it onto the battlefield", jalirasShowRevealBody)
	trigger.Players = []uuid.UUID{tailCarry0(o)}
	return trigger.Apply(ctx)
})

// jalirasShowRevealBody is the reflexive trigger: the destroyed creature's
// controller reveals until a creature card, puts it onto the battlefield and
// the rest on the bottom in a random order. A registered body (ADR 0041 P9);
// an on-disk key, never renamed, never reused.
var jalirasShowRevealBody = game.ReflexiveBody("jaliras-show/reveal-creature", func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	ctx := NewContext(g, item)
	for _, ref := range ctx.Payload() {
		if ref.Kind != game.TargetPlayer {
			continue
		}
		run, hit := revealUntil(ctx, ref.ID, game.Card.IsCreature, "Jalira's Show — revealed until a creature card")
		return PutFromLibraryOntoBattlefield{
			Player:   ref.ID,
			Cards:    run,
			Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
			All:      true,
			ThenTail: PutRestOnBottomInRandomOrderTail,
		}.Apply(ctx)
	}
	return nil
}, nil)
