package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reanimate — Sorcery {B}:
//
//	"Put target creature card from a graveyard onto the battlefield
//	under your control. You lose life equal to that card's mana
//	value."
//
// One black mana for anything in any graveyard, priced in life. The
// first reanimation spell in the catalog, and the card that proves
// ReturnFromGraveyard{Dest: ZoneBattlefield} is a real destination
// rather than a scaffold.
//
// The mana value is read BEFORE the move, the same discipline
// Swords to Plowshares and Feed the Swarm use: afterwards the card
// is a battlefield permanent and the graveyard lookup is gone.
//
// SANDBOX SIMPLIFICATIONS — three, all in the weaker-than-printed
// direction, and all worth reading before this card is trusted in a
// real game:
//
//  1. The target is restricted to YOUR OWN graveyard. Printed, it is
//     "a graveyard" — any of them. ReturnFromGraveyardForEffect
//     routes the card to the battlefield WITHOUT stamping a new
//     controller, so a creature lifted out of an opponent's
//     graveyard would arrive under their control and the caster
//     would have paid a card and a pile of life to hand them their
//     own creature back. Narrowing the target clause is the honest
//     fix until a change-of-control seam exists; "Mind Control /
//     aura attach is deferred" is the same missing machinery
//     (layer_listener.go).
//
//     The residual edge case: a card you OWN that died while an
//     opponent controlled it still carries their controller field in
//     your graveyard, and reanimating it returns it under their
//     control. That needs the same seam.
//
//  2. ETB triggers declared via Spec.OnETB do NOT fire on the
//     reanimated creature. ReturnFromGraveyardForEffect emits
//     EventETB but never calls fireETBHookLocked — the engine has
//     TWO independently-wired ETB mechanisms and this path fires
//     only one of them. Cards whose ETB is declared as Triggered +
//     EventETB (the S19 style, which is most of the catalog) work
//     correctly; the older OnETB hook silently does nothing. This
//     is an engine bug, not a property of this card — see
//     docs/decklists/top-100-commander-staples.md.
//
//  3. The reanimated permanent bypasses the CR 614 replacement
//     pipeline entirely, so a creature with a self-"enters tapped"
//     or "enters with counters" replacement enters untapped and
//     bare. Same root cause as (2): the direct-move helper does not
//     run the entry pipeline.
func init() {
	Register(Spec{
		OracleID: "a044474a-cd72-4e9d-bd8d-a08f2de9cdc0",
		Name:     "Reanimate",
		Targets: TargetCardInGraveyard("target creature card in your graveyard",
			Creature(), YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			cost := manaValueOfCard(card)
			if err := (ReturnFromGraveyard{
				Target: target,
				Dest:   game.ZoneBattlefield,
			}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -cost)
		},
	})
}
