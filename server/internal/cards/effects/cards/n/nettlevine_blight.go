package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nettlevine Blight — Enchantment — Aura {4}{B}{B} (Lorwyn, S82 OLD, #5593):
//
//	"Enchant creature or land
//	 Enchanted permanent has "At the beginning of your end step, sacrifice this
//	 permanent and attach Nettlevine Blight to a creature or land you control.""
//
// Inevitable End's granted end-step trigger: "your" is the enchanted
// permanent's controller. It sacrifices the host, then that player picks a
// creature or land they control as it resolves and the Aura (named by the
// grant, game.StackItem.GrantedBy) is attached to it. With nothing to pick, or
// with the Aura gone, it stays unattached and the state-based action puts it
// into the graveyard (CR 704.5m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11132bbe-985d-44fa-a5b4-4b4e0dc3a7cf",
		Name:         "Nettlevine Blight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant creature or land", effects.Or(effects.Creature(), effects.Land())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: nettlevineBlightGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Nettlevine Blight — sacrifice this permanent and attach Nettlevine Blight to a creature or land you control",
					effects.S82OldB553SacrificeThisAndReattachTheAura),
			},
			Text: "At the beginning of your end step, sacrifice this permanent and attach Nettlevine Blight to a creature or land you control.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(nettlevineBlightGrant)},
	})
}

// nettlevineBlightGrant is the granted bundle's key, an on-disk identity: never
// renamed, never reused.
const nettlevineBlightGrant = "nettlevine-blight/end-step-sacrifice-and-reattach"
