package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PutThisPermanentOnTopOfOwnersLibrary is "Put this <permanent> on top of its
// owner's library" as an activated ability's whole effect (Wayward Soul,
// Soaring Hope). It does nothing once the
// source has left or come back as a new object (CR 400.7).
func PutThisPermanentOnTopOfOwnersLibrary(g *game.Game, item *game.StackItem) error {
	if SourceIsNewObject(g, item) || !OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return g.TuckToLibraryForEffect(item.SourceCardID, false)
}

// s82OldB553NettlevineBlightOracle is Nettlevine Blight's oracle id.
const s82OldB553NettlevineBlightOracle = "11132bbe-985d-44fa-a5b4-4b4e0dc3a7cf"

// s82OldB553CreatureOrLandYouControl is the offer of "a creature or land you
// control" as one seat's candidates, exactly one to pick. The pick is made
// before the enchanted permanent is sacrificed (a prompt in between would let
// the state-based action put the unattached Aura into the graveyard), so a
// permanent a Nettlevine Blight is on is left out: that is the one the
// ability sacrifices, which is not there to be chosen once it has gone. (A
// second Blighted permanent of yours is left out too; its own trigger
// sacrifices it this same end step.)
func s82OldB553CreatureOrLandYouControl(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	blighted := map[uuid.UUID]bool{}
	cards := g.BattlefieldCardsForEffect()
	for _, c := range cards {
		if c.OracleID == s82OldB553NettlevineBlightOracle && c.AttachedTo.Kind == game.TargetCard {
			blighted[c.AttachedTo.ID] = true
		}
	}
	var out []uuid.UUID
	for _, c := range cards {
		if c.Controller == of && (c.IsCreature() || c.IsLand()) && !blighted[c.InstanceID] {
			out = append(out, c.InstanceID)
		}
	}
	return out, 1, 1
}

// s82OldB553SacrificeHostAndAttach is the rest of Nettlevine Blight's granted
// ability once the permanent is chosen: the enchanted permanent is sacrificed
// and the Aura that granted the ability, unattached by that, is attached to the
// chosen one (CR 303.4f) while both are still on the battlefield.
var s82OldB553SacrificeHostAndAttach = RegisterPermanentsPick("prompt-run/sacrifice-enchanted-and-attach-granting-aura", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		if err := SacrificeThisIfStillOnBattlefield(ctx.Game, ctx.Item); err != nil {
			return err
		}
		aura := uuid.Nil
		if len(r.Carry) > 0 {
			aura = r.Carry[0]
		}
		for _, id := range picked.Cards() {
			if aura != uuid.Nil && OnBattlefield(ctx.Game, aura) && OnBattlefield(ctx.Game, id) {
				return ctx.Game.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
		}
		return nil
	},
})

// S82OldB553SacrificeThisAndReattachTheAura is Nettlevine Blight's granted
// "sacrifice this permanent and attach Nettlevine Blight to a creature or land
// you control": its controller picks (no target) and the sacrifice and the
// attach follow.
func S82OldB553SacrificeThisAndReattachTheAura(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question:   "Attach Nettlevine Blight to a creature or land you control",
		Candidates: s82OldB553CreatureOrLandYouControl,
		Keyed:      s82OldB553SacrificeHostAndAttach,
		Carry:      []uuid.UUID{NewContext(g, item).Grantor()},
	}.Apply(NewContext(g, item))
}
