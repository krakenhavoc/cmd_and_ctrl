package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trickster's Talisman — Artifact — Equipment {U} (Adventures in the
// Forgotten Realms, #5499):
//
//	"Invoke Duplicity — Equipped creature gets +1/+1 and has "Whenever this
//	 creature deals combat damage to a player, you may sacrifice
//	 Trickster's Talisman. If you do, create a token that's a copy of this
//	 creature."
//	 Equip {2}"
//
// Spare Dagger's shape: a granted trigger on the creature that names the
// Equipment that gave it (StackItem.GrantedBy) and asks the "you may
// sacrifice" as it resolves, the Talisman riding the prompt with its epoch.
// "If you do" is not a reflexive trigger: the copy is made as part of the
// same resolution once the Talisman really left the battlefield. The token
// is a copy of the creature's copiable values as it resolves, or as it last
// existed if it has left (CR 707.2, CR 608.2h), so it has neither the +1/+1
// nor the granted ability (those are not copiable).
//
// Invoke Duplicity is an ability word (CR 207.2c) with no rules meaning.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17469f58-286e-4410-b717-5908d24eecdd",
		Name:         "Trickster's Talisman",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(tricksterTalismanGrant),
		},
		Grants: []AbilityGrant{{
			Key: tricksterTalismanGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Trickster's Talisman — you may sacrifice Trickster's Talisman; if you do, create a token that's a copy of this creature", tricksterTalismanDamage),
			},
			Text: "Whenever this creature deals combat damage to a player, you may sacrifice Trickster's Talisman. If you do, create a token that's a copy of this creature.",
		}},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// tricksterTalismanGrant is the bundle the Talisman gives its creature. An
// on-disk identity: never renamed.
const tricksterTalismanGrant = "tricksters-talisman/sacrifice-for-copy"

// tricksterTalismanDamage offers to sacrifice the Talisman that granted the
// trigger, while its controller still controls it on the battlefield.
func tricksterTalismanDamage(g *game.Game, item *game.StackItem) error {
	talisman, ok := g.LookupCardForEffect(item.GrantedBy)
	if !ok || !onBattlefield(g, talisman.InstanceID) || talisman.Controller != item.Controller {
		return nil
	}
	return MayChoice{
		Question: "Trickster's Talisman — sacrifice Trickster's Talisman to create a token that's a copy of this creature?",
		YesLabel: "Sacrifice Trickster's Talisman",
		NoLabel:  "Don't",
		Then:     tricksterTalismanAnswer,
		Objects:  []game.ObjectRef{{ID: talisman.InstanceID, Epoch: talisman.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// tricksterTalismanAnswer is the "yes": sacrifice the carried Talisman while
// it is still that object under the answering player's control. Keyed (#4291).
var tricksterTalismanAnswer = MayYes("confirm/tricksters-talisman-sacrifice", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Objects) == 0 {
		return nil
	}
	ref := r.Objects[0]
	talisman, ok := ctx.Game.LookupCardForEffect(ref.ID)
	if !ok || talisman.ObjectEpoch != ref.Epoch || !onBattlefield(ctx.Game, ref.ID) || talisman.Controller != ctx.Controller() {
		return nil
	}
	return SacrificePermanent{Target: ref.ID, ThenTail: ctx.ItemTail(tricksterTalismanSacrificed)}.Apply(ctx)
})

// tricksterTalismanSacrificed is "if you do": the copy, once the Talisman
// really left the battlefield. An on-disk key: never renamed, never reused.
var tricksterTalismanSacrificed = ifItWentTail("tricksters-talisman/sacrificed", func(ctx *Context) error {
	return leitmotifComposerCopy(ctx.Game, ctx.Item)
})
